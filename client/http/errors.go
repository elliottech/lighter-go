package http

import (
	"errors"
	"fmt"
)

// APIError represents an error returned by the Lighter API
type APIError struct {
	Code       int32  `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"` // HTTP status code
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.StatusCode != 0 && e.StatusCode != 200 {
		return fmt.Sprintf("API error (HTTP %d, code %d): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error (code %d): %s", e.Code, e.Message)
}

// IsNotFound returns true if this is a not found error
func (e *APIError) IsNotFound() bool {
	return e.Code == 404 || e.StatusCode == 404
}

// IsUnauthorized returns true if this is an authorization error
func (e *APIError) IsUnauthorized() bool {
	return e.Code == 401 || e.StatusCode == 401
}

// IsRateLimited returns true if this is a rate limit error
func (e *APIError) IsRateLimited() bool {
	return e.Code == 429 || e.StatusCode == 429
}

// IsBadRequest returns true if this is a bad request error
func (e *APIError) IsBadRequest() bool {
	return e.Code == 400 || e.StatusCode == 400
}

// IsServerError returns true if this is a server error
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500
}

// Common sentinel errors
var (
	ErrNotFound       = &APIError{Code: 404, Message: "not found"}
	ErrUnauthorized   = &APIError{Code: 401, Message: "unauthorized"}
	ErrForbidden      = &APIError{Code: 403, Message: "forbidden"}
	ErrRateLimited    = &APIError{Code: 429, Message: "rate limited"}
	ErrBadRequest     = &APIError{Code: 400, Message: "bad request"}
	ErrInternalServer = &APIError{Code: 500, Message: "internal server error"}
)

// NewAPIError creates a new APIError from code and message
func NewAPIError(code int32, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// NewAPIErrorWithStatus creates a new APIError with HTTP status code
func NewAPIErrorWithStatus(code int32, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// IsAPIError checks if the error is an APIError and returns it
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// AuthError represents an authentication error
type AuthError struct {
	Reason string
}

// Error implements the error interface
func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication error: %s", e.Reason)
}

// Common auth errors
var (
	ErrAuthTokenExpired  = &AuthError{Reason: "auth token expired"}
	ErrAuthTokenInvalid  = &AuthError{Reason: "auth token invalid"}
	ErrAuthTokenMissing  = &AuthError{Reason: "auth token required"}
)

// ConnectionError represents a connection error
type ConnectionError struct {
	Err error
}

// Error implements the error interface
func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %v", e.Err)
}

// Unwrap returns the underlying error
func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// ValidationError represents a request validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
