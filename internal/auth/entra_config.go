// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// entra_config.go — read/write Microsoft Entra ID SSO settings stored in the
// Configuration table. The OIDC flow itself is added in a follow-up PR; this
// file only handles persisted configuration.

package auth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
	"github.com/Frimurare/WulfVault/internal/secrets"
)

// Configuration keys used by Entra ID SSO.
const (
	cfgEntraEnabled         = "entra_enabled"
	cfgEntraForceLocalOnly  = "entra_force_local_only"
	cfgEntraTenantID        = "entra_tenant_id"
	cfgEntraClientID        = "entra_client_id"
	cfgEntraClientSecretEnc = "entra_client_secret_enc"
	cfgEntraRedirectURI     = "entra_redirect_uri"
	cfgEntraAutoProvision   = "entra_auto_provision"
	cfgEntraDefaultRole     = "entra_default_role"
	cfgEntraAllowedDomains  = "entra_allowed_domains"
)

// EntraConfig is the admin-facing settings struct. ClientSecret is plaintext
// in this struct (encrypted only at rest in the Configuration table).
type EntraConfig struct {
	Enabled         bool
	ForceLocalOnly  bool
	TenantID        string
	ClientID        string
	ClientSecret    string
	RedirectURI     string
	AutoProvision   bool
	DefaultRole     models.UserRank
	AllowedDomains  []string // lower-cased email domains; empty = any
}

// ShouldShowSSOButton reports whether the login page should expose the
// "Sign in with Microsoft" button. ForceLocalOnly overrides Enabled.
func (c *EntraConfig) ShouldShowSSOButton() bool {
	return c.Enabled && !c.ForceLocalOnly
}

// LoadEntraConfig reads the Entra settings from the Configuration table.
// Missing keys produce zero-values. Returns an error only on DB or decryption
// failures.
func LoadEntraConfig(db *database.Database) (*EntraConfig, error) {
	cfg := &EntraConfig{
		AutoProvision: true,                  // sensible default — admin can disable
		DefaultRole:   models.UserLevelUser,  // least-privileged by default
	}

	getBool := func(key string, dst *bool) error {
		v, err := db.GetConfigValue(key)
		if err != nil {
			return fmt.Errorf("read %s: %w", key, err)
		}
		if v == "" {
			return nil // keep default
		}
		*dst = v == "1" || strings.EqualFold(v, "true")
		return nil
	}

	getStr := func(key string, dst *string) error {
		v, err := db.GetConfigValue(key)
		if err != nil {
			return fmt.Errorf("read %s: %w", key, err)
		}
		*dst = v
		return nil
	}

	if err := getBool(cfgEntraEnabled, &cfg.Enabled); err != nil {
		return nil, err
	}
	if err := getBool(cfgEntraForceLocalOnly, &cfg.ForceLocalOnly); err != nil {
		return nil, err
	}
	if err := getStr(cfgEntraTenantID, &cfg.TenantID); err != nil {
		return nil, err
	}
	if err := getStr(cfgEntraClientID, &cfg.ClientID); err != nil {
		return nil, err
	}
	if err := getStr(cfgEntraRedirectURI, &cfg.RedirectURI); err != nil {
		return nil, err
	}
	if err := getBool(cfgEntraAutoProvision, &cfg.AutoProvision); err != nil {
		return nil, err
	}

	// Default role: stored as the numeric UserRank.
	if v, err := db.GetConfigValue(cfgEntraDefaultRole); err == nil && v != "" {
		if n, perr := strconv.Atoi(v); perr == nil {
			cfg.DefaultRole = models.UserRank(n)
		}
	}

	// Allowed domains: comma-separated, lower-cased on the way in.
	if v, err := db.GetConfigValue(cfgEntraAllowedDomains); err == nil && v != "" {
		for _, d := range strings.Split(v, ",") {
			d = strings.ToLower(strings.TrimSpace(d))
			if d != "" {
				cfg.AllowedDomains = append(cfg.AllowedDomains, d)
			}
		}
	}

	// Decrypt the client secret if present.
	if enc, err := db.GetConfigValue(cfgEntraClientSecretEnc); err == nil && enc != "" {
		key, kerr := secrets.GetOrCreateMasterKey(db)
		if kerr != nil {
			return nil, fmt.Errorf("master key: %w", kerr)
		}
		plain, derr := secrets.Decrypt(enc, key)
		if derr != nil {
			return nil, fmt.Errorf("decrypt client secret: %w", derr)
		}
		cfg.ClientSecret = plain
	}

	return cfg, nil
}

// SaveEntraConfig persists the Entra settings. ClientSecret is encrypted with
// the install's master key before writing. An empty ClientSecret preserves the
// existing one (so editing the page without re-entering the secret is safe).
func SaveEntraConfig(db *database.Database, cfg *EntraConfig) error {
	boolStr := func(b bool) string {
		if b {
			return "1"
		}
		return "0"
	}

	if err := db.SetConfigValue(cfgEntraEnabled, boolStr(cfg.Enabled)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgEntraForceLocalOnly, boolStr(cfg.ForceLocalOnly)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgEntraTenantID, strings.TrimSpace(cfg.TenantID)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgEntraClientID, strings.TrimSpace(cfg.ClientID)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgEntraRedirectURI, strings.TrimSpace(cfg.RedirectURI)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgEntraAutoProvision, boolStr(cfg.AutoProvision)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgEntraDefaultRole, strconv.Itoa(int(cfg.DefaultRole))); err != nil {
		return err
	}

	// Normalize allowed domains: lower-case, deduplicate.
	seen := map[string]bool{}
	cleaned := make([]string, 0, len(cfg.AllowedDomains))
	for _, d := range cfg.AllowedDomains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == "" || seen[d] {
			continue
		}
		seen[d] = true
		cleaned = append(cleaned, d)
	}
	if err := db.SetConfigValue(cfgEntraAllowedDomains, strings.Join(cleaned, ",")); err != nil {
		return err
	}

	// Empty incoming secret = keep existing (admin didn't retype it).
	if strings.TrimSpace(cfg.ClientSecret) != "" {
		key, err := secrets.GetOrCreateMasterKey(db)
		if err != nil {
			return fmt.Errorf("master key: %w", err)
		}
		enc, err := secrets.Encrypt(cfg.ClientSecret, key)
		if err != nil {
			return fmt.Errorf("encrypt client secret: %w", err)
		}
		if err := db.SetConfigValue(cfgEntraClientSecretEnc, enc); err != nil {
			return err
		}
	}

	return nil
}
