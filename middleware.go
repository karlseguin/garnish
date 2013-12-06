package garnish

type Middleware interface {
	Name() string
	Run(context Context, next Next) Response
}

type MiddlewareFactory interface {
	Create() (Middleware, error)
}

type Next func(context Context) Response

type MiddlewareWrapper struct {
	next       *MiddlewareWrapper
	middleware Middleware
	logger     Logger
}

func newMiddlewareWrapper(config *Configuration, index int) (*MiddlewareWrapper, error) {
	factory := config.middlewareFactories[index]
	middleware, err := factory.Create()
	if err != nil {
		return nil, err
	}
	return &MiddlewareWrapper{
		logger:     config.Logger,
		middleware: middleware,
	}, nil
}

func (wrapper *MiddlewareWrapper) Yield(context Context) Response {
	name := wrapper.middleware.Name()
	wrapper.logger.Infof(context, "<%v>", name)
	defer wrapper.logger.Infof(context, "</%v>", name)
	var next Next
	if wrapper.next != nil {
		next = wrapper.next.Yield
	}
	return wrapper.middleware.Run(context, next)
}

type notFoundMiddleware struct{}

func (m *notFoundMiddleware) Name() string {
	return "_notFound"
}

func (m *notFoundMiddleware) Configure(config interface{}) error {
	return nil
}

func (m *notFoundMiddleware) Run(context Context, next Next) Response {
	return NotFound
}

func FakeNext(r Response) Next {
	return func(context Context) Response {
		return r
	}
}
