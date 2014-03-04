// Middleware package which makes request to
// the upstream service
package upstream

import (
	"bytes"
	"errors"
	"github.com/karlseguin/garnish/gc"
	"io"
	"net/http"
)

type Upstream struct {
	*Configuration
}

func (u *Upstream) Name() string {
	return "upstream"
}

func (u *Upstream) Run(context gc.Context, next gc.Next) gc.Response {
	route := context.Route()
	server, ok := u.routeLookup[route.Name]
	if ok == false {
		return context.Fatal(errors.New("Upstream not configured for route: " + route.Name))
	}

	request := u.prepareRequest(context, server)
	context.Infof("request to %v", request.URL.String())
	r, err := server.Transport.RoundTrip(request)
	if err != nil {
		return context.Fatal(err)
	}
	defer r.Body.Close()
	length := int(r.ContentLength)
	defer context.Infof("%d response %d bytes", r.StatusCode, length)
	if length > 0 && length < server.PoolItemSize {
		buffer := server.Pool.Checkout()
		buffer.ReadFrom(r.Body)
		return &gc.ClosableResponse{
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
		buffer := bytes.NewBuffer(make([]byte, 0, server.PoolItemSize))
		io.Copy(buffer, r.Body)
		body = buffer.Bytes()
		length = len(body)
	}
	return gc.RespondH(body, r.Header).Status(r.StatusCode)
}

func (u *Upstream) prepareRequest(context gc.Context, server *Server) *http.Request {
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
