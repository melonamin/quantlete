// Package services provides the business logic layer shared between
// HTTP handlers and WASM handlers.
package services

import (
	"errors"
	"fmt"
)

// Sentinel errors for service layer.
// Handlers translate these to appropriate responses (HTTP status codes / JSON errors).
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("access denied")
	ErrBadRequest   = errors.New("bad request")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal error")
)

// ServiceError wraps an error with additional context.
type ServiceError struct {
	Err     error
	Message string
}

func (e *ServiceError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// Wrap creates a ServiceError with context message.
func Wrap(err error, msg string) error {
	return &ServiceError{Err: err, Message: msg}
}

// Wrapf creates a ServiceError with formatted context message.
func Wrapf(err error, format string, args ...interface{}) error {
	return &ServiceError{Err: err, Message: fmt.Sprintf(format, args...)}
}

// NotFound returns a wrapped ErrNotFound with context.
func NotFound(resource string) error {
	return Wrapf(ErrNotFound, "%s not found", resource)
}

// BadRequest returns a wrapped ErrBadRequest with context.
func BadRequest(msg string) error {
	return Wrap(ErrBadRequest, msg)
}

// Conflict returns a wrapped ErrConflict with context.
func Conflict(msg string) error {
	return Wrap(ErrConflict, msg)
}
