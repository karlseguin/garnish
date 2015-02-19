package gc

import (
	"net/http"
)

// The User Agent to send to the upstream
var DefaultUserAgent = []string{""}

// Tweaks request `out` before sending it to the upstream
type RequestTweaker func(in *Request, out *http.Request)

type Upstream struct {
	Name      string
	Address   string
	Transport *http.Transport
	Headers   []string
	Tweaker   RequestTweaker
}
