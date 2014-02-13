package core

import (
	"github.com/karlseguin/gofake"
)

type Logger interface {
	// Log an informational message using the specified format
	Infof(format string, v ...interface{})

	// Log an informational message.
	Info(v ...interface{})

	// Log an error message using the specified format
	Errorf(format string, v ...interface{})

	// Log an error message
	Error(v ...interface{})

	// Whether info messages should be logged
	// Exposing this can allow implementers to take shortcuts and skip processing
	LogInfo() bool
}

type FakeLogger struct {
	gofake.Fake
}

func newFakeLogger() *FakeLogger {
	return &FakeLogger{gofake.New()}
}

func (f *FakeLogger) Infof(format string, v ...interface{}) {
	f.Called(format, v)
}

func (f *FakeLogger) Info(v ...interface{}) {
	f.Called(v)
}

func (f *FakeLogger) Errorf(format string, v ...interface{}) {
	f.Called(format, v)
}

func (f *FakeLogger) Error(v ...interface{}) {
	f.Called(v)
}

func (f *FakeLogger) LogInfo() bool {
	returns := f.Called()
	return returns.Bool(0, true)
}
