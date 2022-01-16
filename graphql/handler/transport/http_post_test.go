package transport_test

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spachava753/fibergql/graphql/handler/testserver"
	"github.com/spachava753/fibergql/graphql/handler/transport"
	"github.com/stretchr/testify/assert"
)

func TestPOST(t *testing.T) {
	h := testserver.New()
	h.AddTransport(transport.POST{})

	t.Run("success", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query":"{ name }"}`)
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, `{"data":{"name":"test"}}`, string(contents))
	})

	t.Run("decode failure", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "POST", "/graphql", "notjson")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, string(contents))
		assert.Equal(t, resp.Header["Content-Type"], "application/json")
		assert.Equal(t, `{"errors":[{"message":"json body could not be decoded: invalid character 'o' in literal null (expecting 'u')"}],"data":null}`, string(contents))
	})

	t.Run("parse failure", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query": "!"}`)
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, string(contents))
		assert.Equal(t, resp.Header["Content-Type"], "application/json")
		assert.Equal(t, `{"errors":[{"message":"Unexpected !","locations":[{"line":1,"column":1}],"extensions":{"code":"GRAPHQL_PARSE_FAILED"}}],"data":null}`, string(contents))
	})

	t.Run("validation failure", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query": "{ title }"}`)
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, string(contents))
		assert.Equal(t, resp.Header["Content-Type"], "application/json")
		assert.Equal(t, `{"errors":[{"message":"Cannot query field \"title\" on type \"Query\".","locations":[{"line":1,"column":3}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}`, string(contents))
	})

	t.Run("invalid variable", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query": "query($id:Int!){find(id:$id)}","variables":{"id":false}}`)
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, string(contents))
		assert.Equal(t, resp.Header["Content-Type"], "application/json")
		assert.Equal(t, `{"errors":[{"message":"cannot use bool as Int","path":["variable","id"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}`, string(contents))
	})

	t.Run("execution failure", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "POST", "/graphql", `{"query": "mutation { name }"}`)
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		assert.Equal(t, resp.Header["Content-Type"], "application/json")
		assert.Equal(t, `{"errors":[{"message":"mutations are not supported"}],"data":null}`, string(contents))
	})

	t.Run("validate content type", func(t *testing.T) {
		doReq := func(handler fiber.Handler, method string, target string, body string, contentType string) *http.Response {
			r := httptest.NewRequest(method, target, strings.NewReader(body))
			if contentType != "" {
				r.Header.Set("Content-Type", contentType)
			}

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

		validContentTypes := []string{
			"application/json",
			"application/json; charset=utf-8",
		}

		for _, contentType := range validContentTypes {
			t.Run(fmt.Sprintf("allow for content type %s", contentType), func(t *testing.T) {
				resp := doReq(h.ServeFiber, "POST", "/graphql", `{"query":"{ name }"}`, contentType)
				contents, err := io.ReadAll(resp.Body)
				if err != nil {
					require.NoError(t, err, "unable to read all of the response body")
				}
				assert.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
				assert.Equal(t, `{"data":{"name":"test"}}`, string(contents))
			})
		}

		invalidContentTypes := []string{
			"",
			"text/plain",

			// These content types are currently not supported, but they are supported by other GraphQL servers, like express-graphql.
			"application/x-www-form-urlencoded",
			"application/graphql",
		}

		for _, tc := range invalidContentTypes {
			t.Run(fmt.Sprintf("reject for content type %s", tc), func(t *testing.T) {
				resp := doReq(h.ServeFiber, "POST", "/graphql", `{"query":"{ name }"}`, tc)
				contents, err := io.ReadAll(resp.Body)
				if err != nil {
					require.NoError(t, err, "unable to read all of the response body")
				}
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode, string(contents))
				assert.Equal(t, fmt.Sprintf(`{"errors":[{"message":"%s"}],"data":null}`, "transport not supported"), string(contents))
			})
		}
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
