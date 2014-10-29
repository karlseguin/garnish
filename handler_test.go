package garnish

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"github.com/karlseguin/garnish/gc"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type HandlerTests struct{}

func Test_Handler(t *testing.T) {
	Expectify(new(HandlerTests), t)
}

func (_ *HandlerTests) NotFoundForUnknownRoute() {
	handler := testHandler()
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, build.Request().Path("/fail").Request)
	Expect(out.Code).To.Equal(404)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
}

func (_ *HandlerTests) NilResponse() {
	out := httptest.NewRecorder()
	logger := NewFakeLogger()
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return nil
	}).Logger(logger).Build("/control")
	handler.ServeHTTP(out, req)
	Expect(out.Code).To.Equal(500)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
	Expect(logger.errors).To.Contain(`[500] "nil response object" "http://local.test/control"`)
}

func (_ *HandlerTests) OkStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(200, "ok")
	}).Build("/control")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "2xx", 2)
}

func (_ *HandlerTests) ErrorStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(401, "err")
	}).Build("/control")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "4xx", 2)
}

func (_ *HandlerTests) FailStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(500, "fail")
	}).Build("/control")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "5xx", 2)
}

func (_ *HandlerTests) SlowStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		req.Start = time.Now().Add(time.Millisecond * -251)
		return gc.Respond(500, "fail")
	}).Build("/control")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "5xx", 2, "slow", 2)
}

func (_ *HandlerTests) NoCacheForPost() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(200, "ok")
	}).Build("/control")
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("ok")
}

func (_ *HandlerTests) NoCacheForDisabledCache() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(200, "ok3")
	}).Build("/nocache")
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("ok3")
}

func (_ *HandlerTests) CachesValues() {
	called := false
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		if called {
			return gc.Respond(200, "fail")
		}
		called = true
		return gc.Respond(200, "res")
	}).Build("/cache")

	handler.ServeHTTP(httptest.NewRecorder(), req)
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("res")
}

func (_ *HandlerTests) SaintMode() {
	called := false
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		if called {
			return gc.Respond(500, "fail")
		}
		called = true
		return gc.Respond(200, "res")
	}).Build("/cache")

	handler.ServeHTTP(httptest.NewRecorder(), req)
	out := httptest.NewRecorder()

	item := handler.Runtime.Cache.Get("/cache", "")
	item.Extend(time.Hour * -1)
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("res")
}

func assertStats(handler *Handler, values ...interface{}) {
	snapshot := handler.Runtime.Routes["control"].Stats.Snapshot()

	for i := 0; i < len(values); i += 2 {
		name, value := values[i].(string), int64(values[i+1].(int))
		Expect(snapshot[name]).To.Equal(value)
	}
}

type RB struct {
	config *Configuration
	catch  gc.MiddlewareHandler
}

func runtime() *RB {
	config := Configure()
	config.Cache()
	config.Stats()
	config.Upstream("test").Address("http://localhost:9001")
	config.Route("control").Get("/control").Upstream("test")
	config.Route("cache").Get("/cache").Upstream("test").CacheTTL(time.Minute)
	config.Route("nocache").Get("/nocache").Upstream("test").CacheTTL(-1)
	return &RB{config, nil}
}

func (rb *RB) Build(path string) (*Handler, *http.Request) {
	if rb.catch != nil {
		rb.config.catch = rb.catch
	}
	runtime := rb.config.Build()
	if runtime == nil {
		panic("configuration build fail")
	}
	// this forces the upstream route to pass our
	// request up the chain (something normally not possible)
	for _, route := range runtime.Routes {
		route.Upstream = nil
	}
	return &Handler{runtime}, build.Request().Path(path).Request
}

func (rb *RB) Catch(catch gc.MiddlewareHandler) *RB {
	rb.catch = catch
	return rb
}

func (rb *RB) Logger(logger gc.Logs) *RB {
	rb.config.Logger(logger)
	return rb
}
