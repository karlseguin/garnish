package garnish

import (
	"fmt"
	"github.com/karlseguin/garnish/core"
	"log"
)

type logger struct {
	info   bool
	logger *log.Logger
}

func (l *logger) Infof(context core.Context, format string, v ...interface{}) {
	if l.info {
		l.printf(context, format, v...)
	}
}

func (l *logger) Info(context core.Context, v ...interface{}) {
	if l.info {
		l.print(context, v...)
	}
}

func (l *logger) Errorf(context core.Context, format string, v ...interface{}) {
	l.printf(context, format, v...)
}

func (l *logger) Error(context core.Context, v ...interface{}) {
	l.print(context, v...)
}

func (l *logger) printf(context core.Context, format string, v ...interface{}) {
	id := "internal"
	location := "internal"
	if context != nil {
		id = context.RequestId()
		location = context.Location()
	}
	l.logger.Println("["+id+"] ["+location+"]", fmt.Sprintf(format, v...))
}

func (l *logger) print(context core.Context, v ...interface{}) {
	id := "internal"
	location := "internal"
	if context != nil {
		id = context.RequestId()
		location = context.Location()
	}
	l.logger.Println("["+id+"] ["+location+"]", fmt.Sprint(v...))
}
