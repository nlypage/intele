package intele

import (
	"sync"
	"time"
)

type SessionData struct {
	UserID      int64                  `json:"user_id"`
	FlowID      string                 `json:"flow_id"`
	CurrentStep string                 `json:"current_step"`
	State       map[string]interface{} `json:"state"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

type Session struct {
	Data      *SessionData
	Flow      *Flow
	cancelled bool
	mu        sync.Mutex
}

func (s *Session) IsCancelled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cancelled
}

func (s *Session) IsExpired() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Data.ExpiresAt != nil && time.Now().After(*s.Data.ExpiresAt)
}
