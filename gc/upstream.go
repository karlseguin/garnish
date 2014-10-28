package gc

import (
	"github.com/karlseguin/dnscache"
	"net/http"
)

type RequestTweaker func(in *Request, out *http.Request)

type Upstream struct {
	Name      string
	Address   string
	Transport *http.Transport
	Resolver  *dnscache.Resolver
	Headers   []string
	Tweaker   RequestTweaker
}
