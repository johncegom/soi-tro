package logger

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is the key type for storing request ID in context
type contextKey string

const requestIDKey contextKey = "request_id"

// NewContext creates a new context with a request ID
func NewContext(ctx context.Context) context.Context {
	requestID := uuid.New().String()
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID extracts the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}
