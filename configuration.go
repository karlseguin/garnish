package garnish

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/garnish/router"
	"log"
	"os"
	"runtime"
	"time"
)

type Configuration struct {
	maxProcs             int
	address              string
	maxHeaderBytes       int
	readTimeout          time.Duration
	middlewareFactories  []core.MiddlewareFactory
	internalErrorMessage string
	notFoundMessage      string
	unauthorizedMessage  string
	router               core.Router
	Logger               core.Logger
}

func Configure() *Configuration {
	return &Configuration{
		maxHeaderBytes:       8192,
		internalErrorMessage: "internal error",
		notFoundMessage:      "not found",
		unauthorizedMessage:  "unauthorized",
		maxProcs:             runtime.NumCPU(),
		readTimeout:          time.Second * 10,
		address:              "tcp://127.0.0.1:6772",
		middlewareFactories:  make([]core.MiddlewareFactory, 0, 1),
		Logger:               &logger{logger: log.New(os.Stdout, "[garnish] ", log.Ldate|log.Lmicroseconds)},
	}
}

// The address to listen on should be in the format [tcp://127.0.0.1:6772]
// With unix socket: unix:/tmp/garnish.sock
func (c *Configuration) Listen(address string) *Configuration {
	c.address = address
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

// Registers the middlewares to use. Middleware will be executed in the order
// which they are supplied.
func (c *Configuration) Middleware(factories ...[]core.MiddlewareFactory) *Configuration {
	// c.middlewareFactories = append(c.middlewareFactories, factory)
	return c
}

// The body to use when replying with a 404 ["not found"]
func (c *Configuration) NotFound(message string) *Configuration {
	c.notFoundMessage = message
	return c
}

// The body to use when replying with a 401 ["unauthorized"]
func (c *Configuration) Unauthorized(message string) *Configuration {
	c.unauthorizedMessage = message
	return c
}

// The body to use when replying with a 500 ["internal error"]
func (c *Configuration) InternalError(message string) *Configuration {
	c.internalErrorMessage = message
	return c
}

// Creates and returns a new router
// As this breaks the chainable configuration, it'll normally be the last
// step in configuration.
func (c *Configuration) NewRouter() core.Router {
	c.router = router.New(c.Logger)
	return c.router
}
