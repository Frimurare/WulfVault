package server

import (
	"context"

	"github.com/Frimurare/Sharecare/internal/models"
)

type contextKey string

const (
	userContextKey            contextKey = "user"
	downloadAccountContextKey contextKey = "download_account"
)

// contextWithUser adds a user to the context
func contextWithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// userFromContext retrieves a user from the context
func userFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}

// contextWithDownloadAccount adds a download account to the context
func contextWithDownloadAccount(ctx context.Context, account *models.DownloadAccount) context.Context {
	return context.WithValue(ctx, downloadAccountContextKey, account)
}

// downloadAccountFromContext retrieves a download account from the context
func downloadAccountFromContext(ctx context.Context) (*models.DownloadAccount, bool) {
	account, ok := ctx.Value(downloadAccountContextKey).(*models.DownloadAccount)
	return account, ok
}
