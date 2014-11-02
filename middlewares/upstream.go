package middlewares

import (
	"github.com/karlseguin/garnish/gc"
)

func Upstream(req *gc.Request, next gc.Middleware) gc.Response {
	upstream := req.Route.Upstream
	if upstream == nil {
		return Catch(req)
	}
	r, err := upstream.RoundTrip(req)
	if err != nil {
		return gc.FatalErr(err)
	}
	defer r.Body.Close()
	buffer := req.Runtime.BytePool.Checkout()
	buffer.ReadFrom(r.Body)
	req.Info("%s | %d | %d", req.URL.String(), r.StatusCode, buffer.Len())
	return gc.RespondH(r.StatusCode, r.Header, buffer)
}
