package intele

import (
	"encoding/json"
	"fmt"
	tele "gopkg.in/telebot.v3"
	"log"
	"sync"
)

// FlowBus manages multiple flows and their sessions.
// It serves as the central coordinator for message routing and session persistence.
type FlowBus struct {
	bot      *tele.Bot
	flows    map[string]*Flow
	storage  Storage
	sessions map[int64]*Session
	mu       sync.RWMutex
}

// NewBus creates a new FlowBus instance with the provided bot and storage.
func NewBus(bot *tele.Bot, storage Storage) *FlowBus {
	return &FlowBus{
		bot:      bot,
		storage:  storage,
		flows:    make(map[string]*Flow),
		sessions: make(map[int64]*Session),
	}
}

// NewFlow creates a new FlowBuilder for the specified flow ID.
func (bus *FlowBus) NewFlow(flowID string) *FlowBuilder {
	return &FlowBuilder{
		bus: bus,
		id:  flowID,
	}
}

// RegisterFlow registers a flow with the bus.
func (bus *FlowBus) RegisterFlow(flow *Flow) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.flows[flow.id] = flow
}

// GetFlow retrieves a flow by its ID.
func (bus *FlowBus) GetFlow(flowID string) (*Flow, bool) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	flow, exists := bus.flows[flowID]
	return flow, exists
}

// Handle processes incoming messages and routes them to appropriate flows.
// This should be registered as a handler for bot messages and callbacks.
func (bus *FlowBus) Handle(c tele.Context) error {
	if c.Sender().IsBot {
		return nil
	}

	userID := c.Sender().ID
	bus.mu.RLock()
	session, exists := bus.sessions[userID]
	bus.mu.RUnlock()
	if !exists {
		return nil
	}

	// Clean up expired or cancelled sessions
	if session.IsCancelled() || session.IsExpired() {
		bus.cleanupSession(userID)
		return nil
	}

	return session.Flow.handleMessage(c, session)
}

// Restore creates middleware that restores sessions from persistent storage.
// This middleware should be registered before message handlers.
func (bus *FlowBus) Restore() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Sender().IsBot {
				return next(c)
			}

			userID := c.Sender().ID
			// Skip if session already exists in memory
			bus.mu.RLock()
			_, exists := bus.sessions[userID]
			bus.mu.RUnlock()
			if exists {
				return next(c)
			}

			// Try to load session from storage
			session, err := bus.LoadSession(userID)
			if err != nil {
				log.Printf("Error loading session for user %d: %v", userID, err)
				return next(c)
			}

			if session != nil {
				if session.IsExpired() {
					bus.storage.Delete(fmt.Sprintf("session:%d", userID))
					return next(c)
				}

				// Add to active sessions if not already present
				bus.mu.Lock()
				if _, alreadyExists := bus.sessions[userID]; !alreadyExists {
					bus.sessions[userID] = session
				}
				bus.mu.Unlock()
			}

			return next(c)
		}
	}
}

// SaveSession persists a session to storage.
func (bus *FlowBus) SaveSession(userID int64, session *Session) error {
	sessionKey := fmt.Sprintf("session:%d", userID)
	jsonData, err := json.Marshal(session.Data)
	if err != nil {
		return &ErrSessionSerialization{UserID: userID, Err: err}
	}
	bus.storage.Set(sessionKey, string(jsonData))
	return nil
}

// LoadSession restores a session from storage.
func (bus *FlowBus) LoadSession(userID int64) (*Session, error) {
	sessionKey := fmt.Sprintf("session:%d", userID)
	jsonStr, exists := bus.storage.GetString(sessionKey)
	if !exists {
		return nil, nil
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(jsonStr), &sessionData); err != nil {
		return nil, &ErrSessionDeserialization{UserID: userID, Err: err}
	}

	flow, exists := bus.flows[sessionData.FlowID]
	if !exists {
		return nil, &ErrFlowNotFound{FlowID: sessionData.FlowID}
	}

	return &Session{
		Data: &sessionData,
		Flow: flow,
	}, nil
}

// CancelSession cancels and removes a user's active session.
func (bus *FlowBus) CancelSession(userID int64) error {
	bus.mu.Lock()
	session, exists := bus.sessions[userID]
	if exists {
		session.mu.Lock()
		session.cancelled = true
		session.mu.Unlock()
		delete(bus.sessions, userID)
	}
	bus.mu.Unlock()

	if exists {
		bus.storage.Delete(fmt.Sprintf("session:%d", userID))
	}
	return nil
}

// GetActiveSession retrieves an active session for a user.
func (bus *FlowBus) GetActiveSession(userID int64) (*Session, bool) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	session, exists := bus.sessions[userID]
	if !exists || session.IsCancelled() || session.IsExpired() {
		return nil, false
	}
	return session, true
}

// cleanupSession removes a session from both memory and storage.
func (bus *FlowBus) cleanupSession(userID int64) {
	bus.mu.Lock()
	delete(bus.sessions, userID)
	bus.mu.Unlock()
	bus.storage.Delete(fmt.Sprintf("session:%d", userID))
}
