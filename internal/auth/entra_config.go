// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// entra_config.go — identity-provider configuration (Entra ID + Generic OIDC).
// File kept under the entra_*.go name for git history continuity; the type
// is provider-agnostic now (IdentityProviderConfig). EntraConfig remains as
// a type alias so existing callers don't need to change.

package auth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
	"github.com/Frimurare/WulfVault/internal/secrets"
)

// Provider type values stored in cfgProviderType.
const (
	ProviderTypeEntra       = "entra"
	ProviderTypeGenericOIDC = "generic_oidc"
)

// Configuration keys. Kept with "entra_" prefix for v6/v7.0 install
// continuity — DBs already have these rows, renaming would require a
// migration with no behavioural benefit.
const (
	cfgEntraEnabled            = "entra_enabled"
	cfgEntraForceLocalOnly     = "entra_force_local_only"
	cfgEntraTenantID           = "entra_tenant_id"
	cfgEntraClientID           = "entra_client_id"
	cfgEntraClientSecretEnc    = "entra_client_secret_enc"
	cfgEntraRedirectURI        = "entra_redirect_uri"
	cfgEntraAutoProvision      = "entra_auto_provision"
	cfgEntraDefaultRole        = "entra_default_role"
	cfgEntraAllowedDomains     = "entra_allowed_domains"
	// Added in v7.1.0 for multi-provider support
	cfgProviderType            = "entra_provider_type"          // "entra" | "generic_oidc"
	cfgProviderDisplayName     = "entra_provider_display_name"  // "Microsoft", "Google", custom
	cfgGenericIssuerURL        = "entra_generic_issuer_url"     // for Generic OIDC
	cfgGenericScopes           = "entra_generic_scopes"         // space-sep, default "openid email profile"
)

// IdentityProviderConfig holds the configured single sign-on provider.
// One provider per install in v7.1 (Entra OR Generic OIDC, not both).
//
// ClientSecret is plaintext in this struct; it is encrypted at rest in
// the Configuration table.
type IdentityProviderConfig struct {
	Enabled        bool
	ForceLocalOnly bool

	// ProviderType selects which provider-specific code path runs.
	// "entra"        → Microsoft Entra ID (uses TenantID, multi-tenant aliases)
	// "generic_oidc" → Any OIDC-compliant provider (uses IssuerURL)
	ProviderType string

	// ProviderDisplayName is shown to end-users (login button, invite email).
	// Defaults to "Microsoft" for entra, "SSO" for generic_oidc.
	ProviderDisplayName string

	// Shared OAuth/OIDC fields
	ClientID     string
	ClientSecret string
	RedirectURI  string

	// Entra-specific (used when ProviderType == "entra")
	// May be a tenant GUID or one of "common", "organizations", "consumers".
	TenantID string

	// Generic-OIDC-specific (used when ProviderType == "generic_oidc")
	// IssuerURL is the OIDC issuer (e.g. "https://accounts.google.com",
	// "https://your-tenant.okta.com", "https://your-keycloak/realms/your-realm").
	IssuerURL string

	// Scopes for Generic OIDC. Space-separated. Defaults to "openid email profile".
	// Entra always uses the same baseline scopes — this field is ignored for it.
	Scopes string

	// Provisioning rules (shared across providers)
	AutoProvision  bool
	DefaultRole    models.UserRank
	AllowedDomains []string // lower-cased email domains; empty = any
}

// EntraConfig is a back-compatibility type alias. Existing callers that
// type-assert *EntraConfig keep working unchanged — under the hood it's
// always an IdentityProviderConfig now.
type EntraConfig = IdentityProviderConfig

// ShouldShowSSOButton reports whether the login page should expose the
// "Sign in with X" button. ForceLocalOnly overrides Enabled.
func (c *IdentityProviderConfig) ShouldShowSSOButton() bool {
	return c.Enabled && !c.ForceLocalOnly
}

// EffectiveProviderType returns ProviderType with the default ("entra")
// substituted for empty string. This lets installs that pre-date v7.1
// continue to behave as Entra without an explicit migration.
func (c *IdentityProviderConfig) EffectiveProviderType() string {
	if c.ProviderType == "" {
		return ProviderTypeEntra
	}
	return c.ProviderType
}

// IsEntra is a convenience for "should I run the Entra-specific code path".
func (c *IdentityProviderConfig) IsEntra() bool {
	return c.EffectiveProviderType() == ProviderTypeEntra
}

