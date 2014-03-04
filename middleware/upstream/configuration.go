package upstream

import (
	"github.com/karlseguin/dnscache"
	"github.com/karlseguin/garnish/gc"
	"time"
)

// Configuration for upstreams middleware
type Configuration struct {
	overriding     string
	forwardHeaders []string
	dnsRefresh     time.Duration
	serverLookup   map[string]*Server
	routeLookup    map[string]*Server
}

func Configure() *Configuration {
	return &Configuration{
		forwardHeaders: make([]string, 0, 1),
		dnsRefresh:     time.Minute,
		serverLookup:   make(map[string]*Server),
		routeLookup:    make(map[string]*Server),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(config gc.Configuration) (gc.Middleware, error) {
	dns := dnscache.New(c.dnsRefresh)
	for _, server := range c.serverLookup {
		if server.resolver == nil {
			server.Resolver(dns.FetchOneString)
		}
		server.Finalize()
	}
	return &Upstream{c}, nil
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

// Defines an upstream server
func (c *Configuration) Add(name, scheme, address string) *Server {
	server := newServer(scheme, address)
	c.serverLookup[name] = server
	return server
}

func (c *Configuration) OverrideFor(route *gc.Route) {
	c.overriding = route.Name
}

func (c *Configuration) Is(name string) {
	c.routeLookup[c.overriding] = c.serverLookup[name]
}
