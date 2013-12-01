package upstream

import (
	"github.com/karlseguin/garnish"
	"time"
)

// Configuration for upstreams middleware
type Configuration struct {
	logger         garnish.Logger
	forwardHeaders []string
	dnsRefresh     time.Duration
	upstreams      map[string]*garnish.Upstream
}

func Configure(base *garnish.Configuration) *Configuration {
	return &Configuration{
		logger:         base.Logger,
		forwardHeaders: make([]string, 0, 1),
		dnsRefresh:     time.Minute,
		upstreams:      make(map[string]*garnish.Upstream),
	}
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
func (c *Configuration) Add(upstream *garnish.Upstream) *Configuration {
	c.upstreams[upstream.Name] = upstream
	return c
}
