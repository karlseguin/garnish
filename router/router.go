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

type RouteInfo struct {
	method      string
	path        string
	route       *core.Route
	constraints map[string]Contraints
}

type Router struct {
	logger   core.Logger
	info     []*RouteInfo
	fallback *core.Route
	routes   map[string]*RouteMap
}

func New(logger core.Logger) *Router {
	return &Router{
		logger: logger,
		info:   make([]*RouteInfo, 0, 16),
	}
}

func (r *Router) Route(context core.Context) (*core.Route, core.Params, core.Response) {
	request := context.RequestIn()

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
					continue
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
func (r *Router) Add(name, method, path string) {
	ri := &RouteInfo{
		method:      method,
		path:        path,
		route:       core.NewRoute(name),
		constraints: make(map[string]Contraints),
	}
	r.info = append(r.info, ri)
}

func (r *Router) Compile() []string {
	routeNames := make(map[string]struct{})
	r.routes = make(map[string]*RouteMap, len(r.info))
	for _, ri := range r.info {
		routeName := ri.route.Name
		if _, exists := routeNames[routeName]; exists {
			panic(fmt.Sprintf("Route names must be unique, %q used twice", routeName))
		} else {
			routeNames[routeName] = struct{}{}
		}
		r.addInfo(ri)
	}
	r.info = nil
	return nil
}

func (r *Router) addInfo(info *RouteInfo) {
	methods := strings.Split(strings.Replace(info.method, "*", "GET,PUT,POST,DELETE,PATCH", -1), ",")
	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		if method == "PURGE" {
			continue //automatically added on a GET
		}
		rm, exists := r.routes[method]
		if exists == false {
			rm = &RouteMap{routes: make(map[string]*RouteMap)}
			r.routes[method] = rm
		}
		r.add(rm, info)

		if method == "GET" {
			rm, exists := r.routes["PURGE"]
			if exists == false {
				rm = &RouteMap{routes: make(map[string]*RouteMap)}
				r.routes["PURGE"] = rm
			}
			r.add(rm, info)
		}
	}
}

func (r *Router) add(root *RouteMap, info *RouteInfo) {
	path := info.path
	route := info.route
	if len(path) == 0 {
		r.addRoot(root, route)
		return
	}

	if path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}

	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if len(path) == 0 {
		r.addRoot(root, route)
		return
	}

	parts := strings.Split(path, "/")
	var leaf *RouteMap
	ok := false
	for _, part := range parts {
		var parameterName string
		if part[0] == ':' {
			parameterName = part[1:]
			part = "*"
		}
		leaf, ok = root.routes[part]
		if ok == false {
			leaf = &RouteMap{
				routes:        make(map[string]*RouteMap),
				parameterName: parameterName,
			}
			if constraints, exists := info.constraints[parameterName]; exists {
				leaf.constraints = constraints
			}
			root.routes[part] = leaf
		}
		root = leaf
	}
	leaf.route = route
}
func (r *Router) addRoot(root *RouteMap, route *core.Route) {
	root.route = route
}
