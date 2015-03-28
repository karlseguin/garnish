package cache

import (
	"bytes"
	. "github.com/karlseguin/expect"
	"gopkg.in/karlseguin/garnish.v1/gc"
	"strconv"
	"testing"
	"time"
)

type CacheTests struct{}

func Test_Cache(t *testing.T) {
	Expectify(new(CacheTests), t)
}

func (_ CacheTests) SetsAValue() {
	cache := New(100000)
	cache.Set("spice", "must", buildResponse("flow"))
	cache.Set("worm", "likes", buildResponse("sand"))
	assertResponse(cache.Get("spice", "must"), "flow")
	assertResponse(cache.Get("worm", "likes"), "sand")
}

func (_ CacheTests) SetsOfNewItemAdjustsSize() {
	cache := New(100000)
	cache.Set("spice", "must", buildResponse("xxx"))
	time.Sleep(time.Millisecond * 10)
	Expect(cache.size).To.Equal(303)
}

func (_ CacheTests) SetOfReplacementAdjustsSize() {
	cache := New(100000)
	cache.Set("spice", "must", buildResponse("xxx"))
	time.Sleep(time.Millisecond * 10)
	cache.Set("spice", "must", buildResponse("x"))
	time.Sleep(time.Millisecond * 10)
	Expect(cache.size).To.Equal(301)
}

func (_ CacheTests) GetsNil() {
	cache := New(100000)
	Expect(cache.Get("harkonnen", "other")).To.Equal(nil)
}

func (_ CacheTests) GCsTheOldestItems() {
	cache := New(2000000)
	for i := 0; i < 1500; i++ {
		id := strconv.Itoa(i)
		cache.Set(id, id, buildResponse(id))
	}
	Expect(cache.size).To.Equal(310186)
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	Expect(cache.Get("999", "999")).To.Equal(nil)
	assertResponse(cache.Get("1000", "1000"), "1000")
	Expect(cache.size).To.Equal(152000)
}

func (_ CacheTests) GetPromotesAValue() {
	cache := New(2000000)
	for i := 0; i < 1500; i++ {
		id := strconv.Itoa(i)
		cache.Set(id, id, buildResponse(id))
	}
	Expect(cache.size).To.Equal(310186)
	cache.Get("1", "1")
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	Expect(cache.Get("999", "999")).To.Equal(nil)
	Expect(cache.Get("1000", "1000")).To.Equal(nil)
	assertResponse(cache.Get("1", "1"), "1")
	Expect(cache.size).To.Equal(151997)
}

func buildResponse(body string) gc.CachedResponse {
	return gc.Respond(200, body).(*gc.NormalResponse)
}

func assertResponse(response gc.CachedResponse, expected string) {
	buffer := new(bytes.Buffer)
	response.Write(nil, buffer)
	Expect(buffer.String()).To.Equal(expected)
}
