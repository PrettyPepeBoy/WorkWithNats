package cache

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"github.com/PrettyPepeBoy/WorkWithNats/pkg/list"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"hash"
	"hash/fnv"
	"sync"
	"unsafe"
)

var (
	threshold int
	remains   int
)

type bucket[K comparable, V any] struct {
	mx           sync.Mutex
	items        map[K]Item[K, V]
	list         *list.List[K]
	elements     []list.Element[K]
	cleanChan    chan struct{}
	threshold    int
	elementIndex int
}

type Item[K comparable, V any] struct {
	element *list.Element[K]
	Data    V
}

type Cache[K comparable, V any] struct {
	buckets       []bucket[K, V]
	bucketsAmount int
	hash          hasher[K]
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	threshold = viper.GetInt("cache.elems.threshold")

	var c Cache[K, V]
	c.bucketsAmount = viper.GetInt("cache.buckets_amount")
	c.buckets = make([]bucket[K, V], c.bucketsAmount)
	for i := range c.buckets {
		c.buckets[i] = *newCacheBucket[K, V]()
	}

	c.hash = *newHasher[K]()

	return &c
}

func newCacheBucket[K comparable, V any]() *bucket[K, V] {
	m := make(map[K]Item[K, V])

	c := &bucket[K, V]{
		items:     m,
		list:      list.NewList[K](),
		cleanChan: make(chan struct{}),
		elements:  make([]list.Element[K], threshold),
		threshold: threshold,
	}
	c.startClearCache()
	return c
}

func (c *bucket[K, V]) putKey(key K, value V) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.elements[c.elementIndex%c.threshold] = list.Element[K]{
		Value: key,
	}

	c.items[key] = Item[K, V]{
		element: c.list.Put(&c.elements[c.elementIndex%c.threshold]),
		Data:    value,
	}
	c.elementIndex++

	if c.list.CheckLength() == c.threshold {
		c.cleanChan <- struct{}{}
	}
}

func (c *Cache[K, V]) PutKey(key K, value V) {
	i := int(c.hash.getHash(key)) % c.bucketsAmount
	c.buckets[i].putKey(key, value)
}

func (c *bucket[K, V]) removeKey(key K) bool {
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

func (c *bucket[K, V]) get(key K) (any, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.list.Remove(item.element)

	c.elements[c.elementIndex%c.threshold] = list.Element[K]{
		Value: key,
	}

	c.items[key] = Item[K, V]{
		element: c.list.Put(&c.elements[c.elementIndex%c.threshold]),
		Data:    c.items[key].Data,
	}
	return item.Data, true
}

func (c *Cache[K, V]) Get(key K) (any, bool) {
	i := int(c.hash.getHash(key)) % c.bucketsAmount
	return c.buckets[i].get(key)
}

type Data[K comparable, V any] struct {
	Key   K
	Value V
}

func (c *bucket[K, V]) dump() func(w *bufio.Writer) {
	return func(w *bufio.Writer) {
		c.mx.Lock()
		defer c.mx.Unlock()

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		for key := range c.items {
			data := Data[K, V]{
				Key:   key,
				Value: c.items[key].Data,
			}

			err := encoder.Encode(data)
			if err != nil {
				logrus.Errorf("failed to encode data, error: %v", err)
				panic(err)
			}

			_, err = w.Write(buf.Bytes())
			if err != nil {
				logrus.Errorf("failed to write data to buffer, error: %v", err)
				panic(err)
			}
		}
	}
}

func (c *Cache[K, V]) GetAllRawData(bufWriter *bufio.Writer) {
	for i := 0; i < c.bucketsAmount; i++ {
		c.buckets[i].dump()(bufWriter)
	}
}

func (c *bucket[K, V]) startClearCache() {
	go c.clearCache()
}

func (c *bucket[K, V]) clearCache() {
	remains = viper.GetInt("cache.elems.remains_after_clean")
	for range c.cleanChan {
		c.mx.Lock()
		e := c.list.Front()
		for count := 0; count < remains; count++ {
			delete(c.items, e.Value)
			e = e.Next()
		}
		c.list.RemoveFront(e.Prev(), remains)
		c.mx.Unlock()
	}
}

type hasher[K comparable] struct {
	size uintptr
	hash hash.Hash32
}

func newHasher[K comparable]() *hasher[K] {
	var tmp K
	return &hasher[K]{
		size: unsafe.Sizeof(tmp),
		hash: fnv.New32(),
	}
}

func (h *hasher[K]) getHash(key K) uint32 {
	ptr := (*byte)(unsafe.Pointer(&key))
	data := unsafe.Slice(ptr, h.size)

	_, err := h.hash.Write(data)
	if err != nil {
		panic(err)
	}

	return h.hash.Sum32()
}
