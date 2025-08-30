package intele

type Step struct {
	id              string
	handler         StepHandler
	onComplete      StepHandler
	middlewares     []Middleware
	callbackUniques []string
	RequiredDeps    []string
}

func NewStep(id string, handler StepHandler) *Step {
	return &Step{
		id:      id,
		handler: handler,
	}
}

func (s *Step) OnComplete(handler StepHandler) *Step {
	s.onComplete = handler
	return s
}

// AssignCallback assigns a callback by its unique identifier to the step.
func (s *Step) AssignCallback(uniques ...string) *Step {
	s.callbackUniques = append(s.callbackUniques, uniques...)
	return s
}

func (s *Step) WithMiddleware(mw Middleware) *Step {
	s.middlewares = append(s.middlewares, mw)
	return s
}

func (s *Step) RequireDeps(deps ...string) *Step {
	s.RequiredDeps = append(s.RequiredDeps, deps...)
	return s
}

func (s *Step) ID() string {
	return s.id
}

func (s *Step) HasCallback(callbackID string) bool {
	for _, id := range s.callbackUniques {
		if id == callbackID {
			return true
		}
	}
	return false
}
