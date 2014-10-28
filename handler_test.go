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
	}).Logger(logger).Build()
	handler.ServeHTTP(out, req)
	Expect(out.Code).To.Equal(500)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
	Expect(logger.errors).To.Contain(`[500] "nil response object" "http://local.test/c"`)
}

func (_ *HandlerTests) OkStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(200, "ok")
	}).Build()
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "2xx", 2)
}

func (_ *HandlerTests) ErrorStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(401, "err")
	}).Build()
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "4xx", 2)
}

func (_ *HandlerTests) FailStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(500, "fail")
	}).Build()
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "5xx", 2)
}

func (_ *HandlerTests) SlowStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		req.Start = time.Now().Add(time.Millisecond * -251)
		return gc.Respond(500, "fail")
	}).Build()
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "5xx", 2, "slow", 2)
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
	config.Stats()
	config.Upstream("test").Address("http://localhost:9001")
	config.Route("control").Get("/c").Upstream("test")
	return &RB{config, nil}
}

func (rb *RB) Build() (*Handler, *http.Request) {
	if rb.catch != nil {
		rb.config.catch = rb.catch
	}
	runtime := rb.config.Build()
	if runtime == nil {
		panic("configuration build fail")
	}
	// this forces the upstream route to pass our
	// request up the chain (something normally not possible)
	runtime.Routes["control"].Upstream = nil
	return &Handler{runtime}, build.Request().Path("/c").Request
}

func (rb *RB) Catch(catch gc.MiddlewareHandler) *RB {
	rb.catch = catch
	return rb
}

func (rb *RB) Logger(logger gc.Logs) *RB {
	rb.config.Logger(logger)
	return rb
}
