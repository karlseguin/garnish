package gc

import (
	"io"
	"net/http"
)

type HydrateLoader func(references *ReferenceFragment) []byte

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
	Data map[string]interface{}
}

func (f *ReferenceFragment) Render(runtime *Runtime) []byte {
	return runtime.Hydrate.Loader(f)
	return nil
}

type Hydrate struct {
	Header string
	Loader HydrateLoader
}

func NewHydrate(loader HydrateLoader, header string) *Hydrate {
	return &Hydrate{
		Header: header,
		Loader: loader,
	}
}

type HydrateResponse struct {
	status    int
	cached    bool
	header    http.Header
	fragments []Fragment
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

func (r *HydrateResponse) Cached() bool {
	return r.cached
}
