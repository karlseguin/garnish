package upstream

import (
	"github.com/karlseguin/bytepool"
	"net"
	"net/http"
	"strings"
)

// DNS lookup handler
type DnsResolver func(address string) (string, error)

// Configuration for an Server service
type Server struct {
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

// Create am upstream Server. scheme can either be http or https or
// If host starts with unix:/, a unix socket is used
func NewServer(name, scheme, host string, routes ...string) *Server {
	return &Server{
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
func (u *Server) ResponsePool(count, size int) *Server {
	u.PoolItemSize = size
	u.Pool = bytepool.New(count, size)
	return u
}

// Disable keepalive to Server [false]
func (u *Server) DisableKeepAlives() *Server {
	u.disableKeepAlives = true
	return u
}

// Maximum number of keep alive connections to keep [32]
func (u *Server) MaxIdleConnectionsPerHost(count int) *Server {
	u.maxIdleConnectionsPerHost = count
	return u
}

func (u *Server) Resolver(resolver DnsResolver) *Server {
	u.resolver = resolver
	return u
}

func (u *Server) Finalize() {
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
