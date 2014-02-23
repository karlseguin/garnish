package garnish

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/karlseguin/garnish/caches/ccache"
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/garnish/middleware/caching"
	"github.com/karlseguin/garnish/middleware/stats"
	"github.com/karlseguin/garnish/middleware/upstream"
	"github.com/karlseguin/garnish/router"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type Configuration struct {
	maxProcs             int
	address              string
	maxHeaderBytes       int
	readTimeout          time.Duration
	middlewareFactories  []core.MiddlewareFactory
	internalErrorMessage string
	notFoundMessage      string
	unauthorizedMessage  string
	router               core.Router
	logger               core.Logger
}

func Configure() *Configuration {
	return &Configuration{
		maxHeaderBytes:       8192,
		internalErrorMessage: "internal error",
		notFoundMessage:      "not found",
		unauthorizedMessage:  "unauthorized",
		maxProcs:             runtime.NumCPU(),
		readTimeout:          time.Second * 10,
		address:              "tcp://127.0.0.1:6772",
		middlewareFactories:  make([]core.MiddlewareFactory, 0, 1),
		logger:               &logger{logger: log.New(os.Stdout, "[garnish] ", log.Ldate|log.Lmicroseconds)},
	}
}

// The address to listen on should be in the format [tcp://127.0.0.1:6772]
// With unix socket: unix:/tmp/garnish.sock
func (c *Configuration) Listen(address string) *Configuration {
	c.address = address
	return c
}

// Maximum size of request headers, [8192]
func (c *Configuration) MaxHeaderBytes(bytes int) *Configuration {
	c.maxHeaderBytes = bytes
	return c
}

// Maximum duration before timing out read of the request [10 seconds]
func (c *Configuration) ReadTimeout(timeout time.Duration) *Configuration {
	c.readTimeout = timeout
	return c
}

// Maximum number of OS Threads to use (GOMAXPROCS) [# of CPUs]
func (c *Configuration) MaxiumOSThreads(count int) *Configuration {
	c.maxProcs = count
	return c
}

// Enable logging info messages
func (c *Configuration) LogInfo() *Configuration {
	c.logger.(*logger).info = true
	return c
}

// Registers the middlewares to use. Middleware will be executed in the order
// which they are supplied.
func (c *Configuration) Middleware(factories ...core.MiddlewareFactory) *Configuration {
	for _, factory := range factories {
		c.middlewareFactories = append(c.middlewareFactories, factory)
	}
	return c
}

// The body to use when replying with a 404 ["not found"]
func (c *Configuration) NotFound(message string) *Configuration {
	c.notFoundMessage = message
	return c
}

// The body to use when replying with a 401 ["unauthorized"]
func (c *Configuration) Unauthorized(message string) *Configuration {
	c.unauthorizedMessage = message
	return c
}

// The body to use when replying with a 500 ["internal error"]
func (c *Configuration) InternalError(message string) *Configuration {
	c.internalErrorMessage = message
	return c
}

// Creates and returns a new router
// As this breaks the chainable configuration, it'll normally be the last
// step in configuration.
func (c *Configuration) NewRouter() core.Router {
	c.router = router.New(c.logger, c.middlewareFactories)
	return c.router
}

// Gets the logger
func (c *Configuration) Logger() core.Logger {
	return c.logger
}

// Gets the router
func (c *Configuration) Router() core.Router {
	return c.router
}

func ConfigureFromJson(bytes []byte) (config *Configuration, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintf("%v", e))
		}
	}()
	var raw map[string]interface{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, err
	}
	config = Configure()
	mapCoreConfig(config, raw)
	configurators := make(map[string]core.MiddlewareFactory)
	if configData := raw["defaults"].(map[string]interface{}); configData != nil {
		mapMiddlewareConfig(config, configData, configurators)
	}
	router := config.NewRouter()
	if configData := raw["routes"].([]interface{}); configData != nil {
		for _, routeConfigData := range configData {
			rcd := routeConfigData.(map[string]interface{})
			r := router.Add(rcd["name"].(string), rcd["method"].(string), rcd["path"].(string))
			for _, middleware := range config.middlewareFactories {
				middleware.OverrideFor(r.Route())
			}
			mapMiddlewareConfig(config, rcd, configurators)
		}
	}
	return config, nil
}

