// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// entra_identity.go — map a verified Entra ID-token to a local User row.
// Lookup order:
//   1. ExternalID match → existing Entra user, sign in.
//   2. Email match + existing user has IdentityProvider="local" → REJECT
//      (an admin must explicitly promote the local account to Entra; auto-binding
//      a local password account to an external IdP without consent would be a
//      privilege-takeover risk if the OIDC tenant is misconfigured).
//   3. Email match + IdentityProvider="entra" + ExternalID empty → bind (heal).
//   4. No match + auto_provision enabled + domain allowed → create user.
//   5. Otherwise → REJECT with a clear message.

package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// ResolveResult describes the outcome of identity resolution.
type ResolveResult struct {
	User       *models.User // resolved or newly-created user
	Created    bool         // true if auto-provisioned in this call
	BoundEmail bool         // true if we attached ExternalID to an existing entra-typed row
}

// ResolveOrProvisionUser maps verified Entra claims to a local User.
// Caller is responsible for creating the session AFTER this returns successfully.
func ResolveOrProvisionUser(db *database.Database, cfg *EntraConfig, claims *EntraClaims) (*ResolveResult, error) {
	if db == nil || cfg == nil || claims == nil {
		return nil, errors.New("entra: nil argument to ResolveOrProvisionUser")
	}

	email := claims.PrimaryEmail()
	extID := claims.ExternalIdentifier()
	if email == "" || extID == "" {
		return nil, errors.New("entra: claims missing email or external id")
	}

	// Domain allow-list (cheap reject before touching DB).
	if len(cfg.AllowedDomains) > 0 {
		at := strings.LastIndex(email, "@")
		if at < 0 {
			return nil, fmt.Errorf("entra: email %q has no domain", email)
		}
		domain := strings.ToLower(email[at+1:])
		ok := false
		for _, d := range cfg.AllowedDomains {
			if d == domain {
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("entra: domain %q not in allowed_domains list", domain)
		}
	}

	// Step 1 — ExternalID match.
	if u, err := db.GetUserByExternalID(models.IdentityProviderEntra, extID); err == nil && u != nil {
		if !u.IsActive {
			return nil, fmt.Errorf("entra: account %s is disabled", u.Email)
		}
		return &ResolveResult{User: u}, nil
	}

	// Step 2/3 — email match.
	if u, err := db.GetUserByEmail(email); err == nil && u != nil {
		if !u.IsActive {
			return nil, fmt.Errorf("entra: account %s is disabled", u.Email)
		}
		if u.IsLocalAccount() {
			// Refuse to silently bind a local account. Surface a clear error so
			// the admin can take an explicit action (promote-to-entra) in the UI.
			return nil, fmt.Errorf("entra: an active local account exists for %s — an admin must convert it to Entra before SSO can be used", email)
		}
		// IdentityProvider == "entra" but ExternalID is unset (e.g. admin
		// pre-created the row). Bind it now and sign in.
		if u.ExternalID == "" {
			if err := db.SetUserExternalID(u.Id, extID); err != nil {
				return nil, fmt.Errorf("entra: failed to bind ExternalID: %w", err)
			}
			u.ExternalID = extID
			return &ResolveResult{User: u, BoundEmail: true}, nil
		}
		// ExternalID is set on this user but does NOT match the incoming claim's
		// OID. Two different Entra identities trying to share an email = refuse.
		return nil, fmt.Errorf("entra: email %s is bound to a different Entra identity", email)
	}

	// Step 4 — auto-provision.
	if !cfg.AutoProvision {
		return nil, fmt.Errorf("entra: account %s is not provisioned and auto-provisioning is disabled — ask your admin to create the user", email)
	}

	name := strings.TrimSpace(claims.Name)
	if name == "" {
		name = email // fall back to email if Entra didn't return a display name
	}

	// Resolve to a unique Name (Users.Name has UNIQUE). Strategy: try the email
	// local-part first, then append numeric suffixes. Loop bound prevents
	// pathological DBs from spinning.
	baseName := uniqueBaseFromEmail(email)
	username := baseName
	for i := 1; i < 50; i++ {
		if existing, _ := db.GetUserByName(username); existing == nil {
			break
		}
		username = fmt.Sprintf("%s%d", baseName, i)
	}

	newUser := &models.User{
		Name:             username,
		Email:            email,
		Password:         "", // Entra users have no local password
		Permissions:      models.UserPermissionNone,
		UserLevel:        cfg.DefaultRole,
		LastOnline:       time.Now().Unix(),
		ResetPassword:    false,
		StorageQuotaMB:   0, // 0 = inherit default-quota at upload time
		StorageUsedMB:    0,
		CreatedAt:        time.Now().Unix(),
		IsActive:         true,
		IdentityProvider: models.IdentityProviderEntra,
		ExternalID:       extID,
	}
	if err := db.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("entra: failed to create auto-provisioned user: %w", err)
	}

	return &ResolveResult{User: newUser, Created: true}, nil
}

// uniqueBaseFromEmail returns the local-part of an email, trimmed to a sane
// length and stripped of characters that would be awkward in a username.
func uniqueBaseFromEmail(email string) string {
	at := strings.IndexByte(email, '@')
	local := email
	if at > 0 {
		local = email[:at]
	}
	// Keep alphanumerics, dots, underscores, hyphens; replace the rest.
	var b strings.Builder
	for _, r := range local {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		out = "entra_user"
	}
	if len(out) > 40 {
		out = out[:40]
	}
	return out
}
