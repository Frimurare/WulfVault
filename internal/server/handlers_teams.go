// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
	"github.com/Frimurare/WulfVault/internal/models"
)

// handleAdminTeams displays the team management page (Admin only)
func (s *Server) handleAdminTeams(w http.ResponseWriter, r *http.Request) {
	_, _ = userFromContext(r.Context())

	teams, err := database.DB.GetAllTeams()
	if err != nil {
		log.Printf("Error fetching teams: %v", err)
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	// Get member count for each team
	var teamInfos []struct {
		*models.Team
		MemberCount int
	}
	for _, team := range teams {
		members, _ := database.DB.GetTeamMembers(team.Id)
		teamInfos = append(teamInfos, struct {
			*models.Team
			MemberCount int
		}{
			Team:        team,
			MemberCount: len(members),
		})
	}

	s.renderAdminTeams(w, teamInfos)
}

// handleAPITeamCreate creates a new team (Admin only)
func (s *Server) handleAPITeamCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	var req struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		StorageQuotaMB int64  `json:"storageQuotaMB"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	if req.StorageQuotaMB == 0 {
		req.StorageQuotaMB = 10240 // Default 10GB
	}

	team := &models.Team{
		Name:           req.Name,
		Description:    req.Description,
		CreatedBy:      user.Id,
		StorageQuotaMB: req.StorageQuotaMB,
		IsActive:       true,
	}

	if err := database.DB.CreateTeam(team); err != nil {
		log.Printf("Error creating team: %v", err)
		http.Error(w, "Error creating team", http.StatusInternalServerError)
		return
	}

	// Log the action
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "TEAM_CREATED",
		EntityType: "Team",
		EntityID:   fmt.Sprintf("%d", team.Id),
		Details:    fmt.Sprintf("{\"name\":\"%s\",\"storage_quota_mb\":%d}", team.Name, team.StorageQuotaMB),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"team":    team,
	})
}

// handleAPITeamUpdate updates a team (Admin only)
func (s *Server) handleAPITeamUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TeamId         int    `json:"teamId"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		StorageQuotaMB int64  `json:"storageQuotaMB"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	team, err := database.DB.GetTeamByID(req.TeamId)
	if err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	team.Name = req.Name
	team.Description = req.Description
	team.StorageQuotaMB = req.StorageQuotaMB

	if err := database.DB.UpdateTeam(team); err != nil {
		log.Printf("Error updating team: %v", err)
		http.Error(w, "Error updating team", http.StatusInternalServerError)
		return
	}

	// Log the action
	user, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "TEAM_UPDATED",
		EntityType: "Team",
		EntityID:   fmt.Sprintf("%d", team.Id),
		Details:    fmt.Sprintf("{\"name\":\"%s\",\"storage_quota_mb\":%d}", team.Name, team.StorageQuotaMB),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"team":    team,
	})
}

// handleAPITeamDelete deletes a team (Admin only)
func (s *Server) handleAPITeamDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TeamId int `json:"teamId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get team details before deletion for audit log
	team, err := database.DB.GetTeamByID(req.TeamId)
	if err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	if err := database.DB.DeleteTeam(req.TeamId); err != nil {
		log.Printf("Error deleting team: %v", err)
		http.Error(w, "Error deleting team", http.StatusInternalServerError)
		return
	}

	// Log the action
	user, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "TEAM_DELETED",
		EntityType: "Team",
		EntityID:   fmt.Sprintf("%d", req.TeamId),
		Details:    fmt.Sprintf("{\"name\":\"%s\"}", team.Name),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleAPITeamMembers returns all members of a team
func (s *Server) handleAPITeamMembers(w http.ResponseWriter, r *http.Request) {
	teamIdStr := r.URL.Query().Get("teamId")
	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	user, _ := userFromContext(r.Context())

	// Check if user is admin or team member
	if !user.IsAdmin() {
		isMember, err := database.DB.IsTeamMember(teamId, user.Id)
		if err != nil || !isMember {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
	}

	members, err := database.DB.GetTeamMembers(teamId)
	if err != nil {
		log.Printf("Error fetching team members: %v", err)
		http.Error(w, "Error fetching members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"members": members,
	})
}

// handleAPITeamAddMember adds a user to a team
func (s *Server) handleAPITeamAddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	var req struct {
		TeamId int `json:"teamId"`
		UserId int `json:"userId"`
		Role   int `json:"role"` // 0=Owner, 1=Admin, 2=Member
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check permission: admin OR team owner/admin
	canManage := false
	if user.IsAdmin() {
		canManage = true
	} else {
		member, err := database.DB.GetTeamMember(req.TeamId, user.Id)
		if err == nil && member.CanManageMembers() {
			canManage = true
		}
	}

	if !canManage {
		http.Error(w, "You don't have permission to add members", http.StatusForbidden)
		return
	}

	// Check if user to add exists
	targetUser, err := database.DB.GetUserByID(req.UserId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Don't add download users to teams
	if targetUser.UserLevel > models.UserLevelUser {
		http.Error(w, "Download users cannot be added to teams", http.StatusBadRequest)
		return
	}

	// Add member
	member := &models.TeamMember{
		TeamId:  req.TeamId,
		UserId:  req.UserId,
		Role:    models.TeamRole(req.Role),
		AddedBy: user.Id,
	}

	if err := database.DB.AddTeamMember(member); err != nil {
		log.Printf("Error adding team member: %v", err)
		http.Error(w, "Error adding member (user may already be in team)", http.StatusInternalServerError)
		return
	}

	// Log the action
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "TEAM_MEMBER_ADDED",
		EntityType: "TeamMember",
		EntityID:   fmt.Sprintf("%d", req.TeamId),
		Details:    fmt.Sprintf("{\"team_id\":%d,\"user_id\":%d,\"user_email\":\"%s\",\"role\":%d}", req.TeamId, req.UserId, targetUser.Email, req.Role),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	// Send invitation email
	team, _ := database.DB.GetTeamByID(req.TeamId)
	if team != nil {
		companyName := s.config.CompanyName
		if companyName == "" {
			companyName = "WulfVault"
		}
		if err := email.SendTeamInvitationEmail(targetUser.Email, team.Name, s.config.ServerURL, companyName); err != nil {
			log.Printf("Warning: Failed to send team invitation email: %v", err)
			// Don't fail the request if email fails
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"member":  member,
	})
}

// handleAPITeamRemoveMember removes a user from a team
func (s *Server) handleAPITeamRemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	var req struct {
		TeamId int `json:"teamId"`
		UserId int `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check permission: admin OR team owner/admin
	canManage := false
	if user.IsAdmin() {
		canManage = true
	} else {
		member, err := database.DB.GetTeamMember(req.TeamId, user.Id)
		if err == nil && member.CanManageMembers() {
			canManage = true
		}
	}

	if !canManage {
		http.Error(w, "You don't have permission to remove members", http.StatusForbidden)
		return
	}

	// Get user details for audit log
	removedUser, err := database.DB.GetUserByID(req.UserId)
	removedUserEmail := "unknown"
	if err == nil {
		removedUserEmail = removedUser.Email
	}

	if err := database.DB.RemoveTeamMember(req.TeamId, req.UserId); err != nil {
		log.Printf("Error removing team member: %v", err)
		http.Error(w, "Error removing member", http.StatusInternalServerError)
		return
	}

	// Log the action
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     "TEAM_MEMBER_REMOVED",
		EntityType: "TeamMember",
		EntityID:   fmt.Sprintf("%d", req.TeamId),
		Details:    fmt.Sprintf("{\"team_id\":%d,\"user_id\":%d,\"user_email\":\"%s\"}", req.TeamId, req.UserId, removedUserEmail),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    true,
		ErrorMsg:   "",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleAPIShareFileToTeam shares a file with a team
func (s *Server) handleAPIShareFileToTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	var req struct {
		FileId string `json:"file_id"`
		TeamId int    `json:"team_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check if user is team member
	isMember, err := database.DB.IsTeamMember(req.TeamId, user.Id)
	if err != nil || !isMember {
		http.Error(w, "You must be a team member to share files", http.StatusForbidden)
		return
	}

	// Check if user owns the file
	file, err := database.DB.GetFileByID(req.FileId)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if file.UserId != user.Id && !user.HasPermissionEditOtherUploads() {
		http.Error(w, "You don't own this file", http.StatusForbidden)
		return
	}

	// Share file to team
	if err := database.DB.ShareFileToTeam(req.FileId, req.TeamId, user.Id); err != nil {
		log.Printf("Error sharing file to team: %v", err)
		http.Error(w, "Error sharing file (file may already be shared)", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleAPIUnshareFileFromTeam removes a file from a team
func (s *Server) handleAPIUnshareFileFromTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := userFromContext(r.Context())

	var req struct {
		FileId string `json:"file_id"`
		TeamId int    `json:"team_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check if user is team member or admin
	if !user.IsAdmin() {
		isMember, err := database.DB.IsTeamMember(req.TeamId, user.Id)
		if err != nil || !isMember {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Also check if user owns the file
		file, err := database.DB.GetFileByID(req.FileId)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		if file.UserId != user.Id {
			http.Error(w, "You don't own this file", http.StatusForbidden)
			return
		}
	}

	if err := database.DB.UnshareFileFromTeam(req.FileId, req.TeamId); err != nil {
		log.Printf("Error unsharing file from team: %v", err)
		http.Error(w, "Error unsharing file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleUserTeams displays user's teams page
func (s *Server) handleUserTeams(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	// Check if viewing a specific team's files
	teamIdStr := r.URL.Query().Get("id")
	if teamIdStr != "" {
		teamId, err := strconv.Atoi(teamIdStr)
		if err != nil {
			http.Error(w, "Invalid team ID", http.StatusBadRequest)
			return
		}

		// Verify user is team member or admin
		if !user.IsAdmin() {
			isMember, err := database.DB.IsTeamMember(teamId, user.Id)
			if err != nil || !isMember {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
		}

		// Get team info
		team, err := database.DB.GetTeamByID(teamId)
		if err != nil {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}

		s.renderTeamFiles(w, user, team)
		return
	}

	// Show all teams
	teams, err := database.DB.GetTeamsByUser(user.Id)
	if err != nil {
		log.Printf("Error fetching user teams: %v", err)
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	s.renderUserTeams(w, user, teams)
}

// handleAPITeamFiles returns all files shared with a team
func (s *Server) handleAPITeamFiles(w http.ResponseWriter, r *http.Request) {
	teamIdStr := r.URL.Query().Get("teamId")
	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	user, _ := userFromContext(r.Context())

	// Check if user is team member or admin
	if !user.IsAdmin() {
		isMember, err := database.DB.IsTeamMember(teamId, user.Id)
		if err != nil || !isMember {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
	}

	teamFiles, err := database.DB.GetTeamFiles(teamId)
	if err != nil {
		log.Printf("Error fetching team files: %v", err)
		http.Error(w, "Error fetching files", http.StatusInternalServerError)
		return
	}

	// Get full file info for each team file
	var files []map[string]interface{}
	for _, tf := range teamFiles {
		file, err := database.DB.GetFileByID(tf.FileId)
		if err != nil {
			continue
		}

		// Get file owner info
		owner, _ := database.DB.GetUserByID(file.UserId)
		ownerName := "Unknown"
		if owner != nil {
			ownerName = owner.Name
		}

		files = append(files, map[string]interface{}{
			"file":       file,
			"sharedBy":   tf.SharedBy,
			"sharedAt":   tf.SharedAt,
			"ownerName":  ownerName,
			"teamFileId": tf.Id,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"files":   files,
	})
}

// handleAPIMyTeams returns all teams the current user is a member of
func (s *Server) handleAPIMyTeams(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())

	teams, err := database.DB.GetTeamsByUser(user.Id)
	if err != nil {
		log.Printf("Error fetching user teams: %v", err)
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"teams":   teams,
	})
}

// handleAPIFileTeams returns all teams that have access to a specific file
func (s *Server) handleAPIFileTeams(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	fileId := r.URL.Query().Get("file_id")

	if fileId == "" {
		http.Error(w, "file_id is required", http.StatusBadRequest)
		return
	}

	// Verify user can access this file
	file, err := database.DB.GetFileByID(fileId)
	if err != nil || file == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Only file owner can see teams
	if file.UserId != user.Id && !user.IsAdmin() {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	teams, err := database.DB.GetTeamsForFile(fileId)
	if err != nil {
		log.Printf("Error fetching file teams: %v", err)
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"teams":   teams,
	})
}

// renderAdminTeams renders the admin teams management page
func (s *Server) renderAdminTeams(w http.ResponseWriter, teams []struct {
	*models.Team
	MemberCount int
}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Manage Teams - ` + s.config.CompanyName + `</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .container {
            max-width: 1400px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .actions {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 24px;
        }
        .btn {
            padding: 12px 24px;
            background: ` + s.getPrimaryColor() + `;
            color: white;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 500;
            cursor: pointer;
            border: none;
            font-size: 14px;
        }
        .btn:hover {
            opacity: 0.9;
        }
        .btn-danger {
            background: #dc2626;
        }
        .btn-secondary {
            background: #6b7280;
            margin-left: 8px;
        }
        table {
            width: 100%;
            background: white;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        th, td {
            padding: 16px;
            text-align: left;
        }
        th {
            background: #f9f9f9;
            font-weight: 600;
            color: #666;
        }
        tr:not(:last-child) td {
            border-bottom: 1px solid #e0e0e0;
        }
        .badge {
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 500;
        }
        .badge-active { background: #d1fae5; color: #065f46; }
        .badge-inactive { background: #fee2e2; color: #991b1b; }
        .action-links a, .action-links button {
            margin-right: 12px;
            color: ` + s.getPrimaryColor() + `;
            text-decoration: none;
            cursor: pointer;
            background: none;
            border: none;
            font-size: 14px;
        }
        .action-links button:hover {
            text-decoration: underline;
        }
        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            align-items: center;
            justify-content: center;
            z-index: 1000;
        }
        .modal.active {
            display: flex;
        }
        .modal-content {
            background: white;
            padding: 32px;
            border-radius: 12px;
            max-width: 500px;
            width: 90%;
            max-height: 90vh;
            overflow-y: auto;
        }
        .modal-content h2 {
            margin-bottom: 20px;
        }
        #membersList {
            max-height: 400px;
            overflow-y: auto;
            margin: 20px 0;
        }
        .form-group {
            margin-bottom: 16px;
        }
        .form-group label {
            display: block;
            margin-bottom: 6px;
            font-weight: 500;
            color: #374151;
        }
        .form-group input, .form-group textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #d1d5db;
            border-radius: 6px;
            font-size: 14px;
        }
        .form-group textarea {
            resize: vertical;
            min-height: 80px;
        }
        .modal-actions {
            display: flex;
            gap: 12px;
            justify-content: flex-end;
            margin-top: 24px;
        }
        .storage-bar {
            width: 100%;
            height: 6px;
            background: #e5e7eb;
            border-radius: 3px;
            overflow: hidden;
            margin-top: 4px;
        }
        .storage-bar-fill {
            height: 100%;
            background: ` + s.getPrimaryColor() + `;
            transition: width 0.3s;
        }

        /* Mobile Responsive Styles */
        @media screen and (max-width: 768px) {
            .container {
                padding: 0 15px !important;
            }
            .actions {
                flex-direction: column;
                align-items: stretch !important;
                gap: 15px;
            }
            .actions h2 {
                font-size: 20px;
            }
            .btn {
                width: 100%;
                text-align: center;
            }
            table {
                border: 0;
                display: block;
                overflow-x: auto;
            }
            table thead {
                display: none;
            }
            table tbody {
                display: block;
            }
            table tr {
                display: block;
                margin-bottom: 20px;
                border: 1px solid #ddd;
                border-radius: 8px;
                padding: 15px;
                background: white;
            }
            table td {
                display: block;
                text-align: right;
                padding: 8px 0;
                border-bottom: 1px solid #eee;
            }
            table td:last-child {
                border-bottom: none;
            }
            table td::before {
                content: attr(data-label);
                float: left;
                font-weight: 600;
                color: #666;
            }
            .action-links {
                display: flex;
                flex-direction: column;
                gap: 8px;
            }
            .action-links button {
                margin: 0 !important;
                padding: 8px 12px;
                background: #f0f0f0;
                border-radius: 4px;
                text-align: center;
                display: block;
                width: 100%;
            }
            .modal-content {
                max-width: 95%;
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `
    <div class="container">
        <div class="actions">
            <h2>📁 Manage Teams</h2>
            <button class="btn" onclick="showCreateModal()">+ Create Team</button>
        </div>

        <table>
            <thead>
                <tr>
                    <th>Team Name</th>
                    <th>Description</th>
                    <th>Members</th>
                    <th>Storage</th>
                    <th>Created</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody id="teamsTable">`

	if len(teams) == 0 {
		html += `
                <tr>
                    <td colspan="7" style="text-align: center; padding: 40px; color: #666;">
                        No teams yet. Click "Create Team" to get started.
                    </td>
                </tr>`
	} else {
		for _, team := range teams {
			statusBadge := `<span class="badge badge-active">Active</span>`
			if !team.IsActive {
				statusBadge = `<span class="badge badge-inactive">Inactive</span>`
			}

			storagePercent := team.GetStoragePercentage()
			storageUsed := fmt.Sprintf("%.1f GB", float64(team.StorageUsedMB)/1024)
			storageTotal := fmt.Sprintf("%.1f GB", float64(team.StorageQuotaMB)/1024)

			html += fmt.Sprintf(`
                <tr>
                    <td data-label="Team Name"><strong>%s</strong></td>
                    <td data-label="Description">%s</td>
                    <td data-label="Members">%d members</td>
                    <td data-label="Storage">
                        %s / %s (%d%%)
                        <div class="storage-bar">
                            <div class="storage-bar-fill" style="width: %d%%"></div>
                        </div>
                    </td>
                    <td data-label="Created">%s</td>
                    <td data-label="Status">%s</td>
                    <td data-label="Actions" class="action-links">
                        <button onclick="window.location.href='/teams?id=%d'">📁 Files</button>
                        <button onclick="viewMembers(%d, '%s')">👥 Members</button>
                        <button onclick="editTeam(%d)">✏️ Edit</button>
                        <button onclick="deleteTeam(%d, '%s')">🗑️ Delete</button>
                    </td>
                </tr>`,
				team.Name, team.Description, team.MemberCount,
				storageUsed, storageTotal, storagePercent, storagePercent,
				team.GetReadableCreatedAt(), statusBadge,
				team.Id, team.Id, team.Name, team.Id, team.Id, team.Name)
		}
	}

	html += `
            </tbody>
        </table>
    </div>

    <!-- Create/Edit Team Modal -->
    <div id="teamModal" class="modal">
        <div class="modal-content">
            <h2 id="modalTitle">Create Team</h2>
            <div class="form-group">
                <label for="teamName">Team Name *</label>
                <input type="text" id="teamName" placeholder="e.g., Prudencia" required>
            </div>
            <div class="form-group">
                <label for="teamDescription">Description</label>
                <textarea id="teamDescription" placeholder="Optional description"></textarea>
            </div>
            <div class="form-group">
                <label for="teamQuota">Storage Quota (MB) *</label>
                <input type="number" id="teamQuota" value="10240" min="1" required>
                <small style="color: #666;">Default: 10240 MB (10 GB)</small>
            </div>
            <div class="modal-actions">
                <button class="btn btn-secondary" onclick="closeModal()">Cancel</button>
                <button class="btn" onclick="saveTeam()">Save</button>
            </div>
        </div>
    </div>

    <!-- Members Modal -->
    <div id="membersModal" class="modal">
        <div class="modal-content" style="max-width: 700px;">
            <h2 id="membersTitle">Team Members</h2>
            <div id="membersList" style="margin: 20px 0;"></div>
            <div class="modal-actions">
                <button class="btn btn-secondary" onclick="closeMembersModal()">Close</button>
                <button class="btn" onclick="showAddMemberForm()">+ Add Member</button>
            </div>
        </div>
    </div>

    <!-- Add Member Modal -->
    <div id="addMemberModal" class="modal">
        <div class="modal-content">
            <h2>Add Team Member</h2>
            <input type="hidden" id="addMemberTeamId">
            <div class="form-group">
                <label>Select User</label>
                <select id="userSelect" style="width: 100%; padding: 10px; border: 1px solid #d1d5db; border-radius: 6px;">
                    <option value="">Loading users...</option>
                </select>
            </div>
            <div class="form-group">
                <label>Role</label>
                <select id="roleSelect" style="width: 100%; padding: 10px; border: 1px solid #d1d5db; border-radius: 6px;">
                    <option value="2">Member</option>
                    <option value="1">Admin</option>
                    <option value="0">Owner</option>
                </select>
            </div>
            <div class="modal-actions">
                <button class="btn btn-secondary" onclick="closeAddMemberModal()">Cancel</button>
                <button class="btn" onclick="addMemberToTeam()">Add Member</button>
            </div>
        </div>
    </div>

    <script>
        let currentTeamId = null;

        function showCreateModal() {
            document.getElementById('modalTitle').textContent = 'Create Team';
            document.getElementById('teamName').value = '';
            document.getElementById('teamDescription').value = '';
            document.getElementById('teamQuota').value = '10240';
            currentTeamId = null;
            document.getElementById('teamModal').classList.add('active');
        }

        function editTeam(teamId) {
            // Fetch team data and populate form
            fetch('/api/teams/members?teamId=' + teamId)
                .then(r => r.json())
                .then(data => {
                    // For now, just show a basic edit form
                    document.getElementById('modalTitle').textContent = 'Edit Team';
                    currentTeamId = teamId;
                    document.getElementById('teamModal').classList.add('active');
                });
        }

        function closeModal() {
            document.getElementById('teamModal').classList.remove('active');
        }

        function saveTeam() {
            const name = document.getElementById('teamName').value.trim();
            const description = document.getElementById('teamDescription').value.trim();
            const quota = parseInt(document.getElementById('teamQuota').value);

            if (!name) {
                alert('Team name is required');
                return;
            }

            const url = currentTeamId ? '/api/admin/teams/update' : '/api/admin/teams/create';
            const body = {
                name: name,
                description: description,
                storageQuotaMB: quota
            };

            if (currentTeamId) {
                body.teamId = currentTeamId;
            }

            fetch(url, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(body)
            })
            .then(r => r.json())
            .then(data => {
                if (data.success) {
                    alert('Team saved successfully!');
                    location.reload();
                } else {
                    alert('Error saving team');
                }
            })
            .catch(err => {
                alert('Error: ' + err.message);
            });
        }

        function deleteTeam(teamId, teamName) {
            if (!confirm('Are you sure you want to delete team "' + teamName + '"?')) {
                return;
            }

            fetch('/api/admin/teams/delete', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({teamId: teamId})
            })
            .then(r => r.json())
            .then(data => {
                if (data.success) {
                    alert('Team deleted successfully!');
                    location.reload();
                } else {
                    alert('Error deleting team');
                }
            });
        }

        function viewMembers(teamId, teamName) {
            document.getElementById('membersTitle').textContent = 'Members of ' + teamName;

            // Store team ID for add member functionality
            currentAddMemberTeamId = teamId;
            document.getElementById('addMemberTeamId').value = teamId;

            fetch('/api/teams/members?teamId=' + teamId)
                .then(r => r.json())
                .then(data => {
                    let html = '<table style="width: 100%;"><thead><tr><th>Name</th><th>Email</th><th>Role</th><th>Joined</th><th>Action</th></tr></thead><tbody>';

                    if (data.members && data.members.length > 0) {
                        data.members.forEach(m => {
                            const role = m.role === 0 ? 'Owner' : m.role === 1 ? 'Admin' : 'Member';
                            const joinedDate = new Date(m.joinedAt * 1000).toLocaleDateString();
                            html += '<tr><td>' + m.userName + '</td><td>' + m.userEmail + '</td><td>' + role + '</td><td>' + joinedDate + '</td><td><button onclick="removeMember(' + teamId + ', ' + m.userId + ', \'' + m.userName + '\')">Remove</button></td></tr>';
                        });
                    } else {
                        html += '<tr><td colspan="5" style="text-align: center; padding: 20px;">No members yet</td></tr>';
                    }

                    html += '</tbody></table>';
                    document.getElementById('membersList').innerHTML = html;
                    document.getElementById('membersModal').classList.add('active');
                });
        }

        function closeMembersModal() {
            document.getElementById('membersModal').classList.remove('active');
        }

        let currentAddMemberTeamId = null;

        function showAddMemberForm() {
            // Get the current team ID from the members modal title or store it when opening
            const teamId = document.getElementById('addMemberTeamId').value || currentAddMemberTeamId;
            if (!teamId) {
                alert('Error: Team ID not found');
                return;
            }

            // Load all users
            fetch('/api/admin/users/list')
                .then(r => r.json())
                .then(data => {
                    const select = document.getElementById('userSelect');
                    select.innerHTML = '<option value="">-- Select a user --</option>';

                    if (data.users && data.users.length > 0) {
                        data.users.forEach(user => {
                            const option = document.createElement('option');
                            option.value = user.id;
                            option.textContent = user.name + ' (' + user.email + ')';
                            select.appendChild(option);
                        });
                    }

                    document.getElementById('addMemberModal').classList.add('active');
                })
                .catch(err => {
                    alert('Error loading users: ' + err.message);
                });
        }

        function closeAddMemberModal() {
            document.getElementById('addMemberModal').classList.remove('active');
        }

        function addMemberToTeam() {
            const teamId = document.getElementById('addMemberTeamId').value || currentAddMemberTeamId;
            const userId = parseInt(document.getElementById('userSelect').value);
            const role = parseInt(document.getElementById('roleSelect').value);

            if (!userId) {
                alert('Please select a user');
                return;
            }

            fetch('/api/teams/add-member', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    teamId: parseInt(teamId),
                    userId: userId,
                    role: role
                })
            })
            .then(r => r.json())
            .then(data => {
                if (data.success) {
                    alert('Member added successfully!');
                    closeAddMemberModal();
                    closeMembersModal();
                    location.reload();
                } else {
                    alert('Error: ' + (data.error || 'Failed to add member'));
                }
            })
            .catch(err => {
                alert('Error: ' + err.message);
            });
        }

        function removeMember(teamId, userId, userName) {
            if (!confirm('Remove ' + userName + ' from this team?')) {
                return;
            }

            fetch('/api/teams/remove-member', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({teamId: teamId, userId: userId})
            })
            .then(r => r.json())
            .then(data => {
                if (data.success) {
                    alert('Member removed!');
                    closeMembersModal();
                    location.reload();
                }
            });
        }
    </script>
    
</body>
</html>`

	w.Write([]byte(html))
}

// renderUserTeams renders the user teams page
func (s *Server) renderUserTeams(w http.ResponseWriter, user *models.User, teams []*models.TeamWithMembers) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>My Teams - ` + s.config.CompanyName + `</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .container {
            max-width: 1400px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .teams-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 24px;
            margin-top: 24px;
        }
        .team-card {
            background: white;
            border-radius: 12px;
            padding: 24px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            cursor: pointer;
            transition: transform 0.2s;
        }
        .team-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        .team-card h3 {
            color: #111;
            margin-bottom: 8px;
        }
        .team-card p {
            color: #666;
            font-size: 14px;
            margin-bottom: 16px;
        }
        .team-stats {
            display: flex;
            gap: 16px;
            font-size: 13px;
            color: #666;
        }
        .badge {
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 500;
            background: #e0e7ff;
            color: #3730a3;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #666;
        }
        .empty-state h3 {
            margin-bottom: 12px;
            color: #333;
        }

        @media screen and (max-width: 768px) {
            .container {
                padding: 0 15px !important;
            }
            .teams-grid {
                grid-template-columns: 1fr !important;
            }
        }
    </style>
</head>
<body>
    ` + s.getHeaderHTML(user, user.IsAdmin()) + `

    <div class="container">
        <h2 style="margin-bottom: 8px;">👥 My Teams</h2>
        <p style="color: #666; margin-bottom: 24px;">Teams you're a member of</p>

        <div class="teams-grid">`

	if len(teams) == 0 {
		html += `
            <div class="empty-state" style="grid-column: 1 / -1;">
                <h3>No teams yet</h3>
                <p>You haven't been added to any teams yet. Contact your administrator to get started.</p>
            </div>`
	} else {
		for _, team := range teams {
			roleText := "Member"
			if team.UserRole == models.TeamRoleOwner {
				roleText = "Owner"
			} else if team.UserRole == models.TeamRoleAdmin {
				roleText = "Admin"
			}

			storageUsed := fmt.Sprintf("%.1f GB", float64(team.StorageUsedMB)/1024)
			storageTotal := fmt.Sprintf("%.1f GB", float64(team.StorageQuotaMB)/1024)

			html += fmt.Sprintf(`
            <div class="team-card" onclick="viewTeamFiles(%d, '%s')">
                <h3>%s</h3>
                <p>%s</p>
                <div class="team-stats">
                    <span>👤 %d members</span>
                    <span>💾 %s / %s</span>
                    <span class="badge">%s</span>
                </div>
            </div>`,
				team.Id, team.Name, team.Name, team.Description,
				team.MemberCount, storageUsed, storageTotal, roleText)
		}
	}

	html += `
        </div>
    </div>

    <script>
        function viewTeamFiles(teamId, teamName) {
            window.location.href = '/teams?id=' + teamId;
        }
    </script>
    
</body>
</html>`

	w.Write([]byte(html))
}

