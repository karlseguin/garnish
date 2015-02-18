package garnish

import (
	"fmt"
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"github.com/karlseguin/garnish/cache"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares"
	"gopkg.in/karlseguin/router.v1"
	"gopkg.in/karlseguin/typed.v1"
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

type RuntimeTests struct {
	h *RuntimeHelper
}

func Test_Handler(t *testing.T) {
	Expectify(&RuntimeTests{helper()}, t)
}

func (r RuntimeTests) NotFoundForUnknownRoute() {
	runtime, req := r.h.Get("/invalid")
	out := httptest.NewRecorder()
	runtime.ServeHTTP(out, req)
	Expect(out.Code).To.Equal(404)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
}

func (r *RuntimeTests) NilResponse() {
	out := httptest.NewRecorder()
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		return nil
	}).Get("/control")
	runtime.ServeHTTP(out, req)
	Expect(out.Code).To.Equal(500)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
	errors := gc.Log.(*gc.FakeLogger).Errors
	Expect(errors[len(errors)-1]).To.Contain(`nil response object`)
}

func (r *RuntimeTests) OkStats() {
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(200, "ok")
	}).Get("/control")
	runtime.ServeHTTP(httptest.NewRecorder(), req)
	runtime.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(runtime, "hits", 2, "2xx", 2)
}

func (r *RuntimeTests) ErrorStats() {
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(401, "err")
	}).Get("/control")
	runtime.ServeHTTP(httptest.NewRecorder(), req)
	runtime.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(runtime, "hits", 2, "4xx", 2)
}

func (r *RuntimeTests) FailStats() {
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(500, "fail")
	}).Get("/control")
	runtime.ServeHTTP(httptest.NewRecorder(), req)
	runtime.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(runtime, "hits", 2, "5xx", 2)
}

func (r *RuntimeTests) SlowStats() {
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		req.Start = time.Now().Add(time.Millisecond * -251)
		return gc.Respond(500, "fail")
	}).Get("/control")
	runtime.ServeHTTP(httptest.NewRecorder(), req)
	runtime.ServeHTTP(httptest.NewRecorder(), req)
	assertStats(runtime, "hits", 2, "5xx", 2, "slow", 2)
}

func (r *RuntimeTests) NoCacheForPost() {
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(200, "ok")
	}).Get("/control")
	out := httptest.NewRecorder()
	runtime.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("ok")
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("")
}

func (r *RuntimeTests) NoCacheForDisabledCache() {
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(200, "ok3")
	}).Get("/nocache")
	out := httptest.NewRecorder()
	runtime.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("ok3")
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("")
}

func (r *RuntimeTests) CachesValues() {
	called := false
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		if called {
			called = true
			return gc.Respond(200, "fail")
		}
		return gc.Respond(200, "res")
	}).Get("/cache")

	out := httptest.NewRecorder()
	runtime.ServeHTTP(out, req)

	out = httptest.NewRecorder()
	runtime.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("res")
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("hit")
}

func (r *RuntimeTests) SaintMode() {
	called := false
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		if called {
			//this actually gets called from grace mode
			//but our assertions don't look at that
			called = true
			return gc.Respond(500, "fail3")
		}
		return gc.Respond(200, "res3")
	}).Get("/cache")

	runtime.ServeHTTP(httptest.NewRecorder(), req)
	out := httptest.NewRecorder()

	item := runtime.Cache.Storage.Get("/cache", "")
	item.Expire(time.Now().Add(time.Hour * -1))
	runtime.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("res3")
}

func (r *RuntimeTests) Purge() {
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		return gc.Respond(200, "res")
	}).Get("/cache")

	// make this independent of other tests, blah
	req.URL.RawQuery += "purge=test"

	out := httptest.NewRecorder()
	req.Method = "PURGE"
	runtime.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("")
	Expect(out.Code).To.Equal(204)

	req.Method = "GET"
	runtime.ServeHTTP(httptest.NewRecorder(), req)

	out = httptest.NewRecorder()
	req.Method = "PURGE"
	runtime.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("")
	Expect(out.Code).To.Equal(200)
}

func (r *RuntimeTests) Authentication() {
	runtime, req := r.h.Get("/noauth")
	out := httptest.NewRecorder()
	runtime.ServeHTTP(out, req)
	Expect(out.Body.String()).To.Equal("")
	Expect(out.Code).To.Equal(401)
}

