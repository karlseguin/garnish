package garnish

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
	var upstream Upstream
	if len(transports) == 1 {
		upstream = &SingleTransportUpstream{
			headers:   headers,
			tweaker:   tweaker,
			transport: transports[0],
		}
	} else {
		upstream = &MultiTransportUpstream{
			headers:    headers,
			tweaker:    tweaker,
			transports: transports,
		}
	}
	return upstream, nil
}

type SingleTransportUpstream struct {
	transport *Transport
	headers   []string
	tweaker   RequestTweaker
}

func (u *SingleTransportUpstream) Headers() []string {
	return u.headers
}

func (u *SingleTransportUpstream) Tweaker() RequestTweaker {
	return u.tweaker
}

func (u *SingleTransportUpstream) Transport() *Transport {
	return u.transport
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
