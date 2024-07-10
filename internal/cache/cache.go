package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/PrettyPepeBoy/WorkWithNats/pkg/list"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"reflect"
	"strconv"
	"sync"
)

type Bucket[K comparable, V any] struct {
	mx        sync.Mutex
	items     map[K]Item[K, V]
	list      *list.List[K]
	cleanChan chan struct{}
	threshold int
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
	threshold := viper.GetInt("cache.elems.threshold")
	var c Cache[K, V]
	c.Buckets = make([]Bucket[K, V], bucketsAmount)
	for i := range c.Buckets {
		c.Buckets[i] = *newCacheBucket[K, V]()
		c.Buckets[i].threshold = threshold
	}

	return &c
}

func newCacheBucket[K comparable, V any]() *Bucket[K, V] {
	m := make(map[K]Item[K, V])
	c := &Bucket[K, V]{
		items:     m,
		list:      list.NewList[K](),
		cleanChan: make(chan struct{}),
	}
	c.StartClearCache()
	return c
}

func (c *Bucket[K, V]) putKey(key K, value V) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.items[key] = Item[K, V]{
		element: c.list.Put(key),
		Data:    value,
	}

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
	c.items[key] = Item[K, V]{
		element: c.list.Put(key),
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

	data := make([]Data[K, V], 0, len(c.items))
	for key := range c.items {
		data = append(data, Data[K, V]{
			Key:   key,
			Value: c.items[key].Data,
		})
	}

	rawByte, err := json.Marshal(data) //не использовать json использовать бинарный формат
	if err != nil {
		return nil, err
	}
	return rawByte, nil
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
