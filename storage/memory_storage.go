package intele

import (
	"github.com/nlypage/intele/v2"
	"sync"
)

type memoryStorage struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMemoryStorage() intele.Storage {
	return &memoryStorage{
		data: make(map[string]interface{}),
	}
}

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

func (ms *memoryStorage) GetInt(key string) (int, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	if !exists {
		return 0, false
	}

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

func (ms *memoryStorage) GetFloat64(key string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	if !exists {
		return 0, false
	}

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

func (ms *memoryStorage) Get(key string) (interface{}, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	return val, exists
}

func (ms *memoryStorage) Set(key string, value interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data[key] = value
}

func (ms *memoryStorage) Delete(key string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.data, key)
}

func (ms *memoryStorage) Has(key string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	_, exists := ms.data[key]
	return exists
}

func (ms *memoryStorage) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data = make(map[string]interface{})
}

func (ms *memoryStorage) Data() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	data := make(map[string]interface{})
	for k, v := range ms.data {
		data[k] = v
	}
	return data
}

func (ms *memoryStorage) SetData(data map[string]interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.data = make(map[string]interface{})
	for k, v := range data {
		ms.data[k] = v
	}
}
