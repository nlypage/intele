package intele

import (
	"sync"
	"time"
)

// SessionData contains the persistent data for a user session.
// This struct is serialized to JSON for storage persistence.
type SessionData struct {
	UserID      int64                  `json:"user_id"`
	FlowID      string                 `json:"flow_id"`
	CurrentStep string                 `json:"current_step"`
	State       map[string]interface{} `json:"state"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// Session represents an active user session within a flow.
// It combines persistent data with runtime state management.
type Session struct {
	Data      *SessionData
	Flow      *Flow
	cancelled bool
	mu        sync.RWMutex
}

// IsCancelled returns true if the session has been cancelled.
func (s *Session) IsCancelled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cancelled
}

// IsExpired returns true if the session has exceeded its timeout.
func (s *Session) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Data.ExpiresAt != nil && time.Now().After(*s.Data.ExpiresAt)
}

// Cancel marks the session as cancelled.
func (s *Session) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelled = true
}

// GetCurrentStep returns the current step ID in a thread-safe manner.
func (s *Session) GetCurrentStep() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Data.CurrentStep
}

// SetCurrentStep updates the current step ID and timestamp.
func (s *Session) SetCurrentStep(stepID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data.CurrentStep = stepID
	s.Data.UpdatedAt = time.Now()
}

// UpdateTimestamp updates the session's last activity timestamp.
func (s *Session) UpdateTimestamp() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data.UpdatedAt = time.Now()
}
