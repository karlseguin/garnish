package garnish

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/nd"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type context struct {
	logger    gc.Logger
	requestId string
	request   *http.Request
	route     *gc.Route
	params    gc.Params
	location  string
	user      gc.User
	query     map[string]string
	startTime time.Time
	body      gc.Body
}

func newContext(req *http.Request, logger gc.Logger) *context {
	now := time.Now()
	id := nd.Guidv4String()
	return &context{
		startTime: now,
		logger:    logger,
		request:   req,
		requestId: id,
		location:  "handler",
		query:     loadQuery(req.URL.RawQuery),
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

func (c *context) Query() map[string]string {
	return c.query
}

func (c *context) StartTime() time.Time {
	return c.startTime
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

func (c *context) Body() gc.Body {
	if c.body != nil {
		return c.body
	}

	method := c.request.Method
	if method == "POST" || method == "PUT" || method == "DELETE" {
		defer c.request.Body.Close()
		c.body = newBody(c.request.Body)
	}
	return c.body
}

// Whatever, the built-in one returns a map[string][]string
// which is as right as it is annoying
func loadQuery(raw string) map[string]string {
	l := len(raw)
	if l == 0 {
		return nil
	}

	i := 0
	for ; raw[i] == '&'; i++ {
	}
	raw = raw[i:]

	query := make(map[string]string)
	for len(raw) > 0 {
		index := strings.IndexByte(raw, '=')
		if index == -1 {
			break
		}
		key := strings.ToLower(raw[:index])
		raw = raw[index+1:]

		index = strings.IndexByte(raw, '&')
		last := false
		if index == -1 {
			index = len(raw)
			last = true
		}
		if unescaped, err := url.QueryUnescape(raw[:index]); err == nil {
			query[key] = unescaped
		}
		if last {
			break
		}
		raw = raw[index+1:]
	}
	return query
}

type body struct {
	*bytepool.Item
}

func newBody(b io.Reader) *body {
	buffer := InputPool.Checkout()
	buffer.ReadFrom(b)
	return &body{buffer}
}
