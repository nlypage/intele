package intele

import (
	"errors"
	"sync"
)

type container struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewContainer() Container {
	return &container{
		data: make(map[string]interface{}),
	}
}

func (c *container) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.data[key]
	if !exists {
		return nil, errors.New("dependency not found: " + key)
	}
	return value, nil
}

func (c *container) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}
