package upstream

import (
	"github.com/karlseguin/dnscache"
	"github.com/karlseguin/garnish"
	"time"
)

// Configuration for upstreams middleware
type Configuration struct {
	logger         garnish.Logger
	forwardHeaders []string
	dnsRefresh     time.Duration
	upstreams      map[string]*Server
}

func Configure(base *garnish.Configuration) *Configuration {
	return &Configuration{
		logger:         base.Logger,
		forwardHeaders: make([]string, 0, 1),
		dnsRefresh:     time.Minute,
		upstreams:      make(map[string]*Server),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create() (garnish.Middleware, error) {
	upstream := &Upstream{c}
	dns := dnscache.New(c.dnsRefresh)
	for _, upstream := range c.upstreams {
		upstream.Resolver(dns.FetchOneString)
		upstream.Finalize()
	}
	return upstream, nil
}

// Forward the specified headers from the input request to the
// upstream request
func (c *Configuration) ForwardHeaders(headerNames ...string) *Configuration {
	c.forwardHeaders = append(c.forwardHeaders, headerNames...)
	return c
}

// Set the frequency to update DNS [1 minute]
func (c *Configuration) DnsRefresh(frequency time.Duration) *Configuration {
	c.dnsRefresh = frequency
	return c
}

// Adds an upstream. The upstream is picked based on the Route's
// upstream, matching by name
func (c *Configuration) Add(upstream *Server) *Configuration {
	c.upstreams[upstream.Name] = upstream
	return c
}
