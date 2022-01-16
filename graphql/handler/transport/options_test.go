package transport_test

import (
	"net/http"
	"testing"

	"github.com/spachava753/fibergql/graphql/handler/testserver"
	"github.com/spachava753/fibergql/graphql/handler/transport"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	h := testserver.New()
	h.AddTransport(transport.Options{})

	t.Run("responds to options requests", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "OPTIONS", "/graphql?query={me{name}}", ``)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "OPTIONS, GET, POST", resp.Header["Allow"])
	})

	t.Run("responds to head requests", func(t *testing.T) {
		resp := doRequest(h.ServeFiber, "HEAD", "/graphql?query={me{name}}", ``)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
