package model

import "errors"

// Sentinel errors returned by the service layer.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrConflict      = errors.New("conflict")
	ErrUnauthorized  = errors.New("unauthorized")
)

// IsSentinel reports whether err wraps one of the public model sentinels.
func IsSentinel(err, target error) bool {
	return errors.Is(err, target)
}

// ValidationError represents a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors holds multiple field-level validation errors.
type ValidationErrors []ValidationError

// HasErrors reports whether any validation errors were collected.
func (ve ValidationErrors) HasErrors() bool { return len(ve) > 0 }
