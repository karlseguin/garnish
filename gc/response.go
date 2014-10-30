package gc

import (
	"net/http"
)

var (
	// The response to send when the route isn't found
	NotFoundResponse = Empty(404)
)

type ByteCloser interface {
	Bytes() []byte
	Close() error
}

// An http response
type Response interface {
	// The body
	Body() []byte

	// The status code
	Status() int

	// The headers
	Header() http.Header

	// Releases any resources associated with the response
	Close()

	// Whether or not the response came from a cached source
	// (affects the cache stat)
	Cached() bool
}

// Builds a response with no body and the given status
func Empty(status int) Response {
	return &NormalResponse{
		status: status,
	}
}

// Builds a response with the given status code and body
// The body can be a string, []byte, of ByteCloser
// Will generate a generic Fatal (500) response for other types
func Respond(status int, body interface{}) Response {
	return RespondH(status, nil, body)
}

// Builds a response with the given status code, headers and body
// The body can be a string, []byte, of ByteCloser.
// Will generate a generic Fatal (500) response for other types
func RespondH(status int, header http.Header, body interface{}) Response {
	switch b := body.(type) {
	case string:
		return &NormalResponse{body: []byte(b), status: status, header: header}
	case []byte:
		return &NormalResponse{body: b, status: status, header: header}
	case ByteCloser:
		return &CloseableResponse{b, status, header}
	default:
		return Fatal("invalid body type")
	}
}

// A standard response. This response is used by the cache.
// It's also used when the upstream didn't provide a Content-Length, or
// whe the Content-Length was greater then the configured BytePool's capacity
type NormalResponse struct {
	body   []byte
	status int
	header http.Header
	cached bool
}

func (r *NormalResponse) Body() []byte {
	return r.body
}

func (r *NormalResponse) Status() int {
	return r.status
}

func (r *NormalResponse) Header() http.Header {
	return r.header
}

func (r *NormalResponse) Close() {}

func (r *NormalResponse) Cached() bool {
	return r.cached
}

// A response with an associated error string to log
type FatalResponse struct {
	Response
	Err string
}

// Generate a response with a 500 error code
// message will be logged to the logger
func Fatal(message string) Response {
	return &FatalResponse{
		Err:      message,
		Response: Empty(500),
	}
}

// Generate a response with a 500 error code
// err will be logged to the logger
func FatalErr(err error) Response {
	return &FatalResponse{
		Err:      err.Error(),
		Response: Empty(500),
	}
}

// A response with a body of type ByteCloser
type CloseableResponse struct {
	body   ByteCloser
	status int
	header http.Header
}

func (r *CloseableResponse) Body() []byte {
	return r.body.Bytes()
}

func (r *CloseableResponse) Status() int {
	return r.status
}

func (r *CloseableResponse) Header() http.Header {
	return r.header
}

func (r *CloseableResponse) Close() {
	r.body.Close()
}

func (r *CloseableResponse) Cached() bool {
	return false
}

// Clones a response
// Used by the cache to turn any other type of response into
// a NormalResponse with its own copy of the body and header
func CloneResponse(r Response) Response {
	h := r.Header()
	clone := &NormalResponse{
		body:   r.Body(),
		status: r.Status(),
		header: make(http.Header, len(h)),
	}
	for k, v := range h {
		clone.header[k] = v
	}
	return clone
}
