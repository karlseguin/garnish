package garnish

import (
	. "github.com/karlseguin/expect"
	"testing"
)

type ResponseTests struct{}

func Test_Response(t *testing.T) {
	Expectify(new(ResponseTests), t)
}

func (_ ResponseTests) Empty() {
	res := Empty(9001)
	Expect(res.Status()).To.Equal(9001)
	Expect(string(res.(*NormalResponse).body)).To.Equal("")
	Expect(len(res.Header())).To.Equal(0)
}
