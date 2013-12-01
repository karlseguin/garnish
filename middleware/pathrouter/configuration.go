package pathrouter

import (
	"github.com/karlseguin/garnish"
	"strings"
)

type RouteMap struct {
	route  *garnish.Route
	routes map[string]*RouteMap
}

// Configuration for router middleware
type Configuration struct {
	logger garnish.Logger
	fallback *garnish.Route
	routes map[string]*RouteMap
}

func Configure(base *garnish.Configuration) *Configuration {
	return &Configuration{
		logger: base.Logger,
		routes: make(map[string]*RouteMap),
	}
}

// The fallback route to use when no match is found
func (c *Configuration) Fallback(route *garnish.Route) *Configuration {
	c.fallback = route
	return c
}

// Adds a route. A method of * will be expanded to include GET, PUT, POST,
// DELETE, PURGE and PATCH
func (c *Configuration) Add(method, path string, route *garnish.Route) *Configuration {
	methods := strings.Split(strings.Replace(method, "*", "GET,PUT,POST,DELETE,PURGE,PATCH", -1), ",")
	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		rm, exists := c.routes[method]
		if exists == false {
			rm = &RouteMap{routes: make(map[string]*RouteMap)}
			c.routes[method] = rm
		}
		c.add(rm, path, route)
	}
	return c
}

func (c *Configuration) add(root *RouteMap, path string, route *garnish.Route) {
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
		leaf, ok = root.routes[part]
		if ok == false {
			leaf = &RouteMap{routes: make(map[string]*RouteMap)}
			root.routes[part] = leaf
		}
		root = leaf
	}
	leaf.route = route
}
func (c *Configuration) addRoot(root *RouteMap, route *garnish.Route) *Configuration {
	root.route = route
	return c
}
