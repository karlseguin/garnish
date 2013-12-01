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

	Route() *Route

	Fatal(err error) Response
}

type context struct {
	requestId  string
	requestIn  *http.Request
	requestOut *http.Request
	route      *Route
	logger     Logger
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

func (c *context) Fatal(err error) Response {
	c.logger.Error(c, err)
	return InternalError
}

// Context Builder is largely available for testing
// with a context from outside of the main package easy
// The builder itself satifies the Context interface
type CB struct {
	requestId string
	request   *http.Request
	upstream  *http.Request
	route     *Route
}

func ContextBuilder() *CB {
	return &CB{
		requestId: "9001!",
		request:   new(http.Request),
		upstream:  new(http.Request),
		route:     new(Route),
	}
}

func (c *CB) SetId(id string) *CB {
	c.requestId = id
	return c
}

func (c *CB) SetRequest(request *http.Request) *CB {
	c.request = request
	return c
}

func (c *CB) RequestId() string {
	return c.requestId
}

func (c *CB) Request() *http.Request {
	return c.request
}

func (c *CB) Upstream() *http.Request {
	return c.upstream
}

func (c *CB) Route() *Route {
	return c.route
}
