package errors

import (
	"errors"
	"net/http"
)

// AppError is a custom application error type that includes additional context
// such as HTTP status code, error code, and user-friendly message.
// It wraps the original error and provides structured information for API responses.
type AppError struct {
	Err        error  // Original error that caused this error
	StatusCode int    // HTTP status code to return to the client
	Code       string // Machine-readable error code (e.g., "BAD_REQUEST")
	Message    string // Human-readable error message
}

// Error returns the error message for the AppError
// If the original error is not nil, it returns the original error's message
// Otherwise, it returns the user-friendly message.
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// New creates a new AppError with the given parameters
// Parameters:
//   - err: Original error that caused this error (can be nil)
//   - statusCode: HTTP status code to return to the client
//   - code: Machine-readable error code (e.g., "BAD_REQUEST")
//   - message: Human-readable error message
// Returns:
//   - *AppError: A new AppError instance
func New(err error, statusCode int, code, message string) *AppError {
	return &AppError{
		Err:        err,
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// NewBadRequest creates a new 400 Bad Request error
// Parameters:
//   - err: Original error that caused this error (can be nil)
//   - message: Human-readable error message
// Returns:
//   - *AppError: A new AppError instance with status code 400
func NewBadRequest(err error, message string) *AppError {
	return New(err, http.StatusBadRequest, "BAD_REQUEST", message)
}

// NewUnauthorized creates a new 401 Unauthorized error
// Parameters:
//   - err: Original error that caused this error (can be nil)
//   - message: Human-readable error message
// Returns:
//   - *AppError: A new AppError instance with status code 401
func NewUnauthorized(err error, message string) *AppError {
	return New(err, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// NewForbidden creates a new 403 Forbidden error
// Parameters:
//   - err: Original error that caused this error (can be nil)
//   - message: Human-readable error message
// Returns:
//   - *AppError: A new AppError instance with status code 403
func NewForbidden(err error, message string) *AppError {
	return New(err, http.StatusForbidden, "FORBIDDEN", message)
}

// NewNotFound creates a new 404 Not Found error
// Parameters:
//   - err: Original error that caused this error (can be nil)
//   - message: Human-readable error message
// Returns:
//   - *AppError: A new AppError instance with status code 404
func NewNotFound(err error, message string) *AppError {
	return New(err, http.StatusNotFound, "NOT_FOUND", message)
}

// NewInternalServerError creates a new 500 Internal Server Error
// Parameters:
//   - err: Original error that caused this error (can be nil)
//   - message: Human-readable error message
// Returns:
//   - *AppError: A new AppError instance with status code 500
func NewInternalServerError(err error, message string) *AppError {
	return New(err, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// IsAppError checks if an error is an AppError
// Parameters:
//   - err: Error to check
// Returns:
//   - *AppError: The AppError instance if the error is an AppError, nil otherwise
//   - bool: True if the error is an AppError, false otherwise
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
