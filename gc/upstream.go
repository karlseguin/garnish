package gc

import (
	"math/rand"
	"net/http"
	"sync"
)

// The User Agent to send to the upstream
var DefaultUserAgent = []string{""}

// Tweaks request `out` before sending it to the upstream
type RequestTweaker func(in *Request, out *http.Request)

type Upstream interface {
	Headers() []string
	Transport() *Transport
	Tweaker() RequestTweaker
}

func CreateUpstream(headers []string, tweaker RequestTweaker, transports []*Transport) (Upstream, error) {
	return &MultiTransportUpstream{
		headers:    headers,
		tweaker:    tweaker,
		transports: transports,
	}, nil
}

type MultiTransportUpstream struct {
	sync.RWMutex
	transports []*Transport
	headers    []string
	tweaker    RequestTweaker
}

func (u *MultiTransportUpstream) Headers() []string {
	return u.headers
}

func (u *MultiTransportUpstream) Tweaker() RequestTweaker {
	return u.tweaker
}

func (u *MultiTransportUpstream) Transport() *Transport {
	defer u.RUnlock()
	u.RLock()
	return u.transports[rand.Intn(len(u.transports))]
}

type Transport struct {
	*http.Transport
	Address string
}
