package gc

import (
	"fmt"
	"time"
)

type Logs interface {
	// Log an informational message using the specified format
	Info(format string, v ...interface{})

	// Log an error message using the specified format
	Warn(format string, v ...interface{})

	// Log an error message using the specified format
	Error(format string, v ...interface{})

	// Enable logging info messages
	Verbose()

	// Returns true if verbose logging is enabled
	IsVerbose() bool
}

// Garnish's global logger
// If you want to log something and have access to a *gc.Request,
// you should consider using its Info and Error methods for more
// context-aware messages
var Log Logs = new(Logger)

type Logger struct {
	info bool
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.info == false {
		return
	}
	l.log("i", format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.log("w", format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
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
