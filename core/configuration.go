package core

type Configuration interface {
	Router() Router
	Logger() Logger
}

var DummyConfig = &dummyConfiguration{newFakeLogger(), newFakeRouter()}

type dummyConfiguration struct {
	logger Logger
	router Router
}

func (c *dummyConfiguration) Logger() Logger {
	return c.logger
}

func (c *dummyConfiguration) Router() Router {
	return c.router
}
