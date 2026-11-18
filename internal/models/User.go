// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

// Role is the account tier of a WulfVault user. A smaller number means more
// authority. The numbers are written to the Users.Userlevel column, so the
// existing values must never be renumbered.
type Role uint8

const (
	RoleSuperAdmin Role = 0
	RoleAdmin      Role = 1
	RoleUser       Role = 2
)

// roleLabels maps each known role to the text shown in exports and the UI.
var roleLabels = map[Role]string{
	RoleSuperAdmin: "Super Admin",
	RoleAdmin:      "Admin",
	RoleUser:       "User",
}

// PermissionSet is the bit field kept in the Users.Permissions column.
// Administrators are saved with every bit set (PermAll) and ordinary users
// with none, so the stored numbers stay readable by older releases.
type PermissionSet uint16

const (
	// PermNone grants nothing beyond what the role allows.
	PermNone PermissionSet = 0
	// permOthersFiles lets the holder change files that another user uploaded.
	// Bit 2 is the position earlier releases used, kept for stored rows.
	permOthersFiles PermissionSet = 4
	// PermAll is written for administrators.
	PermAll PermissionSet = 0xFF
)

// Identity provider names stored in Users.IdentityProvider.
const (
	IdentityProviderLocal = "local"
	IdentityProviderEntra = "entra"
)

// User is one row of the Users table: a person who can sign in and upload.
// The JSON names are part of the REST API and must stay as they are.
type User struct {
	// Identity
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	// Access
	Role          Role          `json:"userLevel"`
	Permissions   PermissionSet `json:"permissions"`
	Password      string        `json:"-"` // bcrypt hash
	ResetPassword bool          `json:"resetPassword"`
	TOTPEnabled   bool          `json:"totpEnabled"`
	TOTPSecret    string        `json:"-"`
	BackupCodes   string        `json:"-"` // JSON array of hashed one-time codes

	// Sign-in source: "local" (or empty) for password accounts, "entra" for
	// Microsoft Entra ID. ExternalID is the provider's subject identifier.
	IdentityProvider string `json:"identityProvider"`
	ExternalID       string `json:"-"`

	// Storage, in megabytes
	StorageQuotaMB int64 `json:"storageQuotaMB"`
	StorageUsedMB  int64 `json:"storageUsedMB"`

	// Lifecycle, as Unix seconds (0 = never / not set)
	CreatedAt     int64  `json:"createdAt"`
	LastOnline    int64  `json:"lastOnline"`
	IsActive      bool   `json:"isActive"`
	DeletedAt     int64  `json:"deletedAt"`
	DeletedBy     string `json:"deletedBy"`     // "user", "admin" or "system"
	OriginalEmail string `json:"originalEmail"` // SHA-256 fingerprint kept after deletion

	// Preferences
	Language string `json:"language"` // "en", "sv"; empty follows the server default
}

// IsAdmin reports whether the account has administrator rights.
func (u *User) IsAdmin() bool {
	return u.Role <= RoleAdmin
}

// RoleLabel returns a readable name for the user's role.
func (u *User) RoleLabel() string {
	if label, ok := roleLabels[u.Role]; ok {
		return label
	}
	return "Invalid"
}

// CanEditOthersFiles reports whether the user may change files owned by
// someone else.
func (u *User) CanEditOthersFiles() bool {
	return u.Permissions&permOthersFiles != 0
}

// IsLocalAccount reports whether the user signs in with a WulfVault password.
func (u *User) IsLocalAccount() bool {
	switch u.IdentityProvider {
	case "", IdentityProviderLocal:
		return true
	}
	return false
}

// IsEntraAccount reports whether the user signs in through Microsoft Entra ID.
func (u *User) IsEntraAccount() bool {
	return u.IdentityProvider == IdentityProviderEntra
}

// HasStorageSpace reports whether a further sizeMB megabytes fit inside the
// user's quota. A quota of zero allows no uploads.
func (u *User) HasStorageSpace(sizeMB int64) bool {
	if u.StorageQuotaMB <= 0 {
		return false
	}
	return u.StorageUsedMB+sizeMB <= u.StorageQuotaMB
}
