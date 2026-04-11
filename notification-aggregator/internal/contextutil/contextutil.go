package contextutil

import (
	"context"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// WithRequestID は Context に Request ID をセットします
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestID は Context から Request ID を取り出します
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}
