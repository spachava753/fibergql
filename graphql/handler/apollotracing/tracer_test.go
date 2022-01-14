package apollotracing_test

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/testserver"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestApolloTracing(t *testing.T) {
	now := time.Unix(0, 0)

	graphql.Now = func() time.Time {
		defer func() {
			now = now.Add(100 * time.Nanosecond)
		}()
		return now
	}

	h := testserver.New()
	h.AddTransport(transport.POST{})
	h.Use(apollotracing.Tracer{})

	resp := doRequest(h.ServeFiber, http.MethodPost, "/graphql", `{"query":"{ name }"}`)
	contents, err := io.ReadAll(resp.Body)
	require.Nil(t, err, "unexpected error reading response body")
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
	var respData struct {
		Extensions struct {
			Tracing apollotracing.TracingExtension `json:"tracing"`
		} `json:"extensions"`
	}
	require.NoError(t, json.Unmarshal(contents, &respData))

	tracing := &respData.Extensions.Tracing

	require.EqualValues(t, 1, tracing.Version)

	require.Zero(t, tracing.StartTime.UnixNano())
	require.EqualValues(t, 900, tracing.EndTime.UnixNano())
	require.EqualValues(t, 900, tracing.Duration)

	require.EqualValues(t, 300, tracing.Parsing.StartOffset)
	require.EqualValues(t, 100, tracing.Parsing.Duration)

	require.EqualValues(t, 500, tracing.Validation.StartOffset)
	require.EqualValues(t, 100, tracing.Validation.Duration)

	require.EqualValues(t, 700, tracing.Execution.Resolvers[0].StartOffset)
	require.EqualValues(t, 100, tracing.Execution.Resolvers[0].Duration)
	require.EqualValues(t, ast.Path{ast.PathName("name")}, tracing.Execution.Resolvers[0].Path)
	require.Equal(t, "Query", tracing.Execution.Resolvers[0].ParentType)
	require.Equal(t, "name", tracing.Execution.Resolvers[0].FieldName)
	require.Equal(t, "String!", tracing.Execution.Resolvers[0].ReturnType)
}

func TestApolloTracing_withFail(t *testing.T) {
	now := time.Unix(0, 0)

	graphql.Now = func() time.Time {
		defer func() {
			now = now.Add(100 * time.Nanosecond)
		}()
		return now
	}

	h := testserver.New()
	h.AddTransport(transport.POST{})
	h.Use(extension.AutomaticPersistedQuery{Cache: lru.New(100)})
	h.Use(apollotracing.Tracer{})

	resp := doRequest(h.ServeFiber, http.MethodPost, "/graphql", `{"operationName":"A","extensions":{"persistedQuery":{"version":1,"sha256Hash":"338bbc16ac780daf81845339fbf0342061c1e9d2b702c96d3958a13a557083a6"}}}`)
	contents, err := io.ReadAll(resp.Body)
	require.Nil(t, err, "unexpected error reading response body")
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
	t.Log(string(contents))
	var respData struct {
		Errors gqlerror.List
	}
	require.NoError(t, json.Unmarshal(contents, &respData))
	require.Len(t, respData.Errors, 1)
	require.Equal(t, "PersistedQueryNotFound", respData.Errors[0].Message)
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
