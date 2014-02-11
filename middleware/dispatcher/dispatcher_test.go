package dispatcher

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestYieldsTheNextMiddlewareIfTheRouteIsntSetup(t *testing.T) {
	spec := gspec.New(t)
	res := newDispatcher().Run(core.ContextBuilder(), core.FakeNext(core.Respond(nil).Status(5555)))
	spec.Expect(res.GetStatus()).ToEqual(5555)
}

func TestYieldsTheNextMiddlewareIfTheRouteReturnsNil(t *testing.T) {
	spec := gspec.New(t)
	context := core.ContextBuilder().SetRoute(&core.Route{Name: "nil"})
	res := newDispatcher().Run(context, core.FakeNext(core.Respond(nil).Status(5554)))
	spec.Expect(res.GetStatus()).ToEqual(5554)
}

func TestReturnsTheDispatchedResponse(t *testing.T) {
	spec := gspec.New(t)
	context := core.ContextBuilder().SetRoute(&core.Route{Name: "ok"})
	res := newDispatcher().Run(context, nil)
	spec.Expect(res.GetStatus()).ToEqual(200)
}

func newDispatcher() *Dispatcher {
	config := Configure().Dispatch(func(action interface{}, context core.Context) core.Response {
		return action.(func() core.Response)()
	})
	config.Action("nil", func() core.Response { return nil })
	config.Action("ok", func() core.Response { return core.Respond(nil).Status(200) })
	d, _ := config.Create(core.DummyConfig)
	return d.(*Dispatcher)
}
