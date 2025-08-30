package intele

import (
	"context"
	tele "gopkg.in/telebot.v3"
	"strings"
	"time"
)

// GetTyped is a helper function for type-safe container access
func GetTyped[T any](container Container, key string) (T, bool) {
	var zero T
	val, err := container.Get(key)
	if err != nil {
		return zero, false
	}

	typed, ok := val.(T)
	return typed, ok
}

// GetStorageTyped is a helper function for type-safe storage access
func GetStorageTyped[T any](storage Storage, key string) (T, bool) {
	var zero T
	val, exists := storage.Get(key)
	if !exists {
		return zero, false
	}

	typed, ok := val.(T)
	return typed, ok
}

// WithTimeout creates a context with timeout for step execution
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// IsValidStepID checks if step ID is valid (non-empty and reasonable length)
func IsValidStepID(stepID string) bool {
	return len(stepID) > 0 && len(stepID) <= 100
}

// IsValidFlowID checks if flow ID is valid (non-empty and reasonable length)
func IsValidFlowID(flowID string) bool {
	return len(flowID) > 0 && len(flowID) <= 100
}

// LoggingMiddleware creates middleware that logs step executions
func LoggingMiddleware(logger func(string, ...interface{})) Middleware {
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
				logger("Step %s completed for user %d in %v", stepID, userID, duration)
			}

			return err
		}
	}
}

func ParseCallbackData(c tele.Context) tele.Context {
	c.Callback().Data = strings.TrimSpace(c.Callback().Data)
	c.Callback().Unique = strings.TrimSpace(c.Callback().Data)

	if strings.Contains(c.Callback().Data, "|") {
		parts := strings.SplitN(c.Callback().Data, "|", 2)
		c.Callback().Unique = parts[0]
		c.Callback().Data = parts[1]
		return c
	}
	c.Callback().Unique = c.Callback().Data
	return c
}
