package intele

import "reflect"

// Step represents a single step in a conversation flow.
// Each step has a handler for displaying content and an optional completion handler for processing user input.
type Step struct {
	id              string
	handler         StepHandler
	onComplete      StepHandler
	middlewares     []Middleware
	callbackUniques []string
	RequiredDeps    map[string]reflect.Type
}

// NewStep creates a new step with the given ID and handler.
// The handler is called to display the step's content to the user.
func NewStep(id string, handler StepHandler) *Step {
	return &Step{
		id:           id,
		handler:      handler,
		RequiredDeps: make(map[string]reflect.Type),
	}
}

// OnComplete sets the completion handler for this step.
// This handler is called to process user input (text or callbacks) after the step is displayed.
func (s *Step) OnComplete(handler StepHandler) *Step {
	s.onComplete = handler
	return s
}

// AssignCallback registers callback unique identifiers that this step can handle.
// When a callback with one of these unique IDs is received, it will be processed by this step.
func (s *Step) AssignCallback(unique ...string) *Step {
	s.callbackUniques = append(s.callbackUniques, unique...)
	return s
}

// WithMiddleware adds middleware that will be applied to this step's handler.
// Middleware is executed in reverse order (last added runs first).
func (s *Step) WithMiddleware(mw Middleware) *Step {
	s.middlewares = append(s.middlewares, mw)
	return s
}

// RequireDeps specifies dependencies and their required interface types for this step.
// The flow will fail to build if any required dependencies are missing or don't implement the specified interfaces.
func (s *Step) RequireDeps(deps DependencyMap) *Step {
	if s.RequiredDeps == nil {
		s.RequiredDeps = make(map[string]reflect.Type)
	}
	for depName, ifaceType := range deps {
		s.RequiredDeps[depName] = ifaceType
	}
	return s
}

// ID returns the unique identifier of this step.
func (s *Step) ID() string {
	return s.id
}

// HasCallback checks if this step can handle the given callback unique ID.
func (s *Step) HasCallback(callbackID string) bool {
	for _, id := range s.callbackUniques {
		if id == callbackID {
			return true
		}
	}
	return false
}