func (r *RuntimeTests) Hydrate() {
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		return gc.RespondH(200, http.Header{"X-Hydrate": []string{"!ref"}}, hydrateBody)
	}).Get("/nocache")

	out := httptest.NewRecorder()
	runtime.ServeHTTP(out, req)
	assertHydrate(out)
	Expect(out.HeaderMap.Get("X-Cache")).To.Equal("")
}

func (r *RuntimeTests) CachedHydrate() {
	called := 0
	runtime, req := r.h.Catch(func(req *gc.Request) gc.Response {
		called += 1
		return gc.RespondH(200, http.Header{"X-Hydrate": []string{"!ref"}}, hydrateBody)
	}).Get("/hcache")

	out := httptest.NewRecorder()
	runtime.ServeHTTP(out, req)
	assertHydrate(out)

	out = httptest.NewRecorder()
	runtime.ServeHTTP(out, req)

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

func assertStats(runtime *gc.Runtime, values ...interface{}) {
	snapshot := runtime.Routes["control"].Stats.Snapshot()

	for i := 0; i < len(values); i += 2 {
		name, value := values[i].(string), int64(values[i+1].(int))
		Expect(snapshot[name]).To.Equal(value)
	}
}

type RuntimeHelper struct {
	runtime *gc.Runtime
}

func helper() *RuntimeHelper {
	gc.Log = gc.NewFakeLogger()
	r := router.New(router.Configure())
	r.AddNamed("cache", "GET", "/cache", nil)
	r.AddNamed("cache", "PURGE", "/cache", nil)
	r.AddNamed("hcache", "GET", "/hcache", nil)
	r.AddNamed("nocache", "GET", "/nocache", nil)
	r.AddNamed("noauth", "GET", "/noauth", nil)
	r.AddNamed("control", "GET", "/control", nil)

	hydr := &middlewares.Hydrate{Header: "X-Hydrate"}
	auth := &middlewares.Auth{
		Handler: func(req *gc.Request) gc.Response {
			if req.URL.Path == "/noauth" {
				return gc.UnauthorizedResponse
			}
			return nil
		},
	}

	e := gc.WrapMiddleware("upst", middlewares.Upstream, nil)
	e = gc.WrapMiddleware("hydr", hydr.Handle, e)
	e = gc.WrapMiddleware("cach", middlewares.Cache, e)
	e = gc.WrapMiddleware("auth", auth.Handle, e)
	e = gc.WrapMiddleware("stat", middlewares.Stats, e)
	runtime := &gc.Runtime{
		Router:   r,
		Executor: e,
		Routes: map[string]*gc.Route{
			"cache": &gc.Route{
				Stats: gc.NewRouteStats(time.Millisecond * 100),
				Cache: gc.NewRouteCache(time.Minute, gc.DefaultCacheKeyLookup),
			},
			"hcache": &gc.Route{
				Stats: gc.NewRouteStats(time.Millisecond * 100),
				Cache: gc.NewRouteCache(time.Minute, gc.DefaultCacheKeyLookup),
			},
			"noauth": &gc.Route{
				Stats: gc.NewRouteStats(time.Millisecond * 100),
				Cache: gc.NewRouteCache(time.Duration(-1), nil),
			},
			"control": &gc.Route{
				Stats: gc.NewRouteStats(time.Millisecond * 100),
				Cache: gc.NewRouteCache(time.Duration(-1), nil),
			},
			"nocache": &gc.Route{
				Stats: gc.NewRouteStats(time.Millisecond * 100),
				Cache: gc.NewRouteCache(time.Duration(-1), nil),
			},
		},
	}

	runtime.Cache = gc.NewCache()
	runtime.Cache.Storage = cache.New(10)
	runtime.Cache.PurgeHandler = func(req *gc.Request, lookup gc.CacheKeyLookup, cache gc.CacheStorage) gc.Response {
		if cache.Delete(lookup(req)) == false {
			return gc.PurgeMissResponse
		}
		return gc.PurgeHitResponse
	}

	runtime.HydrateLoader = func(reference gc.ReferenceFragment) []byte {
		return []byte(fmt.Sprintf(`{"i": %q, "p": %q}`, reference.String("id"), reference.String("type")))
	}

	return &RuntimeHelper{
		runtime: runtime,
	}
}

func (r *RuntimeHelper) Catch(catch gc.Middleware) *RuntimeHelper {
	middlewares.Catch = catch
	return r
}

func (r *RuntimeHelper) Get(path string) (*gc.Runtime, *http.Request) {
	//snapshotting resets the stats
	//this gives each test empty stats to start with
	for _, route := range r.runtime.Routes {
		route.Stats.Snapshot()
	}
	return r.runtime, build.Request().Path(path).Request
}
