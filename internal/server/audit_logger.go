// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// AuditLogger provides centralized audit logging functionality
type AuditLogger struct {
	server *Server
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(s *Server) *AuditLogger {
	return &AuditLogger{server: s}
}

// LogAction logs an action to the audit log
func (al *AuditLogger) LogAction(
	userID int64,
	userEmail string,
	action string,
	entityType string,
	entityID string,
	details map[string]interface{},
	r *http.Request,
	success bool,
	errorMsg string,
) error {
	entry := &database.AuditLogEntry{
		UserID:     userID,
		UserEmail:  userEmail,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    database.CreateAuditDetails(details),
		Success:    success,
		ErrorMsg:   errorMsg,
	}

	// Add IP and User-Agent if request provided
	if r != nil {
		entry.IPAddress = al.getClientIP(r)
		entry.UserAgent = r.UserAgent()
	}

	return database.DB.LogAction(entry)
}

// LogUserAction logs an action performed by a user
func (al *AuditLogger) LogUserAction(user *models.User, action, entityType, entityID string, details map[string]interface{}, r *http.Request) error {
	return al.LogAction(
		int64(user.Id),
		user.Email,
		action,
		entityType,
		entityID,
		details,
		r,
		true,
		"",
	)
}

// LogUserActionWithStatus logs an action with custom success/error status
func (al *AuditLogger) LogUserActionWithStatus(
	user *models.User,
	action, entityType, entityID string,
	details map[string]interface{},
	r *http.Request,
	success bool,
	errorMsg string,
) error {
	return al.LogAction(
		int64(user.Id),
		user.Email,
		action,
		entityType,
		entityID,
		details,
		r,
		success,
		errorMsg,
	)
}

// LogSystemAction logs a system-initiated action (no user context)
func (al *AuditLogger) LogSystemAction(action, entityType, entityID string, details map[string]interface{}) error {
	return al.LogAction(
		0,
		"system",
		action,
		entityType,
		entityID,
		details,
		nil,
		true,
		"",
	)
}

// LogLoginAttempt logs a login attempt (success or failure)
func (al *AuditLogger) LogLoginAttempt(email string, success bool, r *http.Request, userID int64, errorMsg string) error {
	action := database.ActionLoginSuccess
	if !success {
		action = database.ActionLoginFailed
	}

	details := map[string]interface{}{
		"email":   email,
		"success": success,
	}

	return al.LogAction(
		userID,
		email,
		action,
		database.EntitySession,
		"",
		details,
		r,
		success,
		errorMsg,
	)
}

// LogLogout logs a user logout
func (al *AuditLogger) LogLogout(user *models.User, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionLogout,
		database.EntitySession,
		"",
		map[string]interface{}{
			"email": user.Email,
		},
		r,
	)
}

// LogFileUpload logs a file upload
func (al *AuditLogger) LogFileUpload(user *models.User, fileID, fileName string, fileSize int64, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionFileUploaded,
		database.EntityFile,
		fileID,
		map[string]interface{}{
			"file_name": fileName,
			"file_size": fileSize,
		},
		r,
	)
}

// LogFileDelete logs a file deletion
func (al *AuditLogger) LogFileDelete(user *models.User, fileID, fileName string, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionFileDeleted,
		database.EntityFile,
		fileID,
		map[string]interface{}{
			"file_name": fileName,
		},
		r,
	)
}

// LogFileRestore logs a file restoration from trash
func (al *AuditLogger) LogFileRestore(user *models.User, fileID, fileName string, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionFileRestored,
		database.EntityFile,
		fileID,
		map[string]interface{}{
			"file_name": fileName,
		},
		r,
	)
}

// LogFilePermanentDelete logs permanent file deletion
func (al *AuditLogger) LogFilePermanentDelete(user *models.User, fileID, fileName string, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionFilePermanentlyDeleted,
		database.EntityFile,
		fileID,
		map[string]interface{}{
			"file_name": fileName,
		},
		r,
	)
}

// LogFileDownload logs a file download
func (al *AuditLogger) LogFileDownload(userID int64, userEmail, fileID, fileName string, fileSize int64, r *http.Request) error {
	return al.LogAction(
		userID,
		userEmail,
		database.ActionFileDownloaded,
		database.EntityFile,
		fileID,
		map[string]interface{}{
			"file_name": fileName,
			"file_size": fileSize,
		},
		r,
		true,
		"",
	)
}

// LogUserCreated logs user creation
func (al *AuditLogger) LogUserCreated(admin *models.User, newUserID int64, newUserEmail string, userLevel int, r *http.Request) error {
	return al.LogUserAction(
		admin,
		database.ActionUserCreated,
		database.EntityUser,
		fmt.Sprintf("%d", newUserID),
		map[string]interface{}{
			"email":      newUserEmail,
			"user_level": userLevel,
		},
		r,
	)
}

