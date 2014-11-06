package gc

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
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
	// The response's content length
	// Should return -1 when unknown
	ContentLength() int

	// Write out the response
	Write(runtime *Runtime, w io.Writer)

	// The status code
	Status() int

	// The headers
	Header() http.Header

	// Returns a cacheable version of this response
	ToCacheable() Response

	// Releases any resources associated with the response
	Close()

	// Whether or not the response came from a cached source
	// (affects the cache stat)
	Cached() bool

	ETag() string
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
	return RespondH(status, make(http.Header), body)
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
	case io.ReadCloser:
		return &StreamingResponse{body: b, status: status, header: header}
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

func (r *NormalResponse) ContentLength() int {
	return len(r.body)
}

func (r *NormalResponse) ETag() string {
	return ETagGenerator(r.body)
}

func (r *NormalResponse) Write(runtime *Runtime, w io.Writer) {
	w.Write(r.body)
}

func (r *NormalResponse) Status() int {
	return r.status
}

func (r *NormalResponse) Header() http.Header {
	return r.header
}

func (r *NormalResponse) Close() {}

func (r *NormalResponse) ToCacheable() Response {
	clone := &NormalResponse{
		body:   r.body,
		status: r.status,
		header: make(http.Header, len(r.header)),
		cached: true,
	}
	for k, v := range r.header {
		clone.header[k] = v
	}
	return clone
}

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

// A standard response. This response is used by the cache.
// It's also used when the upstream didn't provide a Content-Length, or
// whe the Content-Length was greater then the configured BytePool's capacity
type StreamingResponse struct {
	bytes  []byte
	body   io.ReadCloser
	status int
	header http.Header
	CL     int64
}

func (r *StreamingResponse) ContentLength() int {
	if r.bytes == nil {
		return int(r.CL)
	}
	return len(r.bytes)
}

func (r *StreamingResponse) Write(runtime *Runtime, w io.Writer) {
	if r.bytes != nil {
		w.Write(r.bytes)
	} else {
		io.Copy(w, r.body)
	}
}

func (r *StreamingResponse) ETag() string {
	if r.bytes != nil {
		return ETagGenerator(r.bytes)
	}
	return ""
}

func (r *StreamingResponse) Status() int {
	return r.status
}

func (r *StreamingResponse) Header() http.Header {
	return r.header
}

func (r *StreamingResponse) ToCacheable() Response {
	if r.bytes == nil {
		r.read()
	}
	clone := &NormalResponse{
		body:   r.bytes,
		status: r.status,
		header: make(http.Header, len(r.header)),
		cached: true,
	}
	for k, v := range r.header {
		clone.header[k] = v
	}
	return clone
}

func (r *StreamingResponse) read() {
	if r.CL > 0 {
		r.bytes = make([]byte, r.CL)
		io.ReadFull(r.body, r.bytes)
		return
	}

	tmp := bytes.NewBuffer(make([]byte, 0, 65536))
	io.Copy(tmp, r.body)
	// read is being called by ToCacheable
	// which will cache our response, let's not waste any space in the cache
	r.bytes = make([]byte, tmp.Len())
	copy(r.bytes, tmp.Bytes())
}

func (r *StreamingResponse) Close() {
	r.body.Close()
	r.body = nil
}

func (r *StreamingResponse) Cached() bool {
	return false
}

var ETagGenerator = func(data []byte) string {
	return `"` + hex.EncodeToString(md5.New().Sum(data)) + `"`
}
