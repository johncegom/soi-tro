package errors

import (
	"fmt"
	"runtime/debug"
)

// AppError represents a custom application error with structured information
type AppError struct {
	Code       string                 // Error code for identification
	Message    string                 // Human-readable error message
	Cause      error                  // Underlying error (if any)
	Context    map[string]any         // Additional context information
	StackTrace string                 // Stack trace for debugging
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError with the given code and message
func New(code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Context:    make(map[string]any),
		StackTrace: string(debug.Stack()),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string) *AppError {
	if err == nil {
		return nil
	}
	
	// If it's already an AppError, just return it
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	
	return &AppError{
		Code:       code,
		Message:    message,
		Cause:      err,
		Context:    make(map[string]any),
		StackTrace: string(debug.Stack()),
	}
}

// With adds context to the error
func (e *AppError) With(key string, value any) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// GetCode returns the error code
func (e *AppError) GetCode() string {
	return e.Code
}

// GetMessage returns the error message
func (e *AppError) GetMessage() string {
	return e.Message
}
