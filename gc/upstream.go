package gc

import (
	"net/http"
	"net/url"
)

// The User Agent to send to the upstream
var DefaultUserAgent = []string{""}

// Tweaks request `out` before sending it to the upstream
type RequestTweaker func(in *Request, out *http.Request)

type Upstream struct {
	Name      string
	Address   string
	Transport *http.Transport
	Headers   []string
	Tweaker   RequestTweaker
}

// Issues a request to the upstream
func (u *Upstream) RoundTrip(in *Request) (*http.Response, error) {
	return u.Transport.RoundTrip(u.createRequest(in))
}

func (u *Upstream) createRequest(in *Request) *http.Request {
	targetUrl, err := url.Parse(u.Address + in.URL.RequestURI())
	if err != nil {
		in.Error("upstream url %s %v", u.Address+in.URL.RequestURI(), err)
		targetUrl = in.URL
	}
	out := &http.Request{
		URL:           targetUrl,
		Close:         false,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Host:          in.Host,
		Body:          in.Body,
		Method:        in.Method,
		ContentLength: in.ContentLength,
		Header:        http.Header{"X-Request-Id": []string{in.Id}, "User-Agent": DefaultUserAgent},
	}

	for _, k := range u.Headers {
		value := in.Header[k]
		if len(value) > 0 {
			out.Header[k] = value
		}
	}

	if u.Tweaker != nil {
		u.Tweaker(in, out)
	}
	return out
}
