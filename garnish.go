package garnish

import (
	"github.com/karlseguin/garnish/gc"
	"net/http"
	"time"
)

func Start(configuration *Configuration) {
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