func ConfigureFromFile(path string) (*Configuration, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ConfigureFromJson(bytes)
}

var configuatorFactories = map[string]func() core.MiddlewareFactory{
	"stats":    func() core.MiddlewareFactory { return stats.Configure() },
	"caching":  func() core.MiddlewareFactory { return caching.Configure() },
	"upstream": func() core.MiddlewareFactory { return upstream.Configure() },
}

func mapMiddlewareConfig(config *Configuration, data map[string]interface{}, configurators map[string]core.MiddlewareFactory) {
	for name, raw := range data {
		name = strings.ToLower(name)
		factory, ok := configuatorFactories[name]
		if ok == false {
			continue
		}

		configurator := configurators[name]
		if configurator == nil {
			configurator = factory()
			config.Middleware(configurator)
			configurators[name] = configurator
		}

		configData := raw.(map[string]interface{})
		switch name {
		case "stats":
			mapStatsConfig(configurator.(*stats.Configuration), configData)
		case "caching":
			mapCachingConfig(configurator.(*caching.Configuration), configData)
		case "upstream":
			mapUpstreamConfig(configurator.(*upstream.Configuration), configData)
		}
	}
}

func mapCoreConfig(config *Configuration, configData map[string]interface{}) {
	for key, value := range configData {
		switch strings.ToLower(key) {
		case "loginfo":
			config.LogInfo()
		case "listen":
			config.Listen(value.(string))
		case "threads":
			config.MaxiumOSThreads(toInt(value))
		case "readtimeout":
			config.ReadTimeout(toDuration(value))
		case "maxheaderbytes":
			config.MaxHeaderBytes(toInt(value))
		}
	}
}

func mapStatsConfig(config *stats.Configuration, configData map[string]interface{}) {
	for key, value := range configData {
		switch strings.ToLower(key) {
		case "samplesize":
			config.SampleSize(toInt(value))
		case "window":
			config.Window(toDuration(value))
		case "output":
			config.Persister(&stats.FilePersister{value.(string)})
		case "treshhold":
			config.Treshhold(toDuration(value))
		case "percentiles":
			floats := value.([]interface{})
			ints := make([]int, len(floats))
			for index, flt := range floats {
				ints[index] = int(flt.(float64))
			}
			config.Percentiles(ints...)
		}
	}
}

func mapCachingConfig(config *caching.Configuration, configData map[string]interface{}) {
	if f, ok := configData["size"].(float64); ok {
		config.Cache(ccache.New(ccache.Configure().Size(uint64(f))))
	}
	for key, value := range configData {
		switch key {
		case "ttl":
			config.TTL(toDuration(value))
		case "grace":
			config.Grace(toDuration(value))
		case "saint":
			config.Saint(toDuration(value))
		}
	}
}

func mapUpstreamConfig(config *upstream.Configuration, configData map[string]interface{}) {
	for key, value := range configData {
		switch strings.ToLower(key) {
		case "forwardheaders":
			config.ForwardHeaders(value.([]string)...)
		case "dnsrefresh":
			config.DnsRefresh(toDuration(value))
		case "servers":
			for _, raw := range value.([]interface{}) {
				serverConfig := raw.(map[string]interface{})
				server := config.Add(serverConfig["name"].(string), serverConfig["scheme"].(string), serverConfig["address"].(string))
				mapServerConfig(server, serverConfig)
			}
		case "upstream":
			config.Is(value.(string))
		}
	}
}

func mapServerConfig(server *upstream.Server, configData map[string]interface{}) {
	for key, value := range configData {
		switch strings.ToLower(key) {
		case "keepalive":
			if value.(bool) == false {
				server.DisableKeepAlives()
			}
		case "maxidle":
			server.MaxIdleConnectionsPerHost(toInt(value))
		}
	}
}

func toInt(value interface{}) int {
	return int(value.(float64))
}

func toDuration(value interface{}) time.Duration {
	return time.Duration(int(value.(float64)))
}
