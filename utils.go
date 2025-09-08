package intele

import (
	"regexp"
	"time"

	tele "gopkg.in/telebot.v3"
)

// GetTyped is a helper function for type-safe dependency retrieval from a container.
// It attempts to cast the retrieved value to the specified type T.
// Returns the zero value and false if the dependency is not found or cannot be cast.
func GetTyped[T any](container Container, key string) (T, bool) {
	var zero T
	val, err := container.Get(key)
	if err != nil {
		return zero, false
	}

	typed, ok := val.(T)
	return typed, ok
}

// GetStorageTyped is a helper function for type-safe value retrieval from storage.
// It attempts to cast the retrieved value to the specified type T.
// Returns the zero value and false if the key is not found or cannot be cast.
func GetStorageTyped[T any](storage Storage, key string) (T, bool) {
	var zero T
	val, exists := storage.Get(key)
	if !exists {
		return zero, false
	}

	typed, ok := val.(T)
	return typed, ok
}

// IsValidStepID validates that a step ID is non-empty and within reasonable length limits.
func IsValidStepID(stepID string) bool {
	return len(stepID) > 0 && len(stepID) <= 100
}

// IsValidFlowID validates that a flow ID is non-empty and within reasonable length limits.
func IsValidFlowID(flowID string) bool {
	return len(flowID) > 0 && len(flowID) <= 100
}

// LoggingMiddleware creates middleware that logs step execution timing and results.
// The provided logger function will be called with formatted messages about step execution.
func LoggingMiddleware(logger func(format string, args ...interface{})) Middleware {
	return func(next StepHandler) StepHandler {
		return func(c tele.Context, ctrl Controller) error {
			start := time.Now()
			session, _ := ctrl.(*controller)
			stepID := session.session.Data.CurrentStep
			userID := session.session.Data.UserID

			logger("Executing step %s for user %d", stepID, userID)

			err := next(c, ctrl)
			duration := time.Since(start)

			if err != nil {
				logger("Step %s failed for user %d after %v: %v", stepID, userID, duration, err)
			} else {
				logger("Step %s execution completed for user %d in %v", stepID, userID, duration)
			}

			return err
		}
	}
}

// Regex pattern for parsing callback data with unique identifier and optional payload
var cbackRx = regexp.MustCompile(`^\f([-\w]+)(\|(.+))?$`)

// ParseCallback parses callback data to extract the unique identifier and payload.
// Telebot uses a special format for callback data that includes both the unique ID
// and optional payload data separated by specific delimiters.
func ParseCallback(c tele.Context) tele.Context {
	if c.Callback() == nil {
		return c
	}

	match := cbackRx.FindAllStringSubmatch(c.Callback().Data, -1)
	if match != nil {
		unique, payload := match[0][1], match[0][3]
		c.Callback().Unique = unique
		c.Callback().Data = payload
	}

	return c
}
