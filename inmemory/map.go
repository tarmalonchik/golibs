package inmemory

import (
	"sync"
)

type InMemory[keyType comparable, valueType any] interface {
	RewriteAll(data map[keyType]valueType)
	Add(key keyType, value valueType)
	Remove(key keyType) map[keyType]valueType
	GetAll() map[keyType]valueType
	Get(key keyType) (value valueType, ok bool)
}

type inMemory[k comparable, v any] struct {
	mutex sync.Mutex
	data  map[k]v
}

func New[keyType comparable, valueType any]() InMemory[keyType, valueType] {
	return &inMemory[keyType, valueType]{
		data: make(map[keyType]valueType),
	}
}

func (c *inMemory[k, v]) RewriteAll(data map[k]v) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = data
}

func (c *inMemory[k, v]) Add(key k, value v) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
}

func (c *inMemory[k, v]) Remove(key k) map[k]v {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
	out := make(map[k]v, len(c.data))
	for mapKey, mapVal := range c.data {
		out[mapKey] = mapVal
	}
	return out
}

func (c *inMemory[k, v]) GetAll() map[k]v {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	out := make(map[k]v, len(c.data))
	for mapKey, mapVal := range c.data {
		out[mapKey] = mapVal
	}
	return out
}

func (c *inMemory[k, v]) Get(key k) (value v, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	value, ok = c.data[key]

	return value, ok
}
