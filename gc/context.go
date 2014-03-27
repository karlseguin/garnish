package gc

import (
	"net/http"
	"net/url"
	"time"
)

// Represents information about the request
type Context interface {
	Logger

	// A unique id for the request
	RequestId() string

	// The incoming http.Request
	Request() *http.Request

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

	// Gets the user
	User() User

	// Sets the user
	SetUser(user User)

	// Gets the querystring parameters
	Query() map[string]string

	// The time the request started at
	StartTime() time.Time
}

// Context Builder is largely available for testing
// with a context from outside of the main package easy
// The builder itself satifies the Context interface
type CB struct {
	Logger
	requestId string
	request   *http.Request
	route     *Route
	params    Params
	location  string
	user      User
	query     map[string]string
	startTime time.Time
}

func ContextBuilder() *CB {
	return &CB{
		startTime: time.Now(),
		Logger:    newFakeLogger(),
		requestId: "9001!",
		request: &http.Request{
			URL:    new(url.URL),
			Header: make(http.Header),
		},
		route:    &Route{Name: "home"},
		params:   make(Params),
		location: "cb",
		query:    make(map[string]string),
	}
}

func (c *CB) SetId(id string) *CB {
	c.requestId = id
	return c
}

func (c *CB) SetParam(key, value string) *CB {
	c.params[key] = value
	return c
}

func (c *CB) SetQuery(key, value string) *CB {
	c.query[key] = value
	return c
}

func (c *CB) SetHeader(key, value string) *CB {
	c.request.Header.Set(key, value)
	return c
}

func (c *CB) SetUrl(u string) *CB {
	c.request.URL, _ = url.Parse(u)
	query := c.request.URL.Query()
	for key, _ := range query {
		c.query[key] = query.Get(key)
	}
	return c
}

func (c *CB) SetRequest(request *http.Request) *CB {
	c.request = request
	return c
}
func (c *CB) SetRoute(route *Route) *CB {
	c.route = route
	return c
}

func (c *CB) RequestId() string {
	return c.requestId
}

func (c *CB) Request() *http.Request {
	return c.request
}

func (c *CB) Route() *Route {
	return c.route
}

func (c *CB) Params() Params {
	return c.params
}

func (c *CB) StartTime() time.Time {
	return c.startTime
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

func (c *CB) SetUser(user User) {
	c.user = user
}

func (c *CB) User() User {
	return c.user
}

func (c *CB) Query() map[string]string {
	return c.query
}
