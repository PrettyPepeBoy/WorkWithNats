package cache

import (
	"sync"
	"time"
)

type Cache struct {
	mx              sync.Mutex
	cleanupInterval time.Duration
	items           map[int]Item
	list            *List
	keysCount       int
}

type Item struct {
	Value []byte
	alive time.Time
}

func NewCache(cleanupInterval time.Duration) *Cache {
	m := make(map[int]Item)
	c := &Cache{
		cleanupInterval: cleanupInterval,
		items:           m,
	}
	c.StartGC()
	return c
}

func (c *Cache) PutKey(key int, value []byte) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.items[key] = Item{
		Value: value,
		alive: time.Now(),
	}
	if c.list == nil {
		c.list = NewList(key)
		c.keysCount++
		return
	}
	c.list.Put(key)
	c.keysCount++
}

func (c *Cache) DeleteKey(key int) {
	c.mx.Lock()
	defer c.mx.Unlock()

	delete(c.items, key)

	c.list.Delete(key)
}

func (c *Cache) Get(key int) ([]byte, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	value, ok := c.items[key]
	if !ok {
		return nil, false
	}

	list := c.list.Delete(key)
	list.Put(key)
	return value.Value, true
}

func (c *Cache) StartGC() {
	go c.GC()
}

func (c *Cache) GC() {
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
