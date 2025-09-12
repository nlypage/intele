// Package intele provides a flow-based conversation framework for Telegram bots.
// It allows building complex multi-step conversations with session management,
// dependency injection, and middleware support.
package intele

import (
	"reflect"
	"time"
)

// FlowBuilder provides a fluent interface for constructing Flow instances.
// It follows the builder pattern to configure various aspects of a flow
// including steps, middleware, timeout, and event handlers.
type FlowBuilder struct {
	bus         *FlowBus
	id          string
	steps       []*Step
	startStep   string
	timeout     time.Duration
	container   Container
	middlewares []Middleware
	onComplete  StepHandler
	onError     ErrorHandler
	onCancel    StepHandler
	onTimeout   StepHandler
}

// Steps adds one or more steps to the flow.
func (b *FlowBuilder) Steps(steps ...*Step) *FlowBuilder {
	b.steps = append(b.steps, steps...)
	return b
}

// StartWith sets the initial step ID for the flow.
// If not specified, the first added step will be used.
func (b *FlowBuilder) StartWith(stepID string) *FlowBuilder {
	b.startStep = stepID
	return b
}

// WithTimeout sets the session timeout for the flow.
// Sessions will automatically expire after this duration.
func (b *FlowBuilder) WithTimeout(timeout time.Duration) *FlowBuilder {
	b.timeout = timeout
	return b
}

// WithContainer sets the dependency injection container for the flow.
func (b *FlowBuilder) WithContainer(container Container) *FlowBuilder {
	b.container = container
	return b
}

// WithMiddleware adds middleware to be executed for all steps in the flow.
func (b *FlowBuilder) WithMiddleware(mw Middleware) *FlowBuilder {
	b.middlewares = append(b.middlewares, mw)
	return b
}

// OnComplete sets the handler to be called when the flow completes successfully.
func (b *FlowBuilder) OnComplete(handler StepHandler) *FlowBuilder {
	b.onComplete = handler
	return b
}

// OnError sets the error handler for the flow.
func (b *FlowBuilder) OnError(handler ErrorHandler) *FlowBuilder {
	b.onError = handler
	return b
}

// OnCancel sets the handler to be called when the flow is cancelled.
func (b *FlowBuilder) OnCancel(handler StepHandler) *FlowBuilder {
	b.onCancel = handler
	return b
}

// OnTimeout sets the handler to be called when the flow times out.
func (b *FlowBuilder) OnTimeout(handler StepHandler) *FlowBuilder {
	b.onTimeout = handler
	return b
}

// Build constructs and validates the Flow instance.
// Returns an error if the flow configuration is invalid.
func (b *FlowBuilder) Build() (*Flow, error) {
	if len(b.steps) == 0 {
		return nil, ErrEmptyFlow
	}

	if !IsValidFlowID(b.id) {
		return nil, &ErrFlowNotFound{FlowID: b.id}
	}

	// Build steps map and validate step IDs
	stepsMap := make(map[string]*Step)
	for _, step := range b.steps {
		if !IsValidStepID(step.ID()) {
			return nil, &ErrStepNotFound{StepID: step.ID(), FlowID: b.id}
		}

		if _, exists := stepsMap[step.ID()]; exists {
			return nil, &ErrDuplicateStep{StepID: step.ID()}
		}

		stepsMap[step.ID()] = step
	}

	// Determine start step
	startStep := b.startStep
	if startStep == "" {
		startStep = b.steps[0].ID()
	}

	if _, exists := stepsMap[startStep]; !exists {
		return nil, &ErrInvalidStartStep{StepID: startStep}
	}

	// Validate dependencies if container is provided
	if b.container != nil {
		for _, step := range b.steps {
			for depName, ifaceType := range step.RequiredDeps {
				val, err := b.container.Get(depName)
				if err != nil {
					return nil, &ErrMissingDependency{
						Dependency: depName,
						StepID:     step.ID(),
					}
				}

				// Check if dependency implements required interface
				valType := reflect.TypeOf(val)
				if !valType.Implements(ifaceType) {
					return nil, &ErrDependencyInterfaceMismatch{
						Dependency: depName,
						StepID:     step.ID(),
						Interface:  ifaceType.String(),
					}
				}
			}
		}
	}

	flow := &Flow{
		bus:         b.bus,
		id:          b.id,
		steps:       stepsMap,
		startStep:   startStep,
		timeout:     b.timeout,
		container:   b.container,
		middlewares: b.middlewares,
		onComplete:  b.onComplete,
		onError:     b.onError,
		onCancel:    b.onCancel,
		onTimeout:   b.onTimeout,
	}

	b.bus.RegisterFlows(flow)
	return flow, nil
}
