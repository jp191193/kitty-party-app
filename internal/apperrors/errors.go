// Package apperrors defines domain-level error types and sentinel values.
package apperrors

import "net/http"

// AppError represents a structured, HTTP-aware application error.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Predefined sentinel errors.
var (
	ErrNotFound   = &AppError{Code: http.StatusNotFound, Message: "resource not found"}
	ErrBadRequest = &AppError{Code: http.StatusBadRequest, Message: "bad request"}
	ErrInternal   = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
	ErrConflict   = &AppError{Code: http.StatusConflict, Message: "resource already exists"}
)

// New creates a custom AppError.
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}
