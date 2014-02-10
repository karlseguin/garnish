package garnish

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/nd"
	"net/http"
)

type context struct {
	requestId  string
	requestIn  *http.Request
	requestOut *http.Request
	route      *core.Route
	logger     core.Logger
	params     core.Params
	location   string
}

func newContext(req *http.Request, logger core.Logger) *context {
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
		location:   "handler",
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

func (c *context) Route() *core.Route {
	return c.route
}

func (c *context) Params() core.Params {
	return c.params
}

func (c *context) Fatal(err error) core.Response {
	c.logger.Error(c, err)
	return InternalError
}

func (c *context) Location() string {
	return c.location
}

func (c *context) SetLocation(location string) {
	c.location = location
}
