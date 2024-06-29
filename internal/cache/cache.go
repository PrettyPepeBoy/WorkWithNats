package cache

import (
	"sync"
)

type Cache[K comparable, V any] struct {
	mx       sync.Mutex
	items    map[K]V
	list     *List[K]
	elemChan chan Element[K]
}

func NewCache[K comparable, V any](listsThreshold int) *Cache[K, V] {
	m := make(map[K]V)
	c := &Cache[K, V]{
		items:    m,
		list:     NewList[K](listsThreshold),
		elemChan: make(chan Element[K]),
	}
	c.StartClearCache()
	return c
}

func (c *Cache[K, V]) PutKey(key K, value V) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.items[key] = value
	c.list.Put(key)
	if c.list.len == c.list.threshold {
		c.elemChan <- c.list.mediterranean
	}
}

func (c *Cache[K, V]) RemoveKey(key K) {
	c.mx.Lock()
	defer c.mx.Unlock()
	delete(c.items, key)
	c.list.Remove(key)
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	value, ok := c.items[key]
	if !ok {
		return value, false
	}

	c.list.Remove(key)
	c.list.Put(key)
	return value, true
}

func (c *Cache[K, V]) GetAllKeys() ([]K, int) {
	keys := make([]K, c.list.len)
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys, c.list.len
}

func (c *Cache[K, V]) StartClearCache() {
	go c.clearCache()
}

func (c *Cache[K, V]) clearCache() {
	for {
		<-c.elemChan
		for e := c.list.front(); e != nil; e.Next() {
			c.RemoveKey(e.Value)
		}
		c.list.PushFront()
		c.list.len = c.list.threshold / 2
	}
}
