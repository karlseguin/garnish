package router

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestConfiguresARootRoute(t *testing.T) {
	spec := gspec.New(t)
	c := Configure(garnish.Configure())
	c.Add("GET", "/", new(garnish.Route))
	c.compile()
	spec.Expect(c.routes["GET"].route).ToNotBeNil()
	spec.Expect(len(c.routes["GET"].routes)).ToEqual(0)
}

func TestConfiguresASimpleRoute(t *testing.T) {
	spec := gspec.New(t)
	c := Configure(garnish.Configure())
	c.Add("GET", "/houses", new(garnish.Route))
	c.compile()
	spec.Expect(c.routes["GET"].routes["houses"].route).ToNotBeNil()
	spec.Expect(len(c.routes["GET"].routes["houses"].routes)).ToEqual(0)
}

func TestConfiguresMultipleRoutes(t *testing.T) {
	spec := gspec.New(t)
	c := Configure(garnish.Configure())
	c.Add("GET", "/houses", &garnish.Route{Name: "ListHouses"})
	c.Add("GET", "/houses/:house/gholas", &garnish.Route{Name: "ListHouseGholas"})
	c.compile()
	spec.Expect(c.routes["GET"].routes["houses"].route).ToNotBeNil()
	spec.Expect(c.routes["GET"].routes["houses"].routes["*"].routes["gholas"].route).ToNotBeNil()
	spec.Expect(c.routes["GET"].routes["houses"].routes["*"].parameterName).ToEqual("house")
}

func TestConfigurationOfRouteMethods(t *testing.T) {
	spec := gspec.New(t)
	c := Configure(garnish.Configure())
	c.Add("*,test", "/houses", new(garnish.Route))
	c.compile()
	spec.Expect(c.routes["GET"]).ToNotBeNil()
	spec.Expect(c.routes["PUT"]).ToNotBeNil()
	spec.Expect(c.routes["POST"]).ToNotBeNil()
	spec.Expect(c.routes["DELETE"]).ToNotBeNil()
	spec.Expect(c.routes["PURGE"]).ToNotBeNil()
	spec.Expect(c.routes["PATCH"]).ToNotBeNil()
	spec.Expect(c.routes["TEST"]).ToNotBeNil()
	spec.Expect(c.routes["OTHER"]).ToBeNil()
}

func TestPanicsIfDuplcateRouteNamesAreUsed(t *testing.T) {
	defer func() {
		if r := recover(); r.(string) != `Route names must be unique, "ListHouses" used twice` {
			t.Error("expecting panic on duplicate route name")
		}
	}()
	c := Configure(garnish.Configure())
	c.Add("GET", "/houses", &garnish.Route{Name: "ListHouses"})
	c.Add("GET", "/houses", &garnish.Route{Name: "ListHouses"})
	c.compile()
}
