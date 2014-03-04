package garnish

import (
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/nd"
	"net/http"
	"strings"
)

type context struct {
	logger    gc.Logger
	requestId string
	request   *http.Request
	route     *gc.Route
	params    gc.Params
	location  string
	user      gc.User
}

func newContext(req *http.Request, logger gc.Logger) *context {
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

func (c *context) Route() *gc.Route {
	return c.route
}

func (c *context) Params() gc.Params {
	return c.params
}

func (c *context) Fatal(err error) gc.Response {
	c.logger.Error(c, err)
	return InternalError
}

func (c *context) Location() string {
	return c.location
}

func (c *context) SetLocation(location string) {
	c.location = location
}

func (c *context) User() gc.User {
	return c.user
}

func (c *context) SetUser(user gc.User) {
	c.user = user
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
