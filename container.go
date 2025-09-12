package intele

import (
	"reflect"
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

// DependencyMap represents a map of dependency names to their required interface types.
// It provides a fluent interface for building dependency requirements.
type DependencyMap map[string]reflect.Type

// NewDeps creates a new empty dependency map.
func NewDeps() DependencyMap {
	return make(DependencyMap)
}

// Require adds a dependency with the specified interface type.
// The ifacePtr should be a pointer to nil interface, e.g., (*MyInterface)(nil).
// Returns the map to support method chaining.
func (dm DependencyMap) Require(depName string, ifacePtr interface{}) DependencyMap {
	dm[depName] = reflect.TypeOf(ifacePtr).Elem()
	return dm
}
