package dispatch

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestYieldsTheNextMiddlewareIfTheRouteIsntSetup(t *testing.T) {
	spec := gspec.New(t)
	res := newDispatch().Run(core.ContextBuilder(), core.FakeNext(core.Respond(nil).Status(5555)))
	spec.Expect(res.GetStatus()).ToEqual(5555)
}

func TestYieldsTheNextMiddlewareIfTheRouteReturnsNil(t *testing.T) {
	spec := gspec.New(t)
	context := core.ContextBuilder().SetRoute(&core.Route{Name: "nil"})
	res := newDispatch().Run(context, core.FakeNext(core.Respond(nil).Status(5554)))
	spec.Expect(res.GetStatus()).ToEqual(5554)
}

func TestReturnsTheDispatchedResponse(t *testing.T) {
	spec := gspec.New(t)
	context := core.ContextBuilder().SetRoute(&core.Route{Name: "ok"})
	res := newDispatch().Run(context, nil)
	spec.Expect(res.GetStatus()).ToEqual(200)
}

func newDispatch() *Dispatch {
	config := Configure().Dispatch(func(action interface{}, context core.Context) core.Response {
		return action.(func() core.Response)()
	})
	config.Action("nil", func() core.Response { return nil })
	config.Action("ok", func() core.Response { return core.Respond(nil).Status(200) })
	d, _ := config.Create(core.DummyConfig)
	return d.(*Dispatch)
}
