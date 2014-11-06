package gc

import (
	"io"
	"net/http"
)

type HydrateLoader func(reference *ReferenceFragment) []byte

type Fragment interface {
	Render(runtime *Runtime) []byte
}

type LiteralFragment []byte

func (f LiteralFragment) Render(runtime *Runtime) []byte {
	return f
}

type ReferenceFragment struct {
	T    string
	Id   string
	Data map[string]string
}

func (f *ReferenceFragment) Render(runtime *Runtime) []byte {
	return runtime.HydrateLoader(f)
	return nil
}

type HydrateResponse struct {
	status    int
	cached    bool
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

func (r *HydrateResponse) Cached() bool {
	return r.cached
}

func (r *HydrateResponse) ToCacheable() Response {
	clone := &HydrateResponse{
		cached:    true,
		status:    r.status,
		header:    make(http.Header, len(r.header)),
		fragments: r.fragments,
	}
	for k, v := range r.header {
		clone.header[k] = v
	}
	return clone
}
