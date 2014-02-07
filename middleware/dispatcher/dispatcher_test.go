package dispatcher

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestYieldsTheNextMiddlewareIfTheRouteIsntSetup(t *testing.T) {
	spec := gspec.New(t)
	res := newDispatcher().Run(garnish.ContextBuilder(), garnish.FakeNext(garnish.Respond(nil).Status(5555)))
	spec.Expect(res.GetStatus()).ToEqual(5555)
}

func TestYieldsTheNextMiddlewareIfTheRouteReturnsNil(t *testing.T) {
	spec := gspec.New(t)
	context := garnish.ContextBuilder().SetRoute(&garnish.Route{Name: "nil"})
	res := newDispatcher().Run(context, garnish.FakeNext(garnish.Respond(nil).Status(5554)))
	spec.Expect(res.GetStatus()).ToEqual(5554)
}

func TestReturnsTheDispatchedResponse(t *testing.T) {
	spec := gspec.New(t)
	context := garnish.ContextBuilder().SetRoute(&garnish.Route{Name: "ok"})
	res := newDispatcher().Run(context, nil)
	spec.Expect(res.GetStatus()).ToEqual(200)
}

func newDispatcher() *Dispatcher {
	config := Configure(garnish.Configure()).Dispatch(func(action interface{}, context garnish.Context) garnish.Response {
		return action.(func() garnish.Response)()
	})
	config.Action("nil", func() garnish.Response { return nil })
	config.Action("ok", func() garnish.Response { return garnish.Respond(nil).Status(200) })
	return &Dispatcher{config}
}
