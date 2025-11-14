// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

import (
	"encoding/json"
	"time"
)

// TeamRole represents the role of a team member
type TeamRole int

const (
	// TeamRoleOwner can manage everything in the team
	TeamRoleOwner TeamRole = 0
	// TeamRoleAdmin can manage members and files
	TeamRoleAdmin TeamRole = 1
	// TeamRoleMember can upload and view files
	TeamRoleMember TeamRole = 2
)

// Team represents a collaborative workspace
type Team struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CreatedBy      int    `json:"createdBy"`
	CreatedAt      int64  `json:"createdAt"`
	StorageQuotaMB int64  `json:"storageQuotaMB"`
	StorageUsedMB  int64  `json:"storageUsedMB"`
	IsActive       bool   `json:"isActive"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	Id       int      `json:"id"`
	TeamId   int      `json:"teamId"`
	UserId   int      `json:"userId"`
	Role     TeamRole `json:"role"`
	JoinedAt int64    `json:"joinedAt"`
	AddedBy  int      `json:"addedBy"`

	// Populated via JOIN
	UserName  string `json:"userName,omitempty"`
	UserEmail string `json:"userEmail,omitempty"`
}

// TeamFile represents a file shared with a team
type TeamFile struct {
	Id       int    `json:"id"`
	FileId   string `json:"fileId"`
	TeamId   int    `json:"teamId"`
	SharedBy int    `json:"sharedBy"`
	SharedAt int64  `json:"sharedAt"`
}

// TeamWithMembers includes team info and member count
type TeamWithMembers struct {
	Team
	MemberCount int      `json:"memberCount"`
	UserRole    TeamRole `json:"userRole"`
}

// GetStoragePercentage returns the storage usage as a percentage (0-100)
func (t *Team) GetStoragePercentage() int {
	if t.StorageQuotaMB == 0 {
		return 0
	}
	return int((t.StorageUsedMB * 100) / t.StorageQuotaMB)
}

// GetStorageRemaining returns the remaining storage in MB
func (t *Team) GetStorageRemaining() int64 {
	if t.StorageQuotaMB == 0 {
		return 0
	}
	remaining := t.StorageQuotaMB - t.StorageUsedMB
	if remaining < 0 {
		return 0
	}
	return remaining
}

// HasStorageSpace returns true if the team has enough storage space for the given file size in MB
func (t *Team) HasStorageSpace(fileSizeMB int64) bool {
	if t.StorageQuotaMB == 0 {
		return false
	}
	return (t.StorageUsedMB + fileSizeMB) <= t.StorageQuotaMB
}

// GetReadableRole returns the role as a human-readable string
func (tm *TeamMember) GetReadableRole() string {
	switch tm.Role {
	case TeamRoleOwner:
		return "Owner"
	case TeamRoleAdmin:
		return "Admin"
	case TeamRoleMember:
		return "Member"
	default:
		return "Unknown"
	}
}

// CanManageMembers returns true if the member can add/remove other members
func (tm *TeamMember) CanManageMembers() bool {
	return tm.Role == TeamRoleOwner || tm.Role == TeamRoleAdmin
}

// CanManageFiles returns true if the member can share files to the team
func (tm *TeamMember) CanManageFiles() bool {
	return tm.Role == TeamRoleOwner || tm.Role == TeamRoleAdmin || tm.Role == TeamRoleMember
}

// ToJson returns the team as a JSON object
func (t *Team) ToJson() string {
	result, err := json.Marshal(t)
	if err != nil {
		return "{}"
	}
	return string(result)
}

// GetReadableCreatedAt returns the creation date as YYYY-MM-DD HH:MM
func (t *Team) GetReadableCreatedAt() string {
	if t.CreatedAt == 0 {
		return "Unknown"
	}
	return time.Unix(t.CreatedAt, 0).Format("2006-01-02 15:04")
}
