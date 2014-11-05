package gc

import (
	. "github.com/karlseguin/expect"
	"testing"
)

type ResponseTests struct{}

func Test_Response(t *testing.T) {
	Expectify(new(ResponseTests), t)
}

func (_ *ResponseTests) Empty() {
	res := Empty(9001)
	Expect(res.Status()).To.Equal(9001)
	Expect(string(res.(*NormalResponse).body)).To.Equal("")
	Expect(len(res.Header())).To.Equal(0)
}

func (_ *ResponseTests) Fatal() {
	res := Fatal("a message")
	Expect(res.Status()).To.Equal(500)
	Expect(len(res.Header())).To.Equal(0)
	Expect(res.(*FatalResponse).Err).ToEqual("a message")
}
