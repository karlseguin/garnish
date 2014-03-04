package ccache

import (
	"github.com/karlseguin/garnish/caches"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/gspec"
	"testing"
)

func TestGetMissFromBucket(t *testing.T) {
	bucket := testBucket()
	gspec.New(t).Expect(bucket.get("invalid", "vary")).ToBeNil()
}

func TestGetHitFromBucket(t *testing.T) {
	bucket := testBucket()
	item := bucket.get("goku", "power")
	assertValue(t, item, "9000")
}

func TestDeleteItemFromBucket(t *testing.T) {
	bucket := testBucket()
	bucket.delete("power")
	gspec.New(t).Expect(bucket.get("power", "power")).ToBeNil()
}

func TestSetsANewBucketItem(t *testing.T) {
	spec := gspec.New(t)
	bucket := testBucket()
	item, new := bucket.set("spice", "must", fakeResponse("flow"))
	assertValue(t, item, "flow")
	item = bucket.get("spice", "must")
	assertValue(t, item, "flow")
	spec.Expect(new).ToEqual(true)
}

func TestSetsAnExistingItem(t *testing.T) {
	spec := gspec.New(t)
	bucket := testBucket()
	item, new := bucket.set("goku", "power", fakeResponse("9002"))
	assertValue(t, item, "9002")
	item = bucket.get("goku", "power")
	assertValue(t, item, "9002")
	spec.Expect(new).ToEqual(false)
}

func TestSetsANewVariance(t *testing.T) {
	spec := gspec.New(t)
	bucket := testBucket()
	item, new := bucket.set("goku", "dbz", fakeResponse("7"))
	assertValue(t, item, "7")
	item = bucket.get("goku", "power")
	assertValue(t, item, "9000")
	item = bucket.get("goku", "dbz")
	assertValue(t, item, "7")
	spec.Expect(new).ToEqual(true)
}

func testBucket() *Bucket {
	b := &Bucket{lookup: make(map[string]*Vary)}
	v := &Vary{lookup: make(map[string]*Item)}
	b.lookup["goku"] = v
	v.lookup["power"] = &Item{
		key:   "goku",
		vary:  "power",
		value: fakeResponse("9000"),
	}
	return b
}

func assertValue(t *testing.T, item *Item, expected string) {
	actual := item.value
	gspec.New(t).Expect(string(actual.GetBody())).ToEqual(expected)
}

func fakeResponse(body string) *caches.CachedResponse {
	return &caches.CachedResponse{
		Response: gc.Respond([]byte(body)),
	}
}
