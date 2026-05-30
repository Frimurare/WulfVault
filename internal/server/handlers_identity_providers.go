// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// handlers_identity_providers.go — admin UI for external SSO providers.
// v7.1 adds Generic OIDC alongside Microsoft Entra ID. The form is
// provider-aware: selecting a type shows/hides the relevant fields.

package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

func (s *Server) handleAdminIdentityProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.renderIdentityProviders(w, "", "")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		s.renderIdentityProviders(w, "", "Failed to parse form: "+err.Error())
		return
	}

	defaultRole, err := strconv.Atoi(r.FormValue("default_role"))
	if err != nil || defaultRole < int(models.UserLevelSuperAdmin) || defaultRole > int(models.UserLevelUser) {
		defaultRole = int(models.UserLevelUser)
	}

	var allowedDomains []string
	for _, raw := range strings.Split(r.FormValue("allowed_domains"), ",") {
		if d := strings.TrimSpace(raw); d != "" {
			allowedDomains = append(allowedDomains, d)
		}
	}

	providerType := strings.TrimSpace(r.FormValue("provider_type"))
	if providerType != auth.ProviderTypeGenericOIDC {
		providerType = auth.ProviderTypeEntra
	}

	cfg := &auth.IdentityProviderConfig{
		Enabled:             r.FormValue("enabled") == "on",
		ForceLocalOnly:      r.FormValue("force_local_only") == "on",
		ProviderType:        providerType,
		ProviderDisplayName: strings.TrimSpace(r.FormValue("provider_display_name")),
		TenantID:            strings.TrimSpace(r.FormValue("tenant_id")),
		ClientID:            strings.TrimSpace(r.FormValue("client_id")),
		ClientSecret:        r.FormValue("client_secret"), // empty = keep existing
		RedirectURI:         strings.TrimSpace(r.FormValue("redirect_uri")),
		IssuerURL:           strings.TrimSpace(r.FormValue("issuer_url")),
		Scopes:              strings.TrimSpace(r.FormValue("scopes")),
		AutoProvision:       r.FormValue("auto_provision") == "on",
		DefaultRole:         models.UserRank(defaultRole),
		AllowedDomains:      allowedDomains,
	}

	if cfg.Enabled && !cfg.ForceLocalOnly {
		if cfg.ClientID == "" {
			s.renderIdentityProviders(w, "", "Client ID is required when SSO is enabled.")
			return
		}
		if cfg.IsEntra() && cfg.TenantID == "" {
			s.renderIdentityProviders(w, "", "Tenant ID is required for Microsoft Entra ID.")
			return
		}
		if cfg.IsGenericOIDC() && cfg.IssuerURL == "" {
			s.renderIdentityProviders(w, "", "Issuer URL is required for Generic OIDC.")
			return
		}
	}

	if err := auth.SaveIdentityProviderConfig(database.DB, cfg); err != nil {
		s.renderIdentityProviders(w, "", "Save failed: "+err.Error())
		return
	}

	admin, _ := userFromContext(r.Context())
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(admin.Id),
		UserEmail:  admin.Email,
		Action:     "IDENTITY_PROVIDER_UPDATED",
		EntityType: "Settings",
		EntityID:   cfg.EffectiveProviderType(),
		Details: fmt.Sprintf(`{"provider_type":%q,"enabled":%v,"force_local_only":%v,"tenant_id_set":%v,"issuer_url_set":%v,"client_id_set":%v,"secret_updated":%v,"auto_provision":%v,"default_role":%d,"allowed_domains":%d}`,
			cfg.EffectiveProviderType(),
			cfg.Enabled, cfg.ForceLocalOnly,
			cfg.TenantID != "", cfg.IssuerURL != "",
			cfg.ClientID != "",
			strings.TrimSpace(cfg.ClientSecret) != "",
			cfg.AutoProvision, cfg.DefaultRole, len(cfg.AllowedDomains)),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Success:   true,
	})

	s.renderIdentityProviders(w, "Identity provider settings saved.", "")
}

