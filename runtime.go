package garnish

import (
	"gopkg.in/karlseguin/bytepool.v3"
	"gopkg.in/karlseguin/dnscache.v1"
	"gopkg.in/karlseguin/router.v1"
	"net/http"
	"strconv"
)

var (
	UnauthorizedResponse = Empty(401)
	fakeRequest          = &Request{Id: "fake"}
)

// Authorization / authentication handler
type AuthHandler func(req *Request) Response

// All the data needed to serve requests
// Built automatically when the garnish.Start() is called
type Runtime struct {
	Address          string
	NotFoundResponse Response
	FatalResponse    Response
	Executor         Middleware
	Upstreams        map[string]Upstream
	Routes           map[string]*Route
	Router           *router.Router
	BytePool         *bytepool.Pool
	StatsWorker      *StatsWorker
	Cache            *Cache
	Resolver         *dnscache.Resolver
	HydrateLoader    HydrateLoader
}

func (r *Runtime) RegisterStats(name string, reporter Reporter) {
	if r.StatsWorker != nil {
		r.StatsWorker.register(name, reporter)
	}
}

func (r *Runtime) ServeHTTP(out http.ResponseWriter, request *http.Request) {
	req := r.route(request)
	if req == nil {
		Log.Infof("404 %s", request.URL)
		r.reply(out, r.NotFoundResponse, fakeRequest)
		return
	}
	req.Infof("%s", req.URL)
	defer req.Close()
	r.reply(out, r.Executor(req), req)
}

func (r *Runtime) reply(out http.ResponseWriter, res Response, req *Request) {
	if res == nil {
		Log.Error("nil response")
		res = r.FatalResponse
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

	if req.hit {
		oh["X-Cache"] = hitHeaderValue
	}
	req.Infof("%d", status)
	out.WriteHeader(status)
	res.Write(r, out)
}

func (r *Runtime) route(req *http.Request) *Request {
	params, action := r.Router.Lookup(req)
	if action == nil {
		return nil
	}
	route, exists := r.Routes[action.Name]
	if exists == false {
		return nil
	}
	request := NewRequest(req, route, params)
	request.Runtime = r
	return request
}

func (o *Runtime) ReplaceWith(n *Runtime) {
	if o.StatsWorker != nil {
		o.StatsWorker.Stop()
	}
	o.Resolver.Stop()
	o.Cache.Storage.SetSize(n.Cache.Storage.GetSize())
	n.Cache.Storage.Stop()
	n.Cache.Storage = o.Cache.Storage
}
