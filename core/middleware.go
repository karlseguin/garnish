package core

type Middleware interface {
	Name() string
	Run(context Context, next Next) Response
}

type MiddlewareFactory interface {
	Create(routeNames []string) (Middleware, error)
	Logger(logger Logger)
}

type Next func(context Context) Response
