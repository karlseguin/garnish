package stats

import (
	"github.com/karlseguin/garnish/core"
	"strconv"
	"time"
)

var worker *Worker

type Persister interface {
	Persist(map[string]Snapshot) error
}

// Configuration for the Stats middleware
type Configuration struct {
	overriding  *Stat
	window      time.Duration
	treshhold   time.Duration
	sampleSize  int64
	sampleSizeF float64
	routeStats  map[string]*Stat
	persister   Persister
	percentiles map[string]float64
}

func Configure() *Configuration {
	return &Configuration{
		window:      time.Second * 5,
		sampleSize:  50,
		sampleSizeF: float64(50),
		treshhold:   time.Millisecond * 500,
		routeStats:  make(map[string]*Stat),
		percentiles: map[string]float64{"50p": 0.5, "75p": 0.75, "95p": 0.95},
		persister:   &FilePersister{"./stats.json"},
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(config core.Configuration) (core.Middleware, error) {
	for name, _ := range config.Router().Routes() {
		if _, ok := c.routeStats[name]; ok == false {
			c.routeStats[name] = newStat(c)
		}
	}
	c.routeStats["?"] = newStat(c)

	if worker != nil {
		worker.stop()
	}

	worker = &Worker{
		logger:     config.Logger(),
		window:     c.window,
		persister:  c.persister,
		routeStats: c.routeStats,
	}
	go worker.start()
	return &Stats{
		routeStats: c.routeStats,
	}, nil

}

// Reservoir sampling is used to report percentiles without having unbound
// growth. The sample size specifies how large a reservoir to use per route.
// A higher value will result in a more accurate report but requires more
// memory. The accuracy suffers from diminishing return, so there's no point
// setting this too high.

// Can be set globally or on a per-route basis

// [500]
func (c *Configuration) SampleSize(size int) *Configuration {
	if c.overriding != nil {
		c.overriding.sampleSize = int64(size)
		c.overriding.sampleSizeF = float64(size)
	} else {
		c.sampleSize = int64(size)
		c.sampleSizeF = float64(size)
	}
	return c
}

// The period of time to group statistics in.

// Can be set globally

// [1 minute]
func (c *Configuration) Window(window time.Duration) *Configuration {
	if c.overriding != nil {
		panic("stats window can only be configured globally")
	}
	c.window = window
	return c
}

// The persister responsible for saving the statistics

// Can be set globally

// [FilePersister("./stats.json")]
func (c *Configuration) Persister(persister Persister) *Configuration {
	if c.overriding != nil {
		panic("stats persister can only be configured globally")
	}
	c.persister = persister
	return c
}

// The percentiles to track

// Can be set globally or on a per-route basis

// [50, 75, 95]
func (c *Configuration) Percentiles(percentiles ...int) *Configuration {
	lookup := make(map[string]float64)
	for _, p := range percentiles {
		lookup[strconv.Itoa(p)+"p"] = float64(p) / 100
	}
	if c.overriding != nil {
		c.overriding.percentiles = lookup
	} else {
		c.percentiles = lookup
	}
	return c
}

// Requests slower than the specified thresshold will be counted as "slow"

// Can be set globally or on a per-route basis

// [500ms]
func (c *Configuration) Treshhold(t time.Duration) *Configuration {
	if c.overriding != nil {
		c.overriding.treshhold = t
	} else {
		c.treshhold = t
	}
	return c
}

func (c *Configuration) OverrideFor(route *core.Route) {
	stat := newStat(c)
	c.routeStats[route.Name] = stat
	c.overriding = stat
}
