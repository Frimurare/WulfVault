// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Frimurare/Sharecare/internal/models"
)

// CreateUser inserts a new user into the database
func (d *Database) CreateUser(user *models.User) error {
	if user.CreatedAt == 0 {
		user.CreatedAt = time.Now().Unix()
	}

	resetPw := 0
	if user.ResetPassword {
		resetPw = 1
	}
	isActive := 1
	if !user.IsActive {
		isActive = 0
	}

	result, err := d.db.Exec(`
		INSERT INTO Users (Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		                   StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Name, user.Email, user.Password, user.Permissions, user.UserLevel, user.LastOnline,
		resetPw, user.StorageQuotaMB, user.StorageUsedMB, user.CreatedAt, isActive,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.Id = int(id)
	return nil
}

// GetUserByID retrieves a user by ID
func (d *Database) GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	var resetPw, isActive, totpEnabled int

	err := d.db.QueryRow(`
		SELECT Id, Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		       StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive, TOTPSecret, TOTPEnabled, BackupCodes
		FROM Users WHERE Id = ?`, id).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password, &user.Permissions, &user.UserLevel,
		&user.LastOnline, &resetPw, &user.StorageQuotaMB, &user.StorageUsedMB,
		&user.CreatedAt, &isActive, &user.TOTPSecret, &totpEnabled, &user.BackupCodes,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.ResetPassword = resetPw == 1
	user.IsActive = isActive == 1
	user.TOTPEnabled = totpEnabled == 1
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (d *Database) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var resetPw, isActive, totpEnabled int

	err := d.db.QueryRow(`
		SELECT Id, Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		       StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive, TOTPSecret, TOTPEnabled, BackupCodes
		FROM Users WHERE Email = ?`, email).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password, &user.Permissions, &user.UserLevel,
		&user.LastOnline, &resetPw, &user.StorageQuotaMB, &user.StorageUsedMB,
		&user.CreatedAt, &isActive, &user.TOTPSecret, &totpEnabled, &user.BackupCodes,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.ResetPassword = resetPw == 1
	user.IsActive = isActive == 1
	user.TOTPEnabled = totpEnabled == 1
	return user, nil
}

// GetUserByName retrieves a user by username
func (d *Database) GetUserByName(name string) (*models.User, error) {
	user := &models.User{}
	var resetPw, isActive, totpEnabled int

	err := d.db.QueryRow(`
		SELECT Id, Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		       StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive, TOTPSecret, TOTPEnabled, BackupCodes
		FROM Users WHERE Name = ?`, name).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password, &user.Permissions, &user.UserLevel,
		&user.LastOnline, &resetPw, &user.StorageQuotaMB, &user.StorageUsedMB,
		&user.CreatedAt, &isActive, &user.TOTPSecret, &totpEnabled, &user.BackupCodes,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.ResetPassword = resetPw == 1
	user.IsActive = isActive == 1
	user.TOTPEnabled = totpEnabled == 1
	return user, nil
}

// GetAllUsers returns all users
func (d *Database) GetAllUsers() ([]*models.User, error) {
	rows, err := d.db.Query(`
		SELECT Id, Name, Email, Password, Permissions, Userlevel, LastOnline, ResetPassword,
		       StorageQuotaMB, StorageUsedMB, CreatedAt, IsActive
		FROM Users ORDER BY Userlevel ASC, LastOnline DESC, Name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var resetPw, isActive int

		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.Permissions,
			&user.UserLevel, &user.LastOnline, &resetPw, &user.StorageQuotaMB, &user.StorageUsedMB,
			&user.CreatedAt, &isActive)
		if err != nil {
			return nil, err
		}

		user.ResetPassword = resetPw == 1
		user.IsActive = isActive == 1
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates an existing user
func (d *Database) UpdateUser(user *models.User) error {
	resetPw := 0
	if user.ResetPassword {
		resetPw = 1
	}
	isActive := 1
	if !user.IsActive {
		isActive = 0
	}

	_, err := d.db.Exec(`
		UPDATE Users SET Name = ?, Email = ?, Password = ?, Permissions = ?, Userlevel = ?,
		                 LastOnline = ?, ResetPassword = ?, StorageQuotaMB = ?, StorageUsedMB = ?,
		                 IsActive = ?
		WHERE Id = ?`,
		user.Name, user.Email, user.Password, user.Permissions, user.UserLevel, user.LastOnline,
		resetPw, user.StorageQuotaMB, user.StorageUsedMB, isActive, user.Id,
	)
	return err
}

// UpdateUserLastOnline updates the last online timestamp
func (d *Database) UpdateUserLastOnline(id int) error {
	_, err := d.db.Exec("UPDATE Users SET LastOnline = ? WHERE Id = ?", time.Now().Unix(), id)
	return err
}

// UpdateUserPassword updates a user's password
func (d *Database) UpdateUserPassword(id int, hashedPassword string) error {
	_, err := d.db.Exec("UPDATE Users SET Password = ? WHERE Id = ?", hashedPassword, id)
	return err
}

// UpdateUserStorage updates a user's storage usage
func (d *Database) UpdateUserStorage(id int, storageUsedMB int64) error {
	_, err := d.db.Exec("UPDATE Users SET StorageUsedMB = ? WHERE Id = ?", storageUsedMB, id)
	return err
}

// DeleteUser deletes a user by ID
// Before deletion, all user's files are moved to trash (soft-deleted)
func (d *Database) DeleteUser(id int, deletedBy int) error {
	// First, soft-delete all files belonging to this user (move to trash)
	if err := d.SoftDeleteUserFiles(id, deletedBy); err != nil {
		return fmt.Errorf("failed to soft-delete user files: %w", err)
	}

	// Then delete the user
	_, err := d.db.Exec("DELETE FROM Users WHERE Id = ?", id)
	return err
}

// GetTotalUsers returns the count of all users
func (d *Database) GetTotalUsers() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM Users").Scan(&count)
	return count, err
}

