package garnish

import (
	"net/http"
	"strconv"
)

type Handler struct {
	router Router
	head   *MiddlewareWrapper
	logger Logger
}

func newHandler(config *Configuration) (*Handler, error) {
	h := &Handler{
		router: config.router,
		logger: config.Logger,
	}

	prev, err := newMiddlewareWrapper(config, 0)
	if err != nil { return nil, err }
	h.head = prev
	for i := 1; i < len(config.middlewareFactories); i++ {
		link, err := newMiddlewareWrapper(config, i)
		if err != nil { return nil, err }
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
	route, response := h.router(context)
	if response != nil {
		h.reply(context, response, output)
	} else if route == nil {
		h.reply(context, NotFound, output)
	} else {
		context.route = route
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
		h.logger.Errorf(context, "%q %d %v", context.RequestIn().URL, status, string(body))
	}

	outHeader["Content-Length"] = []string{strconv.Itoa(len(body))}
	output.WriteHeader(status)
	output.Write(body)
}
