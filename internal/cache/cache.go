package cache

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/PrettyPepeBoy/WorkWithNats/pkg/list"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"hash"
	"hash/fnv"
	"io"
	"sync"
	"unsafe"
)

var (
	threshold int
	remains   int
)

type Marshaller interface {
	Marshal() ([]byte, error)
}

type Key interface {
	Marshaller
	comparable
}

type bucket[K Key, V Marshaller] struct {
	mx                sync.Mutex
	items             map[K]Item[K, V]
	list              *list.List[K]
	elements          []list.Element[K]
	cleanChan         chan struct{}
	threshold         int
	index             int
	remainsAfterClear int
}

type Item[K Key, V Marshaller] struct {
	element *list.Element[K]
	Data    V
	index   int
}

type Cache[K Key, V Marshaller] struct {
	buckets       []bucket[K, V]
	bucketsAmount int
	hash          hasher[K]
}

func NewCache[K Key, V Marshaller]() *Cache[K, V] {
	threshold = viper.GetInt("cache.elems.threshold")

	var c Cache[K, V]
	c.bucketsAmount = viper.GetInt("cache.buckets-amount")
	c.buckets = make([]bucket[K, V], c.bucketsAmount)
	for i := range c.buckets {
		c.buckets[i] = *newCacheBucket[K, V]()
	}

	c.hash = *newHasher[K]()

	return &c
}

func newCacheBucket[K Key, V Marshaller]() *bucket[K, V] {
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

	c.elements[c.index] = list.Element[K]{
		Value: key,
	}

	c.items[key] = Item[K, V]{
		element: c.list.Put(&c.elements[c.index]),
		Data:    value,
		index:   c.index,
	}
	c.index++

	if c.index == c.threshold {
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

func (c *bucket[K, V]) get(key K) (Marshaller, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.list.Remove(item.element)

	c.elements[item.index] = list.Element[K]{
		Value: key,
	}

	c.items[key] = Item[K, V]{
		element: c.list.Put(&c.elements[item.index]),
		Data:    item.Data,
		index:   item.index,
	}

	return item.Data, true
}

func (c *Cache[K, V]) Get(key K) (Marshaller, bool) {
	i := int(c.hash.getHash(key)) % c.bucketsAmount
	return c.buckets[i].get(key)
}

type Data[K Key, V Marshaller] struct {
	Key   K
	Value V
}

func (c *bucket[K, V]) dump() func(w *bufio.Writer) {
	return func(w *bufio.Writer) {
		c.mx.Lock()
		defer c.mx.Unlock()

		var buf bytes.Buffer
		enc := newEncoder[K, V](&buf)

		for key := range c.items {
			data := Data[K, V]{
				Key:   key,
				Value: c.items[key].Data,
			}

			err := enc.encode(data)
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

		if c.index == c.threshold {
			c.index = c.remainsAfterClear
		}

		c.list.RemoveFront(e.Prev())
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

type encoder[K Key, V Marshaller] struct {
	writer io.Writer
}

func newEncoder[K Key, V Marshaller](writer io.Writer) *encoder[K, V] {
	return &encoder[K, V]{writer: writer}
}

func (enc *encoder[K, V]) encode(item Data[K, V]) error {
	rawByteValue, err := item.Value.Marshal()
	if err != nil {
		return err
	}

	_, err = enc.writer.Write(rawByteValue)
	if err != nil {
		return err
	}

	rawByteKey, err := item.Key.Marshal()
	if err != nil {
		panic(err)
	}

	_, err = enc.writer.Write(rawByteKey)
	if err != nil {
		return err
	}

	return nil
}

type Int int

func (i Int) Marshal() ([]byte, error) {
	b := make([]byte, 0, 8)
	binary.BigEndian.AppendUint64(b, uint64(i))
	return b, nil
}

type ByteSlc []byte

func (slc ByteSlc) Marshal() ([]byte, error) {
	b := make(ByteSlc, 0, 256)
	b = append(b, slc...)
	return b, nil
}

type Uint32 uint32

func (ui Uint32) Marshal() ([]byte, error) {
	b := make([]byte, 0, 4)
	binary.BigEndian.AppendUint32(b, uint32(ui))
	return b, nil
}

type Uint64 uint64

func (ui Uint64) Marshal() ([]byte, error) {
	b := make([]byte, 0, 8)
	binary.BigEndian.AppendUint64(b, uint64(ui))
	return b, nil
}
