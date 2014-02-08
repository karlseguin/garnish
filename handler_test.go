package garnish

import (
	"github.com/karlseguin/gspec"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRepliesWithTheRoutersResponse(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(Route), nil, Respond(nil).Status(401), new(nilMiddleware))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, new(http.Request))
	spec.Expect(rec.Code).ToEqual(401)
}

func TestHandlerRepliesWithNotFoundIfRouteIsNotSet(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(nil, nil, nil, new(nilMiddleware))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, new(http.Request))
	spec.Expect(rec.Code).ToEqual(404)
}

func TestHandlerCallsTheMiddlewareChain(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(Route), nil, nil, new(nextMiddleware), newResponseMiddleware(201, "ok", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, new(http.Request))
	spec.Expect(rec.Code).ToEqual(201)
	spec.Expect(rec.Body.String()).ToEqual("ok")
}

func TestHandlerWritesTheContentLength(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(Route), nil, nil, newResponseMiddleware(201, "it's over 9000", http.Header{"Content-Length": []string{"32"}}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, new(http.Request))
	spec.Expect(rec.HeaderMap.Get("Content-Length")).ToEqual("14")
}

func TestHandlerWritesHeaders(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(Route), nil, nil, newResponseMiddleware(201, "it's over 9000", http.Header{"X-Test": []string{"leto"}}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, new(http.Request))
	spec.Expect(rec.HeaderMap.Get("X-Test")).ToEqual("leto")
}

func TestLogsInternalServerErrors(t *testing.T) {
	h := buildHandler(new(Route), nil, nil, newResponseMiddleware(505, "error", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, gspec.Request().Url("http://fake.garnish.io/fail").Req)
	h.logger.(*FakeLogger).Assert(t, FakeLogMessage{false, `"http://fake.garnish.io/fail" 505 error`})
}

func buildHandler(route *Route, params Params, response Response, middlewares ...MiddlewareFactory) *Handler {
	router := &FakeRouter{route, params, response}
	config := Configure().Router(router)
	for _, middleware := range middlewares {
		config.Middleware(middleware)
	}
	config.Logger = new(FakeLogger)
	handler, _ := newHandler(config)
	return handler
}

type FakeRouter struct {
	route *Route
	params Params
	response Response
}

func (r *FakeRouter) Route(context Context) (*Route, Params, Response) {
	return r.route, r.params, r.response
}

func (r *FakeRouter) RouteNames() []string {
	return []string{}
}

type nilMiddleware struct{}

func (m *nilMiddleware) Name() string {
	return "nil"
}

func (m *nilMiddleware) Run(context Context, next Next) Response {
	return nil
}

func (m *nilMiddleware) Create(routeNames []string) (Middleware, error) {
	return m, nil
}

type nextMiddleware struct{}

func (m *nextMiddleware) Name() string {
	return "next"
}

func (m *nextMiddleware) Run(context Context, next Next) Response {
	return next(context)
}

func (m *nextMiddleware) Create(routeNames []string) (Middleware, error) {
	return m, nil
}

type responseMiddleware struct {
	response Response
}

func newResponseMiddleware(status int, body string, header http.Header) MiddlewareFactory {
	r := Respond([]byte(body)).Status(status)
	for k, v := range header {
		r.Header(k, v[0])
	}
	return &responseMiddleware{
		response: r,
	}
}

func (m *responseMiddleware) Name() string {
	return "response"
}

func (m *responseMiddleware) Run(context Context, next Next) Response {
	return m.response
}

func (m *responseMiddleware) Create(routeNames []string) (Middleware, error) {
	return m, nil
}
