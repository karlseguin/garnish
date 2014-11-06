package garnish

import (
	"fmt"
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares"
	"github.com/karlseguin/typed"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var hydrateBody = `
{
  "page":1,
  "total": 54,
  "results": [
    {"!ref": {"id": "9001p", "type": "product"}},
    {"!ref": {"id": "9002p", "type": "product"}},
    {"!ref": {"id": "9003p", "type": "cats"}}
  ]
}`

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
	errors := ht.h.logger.(*FakeLogger).errors
	Expect(errors[len(errors)-1]).To.Contain(`nil response object`)
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

	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	// Expect(out.HeaderMap.Get("ETag")).To.Equal(`"726573d41d8cd98f00b204e9800998ecf8427e"`)

	out = httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("res")
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("hit")
	// Expect(out.HeaderMap.Get("ETag")).To.Equal(`"726573d41d8cd98f00b204e9800998ecf8427e"`)
}

func (ht *HandlerTests) IfNoneMatch() {
	Skip("etags diabled for now")
	called := false
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		if called {
			return gc.Respond(200, "fail2")
		}
		called = true
		return gc.Respond(200, "res2")
	}).Get("/cache")

	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	req.Header.Set("If-None-Match", out.HeaderMap.Get("ETag"))
	out = httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Code).To.Equal(304)
	Expect(out.Body.String()).To.Equal("")
}

func (ht *HandlerTests) SaintMode() {
	called := false
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		if called {
			//this actually gets called from grace mode
			//but our assertions don't look at that
			return gc.Respond(500, "fail3")
		}
		called = true
		return gc.Respond(200, "res3")
	}).Get("/cache")

	handler.ServeHTTP(httptest.NewRecorder(), req)
	out := httptest.NewRecorder()

	item := handler.Runtime.Cache.Get("/cache", "")
	item.Extend(time.Hour * -1)
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("res3")
}

func (ht *HandlerTests) Purge() {
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(200, "res")
	}).Get("/cache")

	// make this independent of other tests, blah
	req.URL.RawQuery += "purge=test"

	out := httptest.NewRecorder()
	req.Method = "PURGE"
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("")
	Expect(out.Code).To.Equal(204)

	req.Method = "GET"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	out = httptest.NewRecorder()
	req.Method = "PURGE"
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("")
	Expect(out.Code).To.Equal(200)
}

func (ht *HandlerTests) Authentication() {
	handler, req := ht.h.Get("/noauth")
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("")
	Expect(out.Code).To.Equal(401)
}

func (ht *HandlerTests) Hydrate() {
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		return gc.RespondH(200, http.Header{"X-Hydrate": []string{"!ref"}}, hydrateBody)
	}).Get("/hydrate")

	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	assertHydrate(out)
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("")
}

func (ht *HandlerTests) CachedHydrate() {
	called := 0
	handler, req := ht.h.Catch(func(req *gc.Request) gc.Response {
		called += 1
		return gc.RespondH(200, http.Header{"X-Hydrate": []string{"!ref"}}, hydrateBody)
	}).Get("/hydrate2")

	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	assertHydrate(out)

	out = httptest.NewRecorder()
	handler.ServeHTTP(out, req)

	Expect(called).To.Equal(1)
	assertHydrate(out)
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("hit")
}

func assertHydrate(out *httptest.ResponseRecorder) {
	Expect(out.Code).To.Equal(200)
	b, _ := typed.Json(out.Body.Bytes())
	Expect(b.Int("page")).To.Equal(1)
	Expect(b.Int("total")).To.Equal(54)
	results := b.Objects("results")
	Expect(results[0].String("i")).To.Equal("9001p")
	Expect(results[0].String("p")).To.Equal("product")
	Expect(results[1].String("i")).To.Equal("9002p")
	Expect(results[1].String("p")).To.Equal("product")
	Expect(results[2].String("i")).To.Equal("9003p")
	Expect(results[2].String("p")).To.Equal("cats")
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
	config.Hydrate(func(reference *gc.ReferenceFragment) []byte {
		return []byte(fmt.Sprintf(`{"i": %q, "p": %q}`, reference.Id, reference.T))
	})

	config.Cache().PurgeHandler(func(req *gc.Request, lookup gc.CacheKeyLookup, cache gc.Purgeable) gc.Response {
		if cache.Delete(lookup(req)) == false {
			return gc.PurgeMissResponse
		}
		return gc.PurgeHitResponse
	})
	config.Auth(func(req *gc.Request) gc.Response {
		if req.URL.Path == "/noauth" {
			return gc.UnauthorizedResponse
		}
		return nil
	})
	config.Stats()
	config.Upstream("test").Address("http://localhost:9001")
	config.Route("control").Get("/control").Upstream("test")
	config.Route("cache").Get("/cache").Upstream("test").CacheTTL(time.Minute)
	config.Route("nocache").Get("/nocache").Upstream("test").CacheTTL(-1)
	config.Route("noauth").Get("/noauth").Upstream("test")
	config.Route("hydrate").Get("/hydrate").Upstream("test")
	config.Route("hydrate2").Get("/hydrate2").Upstream("test").CacheTTL(time.Minute)

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
