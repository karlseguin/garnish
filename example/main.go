package main

import (
	"fmt"
	"gopkg.in/karlseguin/garnish.v1"
	"gopkg.in/karlseguin/garnish.v1/gc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	runtime, err := loadRuntime()
	if err != nil {
		panic(err)
	}
	if err := runtime.Cache.Load("cache.save"); err != nil {
		fmt.Println("failed to restore cache", err)
	}

	go garnish.Start(runtime)

	go func() {
		sigusr2 := make(chan os.Signal, 1)
		signal.Notify(sigusr2, syscall.SIGUSR2)
		for {
			<-sigusr2
			runtime, err := loadRuntime()
			if err != nil {
				fmt.Println("invalid configuration, did not reload.", err)
			} else {
				garnish.Reload(runtime)
				fmt.Println("reloaded")
			}
		}
	}()

	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, syscall.SIGQUIT)
	<-sigquit
	if err := runtime.Cache.Save("cache.save", 5000, time.Second*10); err != nil {
		fmt.Println("failed to save cache", err)
	}
}

func loadRuntime() (*garnish.Runtime, error) {
	config, err := gc.LoadConfig("sample.toml")
	if err != nil {
		return nil, err
	}
	config.Hydrate(HydrateLoader)
	config.Stats().FileName("stats.json").Slow(time.Millisecond * 100)
	config.Cache().Grace(time.Minute).PurgeHandler(PurgeHandler)
	config.NotFound(garnish.Json(404, `{"error":"not found", "code":404}`))
	config.Fatal(garnish.Json(500, `{"error":"server error", "code":500}`))

	config.Upstream("users").Address("http://localhost:3000").KeepAlive(8)
	config.Route("users").Get("/v1/users").Upstream("users").CacheTTL(time.Minute)
	config.Route("ping").Get("/v1/ping").Handler(func(reg *garnish.Request, next garnish.Middleware) garnish.Response {
		return garnish.Respond(200, "ok")
	})
	return config.Build()
}

func HydrateLoader(fragment garnish.ReferenceFragment) []byte {
	return []byte(`{"id": "hyd-` + fragment.String("id") + `"}`)
}

func PurgeHandler(req *garnish.Request, lookup garnish.CacheKeyLookup, cache garnish.CacheStorage) garnish.Response {
	primary, secondary := lookup(req)
	if cache.Delete(primary, secondary) {
		return garnish.PurgeHitResponse
	}
	return garnish.PurgeMissResponse
}
