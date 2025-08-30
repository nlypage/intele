package intele

import (
	"encoding/json"
	"fmt"
	tele "gopkg.in/telebot.v3"
	"sync"
)

type FlowBus struct {
	bot      *tele.Bot
	flows    map[string]*Flow
	storage  Storage
	sessions map[int64]*Session
	mu       sync.RWMutex
}

func NewBus(bot *tele.Bot, storage Storage) *FlowBus {
	return &FlowBus{
		bot:      bot,
		storage:  storage,
		flows:    make(map[string]*Flow),
		sessions: make(map[int64]*Session),
	}
}

func (bus *FlowBus) NewFlow(flowID string) *FlowBuilder {
	return &FlowBuilder{
		bus: bus,
		id:  flowID,
	}
}

func (bus *FlowBus) RegisterFlow(flow *Flow) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.flows[flow.id] = flow
}

func (bus *FlowBus) GetFlow(flowID string) (*Flow, bool) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	flow, exists := bus.flows[flowID]
	return flow, exists
}

func (bus *FlowBus) Handle(c tele.Context) error {
	userID := c.Sender().ID

	bus.mu.RLock()
	session, exists := bus.sessions[userID]
	bus.mu.RUnlock()

	if !exists {
		return nil
	}

	return session.Flow.handleMessage(c, session)
}

func (bus *FlowBus) Restore() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			userID := c.Sender().ID

			session, err := bus.LoadSession(userID)
			if err != nil {
				return next(c)
			}

			if session != nil {
				if session.IsExpired() {
					bus.storage.Delete(fmt.Sprintf("session:%d", userID))
					return next(c)
				}

				bus.mu.Lock()
				bus.sessions[userID] = session
				bus.mu.Unlock()
			}

			return next(c)
		}
	}
}

func (bus *FlowBus) SaveSession(userID int64, session *Session) error {
	sessionKey := fmt.Sprintf("session:%d", userID)
	jsonData, err := json.Marshal(session.Data)
	if err != nil {
		return err
	}

	bus.storage.Set(sessionKey, string(jsonData))
	return nil
}

func (bus *FlowBus) LoadSession(userID int64) (*Session, error) {
	sessionKey := fmt.Sprintf("session:%d", userID)
	jsonStr, exists := bus.storage.GetString(sessionKey)
	if !exists {
		return nil, nil
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(jsonStr), &sessionData); err != nil {
		return nil, err
	}

	flow, exists := bus.flows[sessionData.FlowID]
	if !exists {
		return nil, fmt.Errorf("flow '%s' not found", sessionData.FlowID)
	}

	return &Session{
		Data: &sessionData,
		Flow: flow,
	}, nil
}

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

func (bus *FlowBus) GetActiveSession(userID int64) (*Session, bool) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	session, exists := bus.sessions[userID]
	if !exists || session.IsCancelled() || session.IsExpired() {
		return nil, false
	}
	return session, true
}
