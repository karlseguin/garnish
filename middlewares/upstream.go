package middlewares

import (
	"bytes"
	"github.com/karlseguin/garnish/gc"
	"io"
	"net/http"
	"net/url"
)

var DefaultUserAgent = []string{""}

func Upstream(req *gc.Request, next gc.Middleware) gc.Response {
	upstream := req.Route.Upstream
	if upstream == nil {
		return next(req)
	}
	request := createRequest(req, upstream)
	r, err := upstream.Transport.RoundTrip(request)
	if err != nil {
		return gc.FatalErr(err)
	}
	defer r.Body.Close()
	length := int(r.ContentLength)
	if length > 0 && length < gc.BytePoolItemSize {
		buffer := gc.BytePool.Checkout()
		buffer.ReadFrom(r.Body)
		return gc.RespondH(r.StatusCode, r.Header, buffer)
	}

	var body []byte
	if length > 0 {
		body = make([]byte, r.ContentLength)
		io.ReadFull(r.Body, body)
	} else if length == -1 {
		buffer := bytes.NewBuffer(make([]byte, 0, gc.BytePoolItemSize))
		io.Copy(buffer, r.Body)
		body = buffer.Bytes()
		length = len(body)
	}
	req.Info("%s | %d | %d", request.URL.String(), r.StatusCode, length)
	return gc.RespondH(r.StatusCode, r.Header, body)
}

func createRequest(in *gc.Request, upstream *gc.Upstream) *http.Request {
	u, err := url.Parse(upstream.Address + in.URL.RequestURI())
	if err != nil {
		in.Error("upstream url %s %v", upstream.Address+in.URL.RequestURI(), err)
		u = in.URL
	}
	out := &http.Request{
		Close:      false,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Host:       in.Host,
		Method:     in.Method,
		URL:        u,
		Header:     http.Header{"X-Request-Id": []string{in.Id}, "User-Agent": DefaultUserAgent},
	}

	for _, k := range upstream.Headers {
		value := in.Header[k]
		if len(value) > 0 {
			out.Header[k] = value
		}
	}

	if upstream.Tweaker != nil {
		upstream.Tweaker(in, out)
	}
	return out
}
