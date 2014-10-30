package main

import (
	"github.com/karlseguin/garnish"
	"time"
)

func main() {
	config := garnish.Configure().Address("127.0.0.1:8080")
	config.Stats().FileName("stats.json").Slow(time.Millisecond * 100)
	config.Cache().Grace(time.Minute)
	config.Upstream("test").Address("http://localhost:4005").KeepAlive(8)
	config.Route("users").Get("/user/:id").Upstream("test").CacheTTL(time.Second * 5)
	garnish.Start(config)
}
