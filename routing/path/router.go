// A general purpose router which routes on the incoming request
// method + path. Routes are split by by /, support * and the longest
// matching route is picked. For example, given these two routes:
//    /sayans/*/skills
//    /sayans/
//
// and given the following request:
//   /sayans/9001/skills
//
// the first route will be picked
package path

import (
	"github.com/karlseguin/garnish"
	"strings"
)

func Register(config *Configuration) garnish.Router {
	r := &Router{config}
	return r.router
}

type Router struct {
	*Configuration
}

func (r *Router) router(context garnish.Context) (*garnish.Route, garnish.Response) {
	request := context.RequestIn()

	rm, ok := r.routes[request.Method]
	if ok == false {
		r.logger.Infof(context, "unknown method %q", request.Method)
		return r.fallback, nil
	}

	path := request.URL.Path
	if path == "/" {
		route := rm.route
		if route == nil {
			route = r.fallback
		}
		return rm.route, nil
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
			rm = node
		}
	}
	return route, nil
}
