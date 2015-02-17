package main

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/gc"
	"time"
)

func main() {
	config := garnish.Configure().Address("127.0.0.1:8080").Debug()
	config.Auth(AuthHandler)
	config.Stats().FileName("stats.json").Slow(time.Millisecond * 100)
	config.Cache().Grace(time.Minute).PurgeHandler(PurgeHandler)
	config.Upstream("test").Address("http://localhost:4005").KeepAlive(8)

	config.Route("users").Get("/users/:id").Upstream("test").CacheTTL(time.Second * 5)
	config.Route("plain").Get("/plain").Upstream("test").CacheTTL(time.Second * 5)
	garnish.Start(config)
}

func AuthHandler(req *gc.Request) gc.Response {
	if false {
		return gc.UnauthorizedResponse
	}
	return nil
}

func PurgeHandler(req *gc.Request, lookup gc.CacheKeyLookup, cache gc.CacheStorage) gc.Response {
	primary, secondary := lookup(req)
	if cache.Delete(primary, secondary) {
		return gc.PurgeHitResponse
	}
	return gc.PurgeMissResponse
}
