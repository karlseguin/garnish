package garnish

import (
	"github.com/karlseguin/gspec"
	"net/http"
	// "net/url"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHandlerRepliesWithTheRoutersResponse(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(Route), Respond(nil).Status(401), nilMiddleware)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, nil)
	spec.Expect(rec.Code).ToEqual(401)
}

func TestHandlerRepliesWithNotFoundIfRouteIsNotSet(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(nil, nil, nilMiddleware)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, nil)
	spec.Expect(rec.Code).ToEqual(404)
}

func TestHandlerCallsTheMiddlewareChain(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(Route), nil, nextMiddleware, responseMiddleware(201, "ok", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, nil)
	spec.Expect(rec.Code).ToEqual(201)
	spec.Expect(rec.Body.String()).ToEqual("ok")
}

func TestHandlerWritesTheContentLength(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(Route), nil, responseMiddleware(201, "it's over 9000", http.Header{"Content-Length": []string{"32"}}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, nil)
	spec.Expect(rec.HeaderMap.Get("Content-Length")).ToEqual("14")
}

func TestHandlerWritesHeaders(t *testing.T) {
	spec := gspec.New(t)
	h := buildHandler(new(Route), nil, responseMiddleware(201, "it's over 9000", http.Header{"X-Test": []string{"leto"}}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, nil)
	spec.Expect(rec.HeaderMap.Get("X-Test")).ToEqual("leto")
}

func TestLogsInternalServerErrors(t *testing.T) {
	h := buildHandler(new(Route), nil, responseMiddleware(505, "error", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, gspec.Request().Url("http://fake.garnish.io/fail").Req)
	h.logger.(*FakeLogger).Assert(t, FakeLogMessage{false, `"http://fake.garnish.io/fail" 505 error`})
}

func buildHandler(route *Route, response Response, middlewares ...Middleware) *Handler {
	router := func(context Context) (*Route, Response) { return route, response }
	config := Configure(router)
	for index, middleware := range middlewares {
		config.Middleware(strconv.Itoa(index), middleware)
	}
	config.Logger = new(FakeLogger)
	return newHandler(config)
}

func nilMiddleware(context Context, next Next) Response {
	return nil
}

func nextMiddleware(context Context, next Next) Response {
	return next(context)
}

func responseMiddleware(status int, body string, header http.Header) Middleware {
	return func(context Context, next Next) Response {
		r := Respond([]byte(body)).Status(status)
		for k, v := range header {
			r.Header(k, v[0])
		}
		return r
	}
}
