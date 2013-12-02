package garnish

import (
	"github.com/karlseguin/bytepool"
	"net/http"
	"strconv"
)

// A pre-built response for a 404
var NotFound = Respond([]byte("not found")).Status(404)

// A pre-built response for a 500
var InternalError = Respond([]byte("internal error")).Status(500)

type Response interface {
	// Get the response's header
	GetHeader() http.Header

	// Get the response's body
	GetBody() []byte

	// Get the response's status
	GetStatus() int

	// Close the response
	Close() error

	// Detaches the response from any underlying resourcs.
	// In cases where Close is a no-op, this should probably
	// return self. Otherwise, the response should do whatever
	// it has to so that it can be long-lived (clone itself into
	// a normal response and close itself)
	Detach() Response
}

// A in-memory response with a chainable API. Should be created
// via the Respond() method
type ResponseBuilder struct {
	H http.Header
	B []byte
	S int
}

// Set the response status
func (b *ResponseBuilder) Status(status int) *ResponseBuilder {
	b.S = status
	return b
}

// Set a cache-control for the specified duration
func (b *ResponseBuilder) Cache(duration int) *ResponseBuilder {
	b.H.Set("Cache-Control", "private,max-age="+strconv.Itoa(duration))
	return b
}

// Set a header
func (b *ResponseBuilder) Header(key, value string) *ResponseBuilder {
	b.H.Set(key, value)
	return b
}

// Get headers
func (r *ResponseBuilder) GetHeader() http.Header {
	return r.H
}

// Get the body
func (r *ResponseBuilder) GetBody() []byte {
	return r.B
}

// Get the status
func (r *ResponseBuilder) GetStatus() int {
	return r.S
}

// Noop
func (r *ResponseBuilder) Close() error {
	return nil
}

func (r *ResponseBuilder) Detach() Response {
	return r
}

// Creates a Response
func Respond(body []byte) *ResponseBuilder {
	return &ResponseBuilder{
		S: 200,
		B: body,
		H: make(http.Header),
	}
}

// A in-memory response with a chainable API which uses a bytepool
// as the body
type ClosableResponse struct {
	H http.Header
	B *bytepool.Item
	S int
}

// Get headers
func (r *ClosableResponse) GetHeader() http.Header {
	return r.H
}

// Get the body
func (r *ClosableResponse) GetBody() []byte {
	return r.B.Bytes()
}

// Get the status
func (r *ClosableResponse) GetStatus() int {
	return r.S
}

func (r *ClosableResponse) Close() error {
	return r.B.Close()
}

func (r *ClosableResponse) Detach() Response {
	defer r.B.Close()
	clone := &ResponseBuilder{
		S: r.S,
		H: r.H,
	}
	clone.B = make([]byte, r.B.Len())
	copy(clone.B, r.B.Bytes())
	return clone
}
