package middlewares

import (
	"gopkg.in/karlseguin/garnish.v1"
	"time"
)

func Stats(req *garnish.Request, next garnish.Handler) garnish.Response {
	res := next(req)
	if res == nil {
		return nil
	}
	elapsed := time.Now().Sub(req.Start)
	req.Route.Stats.Hit(res, elapsed)
	req.Infof("%d µs", elapsed/1000)
	return res
}
