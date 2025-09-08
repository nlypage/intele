package intele

import (
	"fmt"
	tele "gopkg.in/telebot.v3"
	"time"
)

// Flow represents a complete conversation flow with multiple steps.
// It manages the execution of steps, middleware, and session state.
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

// ID returns the unique identifier of this flow.
func (f *Flow) ID() string {
	return f.id
}

// SetTimeout updates the session timeout for this flow.
func (f *Flow) SetTimeout(timeout time.Duration) {
	f.timeout = timeout
}

// SetContainer updates the dependency injection container and validates dependencies.
func (f *Flow) SetContainer(container Container) error {
	// Validate that all required dependencies are available
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

// Start initiates a new flow session for a user.
// Creates a new session, saves it to storage, and executes the start step.
func (f *Flow) Start(c tele.Context) error {
	if c.Sender().IsBot {
		return nil
	}

	userID := c.Sender().ID
	// Check for existing active session
	f.bus.mu.Lock()
	if _, exists := f.bus.sessions[userID]; exists {
		// Clean up previous session
		delete(f.bus.sessions, userID)
	}
	f.bus.mu.Unlock()

	// Create new session
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

	// Set expiration if timeout is configured
	if f.timeout > 0 {
		expiresAt := time.Now().Add(f.timeout)
		session.Data.ExpiresAt = &expiresAt
	}

	// Save to storage first
	if err := f.bus.SaveSession(userID, session); err != nil {
		return &ErrSessionSave{UserID: userID, Err: err}
	}

	// Add to active sessions
	f.bus.mu.Lock()
	f.bus.sessions[userID] = session
	f.bus.mu.Unlock()

	// Execute start step
	return f.executeStep(c, session, true)
}

// handleMessage processes incoming messages for active sessions.
func (f *Flow) handleMessage(c tele.Context, session *Session) error {
	// Validate session state
	if session.IsCancelled() {
		return nil
	}

	if session.IsExpired() {
		return f.handleTimeout(c, session)
	}

	// Execute current step without showing handler (process user input)
	return f.executeStep(c, session, false)
}

// executeStep runs a step with its middleware chain and handles callbacks/input.
func (f *Flow) executeStep(c tele.Context, session *Session, showHandler bool) error {
	if session.IsCancelled() || session.IsExpired() {
		return nil
	}

	step, exists := f.steps[session.Data.CurrentStep]
	if !exists {
		return &ErrStepNotFound{
			StepID: session.Data.CurrentStep,
			FlowID: f.id,
		}
	}

	ctrl := &controller{
		bus:     f.bus,
		session: session,
		teleCtx: c,
	}

	// Build middleware chain
	handler := step.handler
	for i := len(step.middlewares) - 1; i >= 0; i-- {
		handler = step.middlewares[i](handler)
	}
	for i := len(f.middlewares) - 1; i >= 0; i-- {
		handler = f.middlewares[i](handler)
	}

	var err error
	if showHandler {
		// Show step message (used for Jump or initial step entry)
		err = handler(c, ctrl)
	} else {
		// Process user input or callbacks
		if c.Callback() != nil {
			err = f.handleCallback(c, step, ctrl)
		} else {
			// Handle text input
			if step.onComplete != nil {
				err = step.onComplete(c, ctrl)
			}
		}
	}

	// Handle errors through error handler if configured
	if err != nil && f.onError != nil {
		return f.onError(c, ctrl, err)
	}

	return err
}

// handleCallback processes callback queries for a step.
func (f *Flow) handleCallback(c tele.Context, step *Step, ctrl *controller) error {
	c = ParseCallback(c)
	// Respond to callback to remove loading state
	if step.HasCallback(c.Callback().Unique) {
		_ = c.Respond()
	}

	// Execute step's onComplete handler for callback processing
	if step.onComplete != nil {
		return step.onComplete(c, ctrl)
	}
	return nil
}

// handleTimeout processes session timeouts.
func (f *Flow) handleTimeout(c tele.Context, session *Session) error {
	ctrl := &controller{
		bus:     f.bus,
		session: session,
		teleCtx: c,
	}

	// Clean up session
	f.bus.mu.Lock()
	delete(f.bus.sessions, session.Data.UserID)
	f.bus.mu.Unlock()
	f.bus.storage.Delete(fmt.Sprintf("session:%d", session.Data.UserID))

	// Call timeout handler if configured
	if f.onTimeout != nil {
		return f.onTimeout(c, ctrl)
	}
	return nil
}
