package configurations

import (
	"gopkg.in/karlseguin/garnish.v1/gc"
	"gopkg.in/karlseguin/garnish.v1/middlewares"
)

type Hydrate struct {
	header string
	loader gc.HydrateLoader
}

func NewHydrate(loader gc.HydrateLoader) *Hydrate {
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

func (h *Hydrate) Build(runtime *gc.Runtime) (*middlewares.Hydrate, error) {
	runtime.HydrateLoader = h.loader
	return &middlewares.Hydrate{
		Header: h.header,
	}, nil
}
