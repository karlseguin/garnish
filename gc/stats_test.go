package gc

import (
	"testing"
	"time"
	. "github.com/karlseguin/expect"
	"os"
	"github.com/karlseguin/typed"
)

type StatsTests struct{}

func Test_Stats(t *testing.T) {
	Expectify(new(StatsTests), t)
}

func (_ *StatsTests) CalculatesThePercentils() {
	s := NewStats(time.Minute)
	for i := 1; i <= 20; i++ {
		s.Hit(Respond(200, ""), time.Millisecond * time.Duration(i))
	}
	snapshot := s.Snapshot()
	Expect(snapshot["75p"]).To.Equal(int64(15250))
	Expect(snapshot["95p"]).To.Equal(int64(19050))
}

func (_ *StatsTests) TracksSlows() {
	s := NewStats(time.Millisecond * 10)
	for i := 1; i <= 20; i++ {
		s.Hit(Respond(200, ""), time.Millisecond * time.Duration(i))
	}
	snapshot := s.Snapshot()
	Expect(snapshot["slow"]).To.Equal(int64(10))
}

func (_ *StatsTests) TracksStatus() {
	s := NewStats(time.Millisecond * 10)
	for i := 298; i < 503; i++ {
		s.Hit(Respond(i, ""), time.Millisecond)
	}
	snapshot := s.Snapshot()
	Expect(snapshot["hits"]).To.Equal(int64(205))
	Expect(snapshot["2xx"]).To.Equal(int64(102))
	Expect(snapshot["4xx"]).To.Equal(int64(100))
	Expect(snapshot["5xx"]).To.Equal(int64(3))
}

func (_ *StatsTests) Resets() {
	s := NewStats(time.Millisecond * 10)
	for i := 298; i < 503; i++ {
		s.Hit(Respond(i, ""), time.Millisecond)
	}
	snapshot := s.Snapshot()
	snapshot = s.Snapshot()
	Expect(snapshot["hits"]).To.Equal(int64(0))
	Expect(snapshot["2xx"]).To.Equal(int64(0))
	Expect(snapshot["4xx"]).To.Equal(int64(0))
	Expect(snapshot["5xx"]).To.Equal(int64(0))
	Expect(snapshot["95p"]).To.Equal(int64(0))
	Expect(snapshot["75p"]).To.Equal(int64(0))
}

func (_ *StatsTests) Persists() {
	defer os.Remove("test_stats.json")
	s := NewStats(time.Millisecond * 350)
	for i := 297; i < 504; i++ {
		s.Hit(Respond(i, ""), time.Millisecond * time.Duration(i))
	}
	runtime := &Runtime{
		StatsFileName: "test_stats.json",
		Routes: map[string]*Route{
			"about": &Route{Stats: s},
		},
	}
	dumpStats(runtime)
	t, _ := typed.JsonFile("test_stats.json")
	t = t.Object("routes").Object("about")
	Expect(t.Int("slow")).To.Equal(153)
	Expect(t.Int("hits")).To.Equal(207)
	Expect(t.Int("2xx")).To.Equal(103)
	Expect(t.Int("4xx")).To.Equal(100)
	Expect(t.Int("5xx")).To.Equal(4)
	Expect(t.Int("75p")).To.Equal(451500)
	Expect(t.Int("95p")).To.Equal(492700)
}
