package gc

import (
	"net/http"
)

// The User Agent to send to the upstream
var DefaultUserAgent = []string{""}

// Tweaks request `out` before sending it to the upstream
type RequestTweaker func(in *Request, out *http.Request)

type Upstream struct {
	Name       string
	Transports []*Transport
	Headers    []string
	Tweaker    RequestTweaker
}

func (u *Upstream) Transport() *Transport {
	return u.Transports[0]
}

type Transport struct {
	weight   int
	fallback bool
	*http.Transport
	Address string
}
