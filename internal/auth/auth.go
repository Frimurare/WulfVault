package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Frimurare/Sharecare/internal/database"
	"github.com/Frimurare/Sharecare/internal/models"
)

const (
	SessionDuration = 24 * time.Hour
	BcryptCost      = 12
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateSessionID generates a random session ID
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session for a user
func CreateSession(userId int) (string, error) {
	sessionId, err := GenerateSessionID()
	if err != nil {
		return "", err
	}

	validUntil := time.Now().Add(SessionDuration).Unix()

	_, err = database.DB.Exec(`
		INSERT INTO Sessions (Id, UserId, ValidUntil)
		VALUES (?, ?, ?)`,
		sessionId, userId, validUntil,
	)
	if err != nil {
		return "", err
	}

	return sessionId, nil
}

// GetUserBySession retrieves a user by session ID
func GetUserBySession(sessionId string) (*models.User, error) {
	var userId int
	var validUntil int64

	err := database.DB.QueryRow(`
		SELECT UserId, ValidUntil FROM Sessions WHERE Id = ?`,
		sessionId,
	).Scan(&userId, &validUntil)

	if err != nil {
		return nil, errors.New("invalid session")
	}

	// Check if session is expired
	if time.Now().Unix() > validUntil {
		// Delete expired session
		database.DB.Exec("DELETE FROM Sessions WHERE Id = ?", sessionId)
		return nil, errors.New("session expired")
	}

	// Get user
	user, err := database.DB.GetUserByID(userId)
	if err != nil {
		return nil, err
	}

	// Update last online
	database.DB.UpdateUserLastOnline(userId)

	return user, nil
}

// DeleteSession deletes a session (logout)
func DeleteSession(sessionId string) error {
	_, err := database.DB.Exec("DELETE FROM Sessions WHERE Id = ?", sessionId)
	return err
}

// CleanupExpiredSessions removes all expired sessions
func CleanupExpiredSessions() error {
	_, err := database.DB.Exec("DELETE FROM Sessions WHERE ValidUntil < ?", time.Now().Unix())
	return err
}

// AuthenticateUser authenticates a user by email/username and password
// Returns User or nil. For download accounts, use AuthenticateDownloadAccount directly.
func AuthenticateUser(emailOrUsername, password string) (*models.User, error) {
	// Try by email first
	user, err := database.DB.GetUserByEmail(emailOrUsername)
	if err != nil {
		// Try by username
		user, err = database.DB.GetUserByName(emailOrUsername)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Check password
	if !CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// AccountType represents the type of account that was authenticated
type AccountType string

const (
	AccountTypeUser           AccountType = "user"
	AccountTypeDownloadAccount AccountType = "download_account"
)

// AuthResult contains information about an authenticated account
type AuthResult struct {
	User            *models.User            // Non-nil if it's a regular user
	DownloadAccount *models.DownloadAccount // Non-nil if it's a download account
	AccountType     AccountType             // Type of account
	AccountID       int                     // ID of the account
	Email           string                  // Email of the account
}

// AuthenticateAnyAccount tries to authenticate as either a User or DownloadAccount
func AuthenticateAnyAccount(email, password string) (*AuthResult, error) {
	// First try regular user authentication
	user, err := AuthenticateUser(email, password)
	if err == nil {
		return &AuthResult{
			User:        user,
			AccountType: AccountTypeUser,
			AccountID:   user.Id,
			Email:       user.Email,
		}, nil
	}

	// If user auth failed, try download account
	downloadAccount, err := AuthenticateDownloadAccount(email, password)
	if err == nil {
		return &AuthResult{
			DownloadAccount: downloadAccount,
			AccountType:     AccountTypeDownloadAccount,
			AccountID:       downloadAccount.Id,
			Email:           downloadAccount.Email,
		}, nil
	}

	// Both failed
	return nil, errors.New("invalid credentials")
}

// AuthenticateDownloadAccount authenticates a download account
func AuthenticateDownloadAccount(email, password string) (*models.DownloadAccount, error) {
	account, err := database.DB.GetDownloadAccountByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if account is active
	if !account.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Check password
	if !CheckPasswordHash(password, account.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Update last used
	database.DB.UpdateDownloadAccountLastUsed(account.Id)

	return account, nil
}

// CreateDownloadAccount creates a new download account
func CreateDownloadAccount(email, password string) (*models.DownloadAccount, error) {
	// Check if account already exists
	existing, _ := database.DB.GetDownloadAccountByEmail(email)
	if existing != nil {
		return nil, errors.New("account already exists")
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	account := &models.DownloadAccount{
		Email:    email,
		Password: hashedPassword,
		IsActive: true,
	}

	err = database.DB.CreateDownloadAccount(account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// CreateDownloadAccountSession creates a session token for a download account
// Returns the email (used as session identifier)
func CreateDownloadAccountSession(accountId int) (string, error) {
	account, err := database.DB.GetDownloadAccountByID(accountId)
	if err != nil {
		return "", err
	}

	// For download accounts, we use their email as the session identifier
	// This is simpler than creating a separate session table
	return account.Email, nil
}

// GetDownloadAccountBySession retrieves a download account by session (email)
func GetDownloadAccountBySession(sessionEmail string) (*models.DownloadAccount, error) {
	account, err := database.DB.GetDownloadAccountByEmail(sessionEmail)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	// Check if account is active
	if !account.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Update last used
	database.DB.UpdateDownloadAccountLastUsed(account.Id)

	return account, nil
}
