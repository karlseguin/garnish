package pathrouter

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestConfiguresARootRoute(t *testing.T) {
	spec := gspec.New(t)
	c := Configure(garnish.Configure()).Add("GET", "/", new(garnish.Route))
	spec.Expect(c.routes["GET"].route).ToNotBeNil()
	spec.Expect(len(c.routes["GET"].routes)).ToEqual(0)
}

func TestConfiguresASimpleRoute(t *testing.T) {
	spec := gspec.New(t)
	c := Configure(garnish.Configure()).Add("GET", "/houses", new(garnish.Route))
	spec.Expect(c.routes["GET"].routes["houses"].route).ToNotBeNil()
	spec.Expect(len(c.routes["GET"].routes["houses"].routes)).ToEqual(0)
}

func TestConfiguresMultipleRoutes(t *testing.T) {
	spec := gspec.New(t)
	c := Configure(garnish.Configure()).Add("GET", "/houses", new(garnish.Route))
	c.Add("GET", "/houses/*/gholas", new(garnish.Route))
	spec.Expect(c.routes["GET"].routes["houses"].route).ToNotBeNil()
	spec.Expect(c.routes["GET"].routes["houses"].routes["*"].routes["gholas"].route).ToNotBeNil()
}

func TestConfigurationOfRouteMethods(t *testing.T) {
	spec := gspec.New(t)
	c := Configure(garnish.Configure()).Add("*,test", "/houses", new(garnish.Route))
	spec.Expect(c.routes["GET"]).ToNotBeNil()
	spec.Expect(c.routes["PUT"]).ToNotBeNil()
	spec.Expect(c.routes["POST"]).ToNotBeNil()
	spec.Expect(c.routes["DELETE"]).ToNotBeNil()
	spec.Expect(c.routes["PURGE"]).ToNotBeNil()
	spec.Expect(c.routes["PATCH"]).ToNotBeNil()
	spec.Expect(c.routes["TEST"]).ToNotBeNil()
	spec.Expect(c.routes["OTHER"]).ToBeNil()
}
