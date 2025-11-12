// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

import (
	"encoding/json"
	"fmt"

	"github.com/Frimurare/Sharecare/internal/totp"
)

// EnableTOTP enables two-factor authentication for a user
func (d *Database) EnableTOTP(userID int, secret string, backupCodes []string) error {
	// Hash all backup codes
	hashedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashed, err := totp.HashBackupCode(code)
		if err != nil {
			return fmt.Errorf("failed to hash backup code: %w", err)
		}
		hashedCodes[i] = hashed
	}

	// Store as JSON
	backupCodesJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	_, err = d.db.Exec(`
		UPDATE Users
		SET TOTPSecret = ?, TOTPEnabled = 1, BackupCodes = ?
		WHERE Id = ?`,
		secret, string(backupCodesJSON), userID)

	if err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}

	return nil
}

// DisableTOTP disables two-factor authentication for a user
func (d *Database) DisableTOTP(userID int) error {
	_, err := d.db.Exec(`
		UPDATE Users
		SET TOTPSecret = '', TOTPEnabled = 0, BackupCodes = ''
		WHERE Id = ?`,
		userID)

	if err != nil {
		return fmt.Errorf("failed to disable TOTP: %w", err)
	}

	return nil
}

// ValidateBackupCode validates a backup code and removes it if valid (one-time use)
func (d *Database) ValidateBackupCode(userID int, code string) (bool, error) {
	// Get user's backup codes
	var backupCodesJSON string
	err := d.db.QueryRow(`SELECT BackupCodes FROM Users WHERE Id = ?`, userID).Scan(&backupCodesJSON)
	if err != nil {
		return false, fmt.Errorf("failed to get backup codes: %w", err)
	}

	if backupCodesJSON == "" {
		return false, nil
	}

	// Parse JSON
	var hashedCodes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &hashedCodes); err != nil {
		return false, fmt.Errorf("failed to unmarshal backup codes: %w", err)
	}

	// Check each code
	for i, hashedCode := range hashedCodes {
		if totp.ValidateBackupCode(code, hashedCode) {
			// Remove used code
			hashedCodes = append(hashedCodes[:i], hashedCodes[i+1:]...)

			// Save updated codes
			updatedJSON, err := json.Marshal(hashedCodes)
			if err != nil {
				return false, fmt.Errorf("failed to marshal updated codes: %w", err)
			}

			_, err = d.db.Exec(`UPDATE Users SET BackupCodes = ? WHERE Id = ?`,
				string(updatedJSON), userID)
			if err != nil {
				return false, fmt.Errorf("failed to update backup codes: %w", err)
			}

			return true, nil
		}
	}

	return false, nil
}

// RegenerateBackupCodes generates new backup codes for a user
func (d *Database) RegenerateBackupCodes(userID int) ([]string, error) {
	// Generate new codes
	codes, err := totp.GenerateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Hash all codes
	hashedCodes := make([]string, len(codes))
	for i, code := range codes {
		hashed, err := totp.HashBackupCode(code)
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		hashedCodes[i] = hashed
	}

	// Store as JSON
	backupCodesJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	_, err = d.db.Exec(`UPDATE Users SET BackupCodes = ? WHERE Id = ?`,
		string(backupCodesJSON), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to save backup codes: %w", err)
	}

	return codes, nil
}

// GetRemainingBackupCodesCount returns the number of remaining backup codes
func (d *Database) GetRemainingBackupCodesCount(userID int) (int, error) {
	var backupCodesJSON string
	err := d.db.QueryRow(`SELECT BackupCodes FROM Users WHERE Id = ?`, userID).Scan(&backupCodesJSON)
	if err != nil {
		return 0, fmt.Errorf("failed to get backup codes: %w", err)
	}

	if backupCodesJSON == "" {
		return 0, nil
	}

	var hashedCodes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &hashedCodes); err != nil {
		return 0, fmt.Errorf("failed to unmarshal backup codes: %w", err)
	}

	return len(hashedCodes), nil
}
