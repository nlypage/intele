package intele

import "time"

// sessionStorage implements the Storage interface for session-specific data.
// It provides thread-safe access to session state with automatic persistence.
type sessionStorage struct {
	session *Session
	bus     *FlowBus
}

// Set stores a value in the session state.
func (ss *sessionStorage) Set(key string, value interface{}) {
	ss.session.mu.Lock()
	defer ss.session.mu.Unlock()

	if ss.session.Data.State == nil {
		ss.session.Data.State = make(map[string]interface{})
	}

	ss.session.Data.State[key] = value
	ss.session.Data.UpdatedAt = time.Now()
}

// GetString retrieves a string value from the session state.
func (ss *sessionStorage) GetString(key string) (string, bool) {
	ss.session.mu.RLock()
	defer ss.session.mu.RUnlock()

	if ss.session.Data.State == nil {
		return "", false
	}

	val, exists := ss.session.Data.State[key]
	if !exists {
		return "", false
	}

	str, ok := val.(string)
	return str, ok
}

// GetInt retrieves an integer value from the session state.
func (ss *sessionStorage) GetInt(key string) (int, bool) {
	ss.session.mu.RLock()
	defer ss.session.mu.RUnlock()

	if ss.session.Data.State == nil {
		return 0, false
	}

	val, exists := ss.session.Data.State[key]
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

// GetBool retrieves a boolean value from the session state.
func (ss *sessionStorage) GetBool(key string) (bool, bool) {
	ss.session.mu.RLock()
	defer ss.session.mu.RUnlock()

	if ss.session.Data.State == nil {
		return false, false
	}

	val, exists := ss.session.Data.State[key]
	if !exists {
		return false, false
	}

	b, ok := val.(bool)
	return b, ok
}

// GetFloat64 retrieves a float64 value from the session state.
func (ss *sessionStorage) GetFloat64(key string) (float64, bool) {
	ss.session.mu.RLock()
	defer ss.session.mu.RUnlock()

	if ss.session.Data.State == nil {
		return 0, false
	}

	val, exists := ss.session.Data.State[key]
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

// Get retrieves any value from the session state.
func (ss *sessionStorage) Get(key string) (interface{}, bool) {
	ss.session.mu.RLock()
	defer ss.session.mu.RUnlock()

	if ss.session.Data.State == nil {
		return nil, false
	}

	val, exists := ss.session.Data.State[key]
	return val, exists
}

// Delete removes a value from the session state.
func (ss *sessionStorage) Delete(key string) {
	ss.session.mu.Lock()
	defer ss.session.mu.Unlock()

	if ss.session.Data.State == nil {
		return
	}

	delete(ss.session.Data.State, key)
	ss.session.Data.UpdatedAt = time.Now()
}

// Has checks if a key exists in the session state.
func (ss *sessionStorage) Has(key string) bool {
	ss.session.mu.RLock()
	defer ss.session.mu.RUnlock()

	if ss.session.Data.State == nil {
		return false
	}

	_, exists := ss.session.Data.State[key]
	return exists
}

// Clear removes all data from the session state.
func (ss *sessionStorage) Clear() {
	ss.session.mu.Lock()
	defer ss.session.mu.Unlock()

	ss.session.Data.State = make(map[string]interface{})
	ss.session.Data.UpdatedAt = time.Now()
}

// Data returns a copy of all session state data.
func (ss *sessionStorage) Data() map[string]interface{} {
	ss.session.mu.RLock()
	defer ss.session.mu.RUnlock()

	if ss.session.Data.State == nil {
		return make(map[string]interface{})
	}

	data := make(map[string]interface{})
	for k, v := range ss.session.Data.State {
		data[k] = v
	}

	return data
}

// SetData replaces all session state data.
func (ss *sessionStorage) SetData(data map[string]interface{}) {
	ss.session.mu.Lock()
	defer ss.session.mu.Unlock()

	ss.session.Data.State = make(map[string]interface{})
	for k, v := range data {
		ss.session.Data.State[k] = v
	}

	ss.session.Data.UpdatedAt = time.Now()
}
