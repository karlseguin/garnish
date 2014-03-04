package garnish

import (
	"errors"
	"github.com/karlseguin/garnish/gc"
	"net/http"
	"strconv"
)

type Handler struct {
	router gc.Router
	logger gc.Logger
	head   *middlewareWrapper
}

func newHandler(config *Configuration) (*Handler, error) {
	if config.router == nil {
		return nil, errors.New("A router must be provided")
	}
	if !config.router.IsValid() {
		return nil, nil
	}
	h := &Handler{
		router: config.router,
		logger: config.logger,
	}
	prev, err := newMiddlewareWrapper(config, 0)
	if err != nil {
		return nil, err
	}
	h.head = prev
	for i := 1; i < len(config.middlewareFactories); i++ {
		link, err := newMiddlewareWrapper(config, i)
		if err != nil {
			return nil, err
		}
		prev.next = link
		prev = link
	}
	prev.next = &middlewareWrapper{middleware: new(NotFoundMiddleware)}

	return h, nil
}

func (h *Handler) ServeHTTP(output http.ResponseWriter, req *http.Request) {
	context := newContext(req, h.logger)
	context.Info(req.URL)
	route, params, response := h.router.Route(context)

	if response != nil {
		h.reply(context, response, output)
	} else if route == nil {
		context.Info("404")
		h.reply(context, NotFound, output)
	} else {
		context.route = route
		context.params = params
		h.reply(context, h.head.Yield(context), output)
	}
}

func (h *Handler) reply(context gc.Context, response gc.Response, output http.ResponseWriter) {
	defer response.Close()
	outHeader := output.Header()
	for k, v := range response.GetHeader() {
		outHeader[k] = v
	}

	body := response.GetBody()
	status := response.GetStatus()

	if status >= 500 {
		if fatal, ok := response.(*gc.FatalResponse); ok {
			context.Errorf("%q - %v", context.Request().URL, fatal.Err)
		} else {
			context.Errorf("%q %d %v", context.Request().URL, status, string(body))
		}
	}

	outHeader["Content-Length"] = []string{strconv.Itoa(len(body))}
	output.WriteHeader(status)
	output.Write(body)
}

type middlewareWrapper struct {
	next       *middlewareWrapper
	middleware gc.Middleware
}

func (wrapper *middlewareWrapper) Yield(context gc.Context) gc.Response {
	defer context.SetLocation(context.Location())
	context.SetLocation(wrapper.middleware.Name())
	var next gc.Next
	if wrapper.next != nil {
		next = wrapper.next.Yield
	}
	return wrapper.middleware.Run(context, next)
}

func newMiddlewareWrapper(config *Configuration, index int) (*middlewareWrapper, error) {
	factory := config.middlewareFactories[index]
	middleware, err := factory.Create(config)
	if err != nil {
		return nil, err
	}
	return &middlewareWrapper{middleware: middleware}, nil
}
