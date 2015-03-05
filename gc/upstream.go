package gc

import (
	"math/rand"
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
	index := rand.Intn(len(u.Transports))
	return u.Transports[index]
}

type Transport struct {
	*http.Transport
	Address string
}
