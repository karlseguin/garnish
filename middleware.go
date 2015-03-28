package garnish

type Handler func(req *Request) Response
type Middleware func(req *Request, next Handler) Response

func WrapMiddleware(name string, m Middleware, next Handler) Handler {
	return func(req *Request) Response {
		old := req.scope
		req.scope = name
		res := m(req, next)
		req.scope = old
		return res
	}
}
