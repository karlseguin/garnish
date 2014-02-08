package stats

import (
	"github.com/karlseguin/garnish"
	"strconv"
	"time"
)

var worker *Worker

type Persister interface {
	Persist(map[string]Snapshot) error
}

// Configuration for the Stats middleware
type Configuration struct {
	logger      garnish.Logger
	window      time.Duration
	sampleSize  int64
	sampleSizeF float64
	routeLookup map[string]*Stat
	persister   Persister
	percentiles map[string]float64
}

func Configure(base *garnish.Configuration) *Configuration {
	return &Configuration{
		logger:      base.Logger,
		window:      time.Second * 5,
		sampleSize:  50,
		sampleSizeF: float64(50),
		percentiles: map[string]float64{"50p": 0.5, "75p": 0.75, "95p": 0.95},
		persister:   &FilePersister{"./stats.json"},
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(routeNames []string) (garnish.Middleware, error) {
	c.routeLookup = make(map[string]*Stat)
	for _, name := range routeNames {
		c.routeLookup[name] = &Stat{
			Configuration: c,
			snapshot:      make(Snapshot),
			samples:       make([]int, c.sampleSize),
		}
	}

	if worker != nil {
		worker.stop()
	}
	stats := &Stats{Configuration: c}
	worker = &Worker{Configuration: c}
	go worker.start()
	return stats, nil

}

// Reservoir sampling is used to report percentiles without having unbound
// growth. The sample size specifies how large a reservoir to use per route.
// A higher value will result in a more accurate report but requires more
// memory. The accuracy suffers from diminishing return, so there's no point
// setting this too high.

// [500]
func (c *Configuration) SampleSize(size int) *Configuration {
	c.sampleSize = int64(size)
	c.sampleSizeF = float64(size)
	return c
}

// The period of time to group statistics in.

// [1 minute]
func (c *Configuration) Window(window time.Duration) *Configuration {
	c.window = window
	return c
}

// The persister responsible for saving the statistics

// [FilePersister("./stats.json")]
func (c *Configuration) Persister(persister Persister) *Configuration {
	c.persister = persister
	return c
}

// The percentiles to track

// [50, 75, 95]
func (c *Configuration) Percentiles(percentiles ...int) *Configuration {
	lookup := make(map[string]float64)
	for _, p := range percentiles {
		lookup[strconv.Itoa(p)+"p"] = float64(p) / 100
	}
	c.percentiles = lookup
	return c
}
