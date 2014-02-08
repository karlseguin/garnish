package stats

import (
	"github.com/karlseguin/garnish"
	"time"
)

type Stats struct {
	*Configuration
}

func (s *Stats) Name() string {
	return "stats"
}

func (s *Stats) Run(context garnish.Context, next garnish.Next) garnish.Response {
	start := time.Now()
	response := next(context)
	elapsed := time.Now().Sub(start)
	s.hit(context, response, elapsed)
	return response
}

func (s *Stats) hit(context garnish.Context, response garnish.Response, elapsed time.Duration) {
	s.logger.Infof(context, "%d ms", elapsed)
	stat, ok := s.routeLookup[context.Route().Name]
	if ok == false {
		s.logger.Errorf(context, "Stats middleware got an unknown route (this should not happen) %s", context.Route().Name)
		return
	}
	stat.hit(response, elapsed)
}
