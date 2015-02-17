package garnish

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/garnish/gc"
	"testing"
)

type ConfigurationTests struct{}

func Test_Configuration(t *testing.T) {
	Expectify(new(ConfigurationTests), t)
}

func (_ ConfigurationTests) FailedBuildWithoutUpstream() {
	logger := gc.NewFakeLogger()
	c := Configure().Logger(logger).DnsTTL(-1)
	Expect(c.Build()).To.Equal(nil)
	Expect(logger.Errors).To.Contain("Atleast one upstream must be configured")
}

func (_ ConfigurationTests) FailedBuildWithMissingUpstreamAddress() {
	logger := gc.NewFakeLogger()
	c := Configure().Logger(logger).DnsTTL(-1)
	c.Upstream("test")
	Expect(c.Build()).To.Equal(nil)
	Expect(logger.Errors).To.Contain(`Upstream test has an invalid address: ""`)
}

func (_ ConfigurationTests) FailedBuildWithInvalidUpstreamAddress() {
	logger := gc.NewFakeLogger()
	c := Configure().Logger(logger).DnsTTL(-1)
	c.Upstream("test1").Address("http://openmymind.net/")
	c.Upstream("test2").Address("128.93.202.0")
	Expect(c.Build()).To.Equal(nil)
	Expect(logger.Errors).To.Contain(`Upstream test2's address should begin with unix:/, http:// or https://`)
}

func (_ ConfigurationTests) FailedBuildWithoutRoute() {
	logger := gc.NewFakeLogger()
	c := Configure().Logger(logger).DnsTTL(-1)
	c.Upstream("test1").Address("http://openmymind.net/")
	Expect(c.Build()).To.Equal(nil)
	Expect(logger.Errors).To.Contain("Atleast one route must be configured")
}
