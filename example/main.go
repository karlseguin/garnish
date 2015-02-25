package main

import (
	"fmt"
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/gc"
	"time"
)

func main() {
	config := garnish.Configure().Address("127.0.0.1:8080").Debug()
	config.Hydrate(HydrateLoader)
	config.Stats().FileName("stats.json").Slow(time.Millisecond * 100)
	config.Cache().Grace(time.Minute).PurgeHandler(PurgeHandler)
	config.Upstream("test").Address("http://localhost:3000").KeepAlive(8)
	config.Route("users").Get("/v1/queue").Upstream("test").CacheTTL(time.Minute)
	runtime, err := config.Build()
	fmt.Println(runtime.Cache.Load("cache.save"))
	if err != nil {
		panic(err)
	}
	go garnish.Start(runtime)
	time.Sleep(time.Second * 10)
	fmt.Println(runtime.Cache.Save("cache.save", 10, time.Second*10))
}

func HydrateLoader(fragment gc.ReferenceFragment) []byte {
	return []byte(`{"id": "hyd-` + fragment.String("id") + `"}`)
}

func PurgeHandler(req *gc.Request, lookup gc.CacheKeyLookup, cache gc.CacheStorage) gc.Response {
	primary, secondary := lookup(req)
	if cache.Delete(primary, secondary) {
		return gc.PurgeHitResponse
	}
	return gc.PurgeMissResponse
}
