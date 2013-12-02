package garnish

import (
	"log"
	"os"
	"runtime"
	"time"
)


type Configuration struct {
	maxProcs             int
	address              string
	maxHeaderBytes       int
	router               Router
	readTimeout          time.Duration
	middlewareFactories          []MiddlewareFactory
	internalErrorMessage string
	notFoundMessage      string

	Logger Logger
}

func Configure() *Configuration {
	return &Configuration{
		maxHeaderBytes:       8192,
		internalErrorMessage: "internal error",
		notFoundMessage:      "not found",
		maxProcs:             runtime.NumCPU(),
		readTimeout:          time.Second * 10,
		address:              "tcp://127.0.0.1:6772",
		middlewareFactories:          make([]MiddlewareFactory, 0, 1),
		Logger:               &logger{logger: log.New(os.Stdout, "[garnish] ", log.Ldate|log.Lmicroseconds)},
	}
}

// The address to listen on should be in the format [tcp://127.0.0.1:6772]
// With unix socket: unix:/tmp/garnish.sock
func (c *Configuration) Listen(address string) *Configuration {
	c.address = address
	return c
}

// The router which takes an http.Request and createa Route
func (c *Configuration) Router(router Router) *Configuration {
	c.router = router
	return c
}

// Maximum size of request headers, [8192]
func (c *Configuration) MaxHeaderBytes(bytes int) *Configuration {
	c.maxHeaderBytes = bytes
	return c
}

// Maximum duration before timing out read of the request [10 seconds]
func (c *Configuration) ReadTimeout(timeout time.Duration) *Configuration {
	c.readTimeout = timeout
	return c
}

// Maximum number of OS Threads to use (GOMAXPROCS) [# of CPUs]
func (c *Configuration) MaxiumOSThreads(count int) *Configuration {
	c.maxProcs = count
	return c
}

// Enable logging info messages
func (c *Configuration) LogInfo() *Configuration {
	c.Logger.(*logger).info = true
	return c
}

// Enable logging info messages
func (c *Configuration) Middleware(factory MiddlewareFactory) *Configuration {
	c.middlewareFactories = append(c.middlewareFactories, factory)
	return c
}

// The body to use when replying with a 404 ["not found"]
func (c *Configuration) NotFound(message string) *Configuration {
	c.notFoundMessage = message
	return c
}

// The body to use when replying with a 500 ["internal error"]
func (c *Configuration) internalError(message string) *Configuration {
	c.internalErrorMessage = message
	return c
}
