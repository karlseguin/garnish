package router

import (
	"fmt"
	"github.com/karlseguin/garnish/core"
	"strings"
)

func (ri *RouteInfo) ParamContraint(param string, values ...string) *RouteInfo {
	s := make(Contraints, len(values))
	for _, value := range values {
		s[value] = struct{}{}
	}
	ri.constraints[param] = s
	return ri
}

// Configuration for router middleware
type Configuration struct {
	logger   core.Logger
	fallback *core.Route
	info     []*RouteInfo
	routes   map[string]*RouteMap
}

func Configure() *Configuration {
	return &Configuration{
		info: make([]*RouteInfo, 0, 10),
	}
}

// The fallback route to use when no match is found
func (c *Configuration) Fallback(route *core.Route) *Configuration {
	c.fallback = route
	return c
}

// Adds a route. A method of * will be expanded to include GET, PUT, POST,
// DELETE, PURGE and PATCH
func (c *Configuration) Add(method, path string, route *core.Route) *RouteInfo {
	ri := &RouteInfo{
		method:      method,
		path:        path,
		route:       route,
		constraints: make(map[string]Contraints),
	}
	c.info = append(c.info, ri)
	return ri
}

func (c *Configuration) compile() *Configuration {
	routeNames := make(map[string]struct{})
	c.routes = make(map[string]*RouteMap, len(c.info))
	for _, ri := range c.info {
		routeName := ri.route.Name
		if _, exists := routeNames[routeName]; exists {
			panic(fmt.Sprintf("Route names must be unique, %q used twice", routeName))
		} else {
			routeNames[routeName] = struct{}{}
		}
		c.addInfo(ri)
	}
	c.info = nil
	return c
}

func (c *Configuration) addInfo(info *RouteInfo) {
	methods := strings.Split(strings.Replace(info.method, "*", "GET,PUT,POST,DELETE,PATCH", -1), ",")
	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		if method == "PURGE" {
			continue //automatically added on a GET
		}
		rm, exists := c.routes[method]
		if exists == false {
			rm = &RouteMap{routes: make(map[string]*RouteMap)}
			c.routes[method] = rm
		}
		c.add(rm, info)

		if method == "GET" {
			rm, exists := c.routes["PURGE"]
			if exists == false {
				rm = &RouteMap{routes: make(map[string]*RouteMap)}
				c.routes["PURGE"] = rm
			}
			c.add(rm, info)
		}
	}
}

func (c *Configuration) add(root *RouteMap, info *RouteInfo) {
	path := info.path
	route := info.route
	if len(path) == 0 {
		c.addRoot(root, route)
		return
	}

	if path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}

	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if len(path) == 0 {
		c.addRoot(root, route)
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
func (c *Configuration) addRoot(root *RouteMap, route *core.Route) *Configuration {
	root.route = route
	return c
}