// IsGenericOIDC is a convenience for "should I run the generic OIDC code path".
func (c *IdentityProviderConfig) IsGenericOIDC() bool {
	return c.EffectiveProviderType() == ProviderTypeGenericOIDC
}

// EffectiveDisplayName returns the configured ProviderDisplayName or a
// sensible default based on the provider type.
func (c *IdentityProviderConfig) EffectiveDisplayName() string {
	if c.ProviderDisplayName != "" {
		return c.ProviderDisplayName
	}
	if c.IsEntra() {
		return "Microsoft"
	}
	return "SSO"
}

// EffectiveScopes returns the configured scopes or the OIDC baseline.
func (c *IdentityProviderConfig) EffectiveScopes() []string {
	s := strings.TrimSpace(c.Scopes)
	if s == "" {
		return []string{"openid", "email", "profile"}
	}
	return strings.Fields(s)
}

// LoadIdentityProviderConfig reads SSO settings from the Configuration table.
// Missing keys produce zero-values with sensible defaults. Returns an error
// only on DB or decryption failures.
func LoadIdentityProviderConfig(db *database.Database) (*IdentityProviderConfig, error) {
	cfg := &IdentityProviderConfig{
		AutoProvision: true,
		DefaultRole:   models.UserLevelUser,
	}

	getBool := func(key string, dst *bool) error {
		v, err := db.GetConfigValue(key)
		if err != nil {
			return fmt.Errorf("read %s: %w", key, err)
		}
		if v == "" {
			return nil
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
	if err := getStr(cfgProviderType, &cfg.ProviderType); err != nil {
		return nil, err
	}
	if err := getStr(cfgProviderDisplayName, &cfg.ProviderDisplayName); err != nil {
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
	if err := getStr(cfgGenericIssuerURL, &cfg.IssuerURL); err != nil {
		return nil, err
	}
	if err := getStr(cfgGenericScopes, &cfg.Scopes); err != nil {
		return nil, err
	}
	if err := getBool(cfgEntraAutoProvision, &cfg.AutoProvision); err != nil {
		return nil, err
	}

	if v, err := db.GetConfigValue(cfgEntraDefaultRole); err == nil && v != "" {
		if n, perr := strconv.Atoi(v); perr == nil {
			cfg.DefaultRole = models.UserRank(n)
		}
	}

	if v, err := db.GetConfigValue(cfgEntraAllowedDomains); err == nil && v != "" {
		for _, d := range strings.Split(v, ",") {
			d = strings.ToLower(strings.TrimSpace(d))
			if d != "" {
				cfg.AllowedDomains = append(cfg.AllowedDomains, d)
			}
		}
	}

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

// LoadEntraConfig is the back-compatibility alias. Identical to
// LoadIdentityProviderConfig — kept so existing callers don't need to change.
func LoadEntraConfig(db *database.Database) (*EntraConfig, error) {
	return LoadIdentityProviderConfig(db)
}

// SaveIdentityProviderConfig persists SSO settings. ClientSecret is encrypted
// before writing. An empty ClientSecret preserves the existing one (so editing
// the page without re-entering the secret is safe).
func SaveIdentityProviderConfig(db *database.Database, cfg *IdentityProviderConfig) error {
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
	// Normalize provider type — anything other than "generic_oidc" → "entra".
	pt := strings.TrimSpace(cfg.ProviderType)
	if pt != ProviderTypeGenericOIDC {
		pt = ProviderTypeEntra
	}
	if err := db.SetConfigValue(cfgProviderType, pt); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgProviderDisplayName, strings.TrimSpace(cfg.ProviderDisplayName)); err != nil {
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
	if err := db.SetConfigValue(cfgGenericIssuerURL, strings.TrimSpace(cfg.IssuerURL)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgGenericScopes, strings.TrimSpace(cfg.Scopes)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgEntraAutoProvision, boolStr(cfg.AutoProvision)); err != nil {
		return err
	}
	if err := db.SetConfigValue(cfgEntraDefaultRole, strconv.Itoa(int(cfg.DefaultRole))); err != nil {
		return err
	}

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

// SaveEntraConfig is the back-compatibility alias.
func SaveEntraConfig(db *database.Database, cfg *EntraConfig) error {
	return SaveIdentityProviderConfig(db, cfg)
}
