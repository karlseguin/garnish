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

func (ct *ConfigurationTests) FailedBuildWithoutUpstream() {
	c := Configure()
	Expect(c.Build()).To.Equal(nil)
	gc.ExpectError("Atleast one upstream must be configured")
}

func (ct *ConfigurationTests) FailedBuildWithMissingUpstreamAddress() {
	c := Configure()
	c.Upstream("test")
	Expect(c.Build()).To.Equal(nil)
	gc.ExpectError(`Upstream test has an invalid address: ""`)
}

func (ct *ConfigurationTests) FailedBuildWithInvalidUpstreamAddress() {
	c := Configure()
	c.Upstream("test1").Address("http://openmymind.net/")
	c.Upstream("test2").Address("128.93.202.0")
	Expect(c.Build()).To.Equal(nil)
	gc.ExpectError(`Upstream test2's address should begin with unix:/, http:// or https://`)
}

func (ct *ConfigurationTests) FailedBuildWithoutRoute() {
	c := Configure()
	c.Upstream("test1").Address("http://openmymind.net/")
	Expect(c.Build()).To.Equal(nil)
	gc.ExpectError("Atleast one route must be configured")
}

func (ct *ConfigurationTests) FailedBuildForUnpathedRoute() {
	c := Configure()
	c.Upstream("test1").Address("http://openmymind.net/")
	c.Route("bad")
	Expect(c.Build()).To.Equal(nil)
	gc.ExpectError(`Route "bad" doesn't have a method+path`)
}

func (ct *ConfigurationTests) FailedBuildForMissingRouteUpstream() {
	c := Configure()
	c.Upstream("test1").Address("http://openmymind.net/")
	c.Route("list_users").Get("/users")
	Expect(c.Build()).To.Equal(nil)
	gc.ExpectError(`Route "list_users" has an unknown upstream ""`)
}

func (ct *ConfigurationTests) FailedBuildForInvalidRouteUpstream() {
	c := Configure()
	c.Upstream("test1").Address("http://openmymind.net/")
	c.Route("list_users").Get("/users").Upstream("invalid")
	Expect(c.Build()).To.Equal(nil)
	gc.ExpectError(`Route "list_users" has an unknown upstream "invalid"`)
}
