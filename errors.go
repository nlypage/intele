package intele

import "fmt"

// ErrFlowNotFound is returned when a requested flow is not found.
type ErrFlowNotFound struct {
	FlowID string
}

func (e ErrFlowNotFound) Error() string {
	return fmt.Sprintf("flow '%s' not found", e.FlowID)
}

// ErrStepNotFound is returned when a requested step is not found in a flow.
type ErrStepNotFound struct {
	StepID string
	FlowID string
}

func (e ErrStepNotFound) Error() string {
	return fmt.Sprintf("step '%s' not found in flow '%s'", e.StepID, e.FlowID)
}

// ErrDuplicateStep is returned when attempting to add a step with an existing ID.
type ErrDuplicateStep struct {
	StepID string
}

func (e ErrDuplicateStep) Error() string {
	return fmt.Sprintf("duplicate step ID: %s", e.StepID)
}

// ErrMissingDependency is returned when a required dependency is not found in the container.
type ErrMissingDependency struct {
	Dependency string
	StepID     string
}

func (e ErrMissingDependency) Error() string {
	return fmt.Sprintf("missing dependency '%s' for step '%s'", e.Dependency, e.StepID)
}

// ErrInvalidStartStep is returned when the specified start step is not found in the flow.
type ErrInvalidStartStep struct {
	StepID string
}

func (e ErrInvalidStartStep) Error() string {
	return fmt.Sprintf("start step '%s' not found", e.StepID)
}

// ErrDependencyNotFound is returned when a dependency is not found in the container.
type ErrDependencyNotFound struct {
	Key string
}

func (e ErrDependencyNotFound) Error() string {
	return fmt.Sprintf("dependency not found: %s", e.Key)
}

// ErrSessionSerialization is returned when session serialization fails.
type ErrSessionSerialization struct {
	UserID int64
	Err    error
}

func (e ErrSessionSerialization) Error() string {
	return fmt.Sprintf("failed to marshal session for user %d: %v", e.UserID, e.Err)
}

// ErrSessionDeserialization is returned when session deserialization fails.
type ErrSessionDeserialization struct {
	UserID int64
	Err    error
}

func (e ErrSessionDeserialization) Error() string {
	return fmt.Sprintf("failed to unmarshal session for user %d: %v", e.UserID, e.Err)
}

// ErrSessionSave is returned when session saving fails.
type ErrSessionSave struct {
	UserID int64
	Err    error
}

func (e ErrSessionSave) Error() string {
	return fmt.Sprintf("failed to save session for user %d: %v", e.UserID, e.Err)
}

// ErrCompletionHandler is returned when completion handler fails.
type ErrCompletionHandler struct {
	FlowID string
	Err    error
}

func (e ErrCompletionHandler) Error() string {
	return fmt.Sprintf("completion handler failed for flow '%s': %v", e.FlowID, e.Err)
}

// ErrCancellationHandler is returned when cancellation handler fails.
type ErrCancellationHandler struct {
	FlowID string
	Err    error
}

func (e ErrCancellationHandler) Error() string {
	return fmt.Sprintf("cancellation handler failed for flow '%s': %v", e.FlowID, e.Err)
}

// ErrSessionCancelled is returned when trying to operate on a cancelled session.
type ErrSessionCancelled struct {
	UserID int64
}

func (e ErrSessionCancelled) Error() string {
	return fmt.Sprintf("cannot operate on cancelled session for user %d", e.UserID)
}

// ErrEmptyFlow is returned when attempting to build a flow without any steps.
var ErrEmptyFlow = fmt.Errorf("flow must have at least one step")

// ErrMessageDeletion is returned when message deletion fails.
type ErrMessageDeletion struct {
	MessageID int
	ChatID    int64
	Err       error
}

func (e ErrMessageDeletion) Error() string {
	return fmt.Sprintf("failed to delete message %d in chat %d: %v", e.MessageID, e.ChatID, e.Err)
}

// ErrMessageSend is returned when message sending fails.
type ErrMessageSend struct {
	ChatID int64
	Err    error
}

func (e ErrMessageSend) Error() string {
	return fmt.Sprintf("failed to send message to chat %d: %v", e.ChatID, e.Err)
}

// ErrMessageCollection is returned when message collection operation fails.
type ErrMessageCollection struct {
	Operation string
	Err       error
}

func (e ErrMessageCollection) Error() string {
	return fmt.Sprintf("message collection %s failed: %v", e.Operation, e.Err)
}
