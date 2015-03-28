package middlewares

import (
	"gopkg.in/karlseguin/garnish.v1/gc"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func Upstream(req *gc.Request, next gc.Middleware) gc.Response {
	r, err := roundTrip(req)
	if err != nil {
		return req.FatalResponseErr("upstream roundtrip", err)
	}
	if r == nil {
		return Catch(req)
	}
	req.Infof("%s | %d | %d", req.URL, r.StatusCode, r.ContentLength)
	r.Header.Del("Connection")
	return gc.Streaming(r.StatusCode, r.Header, r.ContentLength, r.Body)
}

func roundTrip(req *gc.Request) (*http.Response, error) {
	upstream := req.Route.Upstream
	if upstream == nil {
		return nil, nil
	}

	transport := upstream.Transport()
	if transport == nil {
		//log?
		return nil, nil
	}
	return transport.RoundTrip(createRequest(req, transport, upstream))
}

func createRequest(in *gc.Request, transport *gc.Transport, upstream gc.Upstream) *http.Request {
	targetUrl, err := url.Parse(transport.Address + in.URL.RequestURI())
	if err != nil {
		in.Errorf("upstream url %s %v", transport.Address+in.URL.RequestURI(), err)
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

	for _, k := range upstream.Headers() {
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

	if tweaker := upstream.Tweaker(); tweaker != nil {
		tweaker(in, out)
	}
	return out
}
