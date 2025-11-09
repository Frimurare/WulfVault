package server

import (
	"context"

	"github.com/Frimurare/Sharecare/internal/models"
)

type contextKey string

const userContextKey contextKey = "user"

// contextWithUser adds a user to the context
func contextWithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// userFromContext retrieves a user from the context
func userFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}
