package garnish

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type HandlerTests struct {
	h *HandlerHelper
}

func Test_Handler(t *testing.T) {
	Expectify(&HandlerTests{newHelper()}, t)
}

func (_ *HandlerTests) NotFoundForUnknownRoute() {
	handler := testHandler()
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, build.Request().Path("/fail").Request)
	Expect(out.Code).To.Equal(404)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
}

func (ht *HandlerTests) NilResponse() {
	out := httptest.NewRecorder()
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		return nil
	}).Get("/control")
	handler.ServeHTTP(out, req)
	Expect(out.Code).To.Equal(500)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
	Expect(ht.h.logger.(*FakeLogger).errors).To.Contain(`[500] "nil response object" "http://local.test/control"`)
}

func (ht *HandlerTests) OkStats() {

	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(200, "ok")
	}).Get("/control")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "2xx", 2)
}

func (ht *HandlerTests) ErrorStats() {
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(401, "err")
	}).Get("/control")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "4xx", 2)
}

func (ht *HandlerTests) FailStats() {
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(500, "fail")
	}).Get("/control")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "5xx", 2)
}

func (ht *HandlerTests) SlowStats() {
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		req.Start = time.Now().Add(time.Millisecond * -251)
		return gc.Respond(500, "fail")
	}).Get("/control")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(handler, "hits", 2, "5xx", 2, "slow", 2)
}

func (ht *HandlerTests) NoCacheForPost() {
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(200, "ok")
	}).Get("/control")
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("ok")
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("")
}

func (ht *HandlerTests) NoCacheForDisabledCache() {
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(200, "ok3")
	}).Get("/nocache")
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("ok3")
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("")
}

func (ht *HandlerTests) CachesValues() {
	called := false
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		if called {
			return gc.Respond(200, "fail")
		}
		called = true
		return gc.Respond(200, "res")
	}).Get("/cache")

	handler.ServeHTTP(httptest.NewRecorder(), req)
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("res")
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("hit")
}

func (ht *HandlerTests) SaintMode() {
	called := false
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		if called {
			return gc.Respond(500, "fail")
		}
		called = true
		return gc.Respond(200, "res")
	}).Get("/cache")

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

type HandlerHelper struct {
	runtime *gc.Runtime
	logger  gc.Logs
}

func newHelper() *HandlerHelper {
	logger := NewFakeLogger()
	config := Configure().Logger(logger).DnsTTL(-1)
	config.Cache()
	config.Stats()
	config.Upstream("test").Address("http://localhost:9001")
	config.Route("control").Get("/control").Upstream("test")
	config.Route("cache").Get("/cache").Upstream("test").CacheTTL(time.Minute)
	config.Route("nocache").Get("/nocache").Upstream("test").CacheTTL(-1)

	rb := &HandlerHelper{
		logger:  logger,
		runtime: config.Build(),
	}
	if rb.runtime == nil {
		panic("configuration build fail")
	}
	for _, route := range rb.runtime.Routes {
		route.Upstream = nil
	}
	return rb
}

func (h *HandlerHelper) Catch(catch gc.Middleware) *HandlerHelper {
	middlewares.Catch = catch
	return h
}

func (h *HandlerHelper) Get(path string) (*Handler, *http.Request) {
	//snapshotting resets the stats
	//this gives each test empty stats to start with
	for _, route := range h.runtime.Routes {
		route.Stats.Snapshot()
	}
	return &Handler{h.runtime}, build.Request().Path(path).Request
}
