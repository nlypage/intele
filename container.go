package intele

import (
	"sync"
)

// container implements the Container interface for dependency injection.
// It provides thread-safe storage and retrieval of dependencies.
type container struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewContainer creates a new dependency injection container.
func NewContainer() Container {
	return &container{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a dependency by key.
// Returns an error if the dependency is not found.
func (c *container) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	if !exists {
		return nil, &ErrDependencyNotFound{Key: key}
	}
	return value, nil
}

// Set stores a dependency with the given key.
func (c *container) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}
