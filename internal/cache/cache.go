package cache

import (
	"sync"
	"time"
)

type Key comparable

type Cache[k Key, v any] struct {
	mx              sync.Mutex
	cleanupInterval time.Duration
	items           map[k]v
	list            *KeysList[k]
	keysCount       int
}

type Item struct {
	Value []byte
	alive time.Time
}

func NewCache[Key comparable, Value any](cleanupInterval time.Duration) *Cache[Key, Value] {
	m := make(map[Key]Value)
	c := &Cache[Key, Value]{
		cleanupInterval: cleanupInterval,
		items:           m,
	}
	c.StartGC()
	return c
}

func (c *Cache[Key, Value]) PutKey(key Key, value Value) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.items[key] = value
	if c.list == nil {
		c.list = NewList(key)
		c.keysCount++
		return
	}
	c.list.Put(key)
	c.keysCount++
}

func (c *Cache[Key, Value]) DeleteKey(key Key) {
	c.mx.Lock()
	defer c.mx.Unlock()

	delete(c.items, key)

	c.list.Delete(key)
}

func (c *Cache[Key, Value]) Get(key Key) (Value, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	value, ok := c.items[key]
	if !ok {
		return value, false
	}

	list := c.list.Delete(key)
	list.Put(key)
	return value, true
}

func (c *Cache[Key, Value]) StartGC() {
	go c.GC()
}

func (c *Cache[Key, Value]) GC() {
	for {
		time.Sleep(c.cleanupInterval)

		c.mx.Lock()
		if c.keysCount == 0 {
			continue
		}

		amountKeysToDelete := c.keysCount / 2
		for i := 0; i < amountKeysToDelete; i++ {
			c.DeleteKey(c.list.Key)
			c.list = c.list.Next
		}

		c.keysCount -= amountKeysToDelete

		c.mx.Unlock()
	}
}
