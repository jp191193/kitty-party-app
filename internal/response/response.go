// Package response provides helpers for writing consistent JSON API responses.
package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"kitty-party-app/internal/apperrors"
)

// APIResponse is the standard envelope for all API replies.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// ErrorBody carries the machine-readable code and a human message.
type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 success response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data})
}

// Created writes a 201 success response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: data})
}

// NoContent writes a 204 response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error writes a structured error response derived from an *apperrors.AppError.
// Falls back to 500 if the error is not an AppError.
func Error(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	var ok bool

	if appErr, ok = err.(*apperrors.AppError); !ok {
		fmt.Printf("UNHANDLED ERROR: %v\n", err)
		appErr = apperrors.ErrInternal
	}

	c.JSON(appErr.Code, APIResponse{
		Success: false,
		Error:   &ErrorBody{Code: appErr.Code, Message: appErr.Message},
	})
}
