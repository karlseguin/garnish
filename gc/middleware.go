package gc

type Middleware func(req *Request, next Middleware) Response

type MiddlewareExecutor func(req *Request) Response

func WrapMiddleware(m Middleware, next Middleware) MiddlewareExecutor {
	return func(req *Request) Response {
		return m(req, next)
	}
}
