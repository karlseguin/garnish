package caching

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/gspec"
	"testing"
	"time"
)

func TestTTLReturnsTheConfiguredTimeForAGoodRespone(t *testing.T) {
	spec := gspec.New(t)
	duration, ok := ttl(&garnish.Caching{TTL: time.Second * 24}, garnish.Respond(nil).Status(200))
	spec.Expect(duration).ToEqual(time.Second * 24)
	spec.Expect(ok).ToEqual(true)
}

func TestTTLReturnsTheHeaderTimeForAGoodRespone(t *testing.T) {
	spec := gspec.New(t)
	duration, ok := ttl(new(garnish.Caching), garnish.Respond(nil).Status(200).Header("Cache-Control", "max-age=33"))
	spec.Expect(duration).ToEqual(time.Second * 33)
	spec.Expect(ok).ToEqual(true)
}

func TestTTLReturnsTheHeaderTimeForAGoodRespone2(t *testing.T) {
	spec := gspec.New(t)
	duration, ok := ttl(new(garnish.Caching), garnish.Respond(nil).Status(200).Header("Cache-Control", "private,max-age=22"))
	spec.Expect(duration).ToEqual(time.Second * 22)
	spec.Expect(ok).ToEqual(true)
}

func TestTTLDoesNotHandleInvalidExpiryTimes(t *testing.T) {
	spec := gspec.New(t)
	_, ok := ttl(new(garnish.Caching), garnish.Respond(nil).Status(200).Header("Cache-Control", "private,max-age=fail"))
	spec.Expect(ok).ToEqual(false)
}

func TestTTLDoesNotHandleInvalidMissingExpiryTime(t *testing.T) {
	spec := gspec.New(t)
	_, ok := ttl(new(garnish.Caching), garnish.Respond(nil).Status(200))
	spec.Expect(ok).ToEqual(false)
}
