package garnish

import (
	"github.com/karlseguin/garnish/core"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
)

type Garnish struct {
	sync.RWMutex
	logger  core.Logger
	handler http.Handler
}

func New() *Garnish {
	return new(Garnish)
}

// Start garnish with the given configuraiton
func (g *Garnish) Start(config *Configuration) bool {
	runtime.GOMAXPROCS(config.maxProcs)

	if len(config.middlewareFactories) == 0 {
		config.logger.Error("must configure at least 1 middleware")
		return false
	}

	InternalError = Respond([]byte(config.internalErrorMessage)).Status(500)
	NotFound = Respond([]byte(config.notFoundMessage)).Status(404)
	Unauthorized = Respond([]byte(config.unauthorizedMessage)).Status(401)

	var protocol = "tcp"
	address := config.address
	if address[0:4] == "unix" {
		address = address[6:]
		os.Remove(address)
		protocol = "unix"
	} else {
		address = address[6:]
	}

	l, err := net.Listen(protocol, address)
	if err != nil {
		config.logger.Error(err)
		return false
	}

	handler, err := newHandler(config)
	if err != nil {
		config.logger.Error(err)
		return false
	}
	if handler == nil {
		return false
	}

	g.handler = handler
	g.logger = config.logger
	s := &http.Server{
		Handler:        g,
		ReadTimeout:    config.readTimeout,
		MaxHeaderBytes: config.maxHeaderBytes,
	}
	config.logger.Infof("listening on %v", config.address)
	config.logger.Error(s.Serve(l))
	return false
}

func (g *Garnish) ServeHTTP(output http.ResponseWriter, req *http.Request) {
	g.RLock()
	defer g.RUnlock()
	g.handler.ServeHTTP(output, req)
}

func (g *Garnish) Reload(config *Configuration) {
	g.logger.Info(nil, "reloading")
	handler, err := newHandler(config)
	if err != nil {
		config.logger.Error(nil, err)
	} else {
		g.Lock()
		g.logger = config.logger
		g.handler = handler
		g.Unlock()
		g.logger.Info(nil, "reloaded")
	}
}
