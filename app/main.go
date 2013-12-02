// Just some test code for now
package main

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/caches/ccache"
	"github.com/karlseguin/garnish/middleware/caching"
	"github.com/karlseguin/garnish/middleware/upstream"
	"github.com/karlseguin/garnish/routing/path"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mainConfig := garnish.Configure().LogInfo()

	routerConfig := path.Configure(mainConfig)
	routerConfig.Add("GET", "/", garnish.NewRoute("root", "openmymind").Cache(cacheKeyGenerator).TTL(time.Second*5))
	routerConfig.Fallback(garnish.NewRoute("fallback", "openmymind"))
	mainConfig.Router(path.Register(routerConfig))

	routerConfig.Fallback(garnish.NewRoute("fallback", "openmymind"))
	mainConfig.Router(path.Register(routerConfig))

	cachingConfig := caching.Configure(mainConfig, ccache.New(ccache.Configure()))
	mainConfig.Middleware(cachingConfig)

	upstreamConfig := upstream.Configure(mainConfig)
	upstreamConfig.Add(garnish.NewUpstream("openmymind", "http", "openmymind.net"))
	mainConfig.Middleware(upstreamConfig)

	g := garnish.New()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR2)
	go func() {
		for {
			<-sig
			g.Reload(mainConfig)
		}
	}()
	g.Start(mainConfig)
}

func cacheKeyGenerator(context garnish.Context) (string, string) {
	return "/", ""
}
