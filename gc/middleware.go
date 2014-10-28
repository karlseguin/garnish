package gc

type MiddlewareHandler func(req *Request, next Middleware) Response

type Middleware func(req *Request) Response

func WrapMiddleware(m MiddlewareHandler, next Middleware) Middleware {
	return func(req *Request) Response {
		return m(req, next)
	}
}
