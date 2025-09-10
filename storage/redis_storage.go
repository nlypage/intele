package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/nlypage/intele/v2"
	"github.com/redis/go-redis/v9"
)

// redisStorage implements the intele.Storage interface using Redis as the backend.
// This implementation provides persistent storage with optional TTL support.
type redisStorage struct {
	client redis.Cmdable
	ctx    context.Context
	prefix string
	ttl    time.Duration
}

// RedisStorageOptions contains configuration options for Redis storage.
type RedisStorageOptions struct {
	// Prefix is added to all keys to avoid collisions (optional, defaults to "intele:")
	Prefix string
	// TTL sets expiration time for all keys (optional)
	TTL time.Duration
	// Context for Redis operations (optional, defaults to context.Background())
	Context context.Context
}

// NewRedisStorage creates a new Redis storage instance.
func NewRedisStorage(client redis.Cmdable, opts *RedisStorageOptions) intele.Storage {
	if opts == nil {
		opts = &RedisStorageOptions{}
	}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	prefix := opts.Prefix
	if prefix == "" {
		prefix = "intele:"
	}

	return &redisStorage{
		client: client,
		ctx:    ctx,
		prefix: prefix,
		ttl:    opts.TTL,
	}
}

// key returns the full Redis key with prefix
func (rs *redisStorage) key(k string) string {
	return rs.prefix + k
}

// GetString retrieves a string value by key.
func (rs *redisStorage) GetString(key string) (string, bool) {
	val, err := rs.client.Get(rs.ctx, rs.key(key)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return val, true
}

// GetInt retrieves an integer value by key.
func (rs *redisStorage) GetInt(key string) (int, bool) {
	val, exists := rs.GetString(key)
	if !exists {
		return 0, false
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return intVal, true
}

// GetBool retrieves a boolean value by key.
func (rs *redisStorage) GetBool(key string) (bool, bool) {
	val, exists := rs.GetString(key)
	if !exists {
		return false, false
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return false, false
	}
	return boolVal, true
}

// GetFloat64 retrieves a float64 value by key.
func (rs *redisStorage) GetFloat64(key string) (float64, bool) {
	val, exists := rs.GetString(key)
	if !exists {
		return 0, false
	}

	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, false
	}
	return floatVal, true
}

// Get retrieves any value by key and unmarshals it from JSON.
func (rs *redisStorage) Get(key string) (interface{}, bool) {
	val, err := rs.client.Get(rs.ctx, rs.key(key)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// If JSON unmarshal fails, return as string
		return val, true
	}
	return result, true
}

// Set stores a value with the given key.
func (rs *redisStorage) Set(key string, value interface{}) {
	var val string

	switch v := value.(type) {
	case string:
		val = v
	case int:
		val = strconv.Itoa(v)
	case int64:
		val = strconv.FormatInt(v, 10)
	case float64:
		val = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		val = strconv.FormatBool(v)
	default:
		// For complex types, marshal to JSON
		if jsonData, err := json.Marshal(value); err == nil {
			val = string(jsonData)
		} else {
			val = fmt.Sprintf("%v", value)
		}
	}

	if rs.ttl > 0 {
		rs.client.Set(rs.ctx, rs.key(key), val, rs.ttl)
	} else {
		rs.client.Set(rs.ctx, rs.key(key), val, 0)
	}
}

// Delete removes a value by key.
func (rs *redisStorage) Delete(key string) {
	rs.client.Del(rs.ctx, rs.key(key))
}

// Has checks if a key exists in storage.
func (rs *redisStorage) Has(key string) bool {
	count, err := rs.client.Exists(rs.ctx, rs.key(key)).Result()
	return err == nil && count > 0
}

// Clear removes all data from storage by deleting keys with the configured prefix.
func (rs *redisStorage) Clear() {
	pattern := rs.prefix + "*"
	keys, err := rs.client.Keys(rs.ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return
	}

	rs.client.Del(rs.ctx, keys...)
}

// Data returns a copy of all stored data.
func (rs *redisStorage) Data() map[string]interface{} {
	pattern := rs.prefix + "*"
	keys, err := rs.client.Keys(rs.ctx, pattern).Result()
	if err != nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	for _, fullKey := range keys {
		// Remove prefix to get original key
		originalKey := fullKey[len(rs.prefix):]

		val, err := rs.client.Get(rs.ctx, fullKey).Result()
		if err != nil {
			continue
		}

		// Try to unmarshal as JSON, fallback to string
		var value interface{}
		if err := json.Unmarshal([]byte(val), &value); err != nil {
			value = val
		}

		result[originalKey] = value
	}

	return result
}

// SetData replaces all stored data with the provided map.
func (rs *redisStorage) SetData(data map[string]interface{}) {
	// Clear existing data first
	rs.Clear()

	// Set new data
	for key, value := range data {
		rs.Set(key, value)
	}
}

// SetTTL updates the TTL for a specific key.
func (rs *redisStorage) SetTTL(key string, ttl time.Duration) error {
	return rs.client.Expire(rs.ctx, rs.key(key), ttl).Err()
}

// GetTTL returns the remaining TTL for a key.
func (rs *redisStorage) GetTTL(key string) (time.Duration, error) {
	return rs.client.TTL(rs.ctx, rs.key(key)).Result()
}
