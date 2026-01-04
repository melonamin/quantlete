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

// ErrorKind categorizes service errors for consistent handling.
type ErrorKind int

const (
	ErrorKindUnknown ErrorKind = iota
	ErrorKindNotFound
	ErrorKindUnauthorized
	ErrorKindForbidden
	ErrorKindBadRequest
	ErrorKindConflict
)

// ErrorKindFor maps an error to a service error category.
func ErrorKindFor(err error) ErrorKind {
	switch {
	case errors.Is(err, ErrNotFound):
		return ErrorKindNotFound
	case errors.Is(err, ErrUnauthorized):
		return ErrorKindUnauthorized
	case errors.Is(err, ErrForbidden):
		return ErrorKindForbidden
	case errors.Is(err, ErrBadRequest):
		return ErrorKindBadRequest
	case errors.Is(err, ErrConflict):
		return ErrorKindConflict
	default:
		return ErrorKindUnknown
	}
}

// ClientMessage returns a safe error message for clients.
// Internal and unknown errors are masked to avoid leaking server details.
func ClientMessage(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrInternal) || ErrorKindFor(err) == ErrorKindUnknown {
		return "internal server error"
	}
	return err.Error()
}

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

// BadRequestf returns a wrapped ErrBadRequest with formatted context.
func BadRequestf(format string, args ...interface{}) error {
	return Wrapf(ErrBadRequest, format, args...)
}

// Conflict returns a wrapped ErrConflict with context.
func Conflict(msg string) error {
	return Wrap(ErrConflict, msg)
}
