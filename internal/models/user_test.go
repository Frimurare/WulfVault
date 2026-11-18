// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

import (
	"encoding/json"
	"testing"
)

// The numbers below are stored in existing databases; changing them would
// silently promote or demote accounts after an upgrade.
func TestStoredValuesAreStable(t *testing.T) {
	if RoleSuperAdmin != 0 || RoleAdmin != 1 || RoleUser != 2 {
		t.Fatalf("role numbers changed: %d %d %d", RoleSuperAdmin, RoleAdmin, RoleUser)
	}
	if PermNone != 0 || PermAll != 255 {
		t.Fatalf("permission sets changed: none=%d all=%d", PermNone, PermAll)
	}
}

func TestCanEditOthersFiles(t *testing.T) {
	cases := []struct {
		stored PermissionSet
		want   bool
	}{
		{0, false},  // ordinary user
		{255, true}, // administrator rows
		{4, true},   // only the "edit other uploads" bit
		{1 | 2, false},
		{0xFB, false}, // everything except bit 2
	}
	for _, c := range cases {
		u := User{Permissions: c.stored}
		if got := u.CanEditOthersFiles(); got != c.want {
			t.Errorf("Permissions=%d: got %v want %v", c.stored, got, c.want)
		}
	}
}

func TestRoles(t *testing.T) {
	cases := []struct {
		role  Role
		admin bool
		label string
	}{
		{RoleSuperAdmin, true, "Super Admin"},
		{RoleAdmin, true, "Admin"},
		{RoleUser, false, "User"},
		{Role(9), false, "Invalid"},
	}
	for _, c := range cases {
		u := User{Role: c.role}
		if u.IsAdmin() != c.admin {
			t.Errorf("role %d: IsAdmin=%v", c.role, u.IsAdmin())
		}
		if u.RoleLabel() != c.label {
			t.Errorf("role %d: label %q want %q", c.role, u.RoleLabel(), c.label)
		}
	}
}

func TestIdentityProvider(t *testing.T) {
	for _, p := range []string{"", "local"} {
		u := User{IdentityProvider: p}
		if !u.IsLocalAccount() || u.IsEntraAccount() {
			t.Errorf("provider %q should be local", p)
		}
	}
	u := User{IdentityProvider: "entra"}
	if u.IsLocalAccount() || !u.IsEntraAccount() {
		t.Error("entra account misclassified")
	}
}

func TestHasStorageSpace(t *testing.T) {
	u := User{StorageQuotaMB: 100, StorageUsedMB: 60}
	if !u.HasStorageSpace(40) {
		t.Error("exact fit should be allowed")
	}
	if u.HasStorageSpace(41) {
		t.Error("over quota should be refused")
	}
	if (&User{}).HasStorageSpace(1) {
		t.Error("zero quota should refuse uploads")
	}
}

// The REST API and the admin pages read these JSON names.
func TestUserJSONNames(t *testing.T) {
	u := User{Id: 7, Name: "n", Email: "e@x", Role: RoleAdmin, Permissions: PermAll,
		Password: "hash", TOTPSecret: "secret", BackupCodes: "[]", ExternalID: "oid"}
	raw, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"id", "name", "email", "userLevel", "permissions", "lastOnline",
		"resetPassword", "storageQuotaMB", "storageUsedMB", "createdAt", "isActive", "deletedAt",
		"deletedBy", "originalEmail", "totpEnabled", "identityProvider", "language"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing JSON key %q", k)
		}
	}
	if m["userLevel"].(float64) != 1 || m["permissions"].(float64) != 255 {
		t.Errorf("numeric encoding changed: %s", raw)
	}
	if len(m) != 17 {
		t.Errorf("unexpected JSON keys (secrets leaking?): %s", raw)
	}
}
