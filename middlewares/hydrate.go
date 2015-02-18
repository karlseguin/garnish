package middlewares

import (
	"bytes"
	"encoding/json"
	"github.com/karlseguin/garnish/gc"
)

type Hydrate struct {
	Header string
}

func (h *Hydrate) Handle(req *gc.Request, next gc.Middleware) gc.Response {
	res := next(req)
	if res == nil {
		return res
	}

	hydrateField := res.Header().Get(h.Header)
	if len(hydrateField) == 0 || res.Status() >= 300 {
		return res
	}
	res.Header().Del(h.Header)
	return h.convert(req, res, hydrateField)
}

func (h *Hydrate) convert(req *gc.Request, res gc.Response, fieldName string) gc.Response {
	body, position := loadBody(req.Runtime, res), 0
	needle := []byte("\"" + fieldName)
	fragments := make([]gc.Fragment, 0, 10)
	for {
		index := bytes.Index(body, needle)
		if index == -1 {
			fragments = append(fragments, gc.LiteralFragment(body[position:]))
			break
		}
		fragments = append(fragments, gc.LiteralFragment(body[position:index-1]))
		body = body[index:]
		start := bytes.IndexRune(body, '{')
		if start == -1 {
			req.Error("invalid hydration start")
			return res
		}
		end := bytes.IndexRune(body, '}')
		if end == -1 {
			req.Error("invalid hydration end")
			return res
		}
		end += 1
		var ref map[string]string
		if err := json.Unmarshal(body[start:end], &ref); err != nil {
			req.Errorf("invalid hydration payload: %v", err)
			return res
		}
		fragments = append(fragments, gc.NewReferenceFragment(ref, end-start))
		body = body[end+1:]
	}
	return gc.NewHydraterResponse(res.Status(), res.Header(), fragments)
}

func loadBody(runtime *gc.Runtime, res gc.Response) []byte {
	l := res.ContentLength()
	if l <= 0 {
		l = 32768
	}
	b := make([]byte, 0, l)
	buffer := bytes.NewBuffer(b)
	res.Write(runtime, buffer)
	return buffer.Bytes()
}
