package gc

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/nd"
	"github.com/karlseguin/params"
	"net/http"
	"time"
)

// Extends an *http.Request
type Request struct {
	scope  string
	body   *bytepool.Bytes
	params params.Params

	// Every request has a unique id. This is forwarded to the upstreams in the X-Request-Id header
	Id string

	// The time the request started at (used by stats to track the time taken to process)
	Start time.Time

	// The actual *http.Request
	*http.Request

	// The route this request is associated with
	Route *Route

	// Garnish's runtime
	Runtime *Runtime
}

func NewRequest(req *http.Request, route *Route, params params.Params) *Request {
	return &Request{
		scope:   "root",
		Request: req,
		Route:   route,
		params:  params,
		Start:   nd.Now(),
		Id:      nd.Guidv4String(),
	}
}

// Params are values extracted from the URL of a route.
// Given a route /users/:id  we can expect a param with a key of "id"
func (r *Request) Params(key string) string {
	return r.params.Get(key)
}

func (r *Request) Body() []byte {
	if r.body == nil {
		if r.Request.Body == nil {
			return nil
		}
		r.body = r.Runtime.BytePool.Checkout()
		r.body.ReadFrom(r.Request.Body)
		r.Request.Body.Close()
	}
	return r.body.Bytes()
}

func (r *Request) Clone() *Request {
	clone := &Request{
		Id:      r.Id,
		Route:   r.Route,
		scope:   r.scope,
		Start:   r.Start,
		Request: r.Request,
		Runtime: r.Runtime,
	}
	if r.params.Len() == 0 {
		clone.params = params.Empty
	} else {
		params := clone.Runtime.Router.ParamPool.Checkout()
		r.params.Each(func(key, value string) {
			params.Set(key, value)
		})
		clone.params = params
	}
	return clone
}

// Used internally to release any resources associated with the request
func (r *Request) Close() {
	r.params.Release()
	if r.body != nil {
		r.body.Close()
	}
}

// Context-aware info message (only displayed if the global configuration
// has Debug logging enabled)
func (r *Request) Info(format string, args ...interface{}) {
	if Log.IsVerbose() {
		Log.Info(r.Id+" | "+r.scope+" | "+format, args...)
	}
}

// Context-aware error message
func (r *Request) Error(format string, args ...interface{}) {
	Log.Error(r.Id+" | "+r.scope+" | "+format, args...)
}
