package garnish

import (
	"github.com/karlseguin/gspec"
	"testing"
)

func TestNotFoundMiddlewareReturnsNotFoundResponse(t *testing.T) {
	spec := gspec.New(t)
	res := new(notFoundMiddleware).Run(nil, nil)
	spec.Expect(res.GetStatus()).ToEqual(404)
	spec.Expect(string(res.GetBody())).ToEqual("not found")
}

func TestMiddlewareWrapperLogsExecution(t *testing.T) {
	spec := gspec.New(t)
	logger, buffer := testLogger(true)
	mw := &MiddlewareWrapper{logger: logger, middleware: new(notFoundMiddleware)}
	mw.Yield(nil)
	spec.Expect(buffer.String()).ToEqual("[internal] + _notFound\n[internal] - _notFound\n")
}
