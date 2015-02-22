package garnish

import (
	. "github.com/karlseguin/expect"
	// "github.com/karlseguin/garnish/gc"
	"testing"
)

type ConfigurationTests struct{}

func Test_Configuration(t *testing.T) {
	Expectify(new(ConfigurationTests), t)
}

func (_ ConfigurationTests) FailedBuildWithoutUpstream() {
	c := Configure().DnsTTL(-1)
	r, err := c.Build()
	Expect(r).To.Equal(nil)
	Expect(err.Error()).To.Contain("Atleast one upstream must be configured")
}

func (_ ConfigurationTests) FailedBuildWithMissingUpstreamAddress() {
	c := Configure().DnsTTL(-1)
	c.Upstream("test")
	r, err := c.Build()
	Expect(r).To.Equal(nil)
	Expect(err.Error()).To.Contain(`Upstream test has an invalid address: ""`)
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
