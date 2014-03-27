package stats

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

type Stats struct {
	routeStats map[string]*Stat
}

func (s *Stats) Name() string {
	return "stats"
}

func (s *Stats) Run(context gc.Context, next gc.Next) gc.Response {
	response := next(context)
	elapsed := time.Now().Sub(context.StartTime())
	s.hit(context, response, elapsed)
	return response
}

func (s *Stats) hit(context gc.Context, response gc.Response, elapsed time.Duration) {
	context.Infof("%d µs", elapsed/1000)
	stat, ok := s.routeStats[context.Route().Name]
	if ok == false {
		stat = s.routeStats["?"]
	}
	stat.hit(response, elapsed)
}
