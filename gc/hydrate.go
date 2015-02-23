package gc

import (
	"gopkg.in/karlseguin/typed.v1"
	"io"
	"net/http"
	"time"
)

type HydrateLoader func(fragment ReferenceFragment) []byte

type Fragment interface {
	Render(runtime *Runtime) []byte
	Size() int
}

type LiteralFragment []byte

func (f LiteralFragment) Render(runtime *Runtime) []byte {
	return f
}

func (f LiteralFragment) Size() int {
	return len(f)
}

type ReferenceFragment struct {
	size int
	typed.Typed
}

func NewReferenceFragment(data map[string]interface{}, size int) ReferenceFragment {
	return ReferenceFragment{
		size:  size + 100,
		Typed: typed.Typed(data),
	}
}

func (f ReferenceFragment) Render(runtime *Runtime) []byte {
	return runtime.HydrateLoader(f)
}

func (f ReferenceFragment) Size() int {
	return f.size
}

type HydrateResponse struct {
	status    int
	size      int
	expires   time.Time
	header    http.Header
	fragments []Fragment
}

func NewHydraterResponse(status int, header http.Header, fragments []Fragment) *HydrateResponse {
	return &HydrateResponse{
		status:    status,
		header:    header,
		fragments: fragments,
	}
}

func (r *HydrateResponse) Write(runtime *Runtime, w io.Writer) {
	for _, fragment := range r.fragments {
		w.Write(fragment.Render(runtime))
	}
}

func (r *HydrateResponse) Status() int {
	return r.status
}

func (r *HydrateResponse) Header() http.Header {
	return r.header
}

func (r *HydrateResponse) Close() {}

func (r *HydrateResponse) ContentLength() int {
	return -1
}

func (r *HydrateResponse) Size() int {
	return r.size
}

func (r *HydrateResponse) Cached() bool {
	return r.expires != zero
}

func (r *HydrateResponse) Expires() time.Time {
	return r.expires
}

func (r *HydrateResponse) Expire(at time.Time) {
	r.expires = at
}

func (r *HydrateResponse) ToCacheable(expires time.Time) CachedResponse {
	r.expires = expires
	r.size = 300 + 200*len(r.header)
	for _, f := range r.fragments {
		r.size += f.Size()
	}
	return r
}

func (r *HydrateResponse) Serialize(serializer Serializer) error {
	serializer.WriteInt(r.status)
	serializeHeader(serializer, r.header)
	// binary.Write(buffer, endianness, len(r.body))
	// buffer.Write(r.body)
	return nil
}
