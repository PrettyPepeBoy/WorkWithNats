package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"github.com/PrettyPepeBoy/WorkWithNats/pkg/list"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"reflect"
	"strconv"
	"sync"
)

type Bucket[K comparable, V any] struct {
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
	Buckets []Bucket[K, V]
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	bucketsAmount := viper.GetInt("cache.buckets_amount")
	var c Cache[K, V]
	c.Buckets = make([]Bucket[K, V], bucketsAmount)
	for i := range c.Buckets {
		c.Buckets[i] = *newCacheBucket[K, V]()
	}

	return &c
}

func newCacheBucket[K comparable, V any]() *Bucket[K, V] {
	threshold := viper.GetInt("cache.elems.threshold")
	m := make(map[K]Item[K, V])
	c := &Bucket[K, V]{
		items:     m,
		list:      list.NewList[K](),
		cleanChan: make(chan struct{}),
		elements:  make([]list.Element[K], threshold),
		threshold: threshold,
	}
	c.StartClearCache()
	return c
}

func (c *Bucket[K, V]) putKey(key K, value V) {
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
	i := hash(key)
	if i == -1 {
		return
	}
	c.Buckets[i].putKey(key, value)
}

func (c *Bucket[K, V]) removeKey(key K) bool {
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

func (c *Bucket[K, V]) get(key K) (any, bool) {
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
	i := hash(key)
	if i == -1 {
		return nil, false
	}

	return c.Buckets[i].get(key)
}

type Data[K comparable, V any] struct {
	Key   K
	Value V
}

func (c *Bucket[K, V]) GetAllRawData() ([]byte, error) {
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
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (c *Bucket[K, V]) StartClearCache() {
	go c.clearCache()
}

func (c *Bucket[K, V]) clearCache() {
	remains := viper.GetInt("cache.elems.remains_after_clean")
	for {
		<-c.cleanChan
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

func hash[K comparable](key K) int {
	h := sha256.New()
	_, err := h.Write([]byte(strconv.Itoa(int(reflect.ValueOf(key).Int()))))
	if err != nil {
		logrus.Fatalf("failed to hash key, error: %v", err)
		return -1
	}

	bucketsAmount := viper.GetInt("cache.buckets_amount")
	str := hex.EncodeToString(h.Sum(nil))
	var count int32

	for _, elem := range str {
		count += elem
	}
	h.Reset()
	return int(count) % bucketsAmount
}
