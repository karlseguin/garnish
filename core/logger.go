package core

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
