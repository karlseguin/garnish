package garnish

import (
	"fmt"
	"log"
)

type Logger interface {
	// Log an informational message. If context is not nil, the
	// RequestId is appended to the message
	Infof(context Context, format string, v ...interface{})

	// Log an error message. If context is not nil, the
	// RequestId is appended to the message
	Errorf(context Context, format string, v ...interface{})

	// Log an error message. If context is not nil, the
	// RequestId is appended to the message
	Error(context Context, v ...interface{})
}

type logger struct {
	info   bool
	logger *log.Logger
}

func (l *logger) Infof(context Context, format string, v ...interface{}) {
	if l.info {
		l.printf(context, format, v...)
	}
}

func (l *logger) Errorf(context Context, format string, v ...interface{}) {
	l.printf(context, format, v...)
}

func (l *logger) Error(context Context, v ...interface{}) {
	id := "internal"
	if context != nil {
		id = context.RequestId()
	}
	l.logger.Println("["+id+"]", fmt.Sprint(v...))
}

func (l *logger) printf(context Context, format string, v ...interface{}) {
	id := "internal"
	if context != nil {
		id = context.RequestId()
	}
	l.logger.Println("["+id+"]", fmt.Sprintf(format, v...))
}
