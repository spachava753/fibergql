package transport

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/spachava753/fibergql/graphql"
	"mime"
)

// POST implements the POST side of the default HTTP transport
// defined in https://github.com/APIs-guru/graphql-over-http#post
type POST struct{}

var _ graphql.Transport = POST{}

func (h POST) Supports(ctx *fiber.Ctx) bool {
	if ctx.GetReqHeaders()["Upgrade"] != "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(ctx.GetReqHeaders()["Content-Type"])
	if err != nil {
		return false
	}

	return ctx.Method() == "POST" && mediaType == "application/json"
}

func (h POST) Do(ctx *fiber.Ctx, exec graphql.GraphExecutor) {
	ctx.Set("Content-Type", "application/json")

	var params *graphql.RawParams
	start := graphql.Now()
	if err := jsonDecode(bytes.NewReader(ctx.Body()), &params); err != nil {
		ctx.Status(fiber.StatusBadRequest)
		writeJsonErrorf(ctx.Response().BodyWriter(), "json body could not be decoded: "+err.Error())
		return
	}
	params.ReadTime = graphql.TraceTiming{
		Start: start,
		End:   graphql.Now(),
	}

	rc, err := exec.CreateOperationContext(ctx.UserContext(), params)
	if err != nil {
		ctx.Status(statusFor(err))
		resp := exec.DispatchError(graphql.WithOperationContext(ctx.UserContext(), rc), err)
		writeJson(ctx.Response().BodyWriter(), resp)
		return
	}
	responses, userCtx := exec.DispatchOperation(ctx.UserContext(), rc)
	writeJson(ctx.Response().BodyWriter(), responses(userCtx))
}
