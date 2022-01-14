package transport_test

import (
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler/testserver"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/stretchr/testify/assert"
)

func TestGET(t *testing.T) {
	h := testserver.New()
	h.AddTransport(transport.GET{})

	t.Run("success", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "GET", "/graphql?query={name}", ``)
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode, string(contents))
		assert.Equal(t, `{"data":{"name":"test"}}`, string(contents))
	})

	t.Run("has json content-type header", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "GET", "/graphql?query={name}", ``)
		assert.Equal(t, "application/json", resp.Header["Content-Type"])
	})

	t.Run("decode failure", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "GET", "/graphql?query={name}&variables=notjson", "")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, string(contents))
		assert.Equal(t, `{"errors":[{"message":"variables could not be decoded"}],"data":null}`, string(contents))
	})

	t.Run("invalid variable", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "GET", `/graphql?query=query($id:Int!){find(id:$id)}&variables={"id":false}`, "")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, string(contents))
		assert.Equal(t, `{"errors":[{"message":"cannot use bool as Int","path":["variable","id"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}`, string(contents))
	})

	t.Run("parse failure", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "GET", "/graphql?query=!", "")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, string(contents))
		assert.Equal(t, `{"errors":[{"message":"Unexpected !","locations":[{"line":1,"column":1}],"extensions":{"code":"GRAPHQL_PARSE_FAILED"}}],"data":null}`, string(contents))
	})

	t.Run("no mutations", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "GET", "/graphql?query=mutation{name}", "")
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			require.NoError(t, err, "unable to read all of the response body")
		}
		assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode, string(contents))
		assert.Equal(t, `{"errors":[{"message":"GET requests only allow query operations"}],"data":null}`, string(contents))
	})
}
