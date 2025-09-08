// Package storage provides implementations of the Storage interface for persistent data storage.
package storage

import (
	"sync"

	"github.com/nlypage/intele/v2"
)

// memoryStorage implements the intele.Storage interface using in-memory storage.
// This implementation is suitable for development and testing, but data will be lost
// when the application restarts. For production use, consider a persistent storage solution.
type memoryStorage struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage() intele.Storage {
	return &memoryStorage{
		data: make(map[string]interface{}),
	}
}

// GetString retrieves a string value by key.
func (ms *memoryStorage) GetString(key string) (string, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	if !exists {
		return "", false
	}

	str, ok := val.(string)
	return str, ok
}

// GetInt retrieves an integer value by key.
func (ms *memoryStorage) GetInt(key string) (int, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	if !exists {
		return 0, false
	}

	// Handle different numeric types
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetBool retrieves a boolean value by key.
func (ms *memoryStorage) GetBool(key string) (bool, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	if !exists {
		return false, false
	}

	b, ok := val.(bool)
	return b, ok
}

// GetFloat64 retrieves a float64 value by key.
func (ms *memoryStorage) GetFloat64(key string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	if !exists {
		return 0, false
	}

	// Handle different numeric types
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// Get retrieves any value by key.
func (ms *memoryStorage) Get(key string) (interface{}, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	return val, exists
}

// Set stores a value with the given key.
func (ms *memoryStorage) Set(key string, value interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data[key] = value
}

// Delete removes a value by key.
func (ms *memoryStorage) Delete(key string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.data, key)
}

// Has checks if a key exists in storage.
func (ms *memoryStorage) Has(key string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	_, exists := ms.data[key]
	return exists
}

// Clear removes all data from storage.
func (ms *memoryStorage) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data = make(map[string]interface{})
}

// Data returns a copy of all stored data.
func (ms *memoryStorage) Data() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	data := make(map[string]interface{})
	for k, v := range ms.data {
		data[k] = v
	}

	return data
}

// SetData replaces all stored data with the provided map.
func (ms *memoryStorage) SetData(data map[string]interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.data = make(map[string]interface{})
	for k, v := range data {
		ms.data[k] = v
	}
}
