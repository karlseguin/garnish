package gc

type Route struct {
	Name     string
	Upstream *Upstream
	Stats    *Stats
}
