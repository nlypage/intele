package intele

import (
	"fmt"
	tele "gopkg.in/telebot.v3"
	"time"
)

type Flow struct {
	bus         *FlowBus
	id          string
	steps       map[string]*Step
	startStep   string
	timeout     time.Duration
	container   Container
	middlewares []Middleware
	onComplete  StepHandler
	onError     ErrorHandler
	onCancel    StepHandler
	onTimeout   StepHandler
}

func (f *Flow) ID() string {
	return f.id
}

func (f *Flow) SetTimeout(timeout time.Duration) {
	f.timeout = timeout
}

func (f *Flow) SetContainer(container Container) error {
	for _, step := range f.steps {
		for _, dep := range step.RequiredDeps {
			if _, err := container.Get(dep); err != nil {
				return &ErrMissingDependency{
					Dependency: dep,
					StepID:     step.ID(),
				}
			}
		}
	}

	f.container = container
	return nil
}

func (f *Flow) Start(c tele.Context) error {
	userID := c.Sender().ID

	f.bus.mu.Lock()
	if existingSession, exists := f.bus.sessions[userID]; exists {
		if !existingSession.IsCancelled() {
			f.bus.mu.Unlock()
			return &ErrActiveSession{
				UserID: userID,
				FlowID: existingSession.Flow.id,
			}
		}
	}
	f.bus.mu.Unlock()

	session := &Session{
		Data: &SessionData{
			UserID:      userID,
			FlowID:      f.id,
			CurrentStep: f.startStep,
			State:       make(map[string]interface{}),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Flow: f,
	}

	if f.timeout > 0 {
		expiresAt := time.Now().Add(f.timeout)
		session.Data.ExpiresAt = &expiresAt
	}

	f.bus.mu.Lock()
	f.bus.sessions[userID] = session
	f.bus.mu.Unlock()

	if err := f.bus.SaveSession(userID, session); err != nil {
		f.bus.mu.Lock()
		delete(f.bus.sessions, userID)
		f.bus.mu.Unlock()
		return fmt.Errorf("failed to save session: %w", err)
	}

	return f.executeStep(c, session)
}

func (f *Flow) handleMessage(c tele.Context, session *Session) error {
	if session.IsCancelled() {
		return nil
	}

	if session.IsExpired() {
		return f.handleTimeout(c, session)
	}

	return f.executeStep(c, session)
}

func (f *Flow) executeStep(c tele.Context, session *Session) error {
	step, exists := f.steps[session.Data.CurrentStep]
	if !exists {
		return &ErrStepNotFound{
			StepID: session.Data.CurrentStep,
			FlowID: f.id,
		}
	}

	if c.Sender().IsBot {
		return nil
	}

	ctrl := &controller{
		bus:     f.bus,
		session: session,
		teleCtx: c,
	}

	// Build handler chain with middlewares
	handler := step.handler
	for i := len(step.middlewares) - 1; i >= 0; i-- {
		handler = step.middlewares[i](handler)
	}
	for i := len(f.middlewares) - 1; i >= 0; i-- {
		handler = f.middlewares[i](handler)
	}

	if c.Callback() != nil {
		cParsed := ParseCallbackData(c)

		if step.HasCallback(cParsed.Callback().Unique) {
			// Callback belongs to current step
			if step.onComplete != nil {
				_ = cParsed.Respond()
				return step.onComplete(cParsed, ctrl)
			}
			return nil
		}

		// Callback from different step - run main handler
		return handler(c, ctrl)
	}

	// Handle text messages
	if err := handler(c, ctrl); err != nil {
		return err
	}

	if step.onComplete != nil {
		return step.onComplete(c, ctrl)
	}
	return nil
}

func (f *Flow) handleTimeout(c tele.Context, session *Session) error {
	ctrl := &controller{
		bus:     f.bus,
		session: session,
	}

	// Clean up session
	f.bus.mu.Lock()
	delete(f.bus.sessions, session.Data.UserID)
	f.bus.mu.Unlock()

	f.bus.storage.Delete(fmt.Sprintf("session:%d", session.Data.UserID))

	if f.onTimeout != nil {
		return f.onTimeout(c, ctrl)
	}

	return nil
}
