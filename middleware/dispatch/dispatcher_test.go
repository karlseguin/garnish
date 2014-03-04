package dispatch

import (
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestYieldsTheNextMiddlewareIfTheRouteIsntSetup(t *testing.T) {
	spec := gspec.New(t)
	res := newDispatch().Run(gc.ContextBuilder(), gc.FakeNext(gc.Respond(nil).Status(5555)))
	spec.Expect(res.GetStatus()).ToEqual(5555)
}

func TestYieldsTheNextMiddlewareIfTheRouteReturnsNil(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRoute(&gc.Route{Name: "nil"})
	res := newDispatch().Run(context, gc.FakeNext(gc.Respond(nil).Status(5554)))
	spec.Expect(res.GetStatus()).ToEqual(5554)
}

func TestReturnsTheDispatchedResponse(t *testing.T) {
	spec := gspec.New(t)
	context := gc.ContextBuilder().SetRoute(&gc.Route{Name: "ok"})
	res := newDispatch().Run(context, nil)
	spec.Expect(res.GetStatus()).ToEqual(200)
}

func newDispatch() *Dispatch {
	config := Configure().Dispatch(func(action interface{}, context gc.Context) gc.Response {
		return action.(func() gc.Response)()
	})
	config.Action("nil", func() gc.Response { return nil })
	config.Action("ok", func() gc.Response { return gc.Respond(nil).Status(200) })
	d, _ := config.Create(gc.DummyConfig)
	return d.(*Dispatch)
}
