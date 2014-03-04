package garnish

import (
	"github.com/karlseguin/garnish/gc"
)

type NotFoundMiddleware struct{}

func (m *NotFoundMiddleware) Name() string {
	return "NotFound"
}

func (m *NotFoundMiddleware) Configure() error {
	return nil
}

func (m *NotFoundMiddleware) Run(context gc.Context, next gc.Next) gc.Response {
	return NotFound
}
