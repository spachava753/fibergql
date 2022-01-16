package client_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	h := func(ctx *fiber.Ctx) error {
		require.Equal(t, `{"query":"user(id:$id){name}","variables":{"id":1}}`, string(ctx.Body()))

		err := json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"name": "bob",
			},
		})
		if err != nil {
			return err
		}
		return nil
	}

	c := client.New(h)

	var resp struct {
		Name string
	}

	c.MustPost("user(id:$id){name}", &resp, client.Var("id", 1))

	require.Equal(t, "bob", resp.Name)
}

func TestClientMultipartFormData(t *testing.T) {
	h := func(ctx *fiber.Ctx) error {
		bodyBytes := ctx.Request().Body()
		require.Contains(t, string(bodyBytes), `Content-Disposition: form-data; name="operations"`)
		require.Contains(t, string(bodyBytes), `{"query":"mutation ($input: Input!) {}","variables":{"file":{}}`)
		require.Contains(t, string(bodyBytes), `Content-Disposition: form-data; name="map"`)
		require.Contains(t, string(bodyBytes), `{"0":["variables.file"]}`)
		require.Contains(t, string(bodyBytes), `Content-Disposition: form-data; name="0"; filename="example.txt"`)
		require.Contains(t, string(bodyBytes), `Content-Type: text/plain`)
		require.Contains(t, string(bodyBytes), `Hello World`)

		ctx.Write([]byte(`{}`))
		return nil
	}

	c := client.New(h)

	var resp struct{}
	c.MustPost("{ id }", &resp,
		func(bd *client.Request) {
			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)
			bodyWriter.WriteField("operations", `{"query":"mutation ($input: Input!) {}","variables":{"file":{}}`)
			bodyWriter.WriteField("map", `{"0":["variables.file"]}`)

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="0"; filename="example.txt"`)
			h.Set("Content-Type", "text/plain")
			ff, _ := bodyWriter.CreatePart(h)
			ff.Write([]byte("Hello World"))
			bodyWriter.Close()

			bd.HTTP.Body = ioutil.NopCloser(bodyBuf)
			bd.HTTP.Header.Set("Content-Type", bodyWriter.FormDataContentType())
			bd.HTTP.Header.Set("Content-Length", strconv.FormatInt(int64(bodyBuf.Len()), 10))
		},
	)
}

func TestAddHeader(t *testing.T) {
	h := func(ctx *fiber.Ctx) error {
		require.Equal(t, "ASDF", ctx.GetReqHeaders()["Test-Key"])

		ctx.Write([]byte(`{}`))
		return nil
	}

	c := client.New(h)

	var resp struct{}
	c.MustPost("{ id }", &resp,
		client.AddHeader("Test-Key", "ASDF"),
	)
}

func TestAddClientHeader(t *testing.T) {
	h := func(ctx *fiber.Ctx) error {
		require.Equal(t, "ASDF", ctx.GetReqHeaders()["Test-Key"])

		ctx.Write([]byte(`{}`))
		return nil
	}

	c := client.New(h, client.AddHeader("Test-Key", "ASDF"))

	var resp struct{}
	c.MustPost("{ id }", &resp)
}

// copied from standard library: https://cs.opensource.google/go/go/+/refs/tags/go1.17.6:src/net/http/internal/ascii/print.go;l=14
func lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// copied from standard library: https://cs.opensource.google/go/go/+/refs/tags/go1.17.6:src/net/http/internal/ascii/print.go;l=14
func equalFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if lower(s[i]) != lower(t[i]) {
			return false
		}
	}
	return true
}

// copied from standard library: https://cs.opensource.google/go/go/+/refs/tags/go1.17.6:src/net/http/request.go;l=939
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !equalFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

func TestBasicAuth(t *testing.T) {
	h := func(ctx *fiber.Ctx) error {
		auth := ctx.Get(fiber.HeaderAuthorization)
		user, pass, ok := parseBasicAuth(auth)
		require.True(t, ok)
		require.Equal(t, "user", user)
		require.Equal(t, "pass", pass)

		ctx.Write([]byte(`{}`))
		return nil
	}

	c := client.New(h)

	var resp struct{}
	c.MustPost("{ id }", &resp,
		client.BasicAuth("user", "pass"),
	)
}

func TestAddCookie(t *testing.T) {
	h := func(ctx *fiber.Ctx) error {
		c := ctx.Cookies("foo")
		require.Equal(t, "value", c)
		ctx.Write([]byte(`{}`))
		return nil
	}

	c := client.New(h)

	var resp struct{}
	c.MustPost("{ id }", &resp,
		client.AddCookie(&http.Cookie{Name: "foo", Value: "value"}),
	)
}

func TestAddExtensions(t *testing.T) {
	h := func(ctx *fiber.Ctx) error {
		b := ctx.Body()
		require.Equal(t, `{"query":"user(id:1){name}","extensions":{"persistedQuery":{"sha256Hash":"ceec2897e2da519612279e63f24658c3e91194cbb2974744fa9007a7e1e9f9e7","version":1}}}`, string(b))
		err := json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"Name": "Bob",
			},
		})
		if err != nil {
			return err
		}
		return nil
	}

	c := client.New(h)

	var resp struct {
		Name string
	}
	c.MustPost("user(id:1){name}", &resp,
		client.Extensions(map[string]interface{}{"persistedQuery": map[string]interface{}{"version": 1, "sha256Hash": "ceec2897e2da519612279e63f24658c3e91194cbb2974744fa9007a7e1e9f9e7"}}),
	)
}
