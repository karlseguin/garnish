package router

import (
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestRouterReturnsNilIfRouteNotFound(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Req
	route, _, res := buildRouter().Route(gc.ContextBuilder().SetRequest(req))
	spec.Expect(route).ToBeNil()
	spec.Expect(res).ToBeNil()
}

func TestRouterMatchesRoot(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/").Req
	route, _, res := buildRouter("*", "root").Route(gc.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("root")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesANestedRoute(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/atreides/mentants.json").Req
	route, params, res := buildRouter("/", "root", "/houses/", "houses", "/houses/:houseName/mentants", "mentants").Route(gc.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("mentants")
	spec.Expect(params["ext"]).ToEqual("json")
	spec.Expect(params["houseName"]).ToEqual("atreides")
	spec.Expect(res).ToBeNil()
}

func TestRoutesMatchesANestedRouteEndingWithAParameter(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/harkonnen.json").Req
	route, params, res := buildRouter("/", "root", "/houses/", "houses", "/houses/:houseName", "showHouse").Route(gc.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("showHouse")
	spec.Expect(params["ext"]).ToEqual("json")
	spec.Expect(params["houseName"]).ToEqual("harkonnen")
	spec.Expect(res).ToBeNil()
}

// func TestRoutesMatchesANestedRouteAFailedConstraint(t *testing.T) {
// 	spec := gspec.New(t)
// 	req := gspec.Request().Url("/root/harkonnen.json").Req

// 	router := New(nil, nil)
// 	router.Add("root", "GET", "/root")
// 	router.Add("something", "GET", "/root/:something").Constrain("something", "one", "two")

// 	route, _, res := router.Route(gc.ContextBuilder().SetRequest(req))
// 	spec.Expect(route).ToBeNil()
// 	spec.Expect(res).ToBeNil()
// }

// func TestRoutesMatchesANestedRouteAPassedConstraint(t *testing.T) {
// 	spec := gspec.New(t)
// 	req := gspec.Request().Url("/root/one.json").Req

// 	router := New(nil, nil)
// 	router.Add("root", "GET", "/root")
// 	router.Add("something", "GET", "/root/:something").Constrain("something", "one", "two")

// 	route, _, res := router.Route(gc.ContextBuilder().SetRequest(req))
// 	spec.Expect(route.Name).ToEqual("something")
// 	spec.Expect(res).ToBeNil()
// }

func TestRoutesDontMatchPartials(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses/atreides/mentant").Req
	route, _, res := buildRouter("/", "root", "/houses/", "houses").Route(gc.ContextBuilder().SetRequest(req))
	spec.Expect(route).ToBeNil()
	spec.Expect(res).ToBeNil()
}

func TestReturnsNilRouteWhenNoRouteMathes(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/worms/dune/count").Req
	route, _, res := buildRouter("/", "root", "/houses/", "houses").Route(gc.ContextBuilder().SetRequest(req))
	spec.Expect(route).ToBeNil()
	spec.Expect(res).ToBeNil()
}

func TestRouterMatchesASimpleRoute(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses").Req
	route, _, res := buildRouter("/houses", "up1").Route(gc.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("up1")
	spec.Expect(res).ToBeNil()
}

func TestRouterMatchsAPrefix(t *testing.T) {
	spec := gspec.New(t)
	req := gspec.Request().Url("/houses").Req
	route, _, res := buildRouter("/house*", "up1").Route(gc.ContextBuilder().SetRequest(req))
	spec.Expect(route.Name).ToEqual("up1")
	spec.Expect(res).ToBeNil()
}

func buildRouter(data ...string) *Router {
	router := New(nil, nil)
	for i := 0; i < len(data); i += 2 {
		router.Add(data[i+1], "GET", data[i])
	}
	return router
}
