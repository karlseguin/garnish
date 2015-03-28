package gc

import (
	"gopkg.in/karlseguin/garnish.v1"
	"gopkg.in/karlseguin/garnish.v1/middlewares"
)

type Hydrate struct {
	header string
	loader garnish.HydrateLoader
}

func NewHydrate(loader garnish.HydrateLoader) *Hydrate {
	return &Hydrate{
		loader: loader,
		header: "X-Hydrate",
	}
}

// The header to look for in the upstream response
// which identifies the hydrate field
// ["X-Hydrate"]
func (h *Hydrate) Header(name string) *Hydrate {
	h.header = name
	return h
}

func (h *Hydrate) Build(runtime *garnish.Runtime) (*middlewares.Hydrate, error) {
	runtime.HydrateLoader = h.loader
	return &middlewares.Hydrate{
		Header: h.header,
	}, nil
}
