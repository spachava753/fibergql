// client is used internally for testing. See readme for alternatives

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
)

type (
	// Client used for testing GraphQL servers. Not for production use.
	Client struct {
		h    fiber.Handler
		opts []Option
	}

	// Option implements a visitor that mutates an outgoing GraphQL request
	//
	// This is the Option pattern - https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
	Option func(bd *Request)

	// Request represents an outgoing GraphQL request
	Request struct {
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables,omitempty"`
		OperationName string                 `json:"operationName,omitempty"`
		Extensions    map[string]interface{} `json:"extensions,omitempty"`
		HTTP          *http.Request          `json:"-"`
	}

	// Response is a GraphQL layer response from a handler.
	Response struct {
		Data       interface{}
		Errors     json.RawMessage
		Extensions map[string]interface{}
	}
)

// New creates a graphql client
// Options can be set that should be applied to all requests made with this client
func New(h fiber.Handler, opts ...Option) *Client {
	p := &Client{
		h:    h,
		opts: opts,
	}

	return p
}

// MustPost is a convenience wrapper around Post that automatically panics on error
func (p *Client) MustPost(query string, response interface{}, options ...Option) {
	if err := p.Post(query, response, options...); err != nil {
		panic(err)
	}
}

// Post sends a http POST request to the graphql endpoint with the given query then unpacks
// the response into the given object.
func (p *Client) Post(query string, response interface{}, options ...Option) error {
	respDataRaw, err := p.RawPost(query, options...)
	if err != nil {
		return err
	}

	// we want to unpack even if there is an error, so we can see partial responses
	unpackErr := unpack(respDataRaw.Data, response)

	if respDataRaw.Errors != nil {
		return RawJsonError{respDataRaw.Errors}
	}
	return unpackErr
}

// RawPost is similar to Post, except it skips decoding the raw json response
// unpacked onto Response. This is used to test extension keys which are not
// available when using Post.
func (p *Client) RawPost(query string, options ...Option) (*Response, error) {
	r, err := p.newRequest(query, options...)
	if err != nil {
		return nil, fmt.Errorf("build: %w", err)
	}

	app := fiber.New()
	app.Post("/", p.h)
	resp, err := app.Test(r)
	if err != nil {
		return nil, fmt.Errorf("unxpected error from fiber test server: %w", err)
	}
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unxpected error from reading fiber test response: %w", err)
	}
	if resp.StatusCode >= fiber.StatusBadRequest {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(contents))
	}

	// decode it into map string first, let mapstructure do the final decode
	// because it can be much stricter about unknown fields.
	respDataRaw := &Response{}
	err = json.Unmarshal(contents, &respDataRaw)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return respDataRaw, nil
}

func (p *Client) newRequest(query string, options ...Option) (*http.Request, error) {
	bd := &Request{
		Query: query,
		HTTP:  httptest.NewRequest(http.MethodPost, "/", nil),
	}
	bd.HTTP.Header.Set("Content-Type", "application/json")

	// per client options from client.New apply first
	for _, option := range p.opts {
		option(bd)
	}
	// per request options
	for _, option := range options {
		option(bd)
	}

	contentType := bd.HTTP.Header.Get("Content-Type")
	switch {
	case regexp.MustCompile(`multipart/form-data; ?boundary=.*`).MatchString(contentType):
		break
	case "application/json" == contentType:
		requestBody, err := json.Marshal(bd)
		if err != nil {
			return nil, fmt.Errorf("encode: %w", err)
		}
		bd.HTTP.Header.Add("Content-Length", strconv.FormatInt(int64(len(requestBody)), 10))
		bd.HTTP.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	default:
		panic("unsupported encoding " + bd.HTTP.Header.Get("Content-Type"))
	}

	return bd.HTTP, nil
}

func unpack(data interface{}, into interface{}) error {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:      into,
		TagName:     "json",
		ErrorUnused: true,
		ZeroFields:  true,
	})
	if err != nil {
		return fmt.Errorf("mapstructure: %w", err)
	}

	return d.Decode(data)
}
