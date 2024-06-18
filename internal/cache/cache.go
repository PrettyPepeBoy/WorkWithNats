package cache

import (
	"TestTaskNats/internal/models"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Cache struct {
	mx              sync.Mutex
	cleanupInterval time.Duration
	items           map[int]Item
}

type Item struct {
	Value models.ProductBody
	alive time.Time
}

func InitCache(cleanupInterval time.Duration) *Cache {
	m := make(map[int]Item)
	c := &Cache{
		cleanupInterval: cleanupInterval,
		items:           m,
	}
	c.StartGC()
	return c
}

func (c *Cache) PutKey(key int, value models.ProductBody) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.items[key] = Item{
		Value: value,
		alive: time.Now(),
	}
}

func (c *Cache) DeleteKey(key int) {
	c.mx.Lock()
	defer c.mx.Unlock()
	delete(c.items, key)
}

func (c *Cache) ShowKey(key int) models.ProductBody {
	c.mx.Lock()
	defer c.mx.Unlock()
	value, ok := c.items[key]
	if !ok {
		logrus.Infof("there is no such key: %v in cacheservice", key)
		return models.ProductBody{}
	}
	logrus.Infof("value of key: %v is: %v", key, value)
	return value.Value
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
