// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)

package database

import (
	"encoding/json"
	"time"
)

// AuditLogEntry represents a single audit log entry
type AuditLogEntry struct {
	ID          int64  `json:"id"`
	Timestamp   int64  `json:"timestamp"`
	UserID      int64  `json:"user_id"`
	UserEmail   string `json:"user_email"`
	Action      string `json:"action"`       // e.g., "USER_CREATED", "FILE_DELETED", "LOGIN_SUCCESS"
	EntityType  string `json:"entity_type"`  // e.g., "User", "File", "Team", "Settings"
	EntityID    string `json:"entity_id"`    // ID of the entity being acted upon
	Details     string `json:"details"`      // JSON with additional context
	IPAddress   string `json:"ip_address"`   // Can be null if IP logging disabled
	UserAgent   string `json:"user_agent"`   // Browser/client info
	Success     bool   `json:"success"`      // Whether action succeeded
	ErrorMsg    string `json:"error_msg"`    // Error message if failed
}

// AuditLogFilter for querying audit logs
type AuditLogFilter struct {
	UserID      int64
	Action      string
	EntityType  string
	StartDate   int64
	EndDate     int64
	SearchTerm  string
	Limit       int
	Offset      int
}

// InitAuditLogTable creates the audit_logs table
func (db *Database) InitAuditLogTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER NOT NULL,
		user_id INTEGER,
		user_email TEXT NOT NULL,
		action TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id TEXT,
		details TEXT,
		ip_address TEXT,
		user_agent TEXT,
		success INTEGER DEFAULT 1,
		error_msg TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_audit_user_id ON audit_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
	CREATE INDEX IF NOT EXISTS idx_audit_entity_type ON audit_logs(entity_type);
	CREATE INDEX IF NOT EXISTS idx_audit_entity_id ON audit_logs(entity_id);
	`

	_, err := db.db.Exec(query)
	return err
}

// LogAction creates an audit log entry
func (db *Database) LogAction(entry *AuditLogEntry) error {
	if entry.Timestamp == 0 {
		entry.Timestamp = time.Now().Unix()
	}

	query := `
	INSERT INTO audit_logs (
		timestamp, user_id, user_email, action, entity_type, entity_id,
		details, ip_address, user_agent, success, error_msg
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.db.Exec(
		query,
		entry.Timestamp,
		entry.UserID,
		entry.UserEmail,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
		entry.Details,
		entry.IPAddress,
		entry.UserAgent,
		entry.Success,
		entry.ErrorMsg,
	)

	return err
}

