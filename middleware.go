package garnish

type Middleware func(context Context, next Next) Response
type Next func(context Context) Response

type MiddlewareWrapper struct {
	next       *MiddlewareWrapper
	middleware Middleware
	logger     Logger
	name       string
}

func (wrapper *MiddlewareWrapper) Yield(context Context) Response {
	wrapper.logger.Infof(context, "+[%v]", wrapper.name)
	defer wrapper.logger.Infof(context, "-[%v]", wrapper.name)
	var next Next
	if wrapper.next != nil {
		next = wrapper.next.Yield
	}
	return wrapper.middleware(context, next)
}

func notFoundMiddleware(context Context, next Next) Response {
	return NotFound
}
