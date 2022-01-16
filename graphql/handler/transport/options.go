package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spachava753/fibergql/graphql"
)

// Options responds to http OPTIONS and HEAD requests
type Options struct{}

var _ graphql.Transport = Options{}

func (o Options) Supports(ctx *fiber.Ctx) bool {
	return ctx.Method() == "HEAD" || ctx.Method() == "OPTIONS"
}

func (o Options) Do(ctx *fiber.Ctx, exec graphql.GraphExecutor) {
	switch ctx.Method() {
	case fiber.MethodOptions:
		ctx.Set("Allow", "OPTIONS, GET, POST")
		ctx.Status(fiber.StatusOK)
	case fiber.MethodHead:
		ctx.Status(fiber.StatusMethodNotAllowed)
	}
}
