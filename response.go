package garnish

import (
	"net/http"
	"strconv"
)

// An interface used for a Closable response
type ByteCloser interface {
	Len() int
	Close() error
	Bytes() []byte
}

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

	//set the response's status
	SetStatus(status int)

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
	Response
}

// Set a cache-control for the specified duration
func (b *ResponseBuilder) Cache(duration int) *ResponseBuilder {
	return b.Header("Cache-Control", "private,max-age="+strconv.Itoa(duration))
}

// Set a header
func (b *ResponseBuilder) Header(key, value string) *ResponseBuilder {
	b.GetHeader().Set(key, value)
	return b
}

// Set a header
func (b *ResponseBuilder) Status(status int) *ResponseBuilder {
	b.Response.SetStatus(status)
	return b
}

// Creates a Response
func Respond(body interface{}) *ResponseBuilder {
	h := make(http.Header)
	switch b := body.(type) {
	case string:
		return &ResponseBuilder{&InMemoryResponse{h, []byte(b), 200}}
	case []byte:
		return &ResponseBuilder{&InMemoryResponse{h, b, 200}}
	case ByteCloser:
		return &ResponseBuilder{&ClosableResponse{h, b, 200}}
	default:
		return &ResponseBuilder{&InMemoryResponse{h, []byte("invalid body"), 500}}
	}
}

type InMemoryResponse struct {
	H http.Header
	B []byte
	S int
}

// Get headers
func (r *InMemoryResponse) GetHeader() http.Header {
	return r.H
}

// Get the body
func (r *InMemoryResponse) GetBody() []byte {
	return r.B
}

// Get the status
func (r *InMemoryResponse) GetStatus() int {
	return r.S
}

// Sets the status
func (r *InMemoryResponse) SetStatus(status int) {
	r.S = status
}

// close the response (noop)
func (r *InMemoryResponse) Close() error {
	return nil
}

// deatches the response from any underlying resources (noop)
func (r *InMemoryResponse) Detach() Response {
	return r
}

// A in-memory response with a chainable API which uses a bytepool
// as the body
type ClosableResponse struct {
	H http.Header
	B ByteCloser
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

// Sets the status
func (r *ClosableResponse) SetStatus(status int) {
	r.S = status
}

// closes the underlying bytepool
func (r *ClosableResponse) Close() error {
	return r.B.Close()
}

// Detaches the response from the underlying bytepool,
// turning this into an InMemoryResponse
func (r *ClosableResponse) Detach() Response {
	defer r.B.Close()
	clone := &InMemoryResponse{
		S: r.S,
		H: r.H,
	}
	clone.B = make([]byte, r.B.Len())
	copy(clone.B, r.B.Bytes())
	return clone
}
