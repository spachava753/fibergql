package extension_test

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spachava753/fibergql/graphql"
	"github.com/spachava753/fibergql/graphql/handler/extension"
	"github.com/spachava753/fibergql/graphql/handler/testserver"
	"github.com/spachava753/fibergql/graphql/handler/transport"
	"github.com/stretchr/testify/require"
)

func TestHandlerComplexity(t *testing.T) {
	h := testserver.New()
	h.Use(&extension.ComplexityLimit{
		Func: func(ctx context.Context, rc *graphql.OperationContext) int {
			if rc.RawQuery == "{ ok: name }" {
				return 4
			}
			return 2
		},
	})
	h.AddTransport(&transport.POST{})
	var stats *extension.ComplexityStats
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		stats = extension.GetComplexityStats(ctx)
		return next(ctx)
	})

	t.Run("below complexity limit", func(t *testing.T) {
		stats = nil
		h.SetCalculatedComplexity(2)
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query":"{ name }"}`)
		contents, err := io.ReadAll(resp.Body)
		require.Nil(t, err, "unexpected error reading response body")
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		require.Equal(t, `{"data":{"name":"test"}}`, string(contents))

		require.Equal(t, 2, stats.ComplexityLimit)
		require.Equal(t, 2, stats.Complexity)
	})

	t.Run("above complexity limit", func(t *testing.T) {
		stats = nil
		h.SetCalculatedComplexity(4)
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query":"{ name }"}`)
		contents, err := io.ReadAll(resp.Body)
		require.Nil(t, err, "unexpected error reading response body")
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		require.Equal(t, `{"errors":[{"message":"operation has complexity 4, which exceeds the limit of 2","extensions":{"code":"COMPLEXITY_LIMIT_EXCEEDED"}}],"data":null}`,
			string(contents))

		require.Equal(t, 2, stats.ComplexityLimit)
		require.Equal(t, 4, stats.Complexity)
	})

	t.Run("within dynamic complexity limit", func(t *testing.T) {
		stats = nil
		h.SetCalculatedComplexity(4)
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query":"{ ok: name }"}`)
		contents, err := io.ReadAll(resp.Body)
		require.Nil(t, err, "unexpected error reading response body")
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		require.Equal(t, `{"data":{"name":"test"}}`, string(contents))

		require.Equal(t, 4, stats.ComplexityLimit)
		require.Equal(t, 4, stats.Complexity)
	})
}

func TestFixedComplexity(t *testing.T) {
	h := testserver.New()
	h.Use(extension.FixedComplexityLimit(2))
	h.AddTransport(&transport.POST{})

	var stats *extension.ComplexityStats
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		stats = extension.GetComplexityStats(ctx)
		return next(ctx)
	})

	t.Run("below complexity limit", func(t *testing.T) {
		h.SetCalculatedComplexity(2)
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query":"{ name }"}`)
		contents, err := io.ReadAll(resp.Body)
		require.Nil(t, err, "unexpected error reading response body")
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		require.Equal(t, `{"data":{"name":"test"}}`, string(contents))

		require.Equal(t, 2, stats.ComplexityLimit)
		require.Equal(t, 2, stats.Complexity)
	})

	t.Run("above complexity limit", func(t *testing.T) {
		h.SetCalculatedComplexity(4)
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query":"{ name }"}`)
		contents, err := io.ReadAll(resp.Body)
		require.Nil(t, err, "unexpected error reading response body")
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		require.Equal(t, `{"errors":[{"message":"operation has complexity 4, which exceeds the limit of 2","extensions":{"code":"COMPLEXITY_LIMIT_EXCEEDED"}}],"data":null}`, string(contents))

		require.Equal(t, 2, stats.ComplexityLimit)
		require.Equal(t, 4, stats.Complexity)
	})

	t.Run("bypass __schema field", func(t *testing.T) {
		h.SetCalculatedComplexity(4)
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{ "operationName":"IntrospectionQuery", "query":"query IntrospectionQuery { __schema { queryType { name } mutationType { name }}}"}`)
		contents, err := io.ReadAll(resp.Body)
		require.Nil(t, err, "unexpected error reading response body")
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		require.Equal(t, `{"data":{"name":"test"}}`, string(contents))

		require.Equal(t, 2, stats.ComplexityLimit)
		require.Equal(t, 0, stats.Complexity)
	})
}

func doRequest(handler fiber.Handler, method string, target string, body string) *http.Response {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	app := fiber.New()

	switch method {
	case fiber.MethodGet:
		app.Get(target, handler)
	case fiber.MethodPost:
		app.Post(target, handler)
	case fiber.MethodDelete:
		app.Delete(target, handler)
	case fiber.MethodPut:
		app.Put(target, handler)
	case fiber.MethodPatch:
		app.Patch(target, handler)
	case fiber.MethodOptions:
		app.Options(target, handler)
	case fiber.MethodHead:
		app.Head(target, handler)
	case fiber.MethodConnect:
		app.Connect(target, handler)
	case fiber.MethodTrace:
		app.Trace(target, handler)
	}
	resp, err := app.Test(r)
	if err != nil {
		panic("unexpected error when running test server")
	}
	return resp
}
