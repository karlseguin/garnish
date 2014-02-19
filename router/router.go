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
	"github.com/karlseguin/garnish/core"
	"strings"
)

type Contraints map[string]struct{}

type RouteMap struct {
	route         *core.Route
	parameterName string
	constraints   Contraints
	fallback      *core.Route
	routes        map[string]*RouteMap
	prefixes      []*Prefix
}

type Segments []*Segment

type Segment struct {
	name          string
	parameterName string
}

type Prefix struct {
	value string
	route *core.Route
}

type Router struct {
	routes      map[string]*RouteMap
	routeLookup map[string]*core.Route
	middlewares []core.MiddlewareFactory
	logger      core.Logger
	valid       bool
}

func New(logger core.Logger, middlewares []core.MiddlewareFactory) *Router {
	return &Router{
		routeLookup: make(map[string]*core.Route),
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

func (r *Router) Routes() map[string]*core.Route {
	return r.routeLookup
}

func (r *Router) Route(context core.Context) (*core.Route, core.Params, core.Response) {
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

	params := make(core.Params)
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
			if node.constraints != nil {
				if _, constrained := node.constraints[part]; constrained == false {
					return nil, nil, nil
				}
			}
			if node.route != nil {
				route = node.route
			}
			params[node.parameterName] = part
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
	return route, params, nil
}

// Adds a route to the router.

// The name of each route must be unique. Use the special name "fallback" to
// designate the route to use when no match is found.

// Method can be a single HTTP method, or a comma separated list. The method "*"
// matches GET, PUT, POST, DELETE, PURGE and PATCH

// The path can include named captures: /product/:id

// This method is not thread safe. It is expected
func (r *Router) Add(name, method, path string) core.RouteConfig {
	if _, exists := r.routeLookup[name]; exists {
		r.logger.Errorf("Route names must be unique, %q used twice", name)
		r.valid = false
	}

	route := core.NewRoute(name)
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

func (r *Router) add(root *RouteMap, route *core.Route, segments Segments) {
	node := root
	if len(segments) == 1 && segments[0].name == "*" {
		root.fallback = route
		return
	}
	var added bool
	for index, segment := range segments {
		name := segment.name
		if name[len(name)-1] == '*' {
			if index < len(segments)-1 {
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
			leaf.parameterName = segment.parameterName
			leaf.route = route
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
	route    *core.Route
	methods  []string
	segments Segments
}

func (r *RouteConfig) Constrain(parameterName string, values ...string) core.RouteConfig {
	for _, method := range r.methods {
		root := r.router.routes[method]
		r.applyConstraint(root, r.segments, parameterName, values...)
	}
	return r
}

func (r *RouteConfig) Override(override func()) core.RouteConfig {
	for _, middleware := range r.router.middlewares {
		middleware.OverrideFor(r.route)
	}
	override()
	return r
}

func (r *RouteConfig) applyConstraint(root *RouteMap, segments Segments, parameterName string, values ...string) {
	node := root
	for _, segment := range segments {
		node = node.routes[segment.name]
		if node.parameterName == parameterName {
			node.constraints = make(Contraints)
			for _, value := range values {
				node.constraints[value] = struct{}{}
			}
			return
		}
	}
	r.router.logger.Errorf("Constraint to parameter %q on route %q does not appear to match a valid parameter", parameterName, r.route.Name)
	r.router.valid = false
}

func segment(path string) Segments {
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
	segments := make(Segments, len(parts))

	for index, part := range parts {
		segment := new(Segment)
		if part[0] == ':' {
			segment.parameterName = part[1:]
			segment.name = "?"
		} else {
			segment.name = part
		}
		segments[index] = segment
	}
	return segments
}

func newRouteMap() *RouteMap {
	return &RouteMap{
		routes:   make(map[string]*RouteMap),
		prefixes: make([]*Prefix, 0, 1),
	}
}
