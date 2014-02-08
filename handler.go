package garnish

import (
	"errors"
	"net/http"
	"strconv"
)

type Handler struct {
	router Router
	head   *MiddlewareWrapper
	logger Logger
}

func newHandler(config *Configuration) (*Handler, error) {
	if config.router == nil {
		return nil, errors.New("A router must be provided")
	}
	h := &Handler{
		router: config.router,
		logger: config.Logger,
	}
	routeNames := config.router.RouteNames()
	prev, err := newMiddlewareWrapper(config, routeNames, 0)
	if err != nil {
		return nil, err
	}
	h.head = prev
	for i := 1; i < len(config.middlewareFactories); i++ {
		link, err := newMiddlewareWrapper(config, routeNames, i)
		if err != nil {
			return nil, err
		}
		prev.next = link
		prev = link
	}
	prev.next = &MiddlewareWrapper{
		logger:     config.Logger,
		middleware: new(notFoundMiddleware),
	}

	return h, nil
}

func (h *Handler) ServeHTTP(output http.ResponseWriter, req *http.Request) {
	context := newContext(req, h.logger)
	h.logger.Infof(context, "+ router %q", req.URL)
	route, params, response := h.router.Route(context)
	defer h.logger.Info(context, "- router")

	if response != nil {
		h.reply(context, response, output)
	} else if route == nil {
		h.logger.Info(context, "404")
		h.reply(context, NotFound, output)
	} else {
		context.route = route
		context.params = params
		h.reply(context, h.head.Yield(context), output)
	}
}

func (h *Handler) reply(context Context, response Response, output http.ResponseWriter) {
	defer response.Close()
	outHeader := output.Header()
	for k, v := range response.GetHeader() {
		outHeader[k] = v
	}

	body := response.GetBody()
	status := response.GetStatus()

	if status >= 500 {
		if fatal, ok := response.(*FatalResponse); ok {
			h.logger.Errorf(context, "%q - %v", context.RequestIn().URL, fatal.err)
		} else {
			LogError(h.logger, context, status, body)
		}
	}

	outHeader["Content-Length"] = []string{strconv.Itoa(len(body))}
	output.WriteHeader(status)
	output.Write(body)
}

func LogError(logger Logger, context Context, status int, body []byte) {
	logger.Errorf(context, "%q %d %v", context.RequestIn().URL, status, string(body))
}
