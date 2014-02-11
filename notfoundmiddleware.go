package garnish

import (
	"github.com/karlseguin/garnish/core"
)

type NotFoundMiddleware struct{}

func (m *NotFoundMiddleware) Name() string {
	return "NotFound"
}

func (m *NotFoundMiddleware) Configure() error {
	return nil
}

func (m *NotFoundMiddleware) Run(context core.Context, next core.Next) core.Response {
	return NotFound
}
