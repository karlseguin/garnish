package cache

import (
	. "github.com/karlseguin/expect"
	"testing"
)

type ListTests struct{}

func Test_List(t *testing.T) {
	Expectify(new(ListTests), t)
}

func (_ ListTests) PushesNewItemToFront() {
	list := NewList()
	cr1, cr2, cr3 := buildListEntry("2"), buildListEntry("1"), buildListEntry("3")
	list.PushToFront(cr2)
	list.PushToFront(cr3)
	list.PushToFront(cr1)
	assertList(list, cr1, cr3, cr2)
}

func (_ ListTests) PushesExistingItems() {
	list := NewList()
	cr1, cr2, cr3 := buildListEntry("2"), buildListEntry("1"), buildListEntry("3")
	list.PushToFront(cr2)
	list.PushToFront(cr2)
	assertList(list, cr2)
	list.PushToFront(cr3)
	assertList(list, cr3, cr2)
	list.PushToFront(cr3)
	assertList(list, cr3, cr2)
	list.PushToFront(cr1)
	assertList(list, cr1, cr3, cr2)
	list.PushToFront(cr2)
	assertList(list, cr2, cr1, cr3)
}

func (_ ListTests) RemovesItemFromTheList() {
	list := NewList()
	cr1, cr2, cr3, cr4, cr5 := buildListEntry("2"), buildListEntry("1"), buildListEntry("3"), buildListEntry("4"), buildListEntry("5")
	list.PushToFront(cr2)
	list.PushToFront(cr3)
	list.PushToFront(cr5)
	list.PushToFront(cr1)
	list.PushToFront(cr4)
	list.Remove(cr5)
	assertList(list, cr4, cr1, cr3, cr2)
	list.Remove(cr4)
	assertList(list, cr1, cr3, cr2)
	list.Remove(cr2)
	assertList(list, cr1, cr3)
	list.Remove(cr1)
	assertList(list, cr3)
	list.Remove(cr3)
	assertList(list)
}

func (_ ListTests) NoopOnRemovingNoExistingItem() {
	list := NewList()
	list.Remove(buildListEntry("x"))
	assertList(list)
}

func buildListEntry(primary string) *Entry {
	return &Entry{Primary: primary}
}

func assertList(list *List, items ...*Entry) {
	Expect(list.head.Primary).To.Equal("")
	Expect(list.tail.Primary).To.Equal("_TAIL_")
	node := list.head
	for _, item := range items {
		node = node.next
		Expect(item).To.Equal(node)
	}
	Expect(node.next.Primary).To.Equal("_TAIL_")
}
