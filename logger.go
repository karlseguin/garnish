package garnish

import (
	"fmt"
	"time"
)

type Logs interface {
	// Log an information message
	Info(message string)
	// Log an informational message using the specified format
	Infof(format string, v ...interface{})

	// Log a warning message
	Warn(message string)
	// Log a warning message using the specified format
	Warnf(format string, v ...interface{})

	// Log an error message
	Error(message string)
	// Log an error message using the specified format
	Errorf(format string, v ...interface{})

	// Enable logging info messages
	Verbose()

	// Returns true if verbose logging is enabled
	IsVerbose() bool
}

// Garnish's global logger
// If you want to log something and have access to a *garnish.Request,
// you should consider using its Info and Error methods for more
// context-aware messages
var Log Logs = new(Logger)

type Logger struct {
	info bool
}

func (l *Logger) Info(message string) {
	if l.info == false {
		return
	}
	l.log("i", message)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.info == false {
		return
	}
	l.log("i", format, v...)
}

func (l *Logger) Warn(message string) {
	l.log("w", message)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.log("w", format, v...)
}

func (l *Logger) Error(message string) {
	l.log("e", message)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.log("e", format, v...)
}

func (l *Logger) log(level, format string, v ...interface{}) {
	fmt.Printf("%s %s ", level, time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Println(fmt.Sprintf(format, v...))
}

func (l *Logger) Verbose() {
	l.info = true
}

func (l *Logger) IsVerbose() bool {
	return l.info
}

type FakeLogger struct {
	Infos  []string
	Warns  []string
	Errors []string
}

func NewFakeLogger() *FakeLogger {
	return &FakeLogger{}
}

func (f *FakeLogger) Info(message string) {
	f.Infos = append(f.Infos, message)
}

func (f *FakeLogger) Infof(format string, v ...interface{}) {
	f.Infos = append(f.Infos, fmt.Sprintf(format, v...))
}

func (f *FakeLogger) Warn(message string) {
	f.Warns = append(f.Warns, message)
}

func (f *FakeLogger) Warnf(format string, v ...interface{}) {
	f.Warns = append(f.Warns, fmt.Sprintf(format, v...))
}

func (f *FakeLogger) Error(message string) {
	f.Errors = append(f.Errors, message)
}

func (f *FakeLogger) Errorf(format string, v ...interface{}) {
	f.Errors = append(f.Errors, fmt.Sprintf(format, v...))
}

func (f *FakeLogger) Verbose() {}

func (f *FakeLogger) IsVerbose() bool {
	return false
}
