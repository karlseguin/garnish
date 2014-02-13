package garnish

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/nd"
	"net/http"
	"strings"
)

type context struct {
	logger    core.Logger
	requestId string
	request   *http.Request
	route     *core.Route
	params    core.Params
	location  string
}

func newContext(req *http.Request, logger core.Logger) *context {
	id := nd.Guidv4String()
	return &context{
		logger:    logger,
		request:   req,
		requestId: id,
		location:  "handler",
	}
}

func (c *context) RequestId() string {
	return c.requestId
}

func (c *context) Request() *http.Request {
	return c.request
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

func (c *context) Info(v ...interface{}) {
	if c.logger.LogInfo() {
		c.infof(strings.Repeat("%q ", len(v)), v...)
	}
}

func (c *context) Infof(format string, v ...interface{}) {
	if c.logger.LogInfo() {
		c.infof(format, v...)
	}
}

func (c *context) infof(format string, v ...interface{}) {
	c.logger.Infof("["+c.requestId+"] ["+c.location+"] "+format, v...)
}

func (c *context) Error(v ...interface{}) {
	c.Errorf(strings.Repeat("%q ", len(v)), v)
}

func (c *context) Errorf(format string, v ...interface{}) {
	c.logger.Errorf("["+c.requestId+"] ["+c.location+"] "+format, v...)
}

func (c *context) LogInfo() bool {
	return c.logger.LogInfo()
}
