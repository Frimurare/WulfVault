package database

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

const (
	AccountTypeUser           = "user"
	AccountTypeDownloadAccount = "download_account"
	ResetTokenDuration        = 1 * time.Hour
)

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	Id          int
	Token       string
	Email       string
	AccountType string
	ExpiresAt   int64
	Used        bool
	CreatedAt   int64
}

// CreatePasswordResetToken creates a new password reset token
func (db *Database) CreatePasswordResetToken(email, accountType string) (string, error) {
	// Generate secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(ResetTokenDuration).Unix()
	createdAt := time.Now().Unix()

	_, err := db.Exec(`
		INSERT INTO PasswordResetTokens (Token, Email, AccountType, ExpiresAt, CreatedAt)
		VALUES (?, ?, ?, ?, ?)`,
		token, email, accountType, expiresAt, createdAt,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetPasswordResetToken retrieves a token by its value
func (db *Database) GetPasswordResetToken(token string) (*PasswordResetToken, error) {
	var resetToken PasswordResetToken
	var used int

	err := db.QueryRow(`
		SELECT Id, Token, Email, AccountType, ExpiresAt, Used, CreatedAt
		FROM PasswordResetTokens
		WHERE Token = ?`,
		token,
	).Scan(&resetToken.Id, &resetToken.Token, &resetToken.Email, &resetToken.AccountType,
		&resetToken.ExpiresAt, &used, &resetToken.CreatedAt)

	if err != nil {
		return nil, errors.New("token not found")
	}

	resetToken.Used = used == 1

	// Check if expired
	if time.Now().Unix() > resetToken.ExpiresAt {
		return nil, errors.New("token expired")
	}

	// Check if already used
	if resetToken.Used {
		return nil, errors.New("token already used")
	}

	return &resetToken, nil
}

// MarkPasswordResetTokenUsed marks a token as used
func (db *Database) MarkPasswordResetTokenUsed(token string) error {
	_, err := db.Exec(`
		UPDATE PasswordResetTokens
		SET Used = 1
		WHERE Token = ?`,
		token,
	)
	return err
}

// CleanupExpiredResetTokens removes expired tokens
func (db *Database) CleanupExpiredResetTokens() error {
	_, err := db.Exec(`
		DELETE FROM PasswordResetTokens
		WHERE ExpiresAt < ?`,
		time.Now().Unix(),
	)
	return err
}

// ResetPasswordWithToken resets a password using a valid token
func (db *Database) ResetPasswordWithToken(token, newPassword string) error {
	// Get token
	resetToken, err := db.GetPasswordResetToken(token)
	if err != nil {
		return err
	}

	// Update password based on account type
	if resetToken.AccountType == AccountTypeUser {
		// Update regular user
		_, err = db.Exec(`
			UPDATE Users
			SET Password = ?
			WHERE Email = ? AND IsActive = 1`,
			newPassword, resetToken.Email,
		)
	} else if resetToken.AccountType == AccountTypeDownloadAccount {
		// Update download account
		_, err = db.Exec(`
			UPDATE DownloadAccounts
			SET Password = ?
			WHERE Email = ? AND IsActive = 1`,
			newPassword, resetToken.Email,
		)
	} else {
		return errors.New("invalid account type")
	}

	if err != nil {
		return err
	}

	// Mark token as used
	return db.MarkPasswordResetTokenUsed(token)
}
