package intele

import (
	"fmt"
	tele "gopkg.in/telebot.v3"
	"time"
)

// controller implements the Controller interface for flow execution control.
type controller struct {
	bus              *FlowBus
	session          *Session
	teleCtx          tele.Context
	storage          Storage
	messageCollector MessageCollector
}

// Storage returns the session storage for this controller.
// Creates a new sessionStorage instance if not already initialized.
func (ctrl *controller) Storage() Storage {
	if ctrl.storage == nil {
		ctrl.storage = &sessionStorage{
			session: ctrl.session,
			bus:     ctrl.bus,
		}
	}
	return ctrl.storage
}

// Container returns the dependency injection container for the current flow.
func (ctrl *controller) Container() Container {
	return ctrl.session.Flow.container
}

// MessageCollector returns the message collector for this controller.
func (ctrl *controller) MessageCollector() MessageCollector {
	if ctrl.messageCollector == nil {
		ctrl.messageCollector = &messageCollector{
			ctrl: ctrl,
		}
	}
	return ctrl.messageCollector
}

// Jump transitions to a different step within the current flow.
// This will execute the target step's handler and show its message.
func (ctrl *controller) Jump(stepID string) error {
	if !IsValidStepID(stepID) {
		return &ErrStepNotFound{StepID: stepID, FlowID: ctrl.session.Flow.id}
	}

	// Update session state atomically
	if err := ctrl.updateSessionStep(stepID); err != nil {
		return err
	}

	// Save session to storage
	if err := ctrl.bus.SaveSession(ctrl.session.Data.UserID, ctrl.session); err != nil {
		return &ErrSessionSave{UserID: ctrl.session.Data.UserID, Err: err}
	}

	// Execute the target step with showHandler=true to display its message
	return ctrl.session.Flow.executeStep(ctrl.teleCtx, ctrl.session, true)
}

// Complete marks the current flow as completed and cleans up the session.
func (ctrl *controller) Complete() error {
	// Call completion handler if defined
	if ctrl.session.Flow.onComplete != nil && ctrl.teleCtx != nil {
		if err := ctrl.session.Flow.onComplete(ctrl.teleCtx, ctrl); err != nil {
			return &ErrCompletionHandler{FlowID: ctrl.session.Flow.id, Err: err}
		}
	}
	return ctrl.cleanupSession()
}

// Cancel cancels the current flow and cleans up the session.
func (ctrl *controller) Cancel() error {
	// Mark session as cancelled
	ctrl.session.mu.Lock()
	ctrl.session.cancelled = true
	ctrl.session.mu.Unlock()

	// Call cancellation handler if defined
	if ctrl.session.Flow.onCancel != nil && ctrl.teleCtx != nil {
		if err := ctrl.session.Flow.onCancel(ctrl.teleCtx, ctrl); err != nil {
			return &ErrCancellationHandler{FlowID: ctrl.session.Flow.id, Err: err}
		}
	}
	return ctrl.cleanupSession()
}

// updateSessionStep atomically updates the current step in the session.
func (ctrl *controller) updateSessionStep(stepID string) error {
	ctrl.session.mu.Lock()
	defer ctrl.session.mu.Unlock()

	if ctrl.session.cancelled {
		return &ErrSessionCancelled{UserID: ctrl.session.Data.UserID}
	}

	// Verify step exists
	if _, exists := ctrl.session.Flow.steps[stepID]; !exists {
		return &ErrStepNotFound{StepID: stepID, FlowID: ctrl.session.Flow.id}
	}

	// Update step and timestamp
	ctrl.session.Data.CurrentStep = stepID
	ctrl.session.Data.UpdatedAt = time.Now()
	return nil
}

// cleanupSession removes the session from both memory and storage.
func (ctrl *controller) cleanupSession() error {
	userID := ctrl.session.Data.UserID

	// Clear collected messages before cleanup
	if ctrl.messageCollector != nil {
		_ = ctrl.messageCollector.Clear(ClearOptions{IgnoreErrors: true})
	}

	// Remove from active sessions
	ctrl.bus.mu.Lock()
	delete(ctrl.bus.sessions, userID)
	ctrl.bus.mu.Unlock()

	// Remove from persistent storage
	ctrl.bus.storage.Delete(fmt.Sprintf("session:%d", userID))

	return nil
}
