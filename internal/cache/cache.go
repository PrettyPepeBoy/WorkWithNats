package cache

import (
	"github.com/PrettyPepeBoy/WorkWithNats/pkg/list"
	"github.com/spf13/viper"
	"sync"
)

type Cache[K comparable, V any] struct {
	mx        sync.Mutex
	items     map[K]Item[K, V]
	list      *list.List[K]
	elemChan  chan struct{}
	threshold int
}

type Item[K comparable, V any] struct {
	element *list.Element[K]
	Data    V
}

func NewCache[K comparable, V any]() []Cache[K, V] {
	bucketsAmount := viper.GetInt("cache.buckets_amount")
	threshold := viper.GetInt("cache.threshold")

	c := make([]Cache[K, V], bucketsAmount)
	for i := range c {
		c[i] = *newCacheBucket[K, V]()
		c[i].threshold = threshold
	}

	return c
}

func newCacheBucket[K comparable, V any]() *Cache[K, V] {
	m := make(map[K]Item[K, V])
	c := &Cache[K, V]{
		items:    m,
		list:     list.NewList[K](),
		elemChan: make(chan struct{}),
	}
	c.StartClearCache()
	return c
}

func (c *Cache[K, V]) PutKey(key K, value V) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.items[key] = Item[K, V]{
		element: c.list.Put(key),
		Data:    value,
	}

	if c.list.Len == c.threshold {
		c.elemChan <- struct{}{}
	}
}

func (c *Cache[K, V]) RemoveKey(key K) bool {
	c.mx.Lock()
	defer c.mx.Unlock()
	item, ok := c.items[key]
	if !ok {
		return false
	}

	delete(c.items, key)
	c.list.Remove(item.element)
	return true
}

func (c *Cache[K, V]) Get(key K) (any, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.list.Remove(item.element)
	c.list.Put(key)
	return item.Data, true
}

func (c *Cache[K, V]) GetAllKeys() []K {
	c.mx.Lock()
	defer c.mx.Unlock()
	keys := make([]K, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

func (c *Cache[K, V]) StartClearCache() {
	go c.clearCache()
}

func (c *Cache[K, V]) clean(e *list.Element[K], count int) *list.Element[K] {
	delete(c.items, e.Value)
	count++
	return e.Next()
}

func (c *Cache[K, V]) clearCache() {
	for {
		<-c.elemChan
		c.mx.Lock()
		var count int
		e := c.list.Front()
		for count > c.threshold/2 {
			e = c.clean(e, count)
		}
		c.list.Remove(e.Prev())
		c.mx.Unlock()
	}
}
