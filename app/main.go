// Just some test code for now
package main

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/core"
	// "github.com/karlseguin/garnish/caches/ccache"
	// "github.com/karlseguin/garnish/middleware/caching"
	// "github.com/karlseguin/garnish/middleware/stats"
	// "github.com/karlseguin/garnish/middleware/upstream"
	// "github.com/karlseguin/garnish/routing/path"
	// "os"
	// "os/signal"
	// "syscall"
	"time"
)

func main() {
	stats, upstream := garnish.Stats, garnish.Upstream
	config := garnish.Configure().LogInfo().Middleware(stats, upstream)
	router := config.NewRouter()

	stats.Percentiles(50, 75, 95)//.Treshold(time.Milliseconds * 25)

	// caching.Generator(cacheKeyGenerator).TTL(time.Seconds * 5)

	upstream.DnsRefresh(time.Minute)
	upstream.Add("openmymind", "http", "openmymind.net")

	router.Add("root", "GET", "/", func() {
		upstream.Is("openmymind")
		stats.Percentiles(50)
	})


	// routerConfig := path.Configure(mainConfig)
	// routerConfig.Add("GET", "/", garnish.NewRoute("root").Cache(cacheKeyGenerator).TTL(time.Second*5))

	// mainConfig.Router(path.Register(routerConfig))

	// statsConfig := stats.Configure(mainConfig).Percentiles(45, 80, 90, 99)
	// mainConfig.Middleware(statsConfig)

	// cachingConfig := caching.Configure(mainConfig, ccache.New(ccache.Configure()))
	// cachingConfig.AuthorizePurge(func(context garnish.Context) bool {
	// 	return true
	// })
	// mainConfig.Middleware(cachingConfig)

	// upstreamConfig := upstream.Configure(mainConfig)
	// upstreamConfig.Add(upstream.NewServer("openmymind", "http", "openmymind.net"), "root", "fallback")
	// mainConfig.Middleware(upstreamConfig)

	g := garnish.New()
	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGUSR2)
	// go func() {
	// 	for {
	// 		<-sig
	// 		g.Reload(mainConfig)
	// 	}
	// }()
	g.Start(config)
}

func cacheKeyGenerator(context core.Context) (string, string) {
	return "/", ""
}
