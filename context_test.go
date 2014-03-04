package garnish

import (
	"github.com/karlseguin/gspec"
	"github.com/karlseguin/nd"
	"net/http"
	"net/url"
	"testing"
)

var emptyRequest = &http.Request{URL: new(url.URL)}

func TestNewContextGetsARequestId(t *testing.T) {
	spec := gspec.New(t)
	nd.ForceGuid("7ea58ddf-bd8d-4f20-071f-01dcb003952a")
	context := newContext(emptyRequest, nil)
	spec.Expect(context.RequestId()).ToEqual("7ea58ddf-bd8d-4f20-071f-01dcb003952a")
}

func TestNewContextReferencesIncomingRequest(t *testing.T) {
	spec := gspec.New(t)
	context := newContext(&http.Request{Method: "TEST", URL: new(url.URL)}, nil)
	spec.Expect(context.Request().Method).ToEqual("TEST")
}

func TestLoadsTheQueryString(t *testing.T) {
	assertQuery(t, "a=1", "a", "1")
	assertQuery(t, "a=1&b=2", "a", "1", "b", "2")
	assertQuery(t, "hello=world&test=", "hello", "world", "test", "")
	assertQuery(t, "x=abc%20%3F%20123", "x", "abc ? 123")
	context := newContext(&http.Request{URL: &url.URL{RawQuery: ""}}, nil)
	gspec.New(t).Expect(len(context.Query())).ToEqual(0)
}

func assertQuery(t *testing.T, raw string, keyValues ...string) {
	spec := gspec.New(t)
	context := newContext(&http.Request{URL: &url.URL{RawQuery: raw}}, nil)
	query := context.Query()
	spec.Expect(len(query)).ToEqual(len(keyValues) / 2)

	for i := 0; i < len(keyValues); i += 2 {
		key, value := keyValues[i], keyValues[i+1]
		spec.Expect(query[key]).ToEqual(value)
	}
}
