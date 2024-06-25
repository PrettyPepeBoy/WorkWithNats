package cache

import "fmt"

type List struct {
	Next     *List
	Previous *List
	Key      int
}

func NewList(key int) *List {
	return &List{Key: key}
}

func (l *List) Put(key int) *List {
	if l.Next == nil {
		l.Next = &List{Key: key,
			Previous: l}
		return l.Next
	}
	return l.Next.Put(key)
}

func (l *List) Delete(key int) *List {
	if l.Key == key {
		l.Previous.Next = l.Next
		l.Next.Previous = l.Previous
		return l.Next
	}

	return l.Next.Delete(key)
}

func (l *List) Find(key int) (*List, bool) {
	if l.Key == key {
		return l, true
	}
	if l.Next != nil {
		return l.Next.Find(key)
	}
	return l, false
}

func (l *List) GetCountNode(count int) *List {
	if count == 0 {
		return l
	}
	return l.GetCountNode(count - 1)
}

func (l *List) PrintList() *List {
	fmt.Println(l.Key)
	if l.Next != nil {
		return l.Next.PrintList()
	}
	return l
}
