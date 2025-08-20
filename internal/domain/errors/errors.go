package errors

import (
	"fmt"
)

// Error types for different scenarios
var (
	ErrNotFound            = &AppError{Code: "NOT_FOUND", Message: "Resource not found"}
	ErrAlreadyExists       = &AppError{Code: "ALREADY_EXISTS", Message: "Resource already exists"}
	ErrInvalidInput        = &AppError{Code: "INVALID_INPUT", Message: "Invalid input provided"}
	ErrUnauthorized        = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized access"}
	ErrForbidden           = &AppError{Code: "FORBIDDEN", Message: "Access forbidden"}
	ErrDatabaseError       = &AppError{Code: "DATABASE_ERROR", Message: "Database operation failed"}
	ErrValidationError     = &AppError{Code: "VALIDATION_ERROR", Message: "Validation failed"}
	ErrInternalError       = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	ErrConnectionError     = &AppError{Code: "CONNECTION_ERROR", Message: "Connection failed"}
	ErrTimeoutError        = &AppError{Code: "TIMEOUT_ERROR", Message: "Operation timed out"}
	ErrConstraintViolation = &AppError{Code: "CONSTRAINT_VIOLATION", Message: "Database constraint violated"}
)

// AppError represents a custom application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
		Err:     e.Err,
	}
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
		Err:     err,
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if err != nil && err.Error() == ErrNotFound.Error() {
		return true
	}
	return false
}

// IsAlreadyExists checks if the error is an already exists error
func IsAlreadyExists(err error) bool {
	if err != nil && err.Error() == ErrAlreadyExists.Error() {
		return true
	}
	return false
}

// IsInvalidInput checks if the error is an invalid input error
func IsInvalidInput(err error) bool {
	if err != nil && err.Error() == ErrInvalidInput.Error() {
		return true
	}
	return false
}

// IsDatabaseError checks if the error is a database error
func IsDatabaseError(err error) bool {
	if err != nil && err.Error() == ErrDatabaseError.Error() {
		return true
	}
	return false
}

// IsConstraintViolation checks if the error is a constraint violation error
func IsConstraintViolation(err error) bool {
	if err != nil && err.Error() == ErrConstraintViolation.Error() {
		return true
	}
	return false
}

// NewNotFound creates a new not found error
func NewNotFound(resource string) *AppError {
	return ErrNotFound.WithDetails(fmt.Sprintf("%s not found", resource))
}

// NewAlreadyExists creates a new already exists error
func NewAlreadyExists(resource string) *AppError {
	return ErrAlreadyExists.WithDetails(fmt.Sprintf("%s already exists", resource))
}

// NewInvalidInput creates a new invalid input error
func NewInvalidInput(field string, value interface{}) *AppError {
	return ErrInvalidInput.WithDetails(fmt.Sprintf("Invalid %s: %v", field, value))
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation string, err error) *AppError {
	return ErrDatabaseError.WithDetails(operation).WithError(err)
}

// NewConstraintViolation creates a new constraint violation error
func NewConstraintViolation(constraint string, details string) *AppError {
	return ErrConstraintViolation.WithDetails(fmt.Sprintf("Constraint '%s' violated: %s", constraint, details))
}

// WrapError wraps an error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	if err.Error() == ErrNotFound.Error() ||
		err.Error() == ErrAlreadyExists.Error() ||
		err.Error() == ErrInvalidInput.Error() ||
		err.Error() == ErrDatabaseError.Error() ||
		err.Error() == ErrConstraintViolation.Error() {
		return err
	}

	// Wrap with database error context
	return NewDatabaseError(context, err)
}
