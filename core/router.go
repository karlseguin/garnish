package core

// Route parameters
type Params map[string]string

// Map an http.Request to a Route
type Router interface {
	Route(context Context) (*Route, Params, Response)
	Add(name, method, path string, override ...interface{})
	Routes() map[string]*Route
	Compile()
}

// Route information
type Route struct {
	Name string
}

func NewRoute(name string) *Route {
	return &Route{Name: name}
}
