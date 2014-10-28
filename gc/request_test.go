package gc

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/nd"
	"testing"
)

type RequestTests struct{}

func Test_Request(t *testing.T) {
	Expectify(new(RequestTests), t)
}

func (r *RequestTests) UniqueId() {
	nd.ForceGuid("7ea58ddf-bd8d-4f20-071f-01dcb003952a")
	req := NewRequest(nil, nil, nil)
	Expect(req.Id).To.Equal("7ea58ddf-bd8d-4f20-071f-01dcb003952a")
}

func (r *RequestTests) StartTime() {
	now := nd.LockTime()
	req := NewRequest(nil, nil, nil)
	Expect(req.Start).To.Equal(now)
}
