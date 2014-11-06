package middlewares

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

//TODO: handle purge
func Cache(req *gc.Request, next gc.Middleware) gc.Response {
	cache := req.Runtime.Cache
	config := req.Route.Cache

	if req.Method == "PURGE" && cache.PurgeHandler != nil {
		if res := cache.PurgeHandler(req, config.KeyLookup, cache); res != nil {
			return res
		}
		return next(req)
	}

	if req.Method != "GET" || config.TTL < 0 {
		return next(req)
	}
	primary, secondary := config.KeyLookup(req)

	item := cache.Get(primary, secondary)
	if item != nil {
		now := time.Now()
		expires := item.Expires()
		if expires.After(now) {
			req.Info("hit")
			return item.Value().(gc.Response)
		}
		if expires.Add(cache.GraceTTL).After(now) {
			req.Info("grace")
			cache.Grace(primary, secondary, req, next)
			return item.Value().(gc.Response)
		}
	}

	req.Info("miss")
	res := next(req)
	if res == nil || res.Status() >= 500 {
		if item == nil || cache.Saint == false {
			return res
		}
		req.Info("saint")
		item.Extend(time.Second * 5)
		return item.Value().(gc.Response)
	}
	cache.Set(primary, secondary, config, res)
	return res
}
