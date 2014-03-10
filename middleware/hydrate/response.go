package hydrate

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/garnish/gc"
	"net/http"
)

type Response struct {
	status   int
	segments []Segment
	header   http.Header
	hydrator Hydrator
	pool     *bytepool.Pool
	buffer   *bytepool.Item
}

func (r *Response) GetHeader() http.Header {
	return r.header
}

func (r *Response) GetBody() []byte {
	r.buffer = r.pool.Checkout()
	for _, segment := range r.segments {
		r.buffer.Write(segment.Render(r.hydrator))
	}
	return r.buffer.Bytes()
}

func (r *Response) GetStatus() int {
	return r.status
}

func (r *Response) SetStatus(status int) {
	r.status = status
}

func (r *Response) Close() error {
	return r.buffer.Close()
}

func (r *Response) Detach() gc.Response {
	return r
}