// GetActiveUsers returns the count of active users
func (d *Database) GetActiveUsers() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM Users WHERE IsActive = 1").Scan(&count)
	return count, err
}

// GetUsersAddedThisMonth returns count of users (including download accounts) created this month
func (d *Database) GetUsersAddedThisMonth() (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	var regularUsers, downloadAccounts int

	// Count regular users
	err := d.db.QueryRow("SELECT COUNT(*) FROM Users WHERE CreatedAt >= ?", startOfMonth).Scan(&regularUsers)
	if err != nil {
		return 0, err
	}

	// Count download accounts
	err = d.db.QueryRow("SELECT COUNT(*) FROM DownloadAccounts WHERE CreatedAt >= ? AND DeletedAt = 0", startOfMonth).Scan(&downloadAccounts)
	if err != nil {
		return 0, err
	}

	return regularUsers + downloadAccounts, nil
}

// GetUsersRemovedThisMonth returns count of download accounts deleted this month
func (d *Database) GetUsersRemovedThisMonth() (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM DownloadAccounts WHERE DeletedAt >= ? AND DeletedAt < ?",
		startOfMonth, now.Unix()).Scan(&count)
	return count, err
}

// GetUserGrowthPercentage returns the user growth percentage for this month
// Includes both Users and DownloadAccounts to match Users Added/Removed statistics
func (d *Database) GetUserGrowthPercentage() (float64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	// Count Users at start of month
	var regularUsersAtStart int
	err := d.db.QueryRow("SELECT COUNT(*) FROM Users WHERE CreatedAt < ?", startOfMonth).Scan(&regularUsersAtStart)
	if err != nil {
		return 0, err
	}

	// Count DownloadAccounts active at start of month
	// (created before month start AND not deleted, OR deleted after month start)
	var downloadAccountsAtStart int
	err = d.db.QueryRow(`
		SELECT COUNT(*) FROM DownloadAccounts
		WHERE CreatedAt < ? AND (DeletedAt = 0 OR DeletedAt >= ?)
	`, startOfMonth, startOfMonth).Scan(&downloadAccountsAtStart)
	if err != nil {
		return 0, err
	}

	usersAtStart := regularUsersAtStart + downloadAccountsAtStart

	// Count Users now
	var regularUsersNow int
	err = d.db.QueryRow("SELECT COUNT(*) FROM Users").Scan(&regularUsersNow)
	if err != nil {
		return 0, err
	}

	// Count active DownloadAccounts now
	var downloadAccountsNow int
	err = d.db.QueryRow("SELECT COUNT(*) FROM DownloadAccounts WHERE DeletedAt = 0").Scan(&downloadAccountsNow)
	if err != nil {
		return 0, err
	}

	usersNow := regularUsersNow + downloadAccountsNow

	// Calculate percentage
	if usersAtStart == 0 {
		return 0, nil
	}

	growth := float64(usersNow-usersAtStart) / float64(usersAtStart) * 100
	return growth, nil
}

// ============================================================================
// SECURITY STATISTICS
// ============================================================================

// Get2FAAdoptionRate returns the percentage of Users and Admins who have enabled 2FA
func (d *Database) Get2FAAdoptionRate() (float64, error) {
	var totalUsers, usersWithTOTP int

	// Count total Users and Admins (exclude Download accounts which don't have AccountType field)
	err := d.db.QueryRow(`
		SELECT COUNT(*)
		FROM Users
		WHERE AccountType IN ('User', 'Admin')
	`).Scan(&totalUsers)
	if err != nil {
		return 0, err
	}

	if totalUsers == 0 {
		return 0, nil
	}

	// Count users with TOTP enabled
	err = d.db.QueryRow(`
		SELECT COUNT(*)
		FROM Users
		WHERE AccountType IN ('User', 'Admin') AND TOTPEnabled = 1
	`).Scan(&usersWithTOTP)
	if err != nil {
		return 0, err
	}

	adoption := float64(usersWithTOTP) / float64(totalUsers) * 100
	return adoption, nil
}

// GetAverageBackupCodesRemaining returns the average number of backup codes remaining per user with 2FA enabled
func (d *Database) GetAverageBackupCodesRemaining() (float64, error) {
	rows, err := d.db.Query(`
		SELECT BackupCodes
		FROM Users
		WHERE TOTPEnabled = 1 AND BackupCodes != ''
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	totalCodes := 0
	userCount := 0

	for rows.Next() {
		var backupCodesJSON string
		if err := rows.Scan(&backupCodesJSON); err != nil {
			continue
		}

		// Count non-empty codes in JSON array (simple count of commas + 1)
		// More accurate would be to parse JSON, but this is lightweight
		if backupCodesJSON != "" && backupCodesJSON != "[]" {
			// Rough estimate: count opening brackets for hashed codes
			// Each code looks like "$2a$12$..." so count "$2a$" occurrences
			count := 0
			for i := 0; i < len(backupCodesJSON)-3; i++ {
				if backupCodesJSON[i:i+4] == "$2a$" {
					count++
				}
			}
			totalCodes += count
			userCount++
		}
	}

	if userCount == 0 {
		return 10.0, nil // Default: assume full set if no one has used any
	}

	return float64(totalCodes) / float64(userCount), rows.Err()
}
