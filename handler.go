package garnish

import (
	"github.com/karlseguin/garnish/gc"
	"net/http"
	"strconv"
)

type Handler struct {
	*gc.Runtime
}

func (h *Handler) ServeHTTP(out http.ResponseWriter, r *http.Request) {
	req := h.route(r)
	if req == nil {
		gc.Log.Infof("404 %s", r.URL)
		h.reply(out, gc.NotFoundResponse, nil)
		return
	}
	req.Infof("%s", req.URL)
	defer req.Close()
	h.reply(out, h.Executor(req), req)
}

func (h *Handler) reply(out http.ResponseWriter, res gc.Response, req *gc.Request) {
	if res == nil {
		res = gc.Fatal("nil response object")
	}

	defer res.Close()
	oh := out.Header()
	status := res.Status()

	for k, v := range res.Header() {
		oh[k] = v
	}
	if cl := res.ContentLength(); cl > -1 {
		oh["Content-Length"] = []string{strconv.Itoa(cl)}
	}

	if req != nil {
		if status >= 500 {
			if fatal, ok := res.(*gc.FatalResponse); ok {
				req.Error(fatal.Err)
			}
		}
		req.Infof("%d", status)
	}
	out.WriteHeader(status)
	res.Write(h.Runtime, out)
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
	request := gc.NewRequest(req, route, params)
	request.Runtime = h.Runtime
	return request
}
