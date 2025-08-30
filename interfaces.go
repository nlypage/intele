package intele

import (
	tele "gopkg.in/telebot.v3"
)

// StepHandler defines the signature for step handler functions
type StepHandler func(c tele.Context, ctrl Controller) error

// ErrorHandler defines the signature for error handling functions
type ErrorHandler func(c tele.Context, ctrl Controller, err error) error

// Middleware defines the signature for middleware functions
type Middleware func(StepHandler) StepHandler

// Controller provides control interface for flow execution
type Controller interface {
	Storage() Storage
	Container() Container
	Jump(stepID string) error
	Complete() error
	Cancel() error
}

// Storage defines the interface for session state persistence
type Storage interface {
	GetString(key string) (string, bool)
	GetInt(key string) (int, bool)
	GetBool(key string) (bool, bool)
	GetFloat64(key string) (float64, bool)
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Has(key string) bool
	Clear()
	Data() map[string]interface{}
	SetData(data map[string]interface{})
}

// Container defines the interface for dependency injection
type Container interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{})
}
