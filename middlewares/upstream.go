package middlewares

import (
	"github.com/karlseguin/garnish/gc"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func Upstream(req *gc.Request, next gc.Middleware) gc.Response {
	r, err := roundTrip(req)
	if err != nil {
		return gc.FatalErr(err)
	}
	if r == nil {
		return Catch(req)
	}
	req.Infof("%s | %d | %d", req.URL, r.StatusCode, r.ContentLength)
	r.Header.Del("Connection")
	return gc.Streaming(r.StatusCode, r.Header, r.ContentLength, r.Body)
}

func roundTrip(req *gc.Request) (*http.Response, error) {
	if upstream := req.Route.Upstream; upstream != nil {
		return upstream.Transport.RoundTrip(createRequest(req, upstream))
	}
	return nil, nil
}

func createRequest(in *gc.Request, upstream *gc.Upstream) *http.Request {
	targetUrl, err := url.Parse(upstream.Address + in.URL.RequestURI())
	if err != nil {
		in.Errorf("upstream url %s %v", upstream.Address+in.URL.RequestURI(), err)
		targetUrl = in.URL
	}
	out := &http.Request{
		URL:           targetUrl,
		Close:         false,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Host:          in.Host,
		Method:        in.Method,
		ContentLength: in.ContentLength,
		Header:        http.Header{"X-Request-Id": []string{in.Id}, "User-Agent": gc.DefaultUserAgent},
	}

	if in.B != nil {
		out.Body = in.B
		in.B = nil //the upstream call will take care of closing it
	} else {
		out.Body = in.Request.Body
	}

	for _, k := range upstream.Headers {
		value := in.Header[k]
		if len(value) > 0 {
			out.Header[k] = value
		}
	}

	if clientIP, _, err := net.SplitHostPort(in.RemoteAddr); err == nil {
		if prior, ok := out.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		out.Header.Set("X-Forwarded-For", clientIP)
	}

	if upstream.Tweaker != nil {
		upstream.Tweaker(in, out)
	}
	return out
}
