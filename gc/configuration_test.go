package gc

import (
	. "github.com/karlseguin/expect"
	"testing"
)

type ConfigurationTests struct{}

func Test_Configuration(t *testing.T) {
	Expectify(new(ConfigurationTests), t)
}

func (_ ConfigurationTests) FailedBuildWithMissingTransport() {
	c := Configure().DnsTTL(-1)
	c.Upstream("test")
	r, err := c.Build()
	Expect(r).To.Equal(nil)
	Expect(err.Error()).To.Contain(`Upstream test doesn't have a configured transport`)
}

func (_ ConfigurationTests) FailedBuildWithInvalidUpstreamAddress() {
	c := Configure().DnsTTL(-1)
	c.Upstream("test1").Address("http://openmymind.net/")
	c.Upstream("test2").Address("128.93.202.0")
	r, err := c.Build()
	Expect(r).To.Equal(nil)
	Expect(err.Error()).To.Contain(`Upstream test2's address should begin with unix:/, http:// or https://`)
}

func (_ ConfigurationTests) FailedBuildWithoutRoute() {
	c := Configure().DnsTTL(-1)
	c.Upstream("test1").Address("http://openmymind.net/")
	r, err := c.Build()
	Expect(r).To.Equal(nil)
	Expect(err.Error()).To.Contain("Atleast one route must be configured")
}
