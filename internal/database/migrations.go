// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"
)

// Placeholder address prefixes used when a row is anonymized. The row ID is
// appended, which keeps the address unique per table without carrying any
// personal data.
const (
	anonymizedUserPrefix     = "deleted-user"
	anonymizedDownloadPrefix = "deleted-download"
	anonymizedEmailDomain    = "@deleted.local"
)

// anonymizedEmail builds the placeholder address stored in the Email column of
// a soft-deleted row.
func anonymizedEmail(prefix string, id int) string {
	return fmt.Sprintf("%s-%d%s", prefix, id, anonymizedEmailDomain)
}

// emailFingerprint returns a SHA-256 digest of a normalised address. It is
// stored instead of the address itself so support can still answer "was this
// the address?" without the address being readable in the database.
func emailFingerprint(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(email))
	return hex.EncodeToString(sum[:])
}

// RunMigrations applies any pending database migrations
func (d *Database) RunMigrations() error {
	// Add soft delete columns to Users table if they don't exist
	if err := d.addColumnIfNotExists("Users", "DeletedAt", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("Users", "DeletedBy", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("Users", "OriginalEmail", "TEXT DEFAULT ''"); err != nil {
		return err
	}

	// Add soft delete columns to DownloadAccounts table if they don't exist
	if err := d.addColumnIfNotExists("DownloadAccounts", "DeletedAt", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("DownloadAccounts", "DeletedBy", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("DownloadAccounts", "OriginalEmail", "TEXT DEFAULT ''"); err != nil {
		return err
	}

	// Add TOTP (Two-Factor Authentication) columns to Users table
	if err := d.addColumnIfNotExists("Users", "TOTPSecret", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("Users", "TOTPEnabled", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("Users", "BackupCodes", "TEXT DEFAULT ''"); err != nil {
		return err
	}

	// Add Mailgun-specific columns to EmailProviderConfig (v6.2.9, issue #30)
	// Without these, GetActiveProvider scan fails on fresh installs because
	// the Go struct expects MailgunDomain / MailgunRegion to exist.
	if err := d.addColumnIfNotExists("EmailProviderConfig", "MailgunDomain", "TEXT"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("EmailProviderConfig", "MailgunRegion", "TEXT DEFAULT 'us'"); err != nil {
		return err
	}

	// Identity provider columns for external SSO (v6.3.0, issue #29).
	// IdentityProvider = "local" (default) or "entra". ExternalID stores the
	// IdP subject (e.g. Entra OID) and must be unique when non-empty.
	if err := d.addColumnIfNotExists("Users", "IdentityProvider", "TEXT NOT NULL DEFAULT 'local'"); err != nil {
		return err
	}
	if err := d.addColumnIfNotExists("Users", "ExternalID", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	// Partial unique index: enforce uniqueness only when ExternalID is set.
	// SQLite supports partial indexes natively.
	if _, err := d.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_external_id
		ON Users(IdentityProvider, ExternalID) WHERE ExternalID != ''`); err != nil {
		return err
	}

	// Per-user interface language (v7.2.0, issue #33). Empty means "follow the
	// server default", which is what every existing row gets, so upgrading an
	// installation changes nothing until a user picks a language.
	if err := d.addColumnIfNotExists("Users", "Language", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	// RemindedAt is in the CREATE TABLE for fresh installs but never had an
	// upgrade migration, so the expiry-reminder query failed with "no such
	// column" on every database created before it was added to the schema.
	if err := d.addColumnIfNotExists("Files", "RemindedAt", "INTEGER DEFAULT 0"); err != nil {
		return err
	}

	// Rewrite rows anonymized by earlier versions, which stored the real
	// address inside the placeholder and in OriginalEmail.
	if err := d.rewriteLegacyAnonymizedRows(); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// rewriteLegacyAnonymizedRows converts rows anonymized by versions up to 7.1.1
// to the current format. Those versions built the placeholder as
// "deleted_user_<address>@deleted.local" and copied the address into
// OriginalEmail, so the address remained readable. Rows already in the current
// format do not match the patterns, which makes the migration idempotent and a
// no-op on databases that never had such rows.
func (d *Database) rewriteLegacyAnonymizedRows() error {
	if err := d.rewriteLegacyAnonymized("Users", "deleted_user_", anonymizedUserPrefix); err != nil {
		return err
	}
	if err := d.rewriteLegacyAnonymized("DownloadAccounts", "deleted_download_", anonymizedDownloadPrefix); err != nil {
		return err
	}
	return d.rewriteLegacyAnonymizedDownloadLogs()
}

// rewriteLegacyAnonymized rewrites one table. GLOB is used instead of LIKE
// because "_" is a wildcard in LIKE but a literal in GLOB.
func (d *Database) rewriteLegacyAnonymized(table, legacyPrefix, prefix string) error {
	rows, err := d.db.Query(
		"SELECT Id, Email, COALESCE(OriginalEmail, '') FROM "+table+" WHERE Email GLOB ?",
		legacyPrefix+"*"+anonymizedEmailDomain)
	if err != nil {
		return err
	}

	type legacyRow struct {
		id            int
		originalEmail string
	}
	var legacy []legacyRow
	for rows.Next() {
		var id int
		var email, originalEmail string
		if err := rows.Scan(&id, &email, &originalEmail); err != nil {
			rows.Close()
			return err
		}
		if extracted := legacyEmailFromPlaceholder(email, legacyPrefix); extracted != "" {
			originalEmail = extracted
		}
		legacy = append(legacy, legacyRow{id: id, originalEmail: originalEmail})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, row := range legacy {
		if _, err := d.db.Exec(
			"UPDATE "+table+" SET Email = ?, OriginalEmail = ? WHERE Id = ?",
			anonymizedEmail(prefix, row.id), emailFingerprint(row.originalEmail), row.id); err != nil {
			return err
		}
	}

	if len(legacy) > 0 {
		log.Printf("Migration completed: re-anonymized %d legacy row(s) in %s", len(legacy), table)
	}
	return nil
}

// rewriteLegacyAnonymizedDownloadLogs cleans up download log rows that were
// stamped with a legacy placeholder. Rows that still reference an account get
// that account's current placeholder; orphaned rows have the address cleared.
func (d *Database) rewriteLegacyAnonymizedDownloadLogs() error {
	legacyMatch := `(Email GLOB 'deleted_download_*` + anonymizedEmailDomain +
		`' OR Email GLOB 'deleted_user_*` + anonymizedEmailDomain + `')`

	if _, err := d.db.Exec(`
		UPDATE DownloadLogs
		SET Email = '` + anonymizedDownloadPrefix + `-' || DownloadAccountId || '` + anonymizedEmailDomain + `'
		WHERE DownloadAccountId IS NOT NULL AND ` + legacyMatch); err != nil {
		return err
	}

	_, err := d.db.Exec(`
		UPDATE DownloadLogs
		SET Email = ''
		WHERE DownloadAccountId IS NULL AND ` + legacyMatch)
	return err
}

// legacyEmailFromPlaceholder recovers the address embedded in a legacy
// placeholder, or "" if the value is not in that format.
func legacyEmailFromPlaceholder(placeholder, legacyPrefix string) string {
	if !strings.HasPrefix(placeholder, legacyPrefix) || !strings.HasSuffix(placeholder, anonymizedEmailDomain) {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(placeholder, legacyPrefix), anonymizedEmailDomain)
}

// addColumnIfNotExists adds a column to a table if it doesn't already exist
func (d *Database) addColumnIfNotExists(tableName, columnName, columnDef string) error {
	// Check if column exists
	query := `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`
	var count int
	err := d.db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return err
	}

	// If column doesn't exist, add it
	if count == 0 {
		alterSQL := "ALTER TABLE " + tableName + " ADD COLUMN " + columnName + " " + columnDef
		_, err := d.db.Exec(alterSQL)
		if err != nil {
			return err
		}
		log.Printf("Added column %s to table %s", columnName, tableName)
	}

	return nil
}

// SoftDeleteUser marks a user as deleted without removing data (GDPR-compliant soft delete)
func (d *Database) SoftDeleteUser(userId int, deletedBy string) error {
	// Get user to store original email
	user, err := d.GetUserByID(userId)
	if err != nil {
		return err
	}

	// If already deleted, skip
	if user.DeletedAt > 0 {
		return nil
	}

	// Replace the address with a placeholder derived from the row ID, and keep
	// only a fingerprint of the original so it cannot be read back.
	placeholder := anonymizedEmail(anonymizedUserPrefix, userId)

	_, err = d.db.Exec(`
		UPDATE Users
		SET Email = ?, OriginalEmail = ?, DeletedAt = ?, DeletedBy = ?, IsActive = 0
		WHERE Id = ?`,
		placeholder, emailFingerprint(user.Email), currentTimestamp(), deletedBy, userId)

	return err
}

// SoftDeleteDownloadAccount marks a download account as deleted (GDPR-compliant soft delete)
func (d *Database) SoftDeleteDownloadAccount(accountId int, deletedBy string) error {
	// Get account to store original email
	account, err := d.GetDownloadAccountByID(accountId)
	if err != nil {
		return err
	}

	// If already deleted, skip
	if account.DeletedAt > 0 {
		return nil
	}

	// Replace the address with a placeholder derived from the row ID, and keep
	// only a fingerprint of the original so it cannot be read back.
	placeholder := anonymizedEmail(anonymizedDownloadPrefix, accountId)

	_, err = d.db.Exec(`
		UPDATE DownloadAccounts
		SET Email = ?, OriginalEmail = ?, DeletedAt = ?, DeletedBy = ?, IsActive = 0
		WHERE Id = ?`,
		placeholder, emailFingerprint(account.Email), currentTimestamp(), deletedBy, accountId)

	// Also anonymize download logs
	_, _ = d.db.Exec(`
		UPDATE DownloadLogs
		SET Email = ?
		WHERE DownloadAccountId = ?`,
		placeholder, accountId)

	return err
}

// PermanentlyDeleteOldUsers permanently deletes users that have been soft-deleted for more than 90 days
func (d *Database) PermanentlyDeleteOldUsers(daysOld int) (int, error) {
	if daysOld <= 0 {
		daysOld = 90
	}

	cutoffTime := currentTimestamp() - int64(daysOld*24*60*60)

	// Get list of users to delete for logging
	rows, err := d.db.Query(`
		SELECT Id, Email, OriginalEmail, DeletedBy
		FROM Users
		WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	deletedCount := 0
	for rows.Next() {
		var id int
		var email, originalEmail, deletedBy string
		if err := rows.Scan(&id, &email, &originalEmail, &deletedBy); err == nil {
			log.Printf("Permanently deleting user: ID=%d, EmailFingerprint=%s, DeletedBy=%s", id, originalEmail, deletedBy)
			deletedCount++
		}
	}

	// Permanently delete users
	result, err := d.db.Exec(`DELETE FROM Users WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// PermanentlyDeleteOldDownloadAccounts permanently deletes download accounts that have been soft-deleted for more than 90 days
func (d *Database) PermanentlyDeleteOldDownloadAccounts(daysOld int) (int, error) {
	if daysOld <= 0 {
		daysOld = 90
	}

	cutoffTime := currentTimestamp() - int64(daysOld*24*60*60)

	// Get list of accounts to delete for logging
	rows, err := d.db.Query(`
		SELECT Id, Email, OriginalEmail, DeletedBy
		FROM DownloadAccounts
		WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	deletedCount := 0
	for rows.Next() {
		var id int
		var email, originalEmail, deletedBy string
		if err := rows.Scan(&id, &email, &originalEmail, &deletedBy); err == nil {
			log.Printf("Permanently deleting download account: ID=%d, EmailFingerprint=%s, DeletedBy=%s", id, originalEmail, deletedBy)
			deletedCount++
		}
	}

	// Delete download logs for these accounts first
	_, _ = d.db.Exec(`
		DELETE FROM DownloadLogs
		WHERE DownloadAccountId IN (
			SELECT Id FROM DownloadAccounts WHERE DeletedAt > 0 AND DeletedAt < ?
		)`, cutoffTime)

	// Permanently delete accounts
	result, err := d.db.Exec(`DELETE FROM DownloadAccounts WHERE DeletedAt > 0 AND DeletedAt < ?`, cutoffTime)
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// currentTimestamp returns the current Unix timestamp
func currentTimestamp() int64 {
	return time.Now().Unix()
}