// renderTeamFiles displays all files shared with a specific team
func (s *Server) renderTeamFiles(w http.ResponseWriter, user *models.User, team *models.Team) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get team files
	teamFiles, err := database.DB.GetTeamFiles(team.Id)
	if err != nil {
		log.Printf("Error fetching team files: %v", err)
		teamFiles = []*models.TeamFile{}
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>` + team.Name + ` - Files - WulfVault</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .container {
            max-width: 1400px;
            margin: 40px auto;
            padding: 0 20px;
        }
        .page-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 24px;
        }
        .page-header h2 {
            color: #1a1a2e;
            font-size: 28px;
        }
        .page-header .subtitle {
            color: #666;
            font-size: 15px;
            margin-top: 4px;
        }
        .back-btn {
            background: #e0e0e0;
            color: #333;
            padding: 10px 20px;
            border: none;
            border-radius: 6px;
            text-decoration: none;
            font-weight: 500;
            cursor: pointer;
            transition: background 0.2s;
        }
        .back-btn:hover {
            background: #d0d0d0;
        }
        .files-table {
            background: white;
            border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.08);
            overflow: hidden;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        thead {
            background: #f8f9fa;
            border-bottom: 2px solid #e0e0e0;
        }
        th {
            padding: 16px;
            text-align: left;
            font-weight: 600;
            color: #1a1a2e;
            font-size: 13px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        td {
            padding: 16px;
            border-top: 1px solid #f0f0f0;
            color: #333;
        }
        tr:hover {
            background: #f9f9f9;
        }
        .file-name {
            font-weight: 500;
            color: ` + s.getPrimaryColor() + `;
        }
        .file-icon {
            margin-right: 8px;
        }
        .btn-download {
            background: ` + s.getPrimaryColor() + `;
            color: white;
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            text-decoration: none;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: opacity 0.2s;
            display: inline-block;
        }
        .btn-download:hover {
            opacity: 0.9;
        }
        .empty-state {
            text-align: center;
            padding: 80px 20px;
            color: #666;
        }
        .empty-state h3 {
            margin-bottom: 12px;
            color: #333;
            font-size: 20px;
        }
        .empty-state p {
            font-size: 15px;
        }

        @media screen and (max-width: 768px) {
            .container {
                padding: 0 15px !important;
            }
            .page-header {
                flex-direction: column;
                align-items: flex-start !important;
                gap: 15px;
            }
            .back-btn {
                width: 100%;
                text-align: center;
            }
            table thead {
                display: none;
            }
            table, table tbody, table tr {
                display: block;
                width: 100%;
            }
            table tr {
                margin-bottom: 15px;
                border: 1px solid #e0e0e0;
                border-radius: 8px;
                padding: 15px;
                background: white;
            }
            table td {
                display: block;
                text-align: left;
                padding: 12px 0;
                border: none;
                position: relative;
                min-height: 35px;
            }
            table td::before {
                content: attr(data-label);
                display: block;
                font-weight: 600;
                color: #666;
                margin-bottom: 4px;
                font-size: 13px;
            }
            table td:last-child {
                padding-left: 0;
                text-align: center;
                padding-top: 15px;
                margin-top: 10px;
                border-top: 1px solid #e0e0e0;
            }
            table td:last-child::before {
                display: none;
            }
        }
    </style>
</head>
<body>
    ` + s.getHeaderHTML(user, user.IsAdmin()) + `

    <div class="container">
        <div class="page-header">
            <div>
                <h2>📁 ` + team.Name + ` - Shared Files</h2>
                <p class="subtitle">Files shared with this team</p>
            </div>
            <a href="/teams" class="back-btn">← Back to Teams</a>
        </div>`

	if len(teamFiles) == 0 {
		html += `
        <div class="empty-state">
            <h3>No files shared yet</h3>
            <p>Files shared with this team will appear here</p>
        </div>`
	} else {
		html += `
        <div class="files-table">
            <table>
                <thead>
                    <tr>
                        <th>File Name</th>
                        <th>Owner</th>
                        <th>Shared By</th>
                        <th>Shared Date</th>
                        <th>Size</th>
                        <th>Downloads</th>
                        <th>Action</th>
                    </tr>
                </thead>
                <tbody>`

		for _, tf := range teamFiles {
			file, err := database.DB.GetFileByID(tf.FileId)
			if err != nil {
				continue
			}

			// Get file owner info
			owner, _ := database.DB.GetUserByID(file.UserId)
			ownerName := "Unknown"
			if owner != nil {
				ownerName = owner.Name
			}

			// Get shared by user info
			sharedByUser, _ := database.DB.GetUserByID(tf.SharedBy)
			sharedByName := "Unknown"
			if sharedByUser != nil {
				sharedByName = sharedByUser.Name
			}

			// Format file size
			var sizeStr string
			if file.SizeBytes < 1024 {
				sizeStr = fmt.Sprintf("%d B", file.SizeBytes)
			} else if file.SizeBytes < 1024*1024 {
				sizeStr = fmt.Sprintf("%.1f KB", float64(file.SizeBytes)/1024)
			} else if file.SizeBytes < 1024*1024*1024 {
				sizeStr = fmt.Sprintf("%.1f MB", float64(file.SizeBytes)/(1024*1024))
			} else {
				sizeStr = fmt.Sprintf("%.1f GB", float64(file.SizeBytes)/(1024*1024*1024))
			}

			// Format shared date
			sharedTime := time.Unix(tf.SharedAt, 0)
			sharedDate := sharedTime.Format("2006-01-02 15:04")

			html += fmt.Sprintf(`
                    <tr>
                        <td data-label="File Name"><span class="file-icon">📄</span><span class="file-name">%s</span></td>
                        <td data-label="Owner">%s</td>
                        <td data-label="Shared By">%s</td>
                        <td data-label="Shared Date">%s</td>
                        <td data-label="Size">%s</td>
                        <td data-label="Downloads">%d</td>
                        <td data-label="Action"><a href="/d/%s" class="btn-download">Download</a></td>
                    </tr>`, file.Name, ownerName, sharedByName, sharedDate, sizeStr, file.DownloadCount, file.Id)
		}

		html += `
                </tbody>
            </table>
        </div>`
	}

	html += `
    </div>
    
</body>
</html>`

	w.Write([]byte(html))
}
