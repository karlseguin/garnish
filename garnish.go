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
func (g *Garnish) Start(config *Configuration) {
	runtime.GOMAXPROCS(config.maxProcs)

	if len(config.middlewareFactories) == 0 {
		config.Logger.Error(nil, "must configure at least 1 middleware")
		os.Exit(1)
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
		config.Logger.Error(nil, err)
		os.Exit(1)
	}

	handler, err := newHandler(config)
	if err != nil {
		config.Logger.Error(nil, err)
		os.Exit(1)
	}
	g.handler = handler
	g.logger = config.Logger
	s := &http.Server{
		Handler:        g,
		ReadTimeout:    config.readTimeout,
		MaxHeaderBytes: config.maxHeaderBytes,
	}
	config.Logger.Infof(nil, "listening on %v", config.address)
	config.Logger.Error(nil, s.Serve(l))
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
		config.Logger.Error(nil, err)
	} else {
		g.Lock()
		g.logger = config.Logger
		g.handler = handler
		g.Unlock()
		g.logger.Info(nil, "reloaded")
	}
}
