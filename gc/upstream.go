package gc

import (
	"github.com/karlseguin/dnscache"
	"net/http"
)

type Upstream struct {
	Name      string
	Address   string
	Transport *http.Transport
	Resolver  *dnscache.Resolver
}
