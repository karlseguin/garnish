package ccache

import (
	"github.com/viki-org/gspec"
	"testing"
)

func TestItemPromotability(t *testing.T) {
	spec := gspec.New(t)
	item := &Item{promotions: 4}
	spec.Expect(item.shouldPromote(5)).ToEqual(true)
	spec.Expect(item.shouldPromote(5)).ToEqual(false)
}
