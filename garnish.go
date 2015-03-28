package garnish

import (
	"net/http"
	"sync/atomic"
	"time"
)

var garnish *Garnish

type Garnish struct {
	*atomic.Value
}

func (g *Garnish) ServeHTTP(out http.ResponseWriter, request *http.Request) {
	g.Load().(http.Handler).ServeHTTP(out, request)
}

func Start(runtime *Runtime) {
	garnish = &Garnish{new(atomic.Value)}
	garnish.Store(runtime)

	s := http.Server{
		Handler:      garnish,
		Addr:         runtime.Address,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	Log.Infof("listening on %s", runtime.Address)
	panic(s.ListenAndServe())
}

func Reload(runtime *Runtime) {
	garnish.Load().(*Runtime).ReplaceWith(runtime)
	garnish.Store(runtime)
}
