package core

import (
	"github.com/karlseguin/gofake"
)

type Logger interface {
	// Log an informational message. If context is not nil, the
	// RequestId is appended to the message
	Infof(context Context, format string, v ...interface{})

	// Log an informational message. If context is not nil, the
	// RequestId is appended to the message
	Info(context Context, v ...interface{})

	// Log an error message. If context is not nil, the
	// RequestId is appended to the message
	Errorf(context Context, format string, v ...interface{})

	// Log an error message. If context is not nil, the
	// RequestId is appended to the message
	Error(context Context, v ...interface{})
}

type FakeLogger struct {
	gofake.Fake
}

func newFakeLogger() *FakeLogger {
	return &FakeLogger{gofake.New()}
}

func (f *FakeLogger) Infof(context Context, format string, v ...interface{}) {
	f.Called(context, format, v)
}

func (f *FakeLogger) Info(context Context, v ...interface{}) {
	f.Called(context, v)
}

func (f *FakeLogger) Errorf(context Context, format string, v ...interface{}) {
	f.Called(context, format, v)
}

func (f *FakeLogger) Error(context Context, v ...interface{}) {
	f.Called(context, v)
}
