package middlewares

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

func Stats(req *gc.Request, next gc.Middleware) gc.Response {
	res := next(req)
	if res == nil {
		return nil
	}
	elapsed := time.Now().Sub(req.Start)
	req.Route.Stats.Hit(res, elapsed)
	req.Infof("%d µs", elapsed/1000)
	return res
}
