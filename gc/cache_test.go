package gc

import (
	"testing"
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"time"
	"net/http"
)

type CacheTests struct {
	request *Request
}

func Test_Cache(t *testing.T) {
	route := &Route{Cache: &RouteCache{TTL: time.Minute}}
	ct := &CacheTests{
		request: NewRequest(build.Request().Request, route, nil),
	}
	Expectify(ct, t)
}

func (_ *CacheTests) RouteSpecificTTL() {
	c := new(Cache)
	ttl := c.ttl(&RouteCache{TTL: time.Minute}, Respond(200, "hello"))
	Expect(ttl).To.Equal(time.Minute)
}

func (_ *CacheTests) TTLForPrivateOnOverride() {
	c := new(Cache)
	ttl := c.ttl(&RouteCache{TTL: time.Second * 2}, RespondH(200, http.Header{"Cache-Control": []string{"private;max-age=10"}}, "hello"))
	Expect(ttl).To.Equal(time.Second * 2)
}

func (_ *CacheTests) HeaderTTL() {
	c := new(Cache)
	ttl := c.ttl(&RouteCache{}, RespondH(200, http.Header{"Cache-Control": []string{"public;max-age=20"}}, "hello"))
	Expect(ttl).To.Equal(time.Second * 20)
}

func (_ *CacheTests) HeaderTTLOnError() {
	c := new(Cache)
	ttl := c.ttl(&RouteCache{TTL: time.Minute}, RespondH(404, http.Header{"Cache-Control": []string{"public;max-age=10"}}, "hello"))
	Expect(ttl).To.Equal(time.Second * 10)
}

func (_ *CacheTests) NoTTLWhenPrivate() {
	c := new(Cache)
	ttl := c.ttl(&RouteCache{}, RespondH(200, http.Header{"Cache-Control": []string{"private;max-age=10"}}, "hello"))
	Expect(ttl).To.Equal(int64(0))
}

func (_ *CacheTests) GraceSingleDownload() {
	c := NewCache(10)
	c.downloads["pk"] = time.Now().Add(time.Minute)
	c.Grace("p", "k", nil, nil)
}

func (ct *CacheTests) GraceForcesOnStaleDownloads() {
	c := NewCache(10)
	c.downloads["pk"] = time.Now().Add(time.Minute * -1)
	called := false
	c.Grace("p", "k", ct.request, func(req *Request) Response 	{
		called = true
		return Respond(200, "ok")
	})
	Expect(called).To.Equal(true)
	Expect(c.downloads).Not.To.Contain("pk")
}

func (ct *CacheTests) GraceDownload() {
	c := NewCache(10)
	c.downloads["abcd"] = time.Now().Add(time.Minute * -1)
	c.Grace("ab", "cd", ct.request, func(req *Request) Response {
		return Respond(200, "ok")
	})
	Expect(string(c.Get("ab", "cd").Value().(Response).Body())).To.Equal("ok")
}
