package hydrate

import (
	"bytes"
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/garnish/core"
)

type Hydrate struct {
	pool *bytepool.Pool
}

func (h *Hydrate) Name() string {
	return "hydrate"
}

func (h *Hydrate) Run(context core.Context, next core.Next) core.Response {
	response := next(context)
	if response == nil {
		return response
	}
	status := response.GetStatus()
	if response.GetHeader().Get("X-Hydrate") != "true" || status < 200 || status > 299 {
		return response
	}
	context.Info("hydrating")
	return h.parse(response)
}

func (h *Hydrate) parse(response core.Response) core.Response {
	body := response.GetBody()
	segments := make([]Segment, 0, 10)
	position := 0
	for len(body) > 0 {
		index := bytes.IndexByte(body, byte('<'))
		if index == -1 {
			segments = append(segments, &LiteralSegment{detach(body)})
			break
		}

		if body[index+1] != '%' {
			break
		}

		if index > 0 {
			segments = append(segments, &LiteralSegment{detach(body[position:index])})
		}

		segment, b := createPlaceholderSegment(body[index+2:])
		if segment != nil {
			segments = append(segments, segment)
		}
		body = b
	}
	r := &Response{
		pool:     h.pool,
		segments: segments,
		status:   response.GetStatus(),
		header:   response.GetHeader(),
	}
	r.header.Del("Content-Length")
	r.header.Del("X-Hydrate")
	response.Close()
	return r
}

func createPlaceholderSegment(data []byte) (Segment, []byte) {
	for i, l := 0, len(data); i < l; i++ {
		if data[i] != ' ' {
			data = data[i:]
			break
		}
	}
	var id []byte
	for i, l := 0, len(data); i < l; i++ {
		c := data[i]
		if c == ' ' || c == '%' {
			id = data[:i]
			data = data[i:]
			break
		}
	}

	for i, l := 0, len(data); i < l; i++ {
		if data[i] == '%' {
			data = data[i+2:]
			break
		}
	}

	var segment *PlaceholderSegment
	if len(id) > 0 {
		segment = &PlaceholderSegment{detach(id)}
	}
	return segment, data
}

func detach(data []byte) []byte {
	clone := make([]byte, len(data))
	copy(clone, data)
	return clone
}
