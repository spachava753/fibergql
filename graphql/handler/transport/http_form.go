package transport

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

// MultipartForm the Multipart request spec https://github.com/jaydenseric/graphql-multipart-request-spec
type MultipartForm struct {
	// MaxUploadSize sets the maximum number of bytes used to parse a request body
	// as multipart/form-data.
	MaxUploadSize int64

	// MaxMemory defines the maximum number of bytes used to parse a request body
	// as multipart/form-data in memory, with the remainder stored on disk in
	// temporary files.
	MaxMemory int64
}

var _ graphql.Transport = MultipartForm{}

func (f MultipartForm) Supports(ctx *fiber.Ctx) bool {
	if ctx.GetReqHeaders()["Upgrade"] != "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(ctx.GetReqHeaders()["Content-Type"])
	if err != nil {
		return false
	}

	return ctx.Method() == "POST" && mediaType == "multipart/form-data"
}

func (f MultipartForm) maxUploadSize() int64 {
	if f.MaxUploadSize == 0 {
		return 32 << 20
	}
	return f.MaxUploadSize
}

func (f MultipartForm) maxMemory() int64 {
	if f.MaxMemory == 0 {
		return 32 << 20
	}
	return f.MaxMemory
}

func (f MultipartForm) Do(ctx *fiber.Ctx, exec graphql.GraphExecutor) {
	ctx.Set("Content-Type", "application/json")

	start := graphql.Now()

	var err error
	if int64(len(ctx.Body())) > f.maxUploadSize() {
		writeJsonError(ctx.Response().BodyWriter(), "failed to parse multipart form, request body too large")
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.Status(fiber.StatusUnprocessableEntity)
		writeJsonError(ctx.Response().BodyWriter(), "failed to parse multipart form")
		return
	}

	var params graphql.RawParams

	operations := form.Value["operations"]
	if len(operations) != 1 {
		ctx.Status(fiber.StatusUnprocessableEntity)
		writeJsonError(ctx.Response().BodyWriter(), "operations form field could not be decoded")
		return
	}
	if err = jsonDecode(strings.NewReader(operations[0]), &params); err != nil {
		ctx.Status(fiber.StatusUnprocessableEntity)
		writeJsonError(ctx.Response().BodyWriter(), "operations form field could not be decoded")
		return
	}

	uploadsMap := map[string][]string{}
	uploadsMapFormValues := form.Value["map"]
	if len(uploadsMapFormValues) != 1 {
		ctx.Status(fiber.StatusUnprocessableEntity)
		writeJsonError(ctx.Response().BodyWriter(), "map form field could not be decoded")
		return
	}
	if err = json.Unmarshal([]byte(uploadsMapFormValues[0]), &uploadsMap); err != nil {
		ctx.Status(fiber.StatusUnprocessableEntity)
		writeJsonError(ctx.Response().BodyWriter(), "map form field could not be decoded")
		return
	}

	var upload graphql.Upload
	for key, paths := range uploadsMap {
		if len(paths) == 0 {
			ctx.Status(fiber.StatusUnprocessableEntity)
			writeJsonErrorf(ctx.Response().BodyWriter(), "invalid empty operations paths list for key %s", key)
			return
		}
		files := form.File[key]
		if len(files) != 1 {
			ctx.Status(fiber.StatusUnprocessableEntity)
			writeJsonErrorf(ctx.Response().BodyWriter(), "expected file header len for key %s: %d", key, len(files))
			return
		}
		firstFile := files[0]
		header := multipart.FileHeader{
			Filename: firstFile.Filename,
			Header:   firstFile.Header,
			Size:     firstFile.Size,
		}
		file, err := firstFile.Open()
		if err != nil {
			ctx.Status(fiber.StatusUnprocessableEntity)
			writeJsonErrorf(ctx.Response().BodyWriter(), "failed to open multipart file for key %s", key)
			return
		}
		defer file.Close()

		if len(paths) == 1 {
			upload = graphql.Upload{
				File:        file,
				Size:        header.Size,
				Filename:    header.Filename,
				ContentType: header.Header.Get("Content-Type"),
			}

			if err := params.AddUpload(upload, key, paths[0]); err != nil {
				ctx.Status(fiber.StatusUnprocessableEntity)
				writeJsonGraphqlError(ctx.Response().BodyWriter(), err)
				return
			}
		} else {
			if int64(len(ctx.Body())) < f.maxMemory() {
				fileBytes, err := ioutil.ReadAll(file)
				if err != nil {
					ctx.Status(fiber.StatusUnprocessableEntity)
					writeJsonErrorf(ctx.Response().BodyWriter(), "failed to read file for key %s", key)
					return
				}
				for _, path := range paths {
					upload = graphql.Upload{
						File:        &bytesReader{s: &fileBytes, i: 0, prevRune: -1},
						Size:        header.Size,
						Filename:    header.Filename,
						ContentType: header.Header.Get("Content-Type"),
					}

					if err := params.AddUpload(upload, key, path); err != nil {
						ctx.Status(fiber.StatusUnprocessableEntity)
						writeJsonGraphqlError(ctx.Response().BodyWriter(), err)
						return
					}
				}
			} else {
				tmpFile, err := ioutil.TempFile(os.TempDir(), "gqlgen-")
				if err != nil {
					ctx.Status(fiber.StatusUnprocessableEntity)
					writeJsonErrorf(ctx.Response().BodyWriter(), "failed to create temp file for key %s", key)
					return
				}
				tmpName := tmpFile.Name()
				defer func() {
					_ = os.Remove(tmpName)
				}()
				_, err = io.Copy(tmpFile, file)
				if err != nil {
					ctx.Status(fiber.StatusUnprocessableEntity)
					if err := tmpFile.Close(); err != nil {
						writeJsonErrorf(ctx.Response().BodyWriter(), "failed to copy to temp file and close temp file for key %s", key)
						return
					}
					writeJsonErrorf(ctx.Response().BodyWriter(), "failed to copy to temp file for key %s", key)
					return
				}
				if err := tmpFile.Close(); err != nil {
					ctx.Status(fiber.StatusUnprocessableEntity)
					writeJsonErrorf(ctx.Response().BodyWriter(), "failed to close temp file for key %s", key)
					return
				}
				for _, path := range paths {
					pathTmpFile, err := os.Open(tmpName)
					if err != nil {
						ctx.Status(fiber.StatusUnprocessableEntity)
						writeJsonErrorf(ctx.Response().BodyWriter(), "failed to open temp file for key %s", key)
						return
					}
					defer pathTmpFile.Close()
					upload = graphql.Upload{
						File:        pathTmpFile,
						Size:        header.Size,
						Filename:    header.Filename,
						ContentType: header.Header.Get("Content-Type"),
					}

					if err := params.AddUpload(upload, key, path); err != nil {
						ctx.Status(fiber.StatusUnprocessableEntity)
						writeJsonGraphqlError(ctx.Response().BodyWriter(), err)
						return
					}
				}
			}
		}
	}

	params.ReadTime = graphql.TraceTiming{
		Start: start,
		End:   graphql.Now(),
	}

	rc, gerr := exec.CreateOperationContext(ctx.UserContext(), &params)
	if gerr != nil {
		resp := exec.DispatchError(graphql.WithOperationContext(ctx.UserContext(), rc), gerr)
		ctx.Status(statusFor(gerr))
		writeJson(ctx.Response().BodyWriter(), resp)
		return
	}
	responses, userCtx := exec.DispatchOperation(ctx.UserContext(), rc)
	writeJson(ctx.Response().BodyWriter(), responses(userCtx))
}
