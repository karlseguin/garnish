package core

type Middleware interface {
	Name() string
	Run(context Context, next Next) Response
}

type MiddlewareFactory interface {
	Create(Configuration) (Middleware, error)
	OverrideFor(*Route)
}

type Next func(context Context) Response


func FakeNext(r Response) Next {
	return func(context Context) Response {
		return r
	}
}
