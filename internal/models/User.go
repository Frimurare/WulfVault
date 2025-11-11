package models

import (
	"encoding/json"
	"time"
)

// UserPermission contains zero or more permissions as uint16
type UserPermission uint16

// User contains information about the Sharecare user
type User struct {
	Id             int            `json:"id" redis:"id"`
	Name           string         `json:"name" redis:"Name"`
	Email          string         `json:"email" redis:"Email"`
	Permissions    UserPermission `json:"permissions" redis:"Permissions"`
	UserLevel      UserRank       `json:"userLevel" redis:"UserLevel"`
	LastOnline     int64          `json:"lastOnline" redis:"LastOnline"`
	Password       string         `json:"-" redis:"Password"`
	ResetPassword  bool           `json:"resetPassword" redis:"ResetPassword"`
	StorageQuotaMB int64          `json:"storageQuotaMB" redis:"StorageQuotaMB"` // Storage quota in MB
	StorageUsedMB  int64          `json:"storageUsedMB" redis:"StorageUsedMB"`   // Current storage used in MB
	CreatedAt      int64          `json:"createdAt" redis:"CreatedAt"`           // Unix timestamp
	IsActive       bool           `json:"isActive" redis:"IsActive"`             // Account active status
	DeletedAt      int64          `json:"deletedAt" redis:"DeletedAt"`           // Unix timestamp, 0 = not deleted
	DeletedBy      string         `json:"deletedBy" redis:"DeletedBy"`           // "user", "admin", or "system"
	OriginalEmail  string         `json:"originalEmail" redis:"OriginalEmail"`   // Store original email before deletion
}

// GetReadableDate returns the date as YYYY-MM-DD HH:MM
func (u *User) GetReadableDate() string {
	if u.LastOnline == 0 {
		return "Never"
	}
	if time.Now().Unix()-u.LastOnline < 120 {
		return "Online"
	}
	return time.Unix(u.LastOnline, 0).Format("2006-01-02 15:04")
}

// GetReadableUserLevel returns the userlevel as a group name
func (u *User) GetReadableUserLevel() string {
	switch u.UserLevel {
	case UserLevelSuperAdmin:
		return "Super Admin"
	case UserLevelAdmin:
		return "Admin"
	case UserLevelUser:
		return "User"
	default:
		return "Invalid"
	}
}

// ToJson returns the user as a JSON object
func (u *User) ToJson() string {
	result, err := json.Marshal(u)
	if err != nil {
		return "{}"
	}
	return string(result)
}

// UserLevelSuperAdmin indicates that this is the single user with the most permissions
const UserLevelSuperAdmin UserRank = 0

// UserLevelAdmin indicates that this user has by default all permissions (unless they affect the super-admin)
const UserLevelAdmin UserRank = 1

// UserLevelUser indicates that this user has only basic permissions by default
const UserLevelUser UserRank = 2

// UserRank indicates the rank that is assigned to the user
type UserRank uint8

// IsSuperAdmin returns true if the user has the Rank UserLevelSuperAdmin
func (u *User) IsSuperAdmin() bool {
	return u.UserLevel == UserLevelSuperAdmin
}

// IsSameUser returns true, if the user has the same ID
func (u *User) IsSameUser(userId int) bool {
	return u.Id == userId
}

const (
	// UserPermReplaceUploads allows to replace uploads
	UserPermReplaceUploads UserPermission = 1 << iota
	// UserPermListOtherUploads allows to also list uploads by other users
	UserPermListOtherUploads
	// UserPermEditOtherUploads allows editing of uploads by other users
	UserPermEditOtherUploads
	// UserPermReplaceOtherUploads allows replacing of uploads by other users
	UserPermReplaceOtherUploads
	// UserPermDeleteOtherUploads allows deleting uploads by other users
	UserPermDeleteOtherUploads
	// UserPermManageLogs allows viewing and deleting logs
	UserPermManageLogs
	// UserPermManageApiKeys allows editing and deleting of API keys by other users
	UserPermManageApiKeys
	// UserPermManageUsers allows creating and editing of users, including granting and revoking permissions
	UserPermManageUsers
)

// UserPermissionNone means that the user has no permissions
const UserPermissionNone UserPermission = 0

// UserPermissionAll means that the user has all permissions
const UserPermissionAll UserPermission = 255

// GrantPermission grants one or more permissions
func (u *User) GrantPermission(permission UserPermission) {
	u.Permissions |= permission
}

// RemovePermission revokes one or more permissions
func (u *User) RemovePermission(permission UserPermission) {
	u.Permissions &^= permission
}

// HasPermission returns true if the key has the permission(s)
func (u *User) HasPermission(permission UserPermission) bool {
	if permission == UserPermissionNone {
		return true
	}
	return (u.Permissions & permission) == permission
}

// HasPermissionReplace returns true if the user has the permission UserPermReplaceUploads
func (u *User) HasPermissionReplace() bool {
	return u.HasPermission(UserPermReplaceUploads)
}

// HasPermissionListOtherUploads returns true if the user has the permission UserPermListOtherUploads
func (u *User) HasPermissionListOtherUploads() bool {
	return u.HasPermission(UserPermListOtherUploads)
}

// HasPermissionEditOtherUploads returns true if the user has the permission UserPermEditOtherUploads
func (u *User) HasPermissionEditOtherUploads() bool {
	return u.HasPermission(UserPermEditOtherUploads)
}

// HasPermissionReplaceOtherUploads returns true if the user has the permission UserPermReplaceOtherUploads
func (u *User) HasPermissionReplaceOtherUploads() bool {
	return u.HasPermission(UserPermReplaceOtherUploads)
}

// HasPermissionDeleteOtherUploads returns true if the user has the permission UserPermDeleteOtherUploads
func (u *User) HasPermissionDeleteOtherUploads() bool {
	return u.HasPermission(UserPermDeleteOtherUploads)
}

// HasPermissionManageLogs returns true if the user has the permission UserPermManageLogs
func (u *User) HasPermissionManageLogs() bool {
	return u.HasPermission(UserPermManageLogs)
}

// HasPermissionManageApi returns true if the user has the permission UserPermManageApiKeys
func (u *User) HasPermissionManageApi() bool {
	return u.HasPermission(UserPermManageApiKeys)
}

// HasPermissionManageUsers returns true if the user has the permission UserPermManageUsers
func (u *User) HasPermissionManageUsers() bool {
	return u.HasPermission(UserPermManageUsers)
}

// GetStoragePercentage returns the storage usage as a percentage (0-100)
func (u *User) GetStoragePercentage() int {
	if u.StorageQuotaMB == 0 {
		return 0
	}
	return int((u.StorageUsedMB * 100) / u.StorageQuotaMB)
}

// GetStorageRemaining returns the remaining storage in MB
func (u *User) GetStorageRemaining() int64 {
	if u.StorageQuotaMB == 0 {
		return 0
	}
	remaining := u.StorageQuotaMB - u.StorageUsedMB
	if remaining < 0 {
		return 0
	}
	return remaining
}

// HasStorageSpace returns true if the user has enough storage space for the given file size in MB
func (u *User) HasStorageSpace(fileSizeMB int64) bool {
	if u.StorageQuotaMB == 0 {
		return false
	}
	return (u.StorageUsedMB + fileSizeMB) <= u.StorageQuotaMB
}

// IsAdmin returns true if the user is an Admin or SuperAdmin
func (u *User) IsAdmin() bool {
	return u.UserLevel == UserLevelAdmin || u.UserLevel == UserLevelSuperAdmin
}
