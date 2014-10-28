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

func (h *HandlerTests) NotFoundForUnknownRoute() {
	handler := testHandler()
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, build.Request().Path("/fail").Request)
	Expect(out.Code).To.Equal(404)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
}

func (h *HandlerTests) NilResponse() {
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

func (h *HandlerTests) OkStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(200, "ok")
	}).Build()
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, 0, "hits", 2, "2xx", 2)
}

func (h *HandlerTests) ErrorStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(401, "err")
	}).Build()
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, 0, "hits", 2, "4xx", 2)
}

func (h *HandlerTests) FailStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		return gc.Respond(500, "fail")
	}).Build()
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, 0, "hits", 2, "5xx", 2)
}

func (h *HandlerTests) SlowStats() {
	handler, req := runtime().Catch(func(req *gc.Request, next gc.Middleware) gc.Response {
		req.Start = time.Now().Add(time.Millisecond * -251)
		return gc.Respond(500, "fail")
	}).Build()
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, 0, "hits", 2, "5xx", 2, "slow", 2)
}

func assertStats(handler *Handler, dflt int64, values ...interface{}) {
	snapshop := handler.Runtime.Routes["control"].Stats.Snapshot()

	expected := make(map[string]int64, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		expected[values[i].(string)] = int64(values[i+1].(int))
	}

	for k, v := range snapshop {
		if e, ok := expected[k]; ok {
			Expect(v).To.Equal(e).Message("stats " + k)
		} else {
			Expect(v).To.Equal(dflt).Message("stats " + k)
		}
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
