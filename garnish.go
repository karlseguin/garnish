package garnish

import (
	"net"
	"net/http"
	"os"
	"runtime"
)

// Start garnish with the given configuraiton
func Start(config *Configuration) {
	runtime.GOMAXPROCS(config.maxProcs)

	if len(config.middlewares) == 0 {
		config.Logger.Error(nil, "must configure at least 1 middleware")
		os.Exit(1)
	}

	InternalError = Respond([]byte(config.internalErrorMessage)).Status(500)
	NotFound = Respond([]byte(config.notFoundMessage)).Status(404)

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

	s := &http.Server{
		Handler:        newHandler(config),
		ReadTimeout:    config.readTimeout,
		MaxHeaderBytes: config.maxHeaderBytes,
	}
	config.Logger.Infof(nil, "listening on %v", config.address)
	config.Logger.Error(nil, s.Serve(l))
}
