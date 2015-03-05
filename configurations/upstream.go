package configurations

import (
	"fmt"
	"github.com/karlseguin/garnish/gc"
	"net"
	"net/http"
	"strings"
	"time"
)

// The default headers to forward to the upstream
var DefaultHeaders = []string{"Content-Length"}

type Upstreams struct {
	upstreams map[string]*Upstream
}

func NewUpstreams() *Upstreams {
	return &Upstreams{
		upstreams: make(map[string]*Upstream),
	}
}

// Used internally
func (u *Upstreams) Add(name string) *Upstream {
	if _, exists := u.upstreams[name]; exists {
		gc.Log.Warnf("Upstream %q already defined. Overwriting.", name)
	}
	one := &Upstream{
		name:        name,
		dnsDuration: time.Minute,
		headers:     DefaultHeaders,
		transports:  make([]*Transport, 0, 2),
	}
	u.upstreams[name] = one
	return one
}

func (u *Upstreams) Build(runtime *gc.Runtime) error {
	if u == nil {
		runtime.Upstreams = make(map[string]gc.Upstream, 0)
		return nil
	}

	upstreams := make(map[string]gc.Upstream, len(u.upstreams))
	for name, one := range u.upstreams {
		upstream, err := one.Build(runtime)
		if err != nil {
			return err
		}
		upstreams[name] = upstream
	}
	runtime.Upstreams = upstreams
	return nil
}

type Upstream struct {
	name        string
	transports  []*Transport
	dnsDuration time.Duration
	headers     []string
	tweaker     gc.RequestTweaker
}

type Transport struct {
	address   string
	keepalive int
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

// the address to connect to. Should begin with unix:/  http://  or https://
// [""]
func (u *Upstream) Address(address string) *Transport {
	transport := &Transport{
		address:   address,
		keepalive: 16,
	}
	u.transports = append(u.transports, transport)
	return transport
}

// the number of connections to keep alive. Set to 0 to disable
// [16]
func (t *Transport) KeepAlive(count uint32) *Transport {
	t.keepalive = int(count)
	return t
}

func (u *Upstream) Build(runtime *gc.Runtime) (gc.Upstream, error) {
	l := len(u.transports)
	if l == 0 {
		return nil, fmt.Errorf("Upstream %s doesn't have a configured transport", u.name)
	}

	transports := make([]*gc.Transport, l)
	for i := 0; i < l; i++ {
		t := u.transports[i]
		if len(t.address) < 8 {
			return nil, fmt.Errorf("Upstream %s has an invalid address: %q", u.name, t.address)
		}
		var domain string
		if t.address[:7] == "http://" {
			domain = t.address[7:]
		} else if t.address[:8] == "https://" {
			domain = t.address[8:]
		}
		if t.address[:6] != "unix:/" && len(domain) == 0 {
			return nil, fmt.Errorf("Upstream %s's address should begin with unix:/, http:// or https://", u.name)
		}

		transport := &http.Transport{
			MaxIdleConnsPerHost: t.keepalive,
			DisableKeepAlives:   t.keepalive == 0,
		}
		if t.address[:6] == "unix:/" {
			transport.Dial = func(network, address string) (net.Conn, error) {
				//strip out the :80 which Go adds
				return net.Dial("unix", address[:len(address)-3])
			}
		} else if strings.Contains(t.address, "localhost") {
			transport.Dial = func(network, address string) (net.Conn, error) {
				return net.Dial(network, address)
			}
		} else {
			transport.Dial = func(network, address string) (net.Conn, error) {
				separator := strings.LastIndex(address, ":")
				ip, _ := runtime.Resolver.FetchOneString(address[:separator])
				return net.Dial(network, ip+address[separator:])
			}
		}

		if u.dnsDuration > 0 && len(domain) > 0 {
			runtime.Resolver.TTL(domain, u.dnsDuration)
		}

		transports[i] = &gc.Transport{
			Transport: transport,
			Address:   t.address,
		}
	}

	return gc.CreateUpstream(u.headers, u.tweaker, transports)
}
