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
		gc.Log.Info("404 %s", r.URL.String())
		h.reply(out, gc.NotFoundResponse, nil)
		return
	}
	req.Info(req.URL.String())
	defer req.Close()
	h.reply(out, h.Executor(req), req)
}

func (h *Handler) reply(out http.ResponseWriter, res gc.Response, req *gc.Request) {
	if res == nil {
		res = gc.Fatal("nil response object")
	}

	etag := res.ETag()
	if len(etag) > 0 && req != nil {
		if nm := req.Header.Get("If-None-Match"); nm == etag {
			res.Close()
			res = gc.NotModifiedResponse
		} else {
			out.Header().Set("ETag", etag)
		}
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
	if etag := res.ETag(); len(etag) > 0 {
		oh["ETag"] = []string{etag}
	}

	if req != nil {
		if status >= 500 {
			if fatal, ok := res.(*gc.FatalResponse); ok {
				req.Error(fatal.Err)
			}
		}
		req.Info("%d", status)
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
