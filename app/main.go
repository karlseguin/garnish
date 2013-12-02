// Just some test code for now
package main

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/routing/path"
	"github.com/karlseguin/garnish/middleware/upstream"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	mainConfig := garnish.Configure().LogInfo()

	routerConfig := path.Configure(mainConfig)
	routerConfig.Add("GET", "/", garnish.NewRoute("root", "openmymind"))
	routerConfig.Fallback(garnish.NewRoute("fallback", "openmymind"))
	mainConfig.Router(path.Register(routerConfig))

	routerConfig.Fallback(garnish.NewRoute("fallback", "openmymind"))
	mainConfig.Router(path.Register(routerConfig))

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
