// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU General Public License v3.0 (GPL-3.0)
// You must retain this notice in any copy or derivative work.

package totp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

const (
	// BackupCodeCount is the number of backup codes to generate
	BackupCodeCount = 10
	// BackupCodeLength is the length of each backup code
	BackupCodeLength = 8
)

// GenerateSecret generates a new TOTP secret for a user
func GenerateSecret(email, issuer string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	return key, nil
}

// GenerateQRCode generates a QR code PNG image for the TOTP secret
func GenerateQRCode(key *otp.Key) ([]byte, error) {
	// Generate QR code at 256x256 pixels with medium error correction
	qrCode, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	return qrCode, nil
}

// ValidateCode validates a TOTP code against a secret
func ValidateCode(code, secret string) bool {
	// Clean the code (remove spaces)
	code = strings.ReplaceAll(code, " ", "")

	// Validate the code (allows for time skew of ±1 period = 30 seconds before/after)
	valid := totp.Validate(code, secret)
	return valid
}

// GenerateBackupCodes generates random backup codes for account recovery
func GenerateBackupCodes() ([]string, error) {
	codes := make([]string, BackupCodeCount)

	for i := 0; i < BackupCodeCount; i++ {
		// Generate random bytes
		randomBytes := make([]byte, BackupCodeLength)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Convert to base32-like string (readable format)
		code := base64.RawStdEncoding.EncodeToString(randomBytes)
		// Take first BackupCodeLength characters and make uppercase
		if len(code) > BackupCodeLength {
			code = code[:BackupCodeLength]
		}
		code = strings.ToUpper(code)

		codes[i] = code
	}

	return codes, nil
}

// HashBackupCode hashes a backup code using bcrypt
func HashBackupCode(code string) (string, error) {
	// Use bcrypt with cost 12 (same as passwords in Sharecare)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(code), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash backup code: %w", err)
	}
	return string(hashedBytes), nil
}

// ValidateBackupCode validates a backup code against a hashed code
func ValidateBackupCode(code, hashedCode string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(code))
	return err == nil
}

// FormatBackupCode formats a backup code for display (e.g., "ABCD-EFGH")
func FormatBackupCode(code string) string {
	if len(code) <= 4 {
		return code
	}
	// Insert hyphen every 4 characters
	var formatted strings.Builder
	for i, char := range code {
		if i > 0 && i%4 == 0 {
			formatted.WriteRune('-')
		}
		formatted.WriteRune(char)
	}
	return formatted.String()
}
