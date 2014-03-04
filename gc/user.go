package gc

type User interface {
	Id() string
	IntId() int
	Permissions() map[string]bool
}
