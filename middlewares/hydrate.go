package middlewares

import (
	"bytes"
	"gopkg.in/karlseguin/garnish.v1"
)

type Hydrate struct {
	Header string
}

func (h *Hydrate) Handle(req *garnish.Request, next garnish.Handler) garnish.Response {
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

func (h *Hydrate) convert(req *garnish.Request, res garnish.Response, fieldName string) garnish.Response {
	body := loadBody(req.Runtime, res)
	fragments := ExtractFragments(body, req, fieldName)
	if fragments == nil {
		return res
	}
	return garnish.NewHydraterResponse(res.Status(), res.Header(), fragments)
}

var ExtractFragments = func(body []byte, req *garnish.Request, fieldName string) []garnish.Fragment {
	position := 0
	needle := []byte("\"" + fieldName)
	fragments := make([]garnish.Fragment, 0, 20)
	for {
		index := bytes.Index(body, needle)
		if index == -1 {
			fragments = append(fragments, garnish.LiteralFragment(body[position:]))
			break
		}
		fragments = append(fragments, garnish.LiteralFragment(body[position:index-1]))
		body = body[index:]
		start := bytes.IndexRune(body, '{')
		if start == -1 {
			req.Error("invalid hydration start")
			return nil
		}
		end := bytes.IndexRune(body, '}')
		if end == -1 {
			req.Error("invalid hydration end")
			return nil
		}
		end += 1
		fragment, err := garnish.NewReferenceFragment(body[start:end])
		if err != nil {
			req.Errorf("invalid hydration payload: %v", err)
			return nil
		}
		fragments = append(fragments, fragment)
		body = body[end+1:]
	}
	return fragments
}

func loadBody(runtime *garnish.Runtime, res garnish.Response) []byte {
	l := res.ContentLength()
	if l <= 0 {
		l = 32768
	}
	b := make([]byte, 0, l)
	buffer := bytes.NewBuffer(b)
	res.Write(runtime, buffer)
	return buffer.Bytes()
}
