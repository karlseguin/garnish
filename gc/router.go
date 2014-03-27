package gc

import (
	"github.com/karlseguin/gofake"
)

// Route parameters
type Params map[string]string

// Map an http.Request to a Route
type Router interface {
	Route(context Context) (*Route, Params, Response)
	Add(name, method, path string) RouteConfig
	Routes() map[string]*Route
	IsValid() bool
}

type RouteConfig interface {
	Override(func()) RouteConfig
	Route() *Route
}

// Route information
type Route struct {
	Name string
}

func NewRoute(name string) *Route {
	return &Route{Name: name}
}

type FakeRouter struct {
	gofake.Fake
}

func newFakeRouter() *FakeRouter {
	return &FakeRouter{gofake.New()}
}

func (f *FakeRouter) Route(context Context) (*Route, Params, Response) {
	f.Called(context)
	return nil, nil, nil
}

func (f *FakeRouter) Add(name, method, path string) RouteConfig {
	f.Called(name, method, path)
	return nil
}

func (f *FakeRouter) Routes() map[string]*Route {
	f.Called()
	return nil
}

func (f *FakeRouter) IsValid() bool {
	f.Called()
	return true
}
