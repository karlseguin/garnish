package garnish

import (
	"github.com/karlseguin/garnish/gc"
	"net/http"
	"time"
)

func Start(runtime *gc.Runtime) {
	s := http.Server{
		Addr:         runtime.Address,
		Handler:      runtime,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	gc.Log.Infof("listening on %s", runtime.Address)
	panic(s.ListenAndServe())
}
