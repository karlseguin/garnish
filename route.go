package garnish

// Map an http.Request to a Route
type Router func(context Context) (*Route, Response)

// Route information
type Route struct {
	Name     string
	Upstream string
}

func NewRoute(name, upstream string) *Route {
	return &Route{
		Name:     name,
		Upstream: upstream,
	}
}
