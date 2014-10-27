package gc

import (
	"github.com/karlseguin/dnscache"
	"net/http"
)

type Upstream struct {
	Name      string
	Transport *http.Transport
	Resolver  *dnscache.Resolver
}
