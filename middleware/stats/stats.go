package stats

import (
	"github.com/karlseguin/garnish/core"
	"time"
)

type Stats struct {
	*Configuration
}

func (s *Stats) Name() string {
	return "stats"
}

func (s *Stats) Run(context core.Context, next core.Next) core.Response {
	start := time.Now()
	response := next(context)
	elapsed := time.Now().Sub(start)
	s.hit(context, response, elapsed)
	return response
}

func (s *Stats) hit(context core.Context, response core.Response, elapsed time.Duration) {
	s.logger.Infof(context, "%d µs", elapsed/1000)
	stat, ok := s.routeStats[context.Route().Name]
	if ok == false {
		stat = s.routeStats["?"]
	}
	stat.hit(response, elapsed)
}
