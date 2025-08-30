package intele

import (
	"time"
)

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

func (b *FlowBuilder) Steps(steps ...*Step) *FlowBuilder {
	b.steps = append(b.steps, steps...)
	return b
}

func (b *FlowBuilder) StartWith(stepID string) *FlowBuilder {
	b.startStep = stepID
	return b
}

func (b *FlowBuilder) WithTimeout(timeout time.Duration) *FlowBuilder {
	b.timeout = timeout
	return b
}

func (b *FlowBuilder) WithContainer(container Container) *FlowBuilder {
	b.container = container
	return b
}

func (b *FlowBuilder) WithMiddleware(mw Middleware) *FlowBuilder {
	b.middlewares = append(b.middlewares, mw)
	return b
}

func (b *FlowBuilder) OnComplete(handler StepHandler) *FlowBuilder {
	b.onComplete = handler
	return b
}

func (b *FlowBuilder) OnError(handler ErrorHandler) *FlowBuilder {
	b.onError = handler
	return b
}

func (b *FlowBuilder) OnCancel(handler StepHandler) *FlowBuilder {
	b.onCancel = handler
	return b
}

func (b *FlowBuilder) OnTimeout(handler StepHandler) *FlowBuilder {
	b.onTimeout = handler
	return b
}

func (b *FlowBuilder) Build() (*Flow, error) {
	if len(b.steps) == 0 {
		return nil, ErrEmptyFlow
	}

	if !IsValidFlowID(b.id) {
		return nil, &ErrFlowNotFound{FlowID: b.id}
	}

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

	startStep := b.startStep
	if startStep == "" {
		startStep = b.steps[0].ID()
	}

	if _, exists := stepsMap[startStep]; !exists {
		return nil, &ErrInvalidStartStep{StepID: startStep}
	}

	if b.container != nil {
		for _, step := range b.steps {
			for _, dep := range step.RequiredDeps {
				if _, err := b.container.Get(dep); err != nil {
					return nil, &ErrMissingDependency{
						Dependency: dep,
						StepID:     step.ID(),
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

	b.bus.RegisterFlow(flow)
	return flow, nil
}
