package garnish

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/gspec"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRepliesWithTheRoutersResponse(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(core.Route), nil, Respond(nil).Status(401), new(nilMiddleware))
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
	h := buildHandler(new(core.Route), nil, nil, new(nextMiddleware), newResponseMiddleware(201, "ok", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, new(http.Request))
	spec.Expect(rec.Code).ToEqual(201)
	spec.Expect(rec.Body.String()).ToEqual("ok")
}

func TestHandlerWritesTheContentLength(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(core.Route), nil, nil, newResponseMiddleware(201, "it's over 9000", http.Header{"Content-Length": []string{"32"}}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, new(http.Request))
	spec.Expect(rec.HeaderMap.Get("Content-Length")).ToEqual("14")
}

func TestHandlerWritesHeaders(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(core.Route), nil, nil, newResponseMiddleware(201, "it's over 9000", http.Header{"X-Test": []string{"leto"}}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, new(http.Request))
	spec.Expect(rec.HeaderMap.Get("X-Test")).ToEqual("leto")
}

func buildHandler(route *core.Route, params core.Params, response core.Response, middlewares ...core.MiddlewareFactory) *Handler {
	config := Configure()
	config.router = &FakeRouter{route, params, response}
	for _, middleware := range middlewares {
		config.Middleware(middleware)
	}
	config.logger = new(FakeLogger)
	handler, _ := newHandler(config)
	return handler
}

type FakeRouter struct {
	route    *core.Route
	params   core.Params
	response core.Response
}

func (r *FakeRouter) Route(context core.Context) (*core.Route, core.Params, core.Response) {
	return r.route, r.params, r.response
}

func (r *FakeRouter) Add(name, method, path string) core.RouteConfig {
	return nil
}

func (r *FakeRouter) Routes() map[string]*core.Route {
	return nil
}

type nilMiddleware struct{}

func (m *nilMiddleware) Name() string {
	return "nil"
}

func (m *nilMiddleware) Run(context core.Context, next core.Next) core.Response {
	return nil
}

func (m *nilMiddleware) Create(config core.Configuration) (core.Middleware, error) {
	return m, nil
}

func (m *nilMiddleware) OverrideFor(route *core.Route) {}

type nextMiddleware struct{}

func (m *nextMiddleware) Name() string {
	return "next"
}

func (m *nextMiddleware) Run(context core.Context, next core.Next) core.Response {
	return next(context)
}

func (m *nextMiddleware) Create(config core.Configuration) (core.Middleware, error) {
	return m, nil
}

func (m *nextMiddleware) OverrideFor(route *core.Route) {}

type responseMiddleware struct {
	response core.Response
}

func newResponseMiddleware(status int, body string, header http.Header) core.MiddlewareFactory {
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

func (m *responseMiddleware) Run(context core.Context, next core.Next) core.Response {
	return m.response
}

func (m *responseMiddleware) Create(config core.Configuration) (core.Middleware, error) {
	return m, nil
}

func (m *responseMiddleware) OverrideFor(route *core.Route) {}
