package gc

type MiddlewareHandler func(req *Request, next Middleware) Response

type Middleware func(req *Request) Response

func WrapMiddleware(name string, m MiddlewareHandler, next Middleware) Middleware {
	return func(req *Request) Response {
		old := req.scope
		req.scope = name
		res := m(req, next)
		req.scope = old
		return res
	}
}
