// Middleware package which makes request to
// the upstream service
package upstream

import (
	"bytes"
	"errors"
	"github.com/karlseguin/garnish/core"
	"io"
	"net/http"
)

type Upstream struct {
	*Configuration
}

func (u *Upstream) Name() string {
	return "upstream"
}

func (u *Upstream) Run(context core.Context, next core.Next) core.Response {
	route := context.Route()
	upstream, ok := u.routeLookup[route.Name]
	if ok == false {
		return context.Fatal(errors.New("Upstream not configured for route: " + route.Name))
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
		return &core.ClosableResponse{
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

	return core.RespondH(body, r.Header).Status(r.StatusCode)
}

func (u *Upstream) prepareRequest(context core.Context, server *Server) *http.Request {
	in := context.Request()
	out := createRequest(context.RequestId())
	if len(out.Host) == 0 {
		out.Host = server.Host
	}
	if len(out.Method) == 0 {
		out.Method = in.Method
	}
	if out.URL == nil {
		out.URL = in.URL
		out.URL.Host = server.Host
		out.URL.Scheme = server.Scheme
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

func createRequest(id string) *http.Request {
	return &http.Request{
		Close:      false,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{"X-Request-Id": []string{id}},
	}
}
