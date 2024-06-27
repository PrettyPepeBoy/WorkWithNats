package cache

import "fmt"

type KeysList[k Key] struct {
	Next     *KeysList[k]
	Previous *KeysList[k]
	Key      k
}

func NewList[k Key](key k) *KeysList[k] {
	return &KeysList[k]{Key: key}
}

func (l *KeysList[k]) Put(key k) *KeysList[k] {
	if l.Next == nil {
		l.Next = &KeysList[k]{Key: key,
			Previous: l}
		return l.Next
	}
	return l.Next.Put(key)
}

func (l *KeysList[k]) Delete(key k) *KeysList[k] {
	if l.Key == key {
		l.Previous.Next = l.Next
		l.Next.Previous = l.Previous
		return l.Next
	}

	return l.Next.Delete(key)
}

func (l *KeysList[k]) Find(key k) (*KeysList[k], bool) {
	if l.Key == key {
		return l, true
	}
	if l.Next != nil {
		return l.Next.Find(key)
	}
	return l, false
}

func (l *KeysList[k]) GetCountNode(count int) *KeysList[k] {
	if count == 0 {
		return l
	}
	return l.GetCountNode(count - 1)
}

func (l *KeysList[k]) PrintList() *KeysList[k] {
	fmt.Println(l.Key)
	if l.Next != nil {
		return l.Next.PrintList()
	}
	return l
}
