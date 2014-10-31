package gc

import (
	"github.com/karlseguin/nd"
	"github.com/karlseguin/params"
	"net/http"
	"time"
)

// Extends an *http.Request
type Request struct {
	scope  string
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

// Used internally to release any resources associated with the request
func (r *Request) Close() {
	r.params.Release()
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
