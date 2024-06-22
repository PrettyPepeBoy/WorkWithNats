package cache

import "fmt"

type List struct {
	Next     *List
	Previous *List
	Key      int
}

func InitList() List {
	return List{}
}

func (l *List) PutInList(key int) *List {
	if l.Next == nil {
		l.Next = &List{Key: key,
			Previous: l}
		return l.Next
	}
	return l.Next.PutInList(key)
}

func (l *List) DeleteFromList(key int) *List {
	if l.Next == nil {
		return l
	}
	if l.Next.Key == key {
		l.Next = l.Next.Next
		if l.Next != nil {
			l.Next.Previous = l
		}
		return l
	}
	return l.Next.DeleteFromList(key)
}

func (l *List) FindInList(key int) (*List, bool) {
	if l.Next == nil {
		return l, false
	}
	if l.Next.Key == key {
		return l.Next, true
	}
	return l.Next.FindInList(key)
}

func (l *List) PrintList() *List {
	fmt.Println(l.Key)
	if l.Next != nil {
		return l.Next.PrintList()
	}
	return l
}
