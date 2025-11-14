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

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/email"
	"github.com/Frimurare/WulfVault/internal/models"
)

// handleAdminTeams displays the team management page (Admin only)
func (s *Server) handleAdminTeams(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())

	teams, err := database.DB.GetAllTeams()
	if err != nil {
		log.Printf("Error fetching teams: %v", err)
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	// Get member count for each team
	type TeamInfo struct {
		*models.Team
		MemberCount int
	}
	var teamInfos []TeamInfo
	for _, team := range teams {
		members, _ := database.DB.GetTeamMembers(team.Id)
		teamInfos = append(teamInfos, TeamInfo{
			Team:        team,
			MemberCount: len(members),
		})
	}

	data := map[string]interface{}{
		"User":  user,
		"Teams": teamInfos,
	}

	renderTemplate(w, "admin-teams", data)
}

// handleAPITeamCreate creates a new team (Admin only)
func (s *Server) handleAPITeamCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromContext(r.Context())

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

	if err := database.DB.DeleteTeam(req.TeamId); err != nil {
		log.Printf("Error deleting team: %v", err)
		http.Error(w, "Error deleting team", http.StatusInternalServerError)
		return
	}

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

	user := getUserFromContext(r.Context())

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

	user := getUserFromContext(r.Context())

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

	user := getUserFromContext(r.Context())

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

	if err := database.DB.RemoveTeamMember(req.TeamId, req.UserId); err != nil {
		log.Printf("Error removing team member: %v", err)
		http.Error(w, "Error removing member", http.StatusInternalServerError)
		return
	}

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

	user := getUserFromContext(r.Context())

	var req struct {
		FileId string `json:"fileId"`
		TeamId int    `json:"teamId"`
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

	user := getUserFromContext(r.Context())

	var req struct {
		FileId string `json:"fileId"`
		TeamId int    `json:"teamId"`
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
	user := getUserFromContext(r.Context())

	teams, err := database.DB.GetTeamsByUser(user.Id)
	if err != nil {
		log.Printf("Error fetching user teams: %v", err)
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":  user,
		"Teams": teams,
	}

	renderTemplate(w, "user-teams", data)
}

// handleAPITeamFiles returns all files shared with a team
func (s *Server) handleAPITeamFiles(w http.ResponseWriter, r *http.Request) {
	teamIdStr := r.URL.Query().Get("teamId")
	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	user := getUserFromContext(r.Context())

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
	user := getUserFromContext(r.Context())

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

// Helper function to render templates
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// This is a placeholder - you'll need to implement proper template rendering
	// For now, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": tmpl,
		"data":     data,
		"message":  fmt.Sprintf("Template rendering not yet implemented for: %s", tmpl),
	})
}
