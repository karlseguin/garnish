package gc

import (
	"github.com/karlseguin/params"
	"net/http"
)

type Request struct {
	*http.Request
	Route  *Route
	params params.Params
}

func NewRequest(req *http.Request, route *Route, params params.Params) *Request {
	return &Request{
		Request: req,
		Route:   route,
		params:  params,
	}
}

func (r *Request) Params(key string) string {
	return r.params.Get(key)
}

func (r *Request) Close() {
	r.params.Release()
}
