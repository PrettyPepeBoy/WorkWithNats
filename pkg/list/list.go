package list

type Element[K comparable] struct {
	next, prev *Element[K]
	list       *List[K]
	Value      K
}

func (e *Element[K]) Next() *Element[K] {
	if n := e.next; e.list != nil && n != &e.list.root {
		return n
	}
	return nil
}

func (e *Element[K]) Prev() *Element[K] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

type List[K comparable] struct {
	root       Element[K]
	lastInsert *Element[K]
}

func NewList[K comparable]() *List[K] {
	return new(List[K]).init()
}

func (l *List[K]) init() *List[K] {
	l.root.prev = &l.root
	l.root.next = &l.root
	l.root.list = l
	l.lastInsert = &l.root
	return l
}

func (l *List[K]) Put(e *Element[K]) *Element[K] {
	e.list = l
	e.prev = l.lastInsert
	e.prev.next = e
	l.lastInsert = e
	return e
}

func (l *List[K]) Remove(e *Element[K]) {
	e.prev.next = e.next
	if e.next != nil {
		e.next.prev = e.prev
	}

	if l.lastInsert.Value == e.Value {
		l.lastInsert = e.prev
	}

	e.next = nil
	e.prev = nil
	e.list = nil
}

func (l *List[K]) RemoveFront(e *Element[K]) {
	e.prev = &l.root
	e.prev.next = e
}

func (l *List[K]) Front() *Element[K] {
	if l == nil {
		return nil
	}
	return l.root.next
}

func (l *List[K]) findElem(v K) *Element[K] {
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value == v {
			return e
		}
	}
	return nil
}
