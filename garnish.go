package garnish

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middleware/stats"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
)

var InputPool = bytepool.New(32, 65536)

type Garnish struct {
	sync.RWMutex
	logger  gc.Logger
	handler *Handler
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

	gc.InternalError = Respond([]byte(config.internalErrorMessage)).Status(500).Header("Content-Type", config.defaultContentType)
	InternalError = gc.InternalError

	gc.NotFound = Respond([]byte(config.notFoundMessage)).Status(404).Header("Content-Type", config.defaultContentType)
	NotFound = gc.NotFound

	gc.Unauthorized = Respond([]byte(config.unauthorizedMessage)).Status(401).Header("Content-Type", config.defaultContentType)
	Unauthorized = gc.Unauthorized

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

	stats.RegisterReporter("inputPool", InputPool.Stats)
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

func (g *Garnish) Reload(config *Configuration) error {
	g.logger.Info("reloading")
	handler, err := newHandler(config)
	if err != nil {
		config.logger.Error(nil, err)
		return err
	}
	g.Lock()
	g.logger = config.logger
	g.handler = handler
	g.Unlock()
	g.logger.Info("reloaded")
	return nil
}

func (g *Garnish) Shutdown() {
	g.handler.shutdown = true
}
