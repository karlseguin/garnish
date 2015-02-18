package middlewares

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

//TODO: handle purge
func Cache(req *gc.Request, next gc.Middleware) gc.Response {
	cache := req.Runtime.Cache
	config := req.Route.Cache

	if req.Method == "PURGE" && cache.PurgeHandler != nil && config != nil {
		if res := cache.PurgeHandler(req, config.KeyLookup, cache.Storage); res != nil {
			return res
		}
		return next(req)
	}

	if req.Method != "GET" || config.TTL < 0 {
		return next(req)
	}
	primary, secondary := config.KeyLookup(req)

	item := cache.Storage.Get(primary, secondary)
	if item != nil {
		now := time.Now()
		expires := item.Expires()
		if expires.After(now) {
			req.Cached("hit")
			return item
		}
		if expires.Add(cache.GraceTTL).After(now) {
			cache.Grace(primary, secondary, req, next)
			req.Cached("grace")
			return item
		}
	}

	req.Info("miss")
	res := next(req)
	if res == nil || res.Status() >= 500 {
		if item == nil || cache.Saint == false {
			return res
		}
		item.Expire(time.Now().Add(time.Second * 5))
		req.Cached("saint")
		return item
	}
	cache.Set(primary, secondary, config, res)
	return res
}
