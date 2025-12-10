package http

import (
	"fmt"
	"net/http"
)

// HTTPError represents an HTTP error with a status code
type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// BadRequest creates a 400 error
func BadRequest(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, message, err)
}

// Unauthorized creates a 401 error
func Unauthorized(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message, err)
}

// Forbidden creates a 403 error
func Forbidden(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusForbidden, message, err)
}

// NotFound creates a 404 error
func NotFound(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusNotFound, message, err)
}

// InternalError creates a 500 error
func InternalError(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message, err)
}
