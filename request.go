package garnish

import (
	"gopkg.in/karlseguin/bytepool.v3"
	"gopkg.in/karlseguin/nd.v1"
	"gopkg.in/karlseguin/params.v2"
	"net/http"
	"time"
)

var (
	EmptyParams = params.New(0)
)

// Extends an *http.Request
type Request struct {
	hit    bool
	scope  string
	params *params.Params

	// the underlying bytepool for the body. Consumers should use the Body() method
	B *bytepool.Bytes

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

	// To be used by consumer as-needed, unused by Garnish itself.
	Context interface{}
}

func NewRequest(req *http.Request, route *Route, params *params.Params) *Request {
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
	s, _ := r.params.Get(key)
	return s
}

// The request's body. This value is only available before the upstream
// is called (at which point it is drained)
func (r *Request) Body() []byte {
	if r.B == nil {
		if r.Request.Body == nil {
			return nil
		}
		r.B = r.Runtime.BytePool.Checkout()
		r.B.ReadFrom(r.Request.Body)
		r.Request.Body.Close()
	}
	return r.B.Bytes()
}

// For now we don't clone the body.
// Clone is only used by the cache/grace right now, what are the chances
// that we want to cache a GET request with a body?
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
		clone.params = EmptyParams
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
	if r.B != nil {
		r.B.Close()
	}
}

func (r *Request) Cached(reason string) {
	r.hit = true
	r.Info(reason)
}

// Context-aware info message (only displayed if the global configuration
// has Debug logging enabled)
func (r *Request) Infof(format string, args ...interface{}) {
	if Log.IsVerbose() {
		Log.Infof(r.Id+" | "+r.scope+" | "+format, args...)
	}
}

// Context-aware error message
func (r *Request) Errorf(format string, args ...interface{}) {
	Log.Errorf(r.Id+" | "+r.scope+" | "+format, args...)
}

func (r *Request) Info(message string) {
	if Log.IsVerbose() {
		Log.Info(r.Id + " | " + r.scope + " | " + message)
	}
}

// Context-aware error message
func (r *Request) Error(message string) {
	Log.Error(r.Id + " | " + r.scope + " | " + message)
}

func (r *Request) FatalResponse(message string) Response {
	r.Error(message)
	return r.Runtime.FatalResponse
}

func (r *Request) FatalResponseErr(message string, err error) Response {
	r.Errorf("%s: %s", message, err)
	return r.Runtime.FatalResponse
}
