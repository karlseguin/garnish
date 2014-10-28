package gc

import (
	"github.com/karlseguin/nd"
	"github.com/karlseguin/params"
	"net/http"
	"time"
)

type Request struct {
	Id    string
	Start time.Time
	*http.Request
	Route  *Route
	params params.Params
}

func NewRequest(req *http.Request, route *Route, params params.Params) *Request {
	return &Request{
		Request: req,
		Route:   route,
		params:  params,
		Start:   nd.Now(),
		Id:      nd.Guidv4String(),
	}
}

func (r *Request) Params(key string) string {
	return r.params.Get(key)
}

func (r *Request) Close() {
	r.params.Release()
}
