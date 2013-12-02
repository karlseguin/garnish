package garnish

import (
	"github.com/karlseguin/gspec"
	"github.com/karlseguin/nd"
	"net/http"
	"testing"
)

func TestNewContextGetsARequestId(t *testing.T) {
	spec := gspec.New(t)
	nd.ForceGuid("7ea58ddf-bd8d-4f20-071f-01dcb003952a")
	context := newContext(new(http.Request), nil)
	spec.Expect(context.RequestId()).ToEqual("7ea58ddf-bd8d-4f20-071f-01dcb003952a")
}

func TestNewContextSetsUpstreamsRequestIdHeader(t *testing.T) {
	spec := gspec.New(t)
	nd.ForceGuid("cea58ddf-bd8d-4f20-071f-01dcb003952c")
	context := newContext(new(http.Request), nil)
	spec.Expect(context.RequestOut().Header.Get("X-Request-Id")).ToEqual("cea58ddf-bd8d-4f20-071f-01dcb003952c")
}

func TestNewContextReferencesIncomingRequest(t *testing.T) {
	spec := gspec.New(t)
	context := newContext(&http.Request{Method: "TEST"}, nil)
	spec.Expect(context.RequestIn().Method).ToEqual("TEST")
}
