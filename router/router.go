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
	"fmt"
	"github.com/karlseguin/garnish/core"
	"strings"
)

type Contraints map[string]struct{}

type RouteMap struct {
	route         *core.Route
	parameterName string
	constraints   Contraints
	routes        map[string]*RouteMap
}

type Segments []*Segment

type Segment struct {
	name          string
	parameterName string
}

type Router struct {
	logger      core.Logger
	fallback    *core.Route
	routes      map[string]*RouteMap
	routeLookup map[string]*core.Route
	middlewares []core.MiddlewareFactory
}

func New(logger core.Logger, middlewares []core.MiddlewareFactory) *Router {
	return &Router{
		logger:      logger,
		routeLookup: make(map[string]*core.Route),
		routes:      make(map[string]*RouteMap),
		middlewares: middlewares,
	}
}

func (r *Router) Routes() map[string]*core.Route {
	return r.routeLookup
}

func (r *Router) Route(context core.Context) (*core.Route, core.Params, core.Response) {
	request := context.Request()

	rm, ok := r.routes[request.Method]
	if ok == false {
		r.logger.Infof(context, "unknown method %q", request.Method)
		return r.fallback, nil, nil
	}

	path := request.URL.Path
	if path == "/" {
		route := rm.route
		if route == nil {
			route = r.fallback
		}
		return rm.route, nil, nil
	}

	params := make(core.Params)
	if extensionIndex := strings.LastIndex(path, "."); extensionIndex != -1 {
		params["ext"] = path[extensionIndex+1:]
		path = path[0:extensionIndex]
	}
	route := r.fallback
	parts := strings.Split(path[1:], "/")
	for _, part := range parts {
		if node, exists := rm.routes[part]; exists {
			if node.route != nil {
				route = node.route
			}
			rm = node
		} else if node, exists := rm.routes["*"]; exists {
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
		panic(fmt.Sprintf("Route names must be unique, %q used twice", name))
	}

	route := core.NewRoute(name)
	r.routeLookup[name] = route

	config := &RouteConfig{router: r, route: route}
	if name == "fallback" {
		r.fallback = route
		return nil
	}

	segments := segment(path)
	methods := strings.Split(strings.Replace(strings.Replace(method, "GET", "GET,PURGE", -1), "*", "GET,PUT,POST,DELETE,PATCH,PURGE", -1), ",")
	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		root, exists := r.routes[method]
		if exists == false {
			root = &RouteMap{routes: make(map[string]*RouteMap)}
			r.routes[method] = root
		}
		if segments == nil {
			if root.route != nil {
				panic(fmt.Sprintf("Multiple root routes are being defined for %q", method))
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

func (r *Router) add(root *RouteMap, route *core.Route, segments Segments) {
	node := root
	var added bool
	for _, segment := range segments {
		name := segment.name
		leaf, exists := node.routes[name]
		if exists == false {
			added = true
			leaf = &RouteMap{routes: make(map[string]*RouteMap)}
			node.routes[name] = leaf
		}
		leaf.parameterName = segment.parameterName
		leaf.route = route
		node = leaf
	}

	if added == false {
		panic(fmt.Sprintf("%q's path appears to duplicate another route", route.Name))
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
	panic(fmt.Sprintf("Constraint to parameter %q on route %q does not appear to match a valid parameter", parameterName, r.route.Name))
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
			segment.name = "*"
		} else {
			segment.name = part
		}
		segments[index] = segment
	}
	return segments
}
