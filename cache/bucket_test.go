package cache

import (
	. "github.com/karlseguin/expect"
	"testing"
)

type BucketTests struct{}

func Test_Bucket(t *testing.T) {
	Expectify(new(BucketTests), t)
}

func (_ BucketTests) GetMiss() {
	bucket := testBucket()
	Expect(bucket.get("invalid", "bb")).To.Equal(nil)
	Expect(bucket.get("power", "devel")).To.Equal(nil)
}

func (_ BucketTests) GetHit() {
	bucket := testBucket()
	assertEntry(bucket, "power", "level", "over 9000!")
}

func (_ BucketTests) Delete() {
	bucket := testBucket()
	bucket.delete("power", "level")
	Expect(bucket.get("power", "level")).To.Equal(nil)
	// assertEntry(bucket, "power", "rating", "high")
}

func (_ BucketTests) DeleteAll() {
	bucket := testBucket()
	bucket.deleteAll("power")
	Expect(bucket.get("power", "level")).To.Equal(nil)
	Expect(bucket.get("power", "rating")).To.Equal(nil)
}

func (_ BucketTests) SetsANewBucketItem() {
	bucket := testBucket()
	entry := buildEntry("flow")
	Expect(bucket.set("spice", "must", entry)).To.Equal(nil)
	assertEntry(bucket, "power", "level", "over 9000!")
	assertEntry(bucket, "spice", "must", "flow")
}

func (_ BucketTests) SetsAnExistingItem() {
	bucket := testBucket()
	entry := buildEntry("9002")
	existing := bucket.set("power", "level", entry)
	assertResponse(existing, "over 9000!")
	assertEntry(bucket, "power", "level", "9002")
}

func assertEntry(bucket *bucket, primary string, secondary string, expected string) {
	assertResponse(bucket.get(primary, secondary), expected)
}

func buildEntry(body string) *Entry {
	return &Entry{
		CachedResponse: buildResponse(body),
	}
}

func testBucket() *bucket {
	b := &bucket{lookup: make(map[string]map[string]*Entry)}
	b.lookup["power"] = map[string]*Entry{
		"level": &Entry{
			CachedResponse: buildEntry("over 9000!"),
			Primary:        "power",
			Secondary:      "level",
		},
		"rating": &Entry{
			CachedResponse: buildEntry("high"),
			Primary:        "power",
			Secondary:      "rating",
		},
	}
	return b
}
