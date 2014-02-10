package core

type Configuration interface {
	Router() Router
	Logger() Logger
}
