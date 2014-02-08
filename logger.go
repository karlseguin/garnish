package garnish

import (
	"fmt"
	"log"
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

type logger struct {
	info   bool
	logger *log.Logger
}

func (l *logger) Infof(context Context, format string, v ...interface{}) {
	if l.info {
		l.printf(context, format, v...)
	}
}

func (l *logger) Info(context Context, v ...interface{}) {
	if l.info {
		l.print(context, v...)
	}
}

func (l *logger) Errorf(context Context, format string, v ...interface{}) {
	l.printf(context, format, v...)
}

func (l *logger) Error(context Context, v ...interface{}) {
	l.print(context, v...)
}

func (l *logger) printf(context Context, format string, v ...interface{}) {
	id := "internal"
	location := "internal"
	if context != nil {
		id = context.RequestId()
		location = context.Location()
	}
	l.logger.Println("["+id+"] ["+location+"]", fmt.Sprintf(format, v...))
}

func (l *logger) print(context Context, v ...interface{}) {
	id := "internal"
	location := "internal"
	if context != nil {
		id = context.RequestId()
		location = context.Location()
	}
	l.logger.Println("["+id+"] ["+location+"]", fmt.Sprint(v...))
}
