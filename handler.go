package garnish

import (
	"github.com/karlseguin/garnish/gc"
	"net/http"
	"strconv"
)

var NotFoundResponse = gc.Empty(404)

type Handler struct {
	*gc.Runtime
}

func (h *Handler) ServeHTTP(out http.ResponseWriter, r *http.Request) {
	req := h.route(r)
	if req == nil {
		reply(out, NotFoundResponse, req)
		return
	}
	defer req.Close()
	reply(out, h.Executor(req), req)
}

func reply(out http.ResponseWriter, res gc.Response, req *gc.Request) {
	if res == nil {
		res = gc.Fatal("nil response object")
	}
	defer res.Close()

	oh := out.Header()
	body := res.Body()
	status := res.Status()

	res.Header().Each(func(k, v string) {
		oh[k] = []string{v}
	})
	oh["Content-Length"] = []string{strconv.Itoa(len(body))}

	if status >= 500 {
		if fatal, ok := res.(*gc.FatalResponse); ok {
			gc.Logger.Error("[500] %q %q", fatal.Err, req.URL)
		} else {
			gc.Logger.Error("[%d] %q %q", status, string(body), req.URL)
		}
	}
	out.WriteHeader(status)
	out.Write(body)
}

func (h *Handler) route(req *http.Request) *gc.Request {
	params, action := h.Router.Lookup(req)
	if action == nil {
		return nil
	}
	route, exists := h.Routes[action.Name]
	if exists == false {
		return nil
	}
	return gc.NewRequest(req, route, params)
}
