package caching

import (
	"github.com/karlseguin/garnish/caches"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/gspec"
	"testing"
	"time"
)

var originalGrace = grace

func TestDoesNotCacheNonGetRequests(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRequest(gspec.Request().Method("POST").Req)
	caching, _ := Configure().Cache(newFakeStorage()).Create(gc.DummyConfig)
	buildCaching(caching, "goku", "power")
	res := caching.Run(context, gc.FakeNext(gc.Respond(nil).Status(123)))
	spec.Expect(res.GetStatus()).ToEqual(123)
}

func TestDoesNotCacheRoutesWhichArentConfiguredForCaching(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRequest(gspec.Request().Method("GET").Req)
	caching, _ := Configure().Cache(newFakeStorage()).Create(gc.DummyConfig)
	buildCaching(caching, "goku", "power").ttl = 0
	res := caching.Run(context, gc.FakeNext(gc.Respond(nil).Status(123)))
	spec.Expect(res.GetStatus()).ToEqual(123)
}

func TestReturnsAFreshResult(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRequest(gspec.Request().Method("GET").Req)
	caching, _ := Configure().Cache(newFakeStorage()).Create(gc.DummyConfig)
	buildCaching(caching, "goku", "power")
	res := caching.Run(context, nil)
	spec.Expect(res.GetStatus()).ToEqual(9001)
}

func TestSkipsCacheOnCallback(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRequest(gspec.Request().Method("GET").Req)
	caching, _ := Configure().Cache(newFakeStorage()).RuntimeSkip(func(context gc.Context) bool { return true }).Create(gc.DummyConfig)
	buildCaching(caching, "goku", "power")
	res := caching.Run(context, gc.FakeNext(gc.Respond(nil).Status(123)))
	spec.Expect(res.GetStatus()).ToEqual(123)
}

func TestReturnsASlightlyStaleResult(t *testing.T) {
	graceCalled := false
	stubGrace(&graceCalled)
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRequest(gspec.Request().Method("GET").Req)
	caching, _ := Configure().Cache(newFakeStorage()).Create(gc.DummyConfig)
	buildCaching(caching, "goku", "level")
	res := caching.Run(context, nil)
	spec.Expect(res.GetStatus()).ToEqual(3)
	spec.Expect(graceCalled).ToEqual(true)
}

func TestReturnsAStaleResultIfTheNewResultIsAnError(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRequest(gspec.Request().Method("GET").Req)
	caching, _ := Configure().Cache(newFakeStorage()).Create(gc.DummyConfig)
	buildCaching(caching, "goku", "age")
	now := time.Now()
	res := caching.Run(context, gc.FakeNext(gc.Respond(nil).Status(500)))
	spec.Expect(res.GetStatus()).ToEqual(20)
	spec.Expect(res.(*caches.CachedResponse).Expires.After(now)).ToEqual(true)
}

func TestReturnsAndCachesAnUpdatedResult(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRequest(gspec.Request().Method("GET").Req)
	caching, _ := Configure().Cache(newFakeStorage()).Create(gc.DummyConfig)
	buildCaching(caching, "goku", "age")
	res := caching.Run(context, gc.FakeNext(gc.Respond([]byte("21")).Status(200).Cache(200)))
	spec.Expect(res.GetStatus()).ToEqual(200)
	spec.Expect(string(caching.(*Caching).cache.Get("goku", "age").GetBody())).ToEqual("21")
}

func TestReturnsAndCachesANewResult(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRequest(gspec.Request().Method("GET").Req)
	caching, _ := Configure().Cache(newFakeStorage()).Create(gc.DummyConfig)
	buildCaching(caching, "goku", "other")
	res := caching.Run(context, gc.FakeNext(gc.Respond([]byte("otherother")).Status(200).Cache(200)))
	spec.Expect(res.GetStatus()).ToEqual(200)
	spec.Expect(string(caching.(*Caching).cache.Get("goku", "other").GetBody())).ToEqual("otherother")
}

