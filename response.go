package garnish

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

var (
	InvalidTypeResponse   = Empty(500)
	EmptyNotFoundResponse = Empty(404)
)

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
	// When detached is true, it's expected that the original
	// response will continue to be used. Detached = false is
	// an optimization for grace mode which discards the original response
	ToCacheable(ttl time.Time) CachedResponse

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

// Builds a response with no body and the given status an headers
func EmptyH(status int, header http.Header) Response {
	return &NormalResponse{
		status: status,
		header: header,
	}
}

// Builds a response with the given status code and body
// The body can be a string, []byte, or io.ReadCloser.
// Will generate a generic Fatal (500) response for other types
func Respond(status int, body interface{}) Response {
	return RespondH(status, make(http.Header), body)
}

// Builds a response with the given status code and body
// The body can be a string, []byte, or io.ReadCloser.
// Will generate a generic Fatal (500) response for other types
// A Json response is the same as a normal response, except that the
// content-type is set to application/json
func Json(status int, body interface{}) Response {
	return RespondH(status, http.Header{"Content-Type": []string{"application/json"}}, body)
}

// Builds a response with the given status code, headers and body
// The body can be a string, []byte, or io.ReadCloser.
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
		Log.Error("invalid response body type")
		return InvalidTypeResponse
	}
}

// A standard response. This response is used by the cache.
// It's also used when the upstream didn't provide a Content-Length, or
// whe the Content-Length was greater then the configured BytePool's capacity
type NormalResponse struct {
	body    []byte
	status  int
	header  http.Header
	expires time.Time
}

func (r *NormalResponse) ContentLength() int {
	return len(r.body)
}

func (r *NormalResponse) Size() int {
	return len(r.body) + 300 + 200*len(r.header)
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

func (r *NormalResponse) ToCacheable(expires time.Time) CachedResponse {
	r.expires = expires
	return r
}

func (r *NormalResponse) Cached() bool {
	return r.expires != zero
}

func (r *NormalResponse) Expires() time.Time {
	return r.expires
}

func (r *NormalResponse) Expire(at time.Time) {
	r.expires = at
}

func (r *NormalResponse) Serialize(serializer Serializer) error {
	serializer.WriteInt(r.status)
	serializeHeader(serializer, r.header)
	serializer.Write(r.body)
	return nil
}

func (r *NormalResponse) Deserialize(deserializer Deserializer) error {
	r.status = deserializer.ReadInt()
	r.header = deserializerHeader(deserializer)
	r.body = deserializer.ReadBytes()
	return nil
}

// A standard response. This response is used by the cache.
// It's also used when the upstream didn't provide a Content-Length, or
// whe the Content-Length was greater then the configured BytePool's capacity
type StreamingResponse struct {
	bytes         []byte
	body          io.ReadCloser
	status        int
	header        http.Header
	contentLength int64
}

func Streaming(status int, header http.Header, contentLength int64, body io.ReadCloser) Response {
	return &StreamingResponse{
		status:        status,
		header:        header,
		contentLength: contentLength,
		body:          body,
	}
}

func (r *StreamingResponse) ContentLength() int {
	if r.bytes == nil {
		return int(r.contentLength)
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

func (r *StreamingResponse) Status() int {
	return r.status
}

func (r *StreamingResponse) Header() http.Header {
	return r.header
}

func (r *StreamingResponse) ToCacheable(expires time.Time) CachedResponse {
	if r.bytes == nil {
		r.read()
	}
	return &NormalResponse{
		body:    r.bytes,
		header:  r.header,
		status:  r.status,
		expires: expires,
	}
}

func (r *StreamingResponse) read() {
	if r.contentLength > 0 {
		r.bytes = make([]byte, r.contentLength)
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

func serializeHeader(serializer Serializer, header http.Header) {
	serializer.WriteInt(len(header))
	for k, v := range header {
		serializer.WriteString(k)
		serializer.WriteString(v[0])
	}
}

func deserializerHeader(deserializer Deserializer) http.Header {
	l := deserializer.ReadInt()
	header := make(http.Header, l)
	for i := 0; i < l; i++ {
		k, v := deserializer.ReadString(), deserializer.ReadString()
		header.Set(k, v)
	}
	return header
}
