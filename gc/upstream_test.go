package gc

import (
	. "github.com/karlseguin/expect"
	"testing"
)

type UpstreamTests struct{}

func Test_Upstream(t *testing.T) {
	Expectify(new(UpstreamTests), t)
}

// *shrug*
func (_ UpstreamTests) RandomlyPicksATransport() {
	u := &Upstream{
		Transports: []*Transport{&Transport{Address: "a"}, &Transport{Address: "b"}},
	}

	hits := map[string]int{"a": 0, "b": 0}
	for i := 0; i < 1000; i++ {
		hits[u.Transport().Address]++
	}
	d := hits["a"] - hits["b"]
	if d < 0 {
		d *= -1
	}
	Expect(d).Less.Than(10)
}
