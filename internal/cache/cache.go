package cache

import (
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Cache struct {
	mx              sync.Mutex
	cleanupInterval time.Duration
	items           map[int]Item
	list            List
}

type Item struct {
	Value []byte
	alive time.Time
}

func NewCache(cleanupInterval time.Duration) *Cache {
	m := make(map[int]Item)
	list := NewList()
	c := &Cache{
		cleanupInterval: cleanupInterval,
		items:           m,
		list:            list,
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
	list, found := c.list.FindInList(key)
	if !found {
		list.PutInList(key)
		return
	}
	list.DeleteFromList(key)
	list.PutInList(key)
}

func (c *Cache) DeleteKey(key int) {
	c.mx.Lock()
	defer c.mx.Unlock()
	delete(c.items, key)
	c.list.DeleteFromList(key)
}

func (c *Cache) ShowKey(key int) ([]byte, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	value, ok := c.items[key]
	if !ok {
		logrus.Infof("there is no such key: %v in cacheservice", key)
		return nil, false
	}
	list := c.list.DeleteFromList(key)
	list.PutInList(key)
	logrus.Infof("value of key: %v is: %v", key, string(value.Value))
	return value.Value, true
}

func (c *Cache) StartGC() {
	go c.GC()
}

func (c *Cache) GC() {
	for {
		time.Sleep(c.cleanupInterval)
		for key, value := range c.items {
			logrus.Infof("check value to delete, key: %v, time lost: %v", key, time.Since(value.alive))
			if time.Since(value.alive) > c.cleanupInterval {
				c.DeleteKey(key)
			}
		}
	}
}