// LogUserUpdated logs user updates
func (al *AuditLogger) LogUserUpdated(admin *models.User, targetUserID int64, targetUserEmail string, changes map[string]interface{}, r *http.Request) error {
	return al.LogUserAction(
		admin,
		database.ActionUserUpdated,
		database.EntityUser,
		fmt.Sprintf("%d", targetUserID),
		map[string]interface{}{
			"email":   targetUserEmail,
			"changes": changes,
		},
		r,
	)
}

// LogUserDeleted logs user deletion
func (al *AuditLogger) LogUserDeleted(admin *models.User, deletedUserID int64, deletedUserEmail string, r *http.Request) error {
	return al.LogUserAction(
		admin,
		database.ActionUserDeleted,
		database.EntityUser,
		fmt.Sprintf("%d", deletedUserID),
		map[string]interface{}{
			"email": deletedUserEmail,
		},
		r,
	)
}

// LogUserActivated logs user activation
func (al *AuditLogger) LogUserActivated(admin *models.User, targetUserID int64, targetUserEmail string, r *http.Request) error {
	return al.LogUserAction(
		admin,
		database.ActionUserActivated,
		database.EntityUser,
		fmt.Sprintf("%d", targetUserID),
		map[string]interface{}{
			"email": targetUserEmail,
		},
		r,
	)
}

// LogUserDeactivated logs user deactivation
func (al *AuditLogger) LogUserDeactivated(admin *models.User, targetUserID int64, targetUserEmail string, r *http.Request) error {
	return al.LogUserAction(
		admin,
		database.ActionUserDeactivated,
		database.EntityUser,
		fmt.Sprintf("%d", targetUserID),
		map[string]interface{}{
			"email": targetUserEmail,
		},
		r,
	)
}

// LogTeamCreated logs team creation
func (al *AuditLogger) LogTeamCreated(user *models.User, teamID int64, teamName string, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionTeamCreated,
		database.EntityTeam,
		fmt.Sprintf("%d", teamID),
		map[string]interface{}{
			"team_name": teamName,
		},
		r,
	)
}

// LogTeamMemberAdded logs adding a member to a team
func (al *AuditLogger) LogTeamMemberAdded(user *models.User, teamID int64, teamName string, memberID int64, memberEmail string, role string, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionTeamMemberAdded,
		database.EntityTeam,
		fmt.Sprintf("%d", teamID),
		map[string]interface{}{
			"team_name":    teamName,
			"member_id":    memberID,
			"member_email": memberEmail,
			"role":         role,
		},
		r,
	)
}

// LogTeamMemberRemoved logs removing a member from a team
func (al *AuditLogger) LogTeamMemberRemoved(user *models.User, teamID int64, teamName string, memberID int64, memberEmail string, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionTeamMemberRemoved,
		database.EntityTeam,
		fmt.Sprintf("%d", teamID),
		map[string]interface{}{
			"team_name":    teamName,
			"member_id":    memberID,
			"member_email": memberEmail,
		},
		r,
	)
}

// LogFileSharedWithTeam logs file sharing with a team
func (al *AuditLogger) LogFileSharedWithTeam(user *models.User, fileID, fileName string, teamID int64, teamName string, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionFileSharedWithTeam,
		database.EntityFile,
		fileID,
		map[string]interface{}{
			"file_name": fileName,
			"team_id":   teamID,
			"team_name": teamName,
		},
		r,
	)
}

// LogSettingsUpdated logs settings changes
func (al *AuditLogger) LogSettingsUpdated(user *models.User, changes map[string]interface{}, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionSettingsUpdated,
		database.EntitySettings,
		"",
		changes,
		r,
	)
}

// LogPasswordChanged logs password change
func (al *AuditLogger) LogPasswordChanged(user *models.User, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.ActionPasswordChanged,
		database.EntityUser,
		fmt.Sprintf("%d", user.Id),
		map[string]interface{}{
			"email": user.Email,
		},
		r,
	)
}

// Log2FAEnabled logs 2FA enablement
func (al *AuditLogger) Log2FAEnabled(user *models.User, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.Action2FAEnabled,
		database.EntityUser,
		fmt.Sprintf("%d", user.Id),
		map[string]interface{}{
			"email": user.Email,
		},
		r,
	)
}

// Log2FADisabled logs 2FA disablement
func (al *AuditLogger) Log2FADisabled(user *models.User, r *http.Request) error {
	return al.LogUserAction(
		user,
		database.Action2FADisabled,
		database.EntityUser,
		fmt.Sprintf("%d", user.Id),
		map[string]interface{}{
			"email": user.Email,
		},
		r,
	)
}

// getClientIP extracts the real client IP from request
func (al *AuditLogger) getClientIP(r *http.Request) string {
	// Try X-Forwarded-For first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Try X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}
