package hydrate

import (
	"net/http"
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/bytepool"
)

type Response struct {
	pool *bytepool.Pool
	status int
	segments []Segment
	header http.Header
}

func (r *Response) GetHeader() http.Header {
	return r.header
}

func (r *Response) GetBody() []byte {
	buffer := r.pool.Checkout()
	defer buffer.Close()
	for _, segment := range r.segments {
		buffer.Write(segment.Render())
	}
	return buffer.Bytes()
}

func (r *Response) GetStatus() int {
	return r.status
}

func (r *Response) SetStatus(status int)  {
	r.status = status
}

func (r *Response) Close() error {
	return nil
}

func (r *Response) Detach() core.Response {
	return r
}
