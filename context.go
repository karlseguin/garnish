package garnish

import (
	"github.com/karlseguin/nd"
	"net/http"
)

// Represents information about the request
type Context interface {
	// A unique id for the request
	RequestId() string

	// The incoming http.Request
	RequestIn() *http.Request

	// The outgoing http.Request
	RequestOut() *http.Request

	// Get the route parameters
	Params() Params

	// Get the request's route
	Route() *Route

	// Generate a fatal response wrapping the specified error
	Fatal(err error) Response

	// Where the context is currently executing
	// used for logging
	Location() string

	// Set the current location for logging
	SetLocation(string)
}

type context struct {
	requestId  string
	requestIn  *http.Request
	requestOut *http.Request
	route      *Route
	logger     Logger
	params     Params
	location   string
}

func newContext(req *http.Request, logger Logger) *context {
	id := nd.Guidv4String()
	requestOut := &http.Request{
		Close:      false,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{"X-Request-Id": []string{id}},
	}
	return &context{
		requestIn:  req,
		requestId:  id,
		requestOut: requestOut,
		logger:     logger,
		location:   "init",
	}
}

func (c *context) RequestId() string {
	return c.requestId
}

func (c *context) RequestIn() *http.Request {
	return c.requestIn
}

func (c *context) RequestOut() *http.Request {
	return c.requestOut
}

func (c *context) Route() *Route {
	return c.route
}

func (c *context) Params() Params {
	return c.params
}

func (c *context) Fatal(err error) Response {
	c.logger.Error(c, err)
	return InternalError
}

func (c *context) Location() string {
	return c.location
}

func (c *context) SetLocation(location string) {
	c.location = location
}

// Context Builder is largely available for testing
// with a context from outside of the main package easy
// The builder itself satifies the Context interface
type CB struct {
	requestId  string
	requestIn  *http.Request
	requestOut *http.Request
	route      *Route
	params     Params
	location   string
}

func ContextBuilder() *CB {
	return &CB{
		requestId:  "9001!",
		requestIn:  new(http.Request),
		requestOut: new(http.Request),
		route:      new(Route),
		params:     make(Params),
		location:   "cb",
	}
}

func (c *CB) SetId(id string) *CB {
	c.requestId = id
	return c
}

func (c *CB) SetRequestIn(request *http.Request) *CB {
	c.requestIn = request
	return c
}
func (c *CB) SetRoute(route *Route) *CB {
	c.route = route
	return c
}

func (c *CB) RequestId() string {
	return c.requestId
}

func (c *CB) RequestIn() *http.Request {
	return c.requestIn
}

func (c *CB) RequestOut() *http.Request {
	return c.requestOut
}

func (c *CB) Route() *Route {
	return c.route
}

func (c *CB) Params() Params {
	return c.params
}

func (c *CB) Fatal(err error) Response {
	return nil
}

func (c *CB) Location() string {
	return c.location
}

func (c *CB) SetLocation(location string) {
	c.location = location
}
