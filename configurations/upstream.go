package configurations

import (
	"github.com/karlseguin/dnscache"
	"github.com/karlseguin/garnish/gc"
	"net"
	"net/http"
	"strings"
	"time"
)

var DefaultHeaders = []string{"Content-Length"}

type Upstreams struct {
	upstreams map[string]*Upstream
}

func NewUpstreams() *Upstreams {
	return &Upstreams{
		upstreams: make(map[string]*Upstream),
	}
}

func (u *Upstreams) Add(name string) *Upstream {
	if _, exists := u.upstreams[name]; exists {
		gc.Log.Warn("Upstram %q already defined. Overwriting.", name)
	}
	one := &Upstream{
		name:        name,
		keepalive:   16,
		dnsDuration: time.Minute,
		headers:     DefaultHeaders,
	}
	u.upstreams[name] = one
	return one
}

func (u *Upstreams) Build(runtime *gc.Runtime) bool {
	ok := true
	upstreams := make(map[string]*gc.Upstream, len(u.upstreams))
	for name, one := range u.upstreams {
		if upstream := one.Build(); upstream != nil {
			upstreams[name] = upstream
		} else {
			ok = false
		}
	}
	runtime.Upstreams = upstreams
	return ok
}

type Upstream struct {
	name        string
	address     string
	keepalive   int
	dnsDuration time.Duration
	headers     []string
	tweaker     gc.RequestTweaker
}

// the address to connect to. Should begin with unix:/  http://  or https://
// [""]
func (u *Upstream) Address(address string) *Upstream {
	u.address = address
	return u
}

// the number of connections to keep alive. Set to 0 to disable
// [16]
func (u *Upstream) KeepAlive(count int) *Upstream {
	u.keepalive = count
	return u
}

// the duration to cache the upstream's dns lookup. Set to 0 to prevent
// garnish from caching this value (even a few seconds can help)
// [1 minute]
func (u *Upstream) DnsCache(duration time.Duration) *Upstream {
	u.dnsDuration = duration
	return u
}

// The headers to copy from the incoming request to the outgoing request
// [Content-Length]
func (u *Upstream) Headers(headers ...string) *Upstream {
	u.headers = headers
	return u
}

// Custom callback to modify the request (out) that will get sent to the upstream
func (u *Upstream) Tweaker(tweaker gc.RequestTweaker) *Upstream {
	u.tweaker = tweaker
	return u
}

func (u *Upstream) Build() *gc.Upstream {
	if len(u.address) < 8 {
		gc.Log.Error("Upstream %s has an invalid address: %q", u.name, u.address)
		return nil
	}
	if u.address[:6] != "unix:/" && u.address[:7] != "http://" && u.address[:8] != "https://" {
		gc.Log.Error("Upstream %s's address should begin with unix:/, http:// or https://", u.name)
		return nil
	}
	upstream := &gc.Upstream{
		Name:    u.name,
		Address: u.address,
		Headers: u.headers,
		Tweaker: u.tweaker,
	}
	if u.dnsDuration > 0 {
		upstream.Resolver = dnscache.New(u.dnsDuration)
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost: u.keepalive,
		DisableKeepAlives:   u.keepalive == 0,
	}
	upstream.Transport = transport
	if u.address[:6] == "unix:/" {
		transport.Dial = func(network, address string) (net.Conn, error) {
			//strip out the :80 which Go adds
			return net.Dial("unix", address[:len(address)-3])
		}
	} else if u.dnsDuration == 0 {
		transport.Dial = func(network, address string) (net.Conn, error) {
			return net.Dial(network, address)
		}
	} else {
		transport.Dial = func(network, address string) (net.Conn, error) {
			separator := strings.LastIndex(address, ":")
			ip, _ := upstream.Resolver.FetchOneString(address[:separator])
			return net.Dial(network, ip+address[separator:])
		}
	}
	return upstream
}
