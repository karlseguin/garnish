package garnish

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"net/http/httptest"
	"testing"
)

type HandlerTests struct{}

func Test_Handler(t *testing.T) {
	Expectify(new(HandlerTests), t)
}

func (h *HandlerTests) NotFoundForUnknownRoute() {
	handler := testHandler()
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, build.Request().Path("/fail").Request)
	Expect(out.Code).To.Equal(404)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("0")
}

func testHandler() *Handler {
	config := Configure()
	config.Upstream("test").Address("http://127.0.0.1:4005").KeepAlive(2)
	config.Route("plain").Get("/plain").Upstream("test")
	runtime := config.Build()
	return &Handler{runtime}
}
