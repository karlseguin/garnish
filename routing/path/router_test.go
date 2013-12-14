package path

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestRouterReturnsNilIfRouteNotFound(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Req
	route, _, res := buildRouter().router(garnish.ContextBuilder().SetRequestIn(req))
	spec.Expect(route).ToBeNil()
	spec.Expect(res).ToBeNil()
}

func TestRouterMatchesRoot(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/").Req
	route, _, res := buildRouter("/", "root").router(garnish.ContextBuilder().SetRequestIn(req))
	spec.Expect(route.Upstream).ToEqual("root")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesANestedRoute(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/atreides/mentant.json").Req
	route, params, res := buildRouter("/", "root", "/houses/", "houses", "/houses/:houseName/mentant/", "mentant").router(garnish.ContextBuilder().SetRequestIn(req))
	spec.Expect(route.Upstream).ToEqual("mentant")
	spec.Expect(params["ext"]).ToEqual("json")
	spec.Expect(params["houseName"]).ToEqual("atreides")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesANestedRouteEndingWithAParameter(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/harkonnen.json").Req
	route, params, res := buildRouter("/", "root", "/houses/", "houses", "/houses/:houseName", "showHouse").router(garnish.ContextBuilder().SetRequestIn(req))
	spec.Expect(route.Upstream).ToEqual("showHouse")
	spec.Expect(params["ext"]).ToEqual("json")
	spec.Expect(params["houseName"]).ToEqual("harkonnen")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesATheClosestMatch(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/atreides/mentant").Req
	route, _, res := buildRouter("/", "root", "/houses/", "houses").router(garnish.ContextBuilder().SetRequestIn(req))
	spec.Expect(route.Upstream).ToEqual("houses")
	spec.Expect(res).ToBeNil()
}

func TestReturnsNilRouteWhenNoRouteMathes(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/worms/dune/count").Req
	route, _, res := buildRouter("/", "root", "/houses/", "houses").router(garnish.ContextBuilder().SetRequestIn(req))
	spec.Expect(route).ToBeNil()
	spec.Expect(res).ToBeNil()
}

func TestRouterMatchesASimpleRoute(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/").Req
	route, _, res := buildRouter("/houses", "up1").router(garnish.ContextBuilder().SetRequestIn(req))
	spec.Expect(route.Upstream).ToEqual("up1")
	spec.Expect(res).ToBeNil()
}

func buildRouter(data ...string) *Router {
	config := Configure(garnish.Configure())
	for i := 0; i < len(data); i += 2 {
		config.Add("GET", data[i], &garnish.Route{Upstream: data[i+1]})
	}
	return &Router{config}
}
