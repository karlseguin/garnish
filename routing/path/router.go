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
package path

import (
	"github.com/karlseguin/garnish"
	"strings"
)

func Register(config *Configuration) garnish.Router {
	r := &Router{config.compile()}
	return r
}

type Router struct {
	*Configuration
}

func (r *Router) RouteNames() []string {
	names := make([]string, 0, len(r.routes))
	for key, _ := range r.routes {
		names = append(names, key)
	}
	return names
}

func (r *Router) Route(context garnish.Context) (*garnish.Route, garnish.Params, garnish.Response) {
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

	params := make(garnish.Params)
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
