package garnish

import (
	"github.com/karlseguin/garnish/gc"
	"github.com/op/go-logging"
	"net/http"
	"os"
	"time"
)

var LogFormat = "%{level:.4s} | %{time:Jan _2 15:04:05.000} | %{message}"

func Start(configuration *Configuration) {
	logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))
	logging.SetFormatter(logging.MustStringFormatter(LogFormat))

	runtime := configuration.Build()
	if runtime == nil {
		return
	}
	s := http.Server{
		Addr:         configuration.address,
		Handler:      &Handler{runtime},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	gc.Logger.Info("listening on %s", configuration.address)
	gc.Logger.Panic(s.ListenAndServe())
}
