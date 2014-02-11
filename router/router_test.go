package router

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestRouterReturnsNilIfRouteNotFound(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Req
	route, _, res := buildRouter().Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route).ToBeNil()
	spec.Expect(res).ToBeNil()
}

func TestRouterMatchesRoot(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/").Req
	route, _, res := buildRouter("/", "root").Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("root")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesANestedRoute(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/atreides/mentant.json").Req
	route, params, res := buildRouter("/", "root", "/houses/", "houses", "/houses/:houseName/mentant/", "mentant").Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("mentant")
	spec.Expect(params["ext"]).ToEqual("json")
	spec.Expect(params["houseName"]).ToEqual("atreides")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesANestedRouteEndingWithAParameter(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/harkonnen.json").Req
	route, params, res := buildRouter("/", "root", "/houses/", "houses", "/houses/:houseName", "showHouse").Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("showHouse")
	spec.Expect(params["ext"]).ToEqual("json")
	spec.Expect(params["houseName"]).ToEqual("harkonnen")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesANestedRouteAFailedConstraint(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/root/harkonnen.json").Req

	config := Configure(garnish.Configure())
	config.Add("GET", "/root", &garnish.Route{Name: "root"})
	config.Add("GET", "/root/:something", &garnish.Route{Name: "something"}).ParamContraint("something", "one", "two")
	r := &Router{config.compile()}

	route, _, res := r.Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("root")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesANestedRouteAPassedConstraint(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/root/one.json").Req

	config := Configure(garnish.Configure())
	config.Add("GET", "/root", &garnish.Route{Name: "root"})
	config.Add("GET", "/root/:something", &garnish.Route{Name: "something"}).ParamContraint("something", "one", "two")
	r := &Router{config.compile()}

	route, _, res := r.Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("something")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesATheClosestMatch(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/atreides/mentant").Req
	route, _, res := buildRouter("/", "root", "/houses/", "houses").Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("houses")
	spec.Expect(res).ToBeNil()
}

func TestReturnsNilRouteWhenNoRouteMathes(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/worms/dune/count").Req
	route, _, res := buildRouter("/", "root", "/houses/", "houses").Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route).ToBeNil()
	spec.Expect(res).ToBeNil()
}

func TestRouterMatchesASimpleRoute(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/").Req
	route, _, res := buildRouter("/houses", "up1").Route(garnish.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("up1")
	spec.Expect(res).ToBeNil()
}

func buildRouter(data ...string) *Router {
	config := Configure(garnish.Configure())
	for i := 0; i < len(data); i += 2 {
		config.Add("GET", data[i], &garnish.Route{Name: data[i+1]})
	}
	return &Router{config.compile()}
}
