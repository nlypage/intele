package intele

import (
	tele "gopkg.in/telebot.v3"
	"time"
)

// StepHandler defines the function signature for step handlers.
// These functions are called to display step content and handle user interactions.
type StepHandler func(c tele.Context, ctrl Controller) error

// ErrorHandler defines the function signature for error handling.
// These functions are called when errors occur during flow execution.
type ErrorHandler func(c tele.Context, ctrl Controller, err error) error

// Middleware defines the function signature for middleware functions.
type Middleware func(StepHandler) StepHandler

// Controller provides the interface for controlling flow execution within step handlers.
// It allows steps to access storage, dependencies, and control flow navigation.
type Controller interface {
	// Storage returns the session storage for persistent state management.
	Storage() Storage

	// Container returns the dependency injection container.
	Container() Container

	// Jump transitions to a different step in the current flow.
	Jump(stepID string) error

	// Complete marks the flow as successfully completed.
	Complete() error

	// Cancel cancels the current flow execution.
	Cancel() error

	// MessageCollector returns the message collector for this controller
	MessageCollector() MessageCollector
}

// Storage defines the interface for session state persistence.
// Implementations should provide thread-safe access to session data.
type Storage interface {
	// GetString retrieves a string value by key.
	GetString(key string) (string, bool)

	// GetInt retrieves an integer value by key.
	GetInt(key string) (int, bool)

	// GetBool retrieves a boolean value by key.
	GetBool(key string) (bool, bool)

	// GetFloat64 retrieves a float64 value by key.
	GetFloat64(key string) (float64, bool)

	// Get retrieves any value by key.
	Get(key string) (interface{}, bool)

	// Set stores a value with the given key.
	Set(key string, value interface{})

	// Delete removes a value by key.
	Delete(key string)

	// Has checks if a key exists.
	Has(key string) bool

	// Clear removes all stored data.
	Clear()

	// Data returns a copy of all stored data.
	Data() map[string]interface{}

	// SetData replaces all stored data.
	SetData(data map[string]interface{})
}

// Container defines the interface for dependency injection.
// It provides a way to register and retrieve dependencies for steps.
type Container interface {
	// Get retrieves a dependency by key.
	Get(key string) (interface{}, error)

	// Set registers a dependency with the given key.
	Set(key string, value interface{})
}

// ClearOptions defines options for clearing collected messages.
type ClearOptions struct {
	// IgnoreErrors will ignore all errors that occurred during deletion
	IgnoreErrors bool
	// ExcludeLast will exclude the last message and don't delete it
	ExcludeLast bool
	// MaxAge if set, only messages older than this duration will be deleted
	MaxAge *time.Duration
}

// CollectedMessage represents a message stored in the collector.
type CollectedMessage struct {
	MessageID int       `json:"message_id"`
	ChatID    int64     `json:"chat_id"`
	Timestamp time.Time `json:"timestamp"`
}

// MessageCollector provides methods for collecting and managing messages.
type MessageCollector interface {
	// Collect adds a message to the collector
	Collect(messageID int, chatID int64) error
	// Send sends a message and automatically collects it
	Send(what interface{}, opts ...interface{}) error
	// GetMessages returns all collected messages
	GetMessages() ([]CollectedMessage, error)
	// Clear deletes all collected messages and cleans the collector
	Clear(opts ClearOptions) error
}
