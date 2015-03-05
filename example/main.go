package main

import (
	"fmt"
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/gc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// config := garnish.Configure().Address("127.0.0.1:8080").Debug()
	// config.Hydrate(HydrateLoader)
	// config.Stats().FileName("stats.json").Slow(time.Millisecond * 100)
	// config.Cache().Grace(time.Minute).PurgeHandler(PurgeHandler)
	//
	// config.Upstream("test").Address("http://localhost:3000").KeepAlive(8)
	// config.Route("users").Get("/v1/users").Upstream("test").CacheTTL(time.Minute)
	// config.Route("ping").Get("/v1/ping").Handler(func(reg *gc.Request, next gc.Middleware) gc.Response {
	// 	return gc.Respond(200, "ok")
	// })
	config, err := garnish.LoadConfig("sample.toml")
	if err != nil {
		panic(err)
	}
	runtime, err := config.Build()
	if err != nil {
		panic(err)
	}
	if err := runtime.Cache.Load("cache.save"); err != nil {
		fmt.Println("failed to restore cache", err)
	}
	go garnish.Start(runtime)
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGQUIT)
	<-s
	if err := runtime.Cache.Save("cache.save", 5000, time.Second*10); err != nil {
		fmt.Println("failed to save cache", err)
	}
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
