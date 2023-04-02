package ctxutil

import "context"

type contextKey int

const (
	_ contextKey = iota
	ctxKeyUserID
)

// WithUserID creates a new context that has username injected
func WithUserID(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, username)
}

// UserID tries to retrieve username from the given context
func UserID(ctx context.Context) string {
	if username, ok := ctx.Value(ctxKeyUserID).(string); ok {
		return username
	}
	return ""
}
