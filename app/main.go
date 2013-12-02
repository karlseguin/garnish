// Just some test code for now
package main

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/middleware/pathrouter"
	"github.com/karlseguin/garnish/middleware/upstream"
)

func main() {
	mainConfig := garnish.Configure().LogInfo()

	routerConfig := pathrouter.Configure(mainConfig)
	routerConfig.Add("GET", "/", garnish.NewRoute("root", "openmymind"))
	routerConfig.Fallback(garnish.NewRoute("fallback", "openmymind"))
	mainConfig.Router(pathrouter.Register(routerConfig))

	upstreamConfig := upstream.Configure(mainConfig)
	upstreamConfig.Add(garnish.NewUpstream("openmymind", "http", "openmymind.net"))

	mainConfig.Middleware(upstreamConfig)
	garnish.Start(mainConfig)
}
