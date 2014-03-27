// A general purpose router which routes on the incoming request
// method + path. Routes are split by by /, support :params
// and the longest matching route is picked. For example, given
// these two routes:
//    /sayans/:name/skills
//    /sayans/
//
// and given the following request:
//   /sayans/goku/skills
//
// the first route will be picked
package router

import (
	"github.com/karlseguin/garnish/gc"
	"strings"
)

type Contraints map[string]struct{}

type RouteMap struct {
	route         *gc.Route
	parameters    []string
	fallback      *gc.Route
	routes        map[string]*RouteMap
	prefixes      []*Prefix
}

type Segments struct {
	parts []string
	parameters []string
}

type Prefix struct {
	value string
	route *gc.Route
}

type Router struct {
	routes      map[string]*RouteMap
	routeLookup map[string]*gc.Route
	middlewares []gc.MiddlewareFactory
	logger      gc.Logger
	valid       bool
}

func New(logger gc.Logger, middlewares []gc.MiddlewareFactory) *Router {
	return &Router{
		routeLookup: make(map[string]*gc.Route),
		routes: map[string]*RouteMap{
			"GET":     newRouteMap(),
			"POST":    newRouteMap(),
			"PUT":     newRouteMap(),
			"DELETE":  newRouteMap(),
			"PURGE":   newRouteMap(),
			"PATCH":   newRouteMap(),
			"OPTIONS": newRouteMap(),
		},
		middlewares: middlewares,
		logger:      logger,
		valid:       true,
	}
}

func (r *Router) Routes() map[string]*gc.Route {
	return r.routeLookup
}

func (r *Router) Route(context gc.Context) (*gc.Route, gc.Params, gc.Response) {
	request := context.Request()

	rm, ok := r.routes[request.Method]
	if ok == false {
		context.Infof("unknown method %q", request.Method)
		return nil, nil, nil
	}

	path := request.URL.Path
	if path == "/" {
		return rm.fallback, nil, nil
	}

	params := make(gc.Params)
	if extensionIndex := strings.LastIndex(path, "."); extensionIndex != -1 {
		params["ext"] = path[extensionIndex+1:]
		path = path[0:extensionIndex]
	}
	route := rm.fallback
	parts := strings.Split(path[1:], "/")
	for _, part := range parts {
		if node, exists := rm.routes[part]; exists {
			if node.route != nil {
				route = node.route
			}
			rm = node
		} else if node, exists := rm.routes["?"]; exists {
			if node.route != nil {
				route = node.route
			}
			rm = node
		} else {
			for _, prefix := range rm.prefixes {
				if strings.HasPrefix(part, prefix.value) {
					return prefix.route, params, nil
				}
			}
			return nil, nil, nil
		}
	}
	for i, l := 0, len(parts); i < l; i++ {
		if parameter := rm.parameters[i]; len(parameter) > 0 {
			params[parameter] = parts[i]
		}
	}
	return route, params, nil
}

// Adds a route to the router.

// The name of each route must be unique. Use the special name "fallback" to
// designate the route to use when no match is found.

// Method can be a single HTTP method, or a comma separated list. The method "*"
// matches GET, PUT, POST, DELETE, PURGE and PATCH

// The path can include named captures: /product/:id

// This method is not thread safe. It is expected
func (r *Router) Add(name, method, path string) gc.RouteConfig {
	if _, exists := r.routeLookup[name]; exists {
		r.logger.Errorf("Route names must be unique, %q used twice", name)
		r.valid = false
	}

	route := gc.NewRoute(name)
	r.routeLookup[name] = route

	config := &RouteConfig{router: r, route: route}
	segments := segment(path)
	methods := strings.Split(strings.Replace(strings.Replace(method, "GET", "GET,PURGE", -1), "*", "GET,PUT,POST,DELETE,PATCH,PURGE", -1), ",")
	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		root, exists := r.routes[method]
		if exists == false {
			root = newRouteMap()
			r.routes[method] = root
		}
		if segments == nil {
			if root.route != nil {
				r.logger.Errorf("Multiple root routes are being defined for %q", method)
				r.valid = false
			}
			root.route = route
		} else {
			r.add(root, route, segments)
		}
	}
	config.methods = methods
	config.segments = segments
	return config
}

func (r *Router) IsValid() bool {
	return r.valid
}

func (r *Router) add(root *RouteMap, route *gc.Route, segments *Segments) {
	node := root
	length := len(segments.parts) - 1
	if len(segments.parts) == 0 && segments.parts[0] == "*" {
		root.fallback = route
		return
	}
	var added bool
	for index, name := range segments.parts {
		if name[len(name)-1] == '*' {
			if index < length {
				r.logger.Errorf("The prefixed route %q is invalid. Nothing should come after '*'", route.Name)
				r.valid = false
			}
			prefix := &Prefix{value: name[:len(name)-1], route: route}
			node.prefixes = append(node.prefixes, prefix)
			added = true
		} else {
			leaf, exists := node.routes[name]
			if exists == false {
				added = true
				leaf = newRouteMap()
				node.routes[name] = leaf
			}
			if index == length {
				leaf.route = route
				leaf.parameters = segments.parameters
			}
			node = leaf
		}
	}
	if added == false {
		r.logger.Errorf("%q's path appears to duplicate another route", route.Name)
		r.valid = false
	}
}

type RouteConfig struct {
	router   *Router
	route    *gc.Route
	methods  []string
	segments *Segments
}

func (r *RouteConfig) Route() *gc.Route {
	return r.route
}

func (r *RouteConfig) Override(override func()) gc.RouteConfig {
	for _, middleware := range r.router.middlewares {
		middleware.OverrideFor(r.route)
	}
	override()
	return r
}

func segment(path string) *Segments {
	if len(path) == 0 || path == "/" {
		return nil //todo
	}
	if path[0] == '/' {
		path = path[1:]
	}
	if path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}

	parts := strings.Split(path, "/")
	segments := &Segments{
		parts: make([]string, len(parts)),
		parameters: make([]string, len(parts)),
	}

	for index, part := range parts {
		if part[0] == ':' {
			segments.parts[index] = "?"
			segments.parameters[index] = part[1:]
		} else {
			segments.parts[index] = part
			segments.parameters[index] = ""
		}
	}
	return segments
}

func newRouteMap() *RouteMap {
	return &RouteMap{
		routes:   make(map[string]*RouteMap),
		prefixes: make([]*Prefix, 0, 1),
	}
}