func TestTTLReturnsTheConfiguredTimeForAGoodRespone(t *testing.T) {
	spec := gspec.New(t)
	duration, ok := ttl(&RouteConfig{ttl: time.Second * 24}, gc.Respond(nil).Status(200))
	spec.Expect(duration).ToEqual(time.Second * 24)
	spec.Expect(ok).ToEqual(true)
}

func TestTTLReturnsTheHeaderTimeForAGoodRespone(t *testing.T) {
	spec := gspec.New(t)
	duration, ok := ttl(new(RouteConfig), gc.Respond(nil).Status(200).Header("Cache-Control", "max-age=33"))
	spec.Expect(duration).ToEqual(time.Second * 33)
	spec.Expect(ok).ToEqual(true)
}

func TestTTLReturnsTheHeaderTimeForAGoodRespone2(t *testing.T) {
	spec := gspec.New(t)
	duration, ok := ttl(new(RouteConfig), gc.Respond(nil).Status(200).Header("Cache-Control", "private,max-age=22"))
	spec.Expect(duration).ToEqual(time.Second * 22)
	spec.Expect(ok).ToEqual(true)
}

func TestTTLDoesNotHandleInvalidExpiryTimes(t *testing.T) {
	spec := gspec.New(t)
	_, ok := ttl(new(RouteConfig), gc.Respond(nil).Status(200).Header("Cache-Control", "private,max-age=fail"))
	spec.Expect(ok).ToEqual(false)
}

func TestTTLDoesNotHandleInvalidMissingExpiryTime(t *testing.T) {
	spec := gspec.New(t)
	_, ok := ttl(new(RouteConfig), gc.Respond(nil).Status(200))
	spec.Expect(ok).ToEqual(false)
}

func buildCaching(caching gc.Middleware, key, vary string) *RouteConfig {
	kg := func(context gc.Context) (string, string) { return key, vary }
	caching.(*Caching).routeConfigs["home"] = &RouteConfig{
		keyGenerator: kg,
		grace:        time.Minute,
		saint:        time.Minute,
		ttl:          time.Second * 10,
	}
	return caching.(*Caching).routeConfigs["home"]
}

func stubGrace(flag *bool) {
	grace = func(c *Caching, key, vary string, context gc.Context, config *RouteConfig, next gc.Next) {
		grace = originalGrace
		*flag = true
	}
}

type FakeStorage struct {
	data map[string]map[string]*caches.CachedResponse
}

func newFakeStorage() caches.Cache {
	return &FakeStorage{
		data: map[string]map[string]*caches.CachedResponse{
			"goku": map[string]*caches.CachedResponse{
				"power": &caches.CachedResponse{
					Expires:  time.Now().Add(time.Minute),
					Response: gc.Respond([]byte("over 9000")).Status(9001),
				},
				"level": &caches.CachedResponse{
					Expires:  time.Now().Add(time.Second * -8),
					Response: gc.Respond([]byte("super super sayan")).Status(3),
				},
				"age": &caches.CachedResponse{
					Expires:  time.Now().Add(time.Hour * -8),
					Response: gc.Respond([]byte("20")).Status(20),
				},
			},
		},
	}
}

func (c *FakeStorage) Get(key, vary string) *caches.CachedResponse {
	main, exists := c.data[key]
	if exists == false {
		return nil
	}
	return main[vary]
}

func (c *FakeStorage) Set(key, vary string, value *caches.CachedResponse) {
	c.data[key][vary] = value
}

func (c *FakeStorage) Delete(key string) bool {
	_, exists := c.data[key]
	delete(c.data, key)
	return exists
}

func (c *FakeStorage) DeleteVary(key, vary string) bool {
	main, exists := c.data[key]
	if exists == false {
		return false
	}
	_, exists = main[vary]
	delete(main, vary)
	return exists
}
