package garnish

import (
	"fmt"
	"log"
)

type logger struct {
	info   bool
	logger *log.Logger
}

func (l *logger) Infof(format string, v ...interface{}) {
	if l.info {
		l.printf(format, v...)
	}
}

func (l *logger) Info(v ...interface{}) {
	if l.info {
		l.print(v...)
	}
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.printf(format, v...)
}

func (l *logger) Error(v ...interface{}) {
	l.print(v...)
}

func (l *logger) LogInfo() bool {
	return l.info
}

func (l *logger) printf(format string, v ...interface{}) {
	l.logger.Println(fmt.Sprintf(format, v...))
}

func (l *logger) print(v ...interface{}) {
	l.logger.Println(v...)
}
