package handler_test

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/testserver"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser"
)

func TestServer(t *testing.T) {
	srv := testserver.New()
	srv.AddTransport(&transport.GET{})

	t.Run("returns an error if no transport matches", func(t *testing.T) {
		resp := post(srv.ServeFiber, "/foo", "application/json")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, `{"errors":[{"message":"transport not supported"}],"data":null}`, string(contents))
	})

	t.Run("calls query on executable schema", func(t *testing.T) {
		resp := get(srv.ServeFiber, "/foo?query={name}")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, `{"data":{"name":"test"}}`, string(contents))
	})

	t.Run("mutations are forbidden", func(t *testing.T) {
		resp := get(srv.ServeFiber, "/foo?query=mutation{name}")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode)
		assert.Equal(t, `{"errors":[{"message":"GET requests only allow query operations"}],"data":null}`, string(contents))
	})

	t.Run("subscriptions are forbidden", func(t *testing.T) {
		resp := get(srv.ServeFiber, "/foo?query=subscription{name}")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode)
		assert.Equal(t, `{"errors":[{"message":"GET requests only allow query operations"}],"data":null}`, string(contents))
	})

	t.Run("invokes operation middleware in order", func(t *testing.T) {
		var calls []string
		srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
			calls = append(calls, "first")
			return next(ctx)
		})
		srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
			calls = append(calls, "second")
			return next(ctx)
		})

		resp := get(srv.ServeFiber, "/foo?query={name}")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		assert.Equal(t, []string{"first", "second"}, calls)
	})

	t.Run("invokes response middleware in order", func(t *testing.T) {
		var calls []string
		srv.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
			calls = append(calls, "first")
			return next(ctx)
		})
		srv.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
			calls = append(calls, "second")
			return next(ctx)
		})

		resp := get(srv.ServeFiber, "/foo?query={name}")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		assert.Equal(t, []string{"first", "second"}, calls)
	})

	t.Run("invokes field middleware in order", func(t *testing.T) {
		var calls []string
		srv.AroundFields(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
			calls = append(calls, "first")
			return next(ctx)
		})
		srv.AroundFields(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
			calls = append(calls, "second")
			return next(ctx)
		})

		resp := get(srv.ServeFiber, "/foo?query={name}")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		assert.Equal(t, []string{"first", "second"}, calls)
	})

	t.Run("get query parse error in AroundResponses", func(t *testing.T) {
		var errors1 gqlerror.List
		var errors2 gqlerror.List
		srv.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
			resp := next(ctx)
			errors1 = graphql.GetErrors(ctx)
			errors2 = resp.Errors
			return resp
		})

		resp := get(srv.ServeFiber, "/foo?query=invalid")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, string(contents))
		assert.Equal(t, 1, len(errors1))
		assert.Equal(t, 1, len(errors2))
	})

	t.Run("query caching", func(t *testing.T) {
		ctx := context.Background()
		cache := &graphql.MapCache{}
		srv.SetQueryCache(cache)
		qry := `query Foo {name}`

		t.Run("cache miss populates cache", func(t *testing.T) {
			resp := get(srv.ServeFiber, "/foo?query="+url.QueryEscape(qry))
			contents, err := io.ReadAll(resp.Body)
			if err != nil {
				require.Nil(t, err, "unexpected error while reading response body from fiber test server")
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, `{"data":{"name":"test"}}`, string(contents))

			cacheDoc, ok := cache.Get(ctx, qry)
			require.True(t, ok)
			require.Equal(t, "Foo", cacheDoc.(*ast.QueryDocument).Operations[0].Name)
		})

		t.Run("cache hits use document from cache", func(t *testing.T) {
			doc, gqlErr := parser.ParseQuery(&ast.Source{Input: `query Bar {name}`})
			require.Nil(t, gqlErr)
			cache.Add(ctx, qry, doc)

			resp := get(srv.ServeFiber, "/foo?query="+url.QueryEscape(qry))
			contents, err := io.ReadAll(resp.Body)
			if err != nil {
				require.Nil(t, err, "unexpected error while reading response body from fiber test server")
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, `{"data":{"name":"test"}}`, string(contents))

			cacheDoc, ok := cache.Get(ctx, qry)
			require.True(t, ok)
			require.Equal(t, "Bar", cacheDoc.(*ast.QueryDocument).Operations[0].Name)
		})
	})
}

func TestErrorServer(t *testing.T) {
	srv := testserver.NewError()
	srv.AddTransport(&transport.GET{})

	t.Run("get resolver error in AroundResponses", func(t *testing.T) {
		var errors1 gqlerror.List
		var errors2 gqlerror.List
		srv.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
			resp := next(ctx)
			errors1 = graphql.GetErrors(ctx)
			errors2 = resp.Errors
			return resp
		})

		resp := get(srv.ServeFiber, "/foo?query={name}")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		assert.Equal(t, 1, len(errors1))
		assert.Equal(t, 1, len(errors2))
	})
}

type panicTransport struct{}

var _ graphql.Transport = (*panicTransport)(nil)

func (t panicTransport) Supports(ctx *fiber.Ctx) bool {
	return true
}

func (h panicTransport) Do(ctx *fiber.Ctx, exec graphql.GraphExecutor) {
	panic(fmt.Errorf("panic in transport"))
}

func TestRecover(t *testing.T) {
	srv := testserver.New()
	srv.AddTransport(&panicTransport{})

	t.Run("recover from panic", func(t *testing.T) {
		resp := get(srv.ServeFiber, "/foo?query={name}")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.Nil(t, err, "unexpected error while reading response body from fiber test server")
		}

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, string(contents))
	})
}

func get(handler fiber.Handler, target string) *http.Response {
	r := httptest.NewRequest("GET", target, nil)
	app := fiber.New()
	parsedUrl, err := url.Parse(target)
	if err != nil {
		panic("expected valid URL")
	}
	app.Get(parsedUrl.Path, handler)
	resp, err := app.Test(r)
	if err != nil {
		panic("encountered error while running fiber test server")
	}
	return resp
}

func post(handler fiber.Handler, target, contentType string) *http.Response {
	r := httptest.NewRequest("POST", target, nil)
	r.Header.Set("Content-Type", contentType)
	app := fiber.New()
	app.All(target, handler)
	resp, err := app.Test(r)
	if err != nil {
		panic("encountered error while running fiber test server")
	}
	return resp
}
