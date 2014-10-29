package gc

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/typed"
	"os"
	"testing"
	"time"
)

type StatsTests struct{}

func Test_Stats(t *testing.T) {
	Expectify(new(StatsTests), t)
}

func (_ *StatsTests) CalculatesThePercentils() {
	s := NewRouteStats(time.Minute)
	for i := 1; i <= 20; i++ {
		s.Hit(Respond(200, ""), time.Millisecond*time.Duration(i))
	}
	snapshot := s.Snapshot()
	Expect(snapshot["75p"]).To.Equal(int64(15250))
	Expect(snapshot["95p"]).To.Equal(int64(19050))
}

func (_ *StatsTests) TracksSlows() {
	s := NewRouteStats(time.Millisecond * 10)
	for i := 1; i <= 20; i++ {
		s.Hit(Respond(200, ""), time.Millisecond*time.Duration(i))
	}
	snapshot := s.Snapshot()
	Expect(snapshot["slow"]).To.Equal(int64(10))
}

func (_ *StatsTests) TracksStatus() {
	s := NewRouteStats(time.Millisecond * 10)
	for i := 298; i < 503; i++ {
		s.Hit(Respond(i, ""), time.Millisecond)
	}
	snapshot := s.Snapshot()
	Expect(snapshot["hits"]).To.Equal(int64(205))
	Expect(snapshot["2xx"]).To.Equal(int64(102))
	Expect(snapshot["4xx"]).To.Equal(int64(100))
	Expect(snapshot["5xx"]).To.Equal(int64(3))
}


func (_ *StatsTests) TracksCache() {
	s := NewRouteStats(time.Millisecond * 10)
	for i := 298; i < 503; i++ {
		r := Respond(i, "")
		r.(*NormalResponse).cached = i % 2 == 0
		s.Hit(r, time.Millisecond)
	}
	snapshot := s.Snapshot()
	Expect(snapshot["cached"]).To.Equal(int64(103))
}

func (_ *StatsTests) Resets() {
	s := NewRouteStats(time.Millisecond * 10)
	for i := 298; i < 503; i++ {
		s.Hit(Respond(i, ""), time.Millisecond)
	}
	snapshot := s.Snapshot()
	snapshot = s.Snapshot()
	Expect(snapshot["hits"]).To.Equal(int64(0))
	Expect(snapshot["cached"]).To.Equal(int64(0))
	Expect(snapshot["2xx"]).To.Equal(int64(0))
	Expect(snapshot["4xx"]).To.Equal(int64(0))
	Expect(snapshot["5xx"]).To.Equal(int64(0))
	Expect(snapshot["95p"]).To.Equal(int64(0))
	Expect(snapshot["75p"]).To.Equal(int64(0))
}

func (_ *StatsTests) Persists() {
	defer os.Remove("test_stats.json")
	s := NewRouteStats(time.Millisecond * 350)
	for i := 297; i < 504; i++ {
		s.Hit(Respond(i, ""), time.Millisecond*time.Duration(i))
	}
	runtime := &Runtime{
		Routes: map[string]*Route{
			"about": &Route{Stats: s},
		},
	}
	sw := NewStatsWorker(runtime, "test_stats.json")
	runtime.StatsWorker = sw

	runtime.RegisterStats("test", func() map[string]int64 {
		return map[string]int64{"abc": 123, "991": 944}
	})

	sw.work()
	t, _ := typed.JsonFile("test_stats.json")
	r := t.Object("routes").Object("about")
	Expect(r.Int("slow")).To.Equal(153)
	Expect(r.Int("hits")).To.Equal(207)
	Expect(r.Int("2xx")).To.Equal(103)
	Expect(r.Int("4xx")).To.Equal(100)
	Expect(r.Int("5xx")).To.Equal(4)
	Expect(r.Int("75p")).To.Equal(451500)
	Expect(r.Int("95p")).To.Equal(492700)

	r = t.Object("runtime")
	// hard to tell for sure what these are going to be...
	Expect(r.Int("gc")).Greater.Than(0)
	Expect(r.Int("go")).Greater.Than(0)

	r = t.Object("other").Object("test")
	Expect(r.Int("abc")).To.Equal(123)
	Expect(r.Int("991")).To.Equal(944)
}
