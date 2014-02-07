package garnish

import (
	"github.com/karlseguin/bytepool"
	"net"
	"net/http"
	"strings"
)

// DNS lookup handler
type DnsResolver func(address string) (string, error)

// Configuration for an upstream service
type Upstream struct {
	Name                      string
	Host                      string
	Scheme                    string
	Pool                      *bytepool.Pool
	PoolItemSize              int
	disableKeepAlives         bool
	maxIdleConnectionsPerHost int
	resolver                  DnsResolver
	Transport                 *http.Transport
}

// Create an upstream service. scheme can either be http or https or
// If host starts with unix:/, a unix socket is used
func NewUpstream(name, scheme, host string, routes ...string) *Upstream {
	return &Upstream{
		Name:                      name,
		Host:                      host,
		Scheme:                    scheme,
		PoolItemSize:              65536,
		Pool:                      bytepool.New(125, 65536),
		disableKeepAlives:         false,
		maxIdleConnectionsPerHost: 32,
	}
}

// Responses which fit within the specified size (as per the response
// Content-Length) are stored in a pre-allocated []byte pool [1024, 65536]
func (u *Upstream) ResponsePool(count, size int) *Upstream {
	u.PoolItemSize = size
	u.Pool = bytepool.New(count, size)
	return u
}

// Disable keepalive to upstream [false]
func (u *Upstream) DisableKeepAlives() *Upstream {
	u.disableKeepAlives = true
	return u
}

// Maximum number of keep alive connections to keep [32]
func (u *Upstream) MaxIdleConnectionsPerHost(count int) *Upstream {
	u.maxIdleConnectionsPerHost = count
	return u
}

func (u *Upstream) Resolver(resolver DnsResolver) *Upstream {
	u.resolver = resolver
	return u
}

func (u *Upstream) Finalize() {
	u.Transport = &http.Transport{
		MaxIdleConnsPerHost: u.maxIdleConnectionsPerHost,
		DisableKeepAlives:   u.disableKeepAlives,
	}
	if u.Host[0:6] == "unix:/" {
		u.Host = u.Host[5:]
		u.Transport.Dial = func(network, address string) (net.Conn, error) {
			return net.Dial("unix", address[:len(address)-3])
		}
	} else {
		u.Transport.Dial = func(network, address string) (net.Conn, error) {
			separator := strings.LastIndex(address, ":")
			ip, _ := u.resolver(address[:separator])
			return net.Dial("tcp", ip+address[separator:])
		}
	}
}
