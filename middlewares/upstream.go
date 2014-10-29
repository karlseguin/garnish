package middlewares

import (
	"bytes"
	"github.com/karlseguin/garnish/gc"
	"io"
)

func Upstream(req *gc.Request, next gc.Middleware) gc.Response {
	upstream := req.Route.Upstream
	if upstream == nil {
		return next(req)
	}
	r, err := upstream.RoundTrip(req)
	if err != nil {
		return gc.FatalErr(err)
	}
	defer r.Body.Close()
	length := int(r.ContentLength)

	runtime := req.Runtime
	capacity := runtime.BytePool.Capacity()
	if length > 0 && length < capacity {
		buffer := runtime.BytePool.Checkout()
		buffer.ReadFrom(r.Body)
		return gc.RespondH(r.StatusCode, r.Header, buffer)
	}

	var body []byte
	if length > 0 {
		body = make([]byte, r.ContentLength)
		io.ReadFull(r.Body, body)
	} else if length == -1 {
		buffer := bytes.NewBuffer(make([]byte, 0, capacity))
		io.Copy(buffer, r.Body)
		body = buffer.Bytes()
		length = len(body)
	}
	req.Info("%s | %d | %d", req.URL.String(), r.StatusCode, length)
	return gc.RespondH(r.StatusCode, r.Header, body)
}