func (s *Server) renderIdentityProviders(w http.ResponseWriter, success, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	cfg, loadErr := auth.LoadIdentityProviderConfig(database.DB)
	if loadErr != nil && errMsg == "" {
		errMsg = "Failed to load current settings: " + loadErr.Error()
		cfg = &auth.IdentityProviderConfig{
			AutoProvision: true,
			DefaultRole:   models.UserLevelUser,
		}
	}

	esc := template.HTMLEscapeString
	checked := func(b bool) string {
		if b {
			return "checked"
		}
		return ""
	}
	selectedAttr := func(b bool) string {
		if b {
			return "selected"
		}
		return ""
	}

	// Default redirect URI hint. v7.1: /auth/oidc/callback is canonical for
	// new installs; /auth/entra/callback still works for legacy.
	defaultRedirect := strings.TrimRight(s.config.ServerURL, "/") + "/auth/oidc/callback"
	currentRedirect := cfg.RedirectURI
	if currentRedirect == "" {
		currentRedirect = defaultRedirect
	}

	secretPlaceholder := "(unchanged — leave blank to keep current value)"
	if cfg.ClientSecret == "" {
		secretPlaceholder = "Paste the client secret value here"
	}

	allowedDomainsStr := strings.Join(cfg.AllowedDomains, ", ")

	banner := ""
	if success != "" {
		banner = `<div class="message success">` + esc(success) + `</div>`
	}
	if errMsg != "" {
		banner = `<div class="message error">` + esc(errMsg) + `</div>`
	}

	isEntra := cfg.IsEntra()
	isGeneric := cfg.IsGenericOIDC()

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="author" content="Ulf Holmström">
    <title>Identity Providers - ` + esc(s.config.CompanyName) + `</title>
    ` + s.getFaviconHTML() + `
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; }
        .container { max-width: 880px; margin: 40px auto; padding: 0 20px; }
        .card { background: white; padding: 30px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); margin-bottom: 20px; }
        h2 { margin-bottom: 10px; color: #222; }
        h3 { margin-top: 24px; margin-bottom: 12px; color: #333; font-size: 18px; }
        p.lead { color: #555; margin-bottom: 18px; font-size: 14px; line-height: 1.5; }
        .form-group { margin-bottom: 18px; }
        label { display: block; margin-bottom: 6px; color: #333; font-weight: 500; font-size: 14px; }
        .help { color: #666; font-size: 12px; margin-top: 4px; line-height: 1.4; }
        input[type="text"], input[type="password"], input[type="url"], select, textarea {
            width: 100%; padding: 10px; border: 2px solid #e0e0e0; border-radius: 6px; font-size: 14px; font-family: inherit;
        }
        input:focus, select:focus, textarea:focus { outline: none; border-color: ` + s.getPrimaryColor() + `; }
        input[type="checkbox"] { margin-right: 8px; transform: scale(1.2); }
        .checkbox-row { display: flex; align-items: center; padding: 8px 0; }
        .btn { padding: 12px 24px; background: ` + s.getPrimaryColor() + `; color: white; border: none; border-radius: 6px; font-size: 16px; font-weight: 600; cursor: pointer; }
        .btn:hover { opacity: 0.92; }
        .btn-secondary { background: #6b7280; margin-left: 8px; }
        .message { padding: 12px 18px; border-radius: 6px; margin-bottom: 20px; font-size: 14px; }
        .message.success { background: #d1fae5; color: #065f46; border: 1px solid #6ee7b7; }
        .message.error   { background: #fee2e2; color: #991b1b; border: 1px solid #fca5a5; }
        .info-box { background: #eff6ff; border-left: 4px solid #3b82f6; padding: 14px 18px; border-radius: 4px; margin-bottom: 16px; font-size: 13px; color: #1e3a8a; line-height: 1.5; }
        .preset-list { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 6px; padding: 12px 18px; margin-top: 8px; font-size: 13px; line-height: 1.7; }
        .preset-list code { background: #fff; padding: 2px 6px; border-radius: 3px; font-size: 12px; font-family: 'SF Mono', Menlo, monospace; }
        code { background: #f3f4f6; padding: 2px 6px; border-radius: 3px; font-size: 12px; font-family: 'SF Mono', Menlo, monospace; }
        .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
        @media (max-width: 640px) { .grid-2 { grid-template-columns: 1fr; } }
    </style>
</head>
<body>
    ` + s.getAdminHeaderHTML("") + `
    <div class="container">
        <h2>Identity Providers</h2>
        <p class="lead">Configure external single sign-on. Local accounts always remain available as the primary mechanism and as a break-glass fallback.</p>
        ` + banner + `

        <div class="card">
            <form method="POST">
                <div class="form-group">
                    <label for="provider_type">Provider type</label>
                    <select id="provider_type" name="provider_type" onchange="toggleProviderFields()">
                        <option value="entra" ` + selectedAttr(isEntra) + `>Microsoft Entra ID (Azure AD)</option>
                        <option value="generic_oidc" ` + selectedAttr(isGeneric) + `>Generic OpenID Connect</option>
                    </select>
                    <div class="help">Entra ID is special-cased with multi-tenant aliases (common / organizations / consumers). Generic OIDC works with any OIDC-compliant provider — Google, Okta, Keycloak, Authelia, Auth0, etc.</div>
                </div>

                <div class="form-group checkbox-row">
                    <input type="checkbox" id="enabled" name="enabled" ` + checked(cfg.Enabled) + `>
                    <label for="enabled" style="margin-bottom:0;">Enable SSO</label>
                </div>

                <div class="form-group checkbox-row">
                    <input type="checkbox" id="force_local_only" name="force_local_only" ` + checked(cfg.ForceLocalOnly) + `>
                    <label for="force_local_only" style="margin-bottom:0;">Force local accounts only (overrides "Enabled" — hides SSO button and rejects callbacks)</label>
                </div>

                <div class="form-group">
                    <label for="provider_display_name">Provider display name (optional)</label>
                    <input type="text" id="provider_display_name" name="provider_display_name" value="` + esc(cfg.ProviderDisplayName) + `" placeholder="Microsoft / Google / Okta / SSO">
                    <div class="help">Shown on the login button ("Sign in with X") and in invite emails. Defaults to "Microsoft" for Entra, "SSO" for Generic OIDC.</div>
                </div>

                <!-- Entra-specific fields -->
                <div id="entra-fields" style="display: ` + showIf(isEntra) + `;">
                    <div class="info-box">
                        <strong>Microsoft Entra ID setup:</strong>
                        <ol style="margin:8px 0 0 22px;">
                            <li>Register a new application in <em>Entra ID → App registrations</em>.</li>
                            <li>Add redirect URI (Web): <code>` + esc(defaultRedirect) + `</code></li>
                            <li>Under <em>Certificates &amp; secrets</em>, create a client secret and copy the <strong>Value</strong>.</li>
                            <li>Note the <em>Application (client) ID</em> and <em>Directory (tenant) ID</em>.</li>
                        </ol>
                    </div>

                    <div class="form-group">
                        <label for="tenant_id">Tenant ID</label>
                        <input type="text" id="tenant_id" name="tenant_id" value="` + esc(cfg.TenantID) + `" placeholder="00000000-0000-0000-0000-000000000000" autocomplete="off">
                        <div class="help">Directory (tenant) ID. May also be <code>common</code>, <code>organizations</code> or <code>consumers</code> for multi-tenant apps.</div>
                    </div>
                </div>

                <!-- Generic-OIDC-specific fields -->
                <div id="generic-fields" style="display: ` + showIf(isGeneric) + `;">
                    <div class="info-box">
                        <strong>Generic OpenID Connect setup:</strong> Any OIDC-compliant provider. Register WulfVault as a confidential client with the redirect URI shown below, then paste the issuer URL, client ID and secret.
                    </div>

                    <div class="form-group">
                        <label for="issuer_url">Issuer URL</label>
                        <input type="url" id="issuer_url" name="issuer_url" value="` + esc(cfg.IssuerURL) + `" placeholder="https://accounts.google.com" autocomplete="off">
                        <div class="help">The OIDC issuer — the URL whose <code>/.well-known/openid-configuration</code> describes the provider. No trailing slash.</div>
                        <div class="preset-list">
                            <strong>Common issuers:</strong><br>
                            Google Workspace: <code>https://accounts.google.com</code><br>
                            Okta: <code>https://&lt;your-tenant&gt;.okta.com</code><br>
                            Keycloak: <code>https://&lt;keycloak-host&gt;/realms/&lt;realm&gt;</code><br>
                            Authelia: <code>https://&lt;authelia-host&gt;</code><br>
                            Auth0: <code>https://&lt;your-tenant&gt;.auth0.com</code>
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="scopes">Scopes (optional)</label>
                        <input type="text" id="scopes" name="scopes" value="` + esc(cfg.Scopes) + `" placeholder="openid email profile">
                        <div class="help">Space-separated. Defaults to <code>openid email profile</code> — the OIDC baseline that returns sub, email and display name.</div>
                    </div>
                </div>

                <h3>OAuth / OIDC credentials</h3>

                <div class="grid-2">
                    <div class="form-group">
                        <label for="client_id">Client ID</label>
                        <input type="text" id="client_id" name="client_id" value="` + esc(cfg.ClientID) + `" placeholder="00000000-0000-0000-0000-000000000000" autocomplete="off">
                        <div class="help">Application / Client ID from the provider.</div>
                    </div>
                    <div class="form-group">
                        <label for="client_secret">Client Secret</label>
                        <input type="password" id="client_secret" name="client_secret" value="" placeholder="` + esc(secretPlaceholder) + `" autocomplete="new-password">
                        <div class="help">Stored encrypted at rest. Leave blank to preserve the existing secret.</div>
                    </div>
                </div>

                <div class="form-group">
                    <label for="redirect_uri">Redirect URI</label>
                    <input type="url" id="redirect_uri" name="redirect_uri" value="` + esc(currentRedirect) + `">
                    <div class="help">Must match exactly what is registered at the provider. Default: <code>` + esc(defaultRedirect) + `</code> — the legacy <code>/auth/entra/callback</code> path also works.</div>
                </div>

                <h3>Provisioning</h3>

                <div class="form-group checkbox-row">
                    <input type="checkbox" id="auto_provision" name="auto_provision" ` + checked(cfg.AutoProvision) + `>
                    <label for="auto_provision" style="margin-bottom:0;">Auto-provision new users on first sign-in</label>
                </div>

                <div class="grid-2">
                    <div class="form-group">
                        <label for="default_role">Default role for new users</label>
                        <select id="default_role" name="default_role">
                            <option value="2" ` + selectedAttr(cfg.DefaultRole == models.UserLevelUser) + `>User</option>
                            <option value="1" ` + selectedAttr(cfg.DefaultRole == models.UserLevelAdmin) + `>Admin</option>
                        </select>
                        <div class="help">Applies when auto-provisioning is enabled. SuperAdmin cannot be auto-assigned.</div>
                    </div>
                    <div class="form-group">
                        <label for="allowed_domains">Allowed email domains</label>
                        <input type="text" id="allowed_domains" name="allowed_domains" value="` + esc(allowedDomainsStr) + `" placeholder="contoso.com, prudsec.se">
                        <div class="help">Comma-separated. Empty = any tenant member allowed.</div>
                    </div>
                </div>

                <div style="margin-top: 24px;">
                    <button type="submit" class="btn">Save Settings</button>
                    <a href="/admin/settings" class="btn btn-secondary" style="text-decoration:none; display:inline-block;">Back to Settings</a>
                </div>
            </form>
        </div>
    </div>

    <script>
        function toggleProviderFields() {
            var type = document.getElementById('provider_type').value;
            document.getElementById('entra-fields').style.display = (type === 'entra') ? '' : 'none';
            document.getElementById('generic-fields').style.display = (type === 'generic_oidc') ? '' : 'none';
        }
    </script>
</body>
</html>`

	_, _ = w.Write([]byte(html))
}

// showIf is a tiny helper to keep the inline-style ternaries readable.
func showIf(condition bool) string {
	if condition {
		return ""
	}
	return "none"
}