// GetAuditLogs retrieves audit logs with optional filtering
func (db *Database) GetAuditLogs(filter *AuditLogFilter) ([]*AuditLogEntry, error) {
	query := `SELECT id, timestamp, user_id, user_email, action, entity_type, entity_id,
	          details, ip_address, user_agent, success, error_msg
	          FROM audit_logs WHERE 1=1`
	args := []interface{}{}

	if filter.UserID > 0 {
		query += " AND user_id = ?"
		args = append(args, filter.UserID)
	}

	if filter.Action != "" {
		query += " AND action = ?"
		args = append(args, filter.Action)
	}

	if filter.EntityType != "" {
		query += " AND entity_type = ?"
		args = append(args, filter.EntityType)
	}

	if filter.StartDate > 0 {
		query += " AND timestamp >= ?"
		args = append(args, filter.StartDate)
	}

	if filter.EndDate > 0 {
		query += " AND timestamp <= ?"
		args = append(args, filter.EndDate)
	}

	if filter.SearchTerm != "" {
		query += " AND (user_email LIKE ? OR action LIKE ? OR details LIKE ? OR entity_id LIKE ?)"
		searchPattern := "%" + filter.SearchTerm + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLogEntry
	for rows.Next() {
		log := &AuditLogEntry{}
		var success int
		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.UserID,
			&log.UserEmail,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.Details,
			&log.IPAddress,
			&log.UserAgent,
			&success,
			&log.ErrorMsg,
		)
		if err != nil {
			return nil, err
		}
		log.Success = success == 1
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetAuditLogCount returns total count of logs matching filter
func (db *Database) GetAuditLogCount(filter *AuditLogFilter) (int, error) {
	query := "SELECT COUNT(*) FROM audit_logs WHERE 1=1"
	args := []interface{}{}

	if filter.UserID > 0 {
		query += " AND user_id = ?"
		args = append(args, filter.UserID)
	}

	if filter.Action != "" {
		query += " AND action = ?"
		args = append(args, filter.Action)
	}

	if filter.EntityType != "" {
		query += " AND entity_type = ?"
		args = append(args, filter.EntityType)
	}

	if filter.StartDate > 0 {
		query += " AND timestamp >= ?"
		args = append(args, filter.StartDate)
	}

	if filter.EndDate > 0 {
		query += " AND timestamp <= ?"
		args = append(args, filter.EndDate)
	}

	if filter.SearchTerm != "" {
		query += " AND (user_email LIKE ? OR action LIKE ? OR details LIKE ? OR entity_id LIKE ?)"
		searchPattern := "%" + filter.SearchTerm + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	var count int
	err := db.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// CleanupOldAuditLogs removes logs older than specified days
func (db *Database) CleanupOldAuditLogs(retentionDays int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays).Unix()
	result, err := db.db.Exec("DELETE FROM audit_logs WHERE timestamp < ?", cutoffTime)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetAuditLogSize returns total size of audit_logs table in bytes
func (db *Database) GetAuditLogSize() (int64, error) {
	var size int64
	err := db.db.QueryRow(`
		SELECT page_count * page_size as size
		FROM pragma_page_count('main'), pragma_page_size
	`).Scan(&size)

	if err != nil {
		// Fallback: count rows and estimate
		var count int
		db.db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&count)
		return int64(count * 500), nil // Estimate 500 bytes per log entry
	}

	return size, nil
}

// CleanupAuditLogsBySize removes oldest logs when size exceeds limit (in bytes)
func (db *Database) CleanupAuditLogsBySize(maxSizeBytes int64) (int64, error) {
	currentSize, err := db.GetAuditLogSize()
	if err != nil {
		return 0, err
	}

	if currentSize <= maxSizeBytes {
		return 0, nil // No cleanup needed
	}

	// Calculate how many rows to delete (approx)
	var totalRows int
	db.db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&totalRows)

	if totalRows == 0 {
		return 0, nil
	}

	avgRowSize := currentSize / int64(totalRows)
	rowsToDelete := (currentSize - maxSizeBytes) / avgRowSize
	if rowsToDelete <= 0 {
		rowsToDelete = 1
	}

	// Delete oldest entries
	result, err := db.db.Exec(`
		DELETE FROM audit_logs
		WHERE id IN (
			SELECT id FROM audit_logs
			ORDER BY timestamp ASC
			LIMIT ?
		)
	`, rowsToDelete)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetAuditLogStats returns summary statistics
func (db *Database) GetAuditLogStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total logs
	var total int
	db.db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&total)
	stats["total_logs"] = total

	// Logs by action type
	rows, err := db.db.Query(`
		SELECT action, COUNT(*) as count
		FROM audit_logs
		GROUP BY action
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actionCounts := make(map[string]int)
	for rows.Next() {
		var action string
		var count int
		rows.Scan(&action, &count)
		actionCounts[action] = count
	}
	stats["top_actions"] = actionCounts

	// Logs in last 24 hours
	last24h := time.Now().Add(-24 * time.Hour).Unix()
	var recent int
	db.db.QueryRow("SELECT COUNT(*) FROM audit_logs WHERE timestamp >= ?", last24h).Scan(&recent)
	stats["logs_last_24h"] = recent

	// Failed actions
	var failed int
	db.db.QueryRow("SELECT COUNT(*) FROM audit_logs WHERE success = 0").Scan(&failed)
	stats["failed_actions"] = failed

	// Database size
	size, _ := db.GetAuditLogSize()
	stats["size_bytes"] = size
	stats["size_mb"] = float64(size) / (1024 * 1024)

	return stats, nil
}

// Helper function to create JSON details
func CreateAuditDetails(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonData)
}

// Action constants for consistent naming
const (
	// User actions
	ActionUserCreated      = "USER_CREATED"
	ActionUserUpdated      = "USER_UPDATED"
	ActionUserDeleted      = "USER_DELETED"
	ActionUserActivated    = "USER_ACTIVATED"
	ActionUserDeactivated  = "USER_DEACTIVATED"
	ActionUserQuotaChanged = "USER_QUOTA_CHANGED"
	ActionUserRoleChanged  = "USER_ROLE_CHANGED"

	// Authentication actions
	ActionLoginSuccess        = "LOGIN_SUCCESS"
	ActionLoginFailed         = "LOGIN_FAILED"
	ActionLogout              = "LOGOUT"
	Action2FAEnabled          = "2FA_ENABLED"
	Action2FADisabled         = "2FA_DISABLED"
	ActionPasswordChanged     = "PASSWORD_CHANGED"
	ActionPasswordResetRequested = "PASSWORD_RESET_REQUESTED"
	ActionPasswordResetCompleted = "PASSWORD_RESET_COMPLETED"

	// File actions
	ActionFileUploaded       = "FILE_UPLOADED"
	ActionFileDeleted        = "FILE_DELETED"
	ActionFileRestored       = "FILE_RESTORED"
	ActionFilePermanentlyDeleted = "FILE_PERMANENTLY_DELETED"
	ActionFileShared         = "FILE_SHARED"
	ActionFileDownloaded     = "FILE_DOWNLOADED"
	ActionFileExpired        = "FILE_EXPIRED"

	// Team actions
	ActionTeamCreated       = "TEAM_CREATED"
	ActionTeamUpdated       = "TEAM_UPDATED"
	ActionTeamDeleted       = "TEAM_DELETED"
	ActionTeamMemberAdded   = "TEAM_MEMBER_ADDED"
	ActionTeamMemberRemoved = "TEAM_MEMBER_REMOVED"
	ActionTeamMemberRoleChanged = "TEAM_MEMBER_ROLE_CHANGED"
	ActionFileSharedWithTeam = "FILE_SHARED_WITH_TEAM"
	ActionFileUnsharedFromTeam = "FILE_UNSHARED_FROM_TEAM"

	// Settings actions
	ActionSettingsUpdated = "SETTINGS_UPDATED"
	ActionBrandingUpdated = "BRANDING_UPDATED"
	ActionEmailConfigUpdated = "EMAIL_CONFIG_UPDATED"
	ActionLogoUploaded    = "LOGO_UPLOADED"
	ActionLogoDeleted     = "LOGO_DELETED"

	// Download account actions
	ActionDownloadAccountCreated   = "DOWNLOAD_ACCOUNT_CREATED"
	ActionDownloadAccountUpdated   = "DOWNLOAD_ACCOUNT_UPDATED"
	ActionDownloadAccountDeleted   = "DOWNLOAD_ACCOUNT_DELETED"
	ActionDownloadAccountActivated = "DOWNLOAD_ACCOUNT_ACTIVATED"
	ActionDownloadAccountDeactivated = "DOWNLOAD_ACCOUNT_DEACTIVATED"

	// File request actions
	ActionFileRequestCreated = "FILE_REQUEST_CREATED"
	ActionFileRequestDeleted = "FILE_REQUEST_DELETED"
	ActionFileRequestUploaded = "FILE_REQUEST_UPLOADED"

	// System actions
	ActionSystemStarted = "SYSTEM_STARTED"
	ActionSystemRestarted = "SYSTEM_RESTARTED"
	ActionDatabaseBackup = "DATABASE_BACKUP"
	ActionAuditLogCleanup = "AUDIT_LOG_CLEANUP"
)

// Entity type constants
const (
	EntityUser            = "User"
	EntityFile            = "File"
	EntityTeam            = "Team"
	EntitySettings        = "Settings"
	EntityDownloadAccount = "DownloadAccount"
	EntityFileRequest     = "FileRequest"
	EntitySession         = "Session"
	EntitySystem          = "System"
)
