package auth

import (
	"context"

	"github.com/liuyukai/agentskills/internal/database"
)

type contextKey string

const userContextKey contextKey = "user"

// WithUser returns a new context with the user attached.
func WithUser(ctx context.Context, user *database.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUser retrieves the authenticated user from request context.
func GetUser(ctx context.Context) *database.User {
	u, _ := ctx.Value(userContextKey).(*database.User)
	return u
}
