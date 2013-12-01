// Middleware package which makes request to
// the upstream service
package upstream

import (
	"bytes"
	"errors"
	"github.com/karlseguin/dnscache"
	"github.com/karlseguin/garnish"
	"io"
	"net/http"
)

type Upstream struct {
	*Configuration
}

func Register(config *Configuration) (string, garnish.Middleware) {
	dns := dnscache.New(config.dnsRefresh)
	u := &Upstream{config}

	for _, upstream := range config.upstreams {
		upstream.Resolver(dns.FetchOneString)
		upstream.Finalize()
	}
	return "upstream", u.run
}

func (u *Upstream) run(context garnish.Context, next garnish.Next) garnish.Response {
	route := context.Route()
	upstream, ok := u.upstreams[route.Upstream]
	if ok == false {
		return context.Fatal(errors.New("Upstream not configured for host: " + route.Upstream))
	}

	request := u.prepareRequest(context, upstream)
	u.logger.Infof(context, "request to %v", request.URL.String())
	r, err := upstream.Transport.RoundTrip(request)
	if err != nil {
		return context.Fatal(err)
	}
	defer r.Body.Close()
	length := int(r.ContentLength)
	defer u.logger.Infof(context, "%d response %d bytes", r.StatusCode, length)
	if length > 0 && length < upstream.PoolItemSize {
		buffer := upstream.Pool.Checkout()
		buffer.ReadFrom(r.Body)
		return &garnish.ClosableResponse{
			S: r.StatusCode,
			B: buffer,
			H: r.Header,
		}
	}

	var body []byte
	if length > 0 {
		body = make([]byte, r.ContentLength)
		io.ReadFull(r.Body, body)
	} else if length == -1 {
		buffer := bytes.NewBuffer(make([]byte, 0, upstream.PoolItemSize))
		io.Copy(buffer, r.Body)
		body = buffer.Bytes()
		length = len(body)
	}

	return &garnish.ResponseBuilder{
		S: r.StatusCode,
		B: body,
		H: r.Header,
	}
}

func (u *Upstream) prepareRequest(context garnish.Context, upstream *garnish.Upstream) *http.Request {
	out := context.RequestOut()
	in := context.RequestIn()
	if len(out.Host) == 0 {
		out.Host = upstream.Host
	}
	if len(out.Method) == 0 {
		out.Method = in.Method
	}
	if out.URL == nil {
		out.URL = in.URL
		out.URL.Host = upstream.Host
		out.URL.Scheme = upstream.Scheme
	}

	inHeader := in.Header
	outHeader := out.Header
	for _, headerName := range u.forwardHeaders {
		value := inHeader.Get(headerName)
		if len(value) != 0 {
			outHeader.Set(headerName, value)
		}
	}
	return out
}
