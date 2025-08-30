package intele

import (
	"fmt"
	tele "gopkg.in/telebot.v3"
	"log"
	"time"
)

type controller struct {
	bus     *FlowBus
	session *Session
	teleCtx tele.Context
}

func (ctrl *controller) Storage() Storage {
	return &sessionStorage{
		session: ctrl.session,
		bus:     ctrl.bus,
	}
}

func (ctrl *controller) Container() Container {
	return ctrl.session.Flow.container
}

func (ctrl *controller) Jump(stepID string) error {
	if !IsValidStepID(stepID) {
		return &ErrStepNotFound{StepID: stepID, FlowID: ctrl.session.Flow.id}
	}

	ctrl.session.mu.Lock()
	defer ctrl.session.mu.Unlock()

	if _, exists := ctrl.session.Flow.steps[stepID]; !exists {
		return &ErrStepNotFound{StepID: stepID, FlowID: ctrl.session.Flow.id}
	}

	ctrl.session.Data.CurrentStep = stepID
	ctrl.session.Data.UpdatedAt = time.Now()

	if err := ctrl.bus.SaveSession(ctrl.session.Data.UserID, ctrl.session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Execute new step immediately
	return ctrl.session.Flow.executeStep(ctrl.teleCtx, ctrl.session)
}

func (ctrl *controller) Complete() error {
	ctrl.session.mu.Lock()
	sessionCopy := ctrl.session
	ctrl.session.mu.Unlock()

	// execute onComplete if exists
	if sessionCopy.Flow.onComplete != nil && ctrl.teleCtx != nil {
		if err := sessionCopy.Flow.onComplete(ctrl.teleCtx, ctrl); err != nil {
			log.Printf("Error in onComplete handler: %v", err)
		}
	}

	ctrl.bus.mu.Lock()
	delete(ctrl.bus.sessions, sessionCopy.Data.UserID)
	ctrl.bus.mu.Unlock()

	ctrl.bus.storage.Delete(fmt.Sprintf("session:%d", sessionCopy.Data.UserID))

	log.Printf("Flow '%s' completed for user %d", sessionCopy.Flow.id, sessionCopy.Data.UserID)
	return nil
}

func (ctrl *controller) Cancel() error {
	ctrl.session.mu.Lock()
	ctrl.session.cancelled = true
	sessionCopy := ctrl.session
	ctrl.session.mu.Unlock()

	// execute onCancel if exists
	if sessionCopy.Flow.onCancel != nil && ctrl.teleCtx != nil {
		if err := sessionCopy.Flow.onCancel(ctrl.teleCtx, ctrl); err != nil {
			return fmt.Errorf("error in onCancel handler: %w", err)
		}
	}

	ctrl.bus.mu.Lock()
	delete(ctrl.bus.sessions, sessionCopy.Data.UserID)
	ctrl.bus.mu.Unlock()

	ctrl.bus.storage.Delete(fmt.Sprintf("session:%d", sessionCopy.Data.UserID))

	return nil
}
