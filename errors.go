package intele

import "fmt"

// ErrFlowNotFound is returned when a flow is not found
type ErrFlowNotFound struct {
	FlowID string
}

func (e ErrFlowNotFound) Error() string {
	return fmt.Sprintf("flow '%s' not found", e.FlowID)
}

// ErrStepNotFound is returned when a step is not found
type ErrStepNotFound struct {
	StepID string
	FlowID string
}

func (e ErrStepNotFound) Error() string {
	return fmt.Sprintf("step '%s' not found in flow '%s'", e.StepID, e.FlowID)
}

// ErrDuplicateStep is returned when trying to add a step with existing ID
type ErrDuplicateStep struct {
	StepID string
}

func (e ErrDuplicateStep) Error() string {
	return fmt.Sprintf("duplicate step ID: %s", e.StepID)
}

// ErrMissingDependency is returned when a required dependency is not found
type ErrMissingDependency struct {
	Dependency string
	StepID     string
}

func (e ErrMissingDependency) Error() string {
	return fmt.Sprintf("missing dependency '%s' for step '%s'", e.Dependency, e.StepID)
}

// ErrActiveSession is returned when trying to start a flow with active session
type ErrActiveSession struct {
	UserID int64
	FlowID string
}

func (e ErrActiveSession) Error() string {
	return fmt.Sprintf("user %d already has active session in flow '%s'", e.UserID, e.FlowID)
}

// ErrEmptyFlow is returned when trying to build a flow without steps
var ErrEmptyFlow = fmt.Errorf("flow must have at least one step")

// ErrInvalidStartStep is returned when start step is not found
type ErrInvalidStartStep struct {
	StepID string
}

func (e ErrInvalidStartStep) Error() string {
	return fmt.Sprintf("start step '%s' not found", e.StepID)
}
