package core

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
