package garnish

import (
	"github.com/karlseguin/gspec"
	"testing"
)

func TestNotFoundMiddlewareReturnsNotFoundResponse(t *testing.T) {
	spec := gspec.New(t)
	res := notFoundMiddleware(nil, nil)
	spec.Expect(res.GetStatus()).ToEqual(404)
	spec.Expect(string(res.GetBody())).ToEqual("not found")
}

func TestMiddlewareWrapperLogsExecution(t *testing.T) {
	spec := gspec.New(t)
	logger, buffer := testLogger(true)
	mw := &MiddlewareWrapper{name: "test-wrap", logger: logger, middleware: notFoundMiddleware}
	mw.Yield(nil)
	spec.Expect(buffer.String()).ToEqual("[internal] +[test-wrap]\n[internal] -[test-wrap]\n")
}
