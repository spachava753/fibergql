package transport

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// GET implements the GET side of the default HTTP transport
// defined in https://github.com/APIs-guru/graphql-over-http#get
type GET struct{}

var _ graphql.Transport = GET{}

func (h GET) Supports(ctx *fiber.Ctx) bool {
	if ctx.GetReqHeaders()["Upgrade"] != "" {
		return false
	}

	return ctx.Method() == "GET"
}

func (h GET) Do(ctx *fiber.Ctx, exec graphql.GraphExecutor) {
	ctx.Set("Content-Type", "application/json")

	raw := &graphql.RawParams{
		Query:         ctx.Query("query"),
		OperationName: ctx.Query("operationName"),
	}
	raw.ReadTime.Start = graphql.Now()

	if variables := ctx.Query("variables"); variables != "" {
		if err := jsonDecode(strings.NewReader(variables), &raw.Variables); err != nil {
			ctx.Status(fiber.StatusBadRequest)
			writeJsonError(ctx.Response().BodyWriter(), "variables could not be decoded")
			return
		}
	}

	if extensions := ctx.Query("extensions"); extensions != "" {
		if err := jsonDecode(strings.NewReader(extensions), &raw.Extensions); err != nil {
			ctx.Status(fiber.StatusBadRequest)
			writeJsonError(ctx.Response().BodyWriter(), "extensions could not be decoded")
			return
		}
	}

	raw.ReadTime.End = graphql.Now()

	rc, err := exec.CreateOperationContext(ctx.UserContext(), raw)
	if err != nil {
		ctx.Status(statusFor(err))
		resp := exec.DispatchError(graphql.WithOperationContext(ctx.UserContext(), rc), err)
		writeJson(ctx.Response().BodyWriter(), resp)
		return
	}
	op := rc.Doc.Operations.ForName(rc.OperationName)
	if op.Operation != ast.Query {
		ctx.Status(fiber.StatusNotAcceptable)
		writeJsonError(ctx.Response().BodyWriter(), "GET requests only allow query operations")
		return
	}

	responses, userCtx := exec.DispatchOperation(ctx.UserContext(), rc)
	writeJson(ctx.Response().BodyWriter(), responses(userCtx))
}

func jsonDecode(r io.Reader, val interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(val)
}

func statusFor(errs gqlerror.List) int {
	switch errcode.GetErrorKind(errs) {
	case errcode.KindProtocol:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusOK
	}
}
