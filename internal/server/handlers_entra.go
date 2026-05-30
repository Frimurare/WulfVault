// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// handlers_entra.go — Microsoft Entra ID OIDC endpoints.
//
// Two routes:
//   GET /auth/entra/login    — kicks off the flow: PKCE + state cookies, redirect
//                               to the tenant's authorize endpoint.
//   GET /auth/entra/callback — handles the redirect back: validate state, exchange
//                               code, verify ID-token, resolve identity, sign in.
//
// Failure modes are surfaced via renderEntraError so end-users see a friendly page
// instead of a JSON blob or raw 500.

package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/Frimurare/WulfVault/internal/auth"
	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

const (
	// Short-lived cookies used to round-trip CSRF state and the PKCE verifier
	// across the user-agent redirect to Entra and back. 10 minutes is generous
	// — Microsoft's authorize step is usually under a minute, but corporate MFA
	// (especially with conditional access prompts) can take longer.
	entraStateCookieName = "entra_state"
	entraPKCECookieName  = "entra_pkce"
	entraStateMaxAge     = 10 * 60 // seconds
)

// handleEntraLogin starts the OIDC flow.
func (s *Server) handleEntraLogin(w http.ResponseWriter, r *http.Request) {
	cfg, err := auth.LoadEntraConfig(database.DB)
	if err != nil || cfg == nil || !cfg.ShouldShowSSOButton() {
		http.NotFound(w, r)
		return
	}

	provider, err := auth.BuildEntraProvider(r.Context(), cfg)
	if err != nil || !provider.Enabled {
		log.Printf("Entra: BuildEntraProvider failed in /login: %v", err)
		s.renderEntraError(w, "Microsoft sign-in is misconfigured. Ask your administrator to verify the Tenant ID, Client ID and Client Secret on the Identity Providers page.")
		return
	}

	pkce, err := auth.GeneratePKCE()
	if err != nil {
		s.renderEntraError(w, "Failed to generate PKCE pair. Try again.")
		return
	}
	state, err := auth.GenerateState()
	if err != nil {
		s.renderEntraError(w, "Failed to generate state token. Try again.")
		return
	}

	secure := r.TLS != nil
	expires := time.Now().Add(time.Duration(entraStateMaxAge) * time.Second)
	http.SetCookie(w, &http.Cookie{
		Name:     entraStateCookieName,
		Value:    state,
		Path:     "/auth/entra/",
		Expires:  expires,
		MaxAge:   entraStateMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     entraPKCECookieName,
		Value:    pkce.Verifier,
		Path:     "/auth/entra/",
		Expires:  expires,
		MaxAge:   entraStateMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	url := provider.AuthCodeURL(state, pkce)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// handleEntraCallback finishes the OIDC flow and creates a session.
func (s *Server) handleEntraCallback(w http.ResponseWriter, r *http.Request) {
	cfg, err := auth.LoadEntraConfig(database.DB)
	if err != nil || cfg == nil || !cfg.ShouldShowSSOButton() {
		http.NotFound(w, r)
		return
	}

	// Entra returns errors via query params, not HTTP status — surface them.
	if errCode := r.URL.Query().Get("error"); errCode != "" {
		desc := r.URL.Query().Get("error_description")
		log.Printf("Entra: callback error code=%s desc=%s", errCode, desc)
		s.renderEntraError(w, fmt.Sprintf("Microsoft sign-in returned an error: %s. %s", errCode, desc))
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		s.renderEntraError(w, "Missing code or state in callback URL. Try signing in again.")
		return
	}

	stateCookie, err := r.Cookie(entraStateCookieName)
	if err != nil || stateCookie.Value == "" {
		s.renderEntraError(w, "Your sign-in took too long or cookies were blocked. Try again with cookies enabled.")
		return
	}
	if stateCookie.Value != state {
		log.Printf("Entra: state mismatch (CSRF or stale tab) — cookie=%q query=%q", stateCookie.Value, state)
		s.renderEntraError(w, "Sign-in state did not match. This protects against cross-site request forgery. Try again from the login page.")
		return
	}

	pkceCookie, err := r.Cookie(entraPKCECookieName)
	if err != nil || pkceCookie.Value == "" {
		s.renderEntraError(w, "PKCE verifier missing — start the sign-in over from the login page.")
		return
	}

	provider, err := auth.BuildEntraProvider(r.Context(), cfg)
	if err != nil || !provider.Enabled {
		log.Printf("Entra: BuildEntraProvider failed in /callback: %v", err)
		s.renderEntraError(w, "Microsoft sign-in is misconfigured. Ask your administrator to verify settings.")
		return
	}

	claims, err := provider.ExchangeCode(r.Context(), code, pkceCookie.Value)
	if err != nil {
		log.Printf("Entra: ExchangeCode failed: %v", err)
		s.renderEntraError(w, "Microsoft did not return a valid identity token. Try again, and contact your administrator if it keeps failing.")
		return
	}

	// Clear the short-lived flow cookies — they've served their purpose.
	clearEntraFlowCookies(w, r.TLS != nil)

	result, err := auth.ResolveOrProvisionUser(database.DB, cfg, claims)
	if err != nil {
		log.Printf("Entra: identity resolution failed for %s: %v", claims.PrimaryEmail(), err)
		s.auditEntraFailure(r, claims.PrimaryEmail(), err.Error())
		s.renderEntraError(w, "Microsoft sign-in succeeded but this WulfVault rejected the account: "+err.Error())
		return
	}

	user := result.User
	sessionID, err := auth.CreateSession(user.Id, 24*time.Hour)
	if err != nil {
		log.Printf("Entra: CreateSession failed for user %d: %v", user.Id, err)
		s.renderEntraError(w, "Sign-in succeeded but the session could not be created. Try again.")
		return
	}

	// Audit log — distinguish auto-provisioning from a normal sign-in.
	action := "ENTRA_LOGIN_SUCCESS"
	if result.Created {
		action = "ENTRA_USER_PROVISIONED"
	} else if result.BoundEmail {
		action = "ENTRA_USER_BOUND"
	}
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     int64(user.Id),
		UserEmail:  user.Email,
		Action:     action,
		EntityType: "Session",
		EntityID:   sessionID,
		Details: fmt.Sprintf(`{"email":"%s","external_id":"%s","tenant":"%s","created":%v,"bound":%v}`,
			user.Email, user.ExternalID, claims.TenantID, result.Created, result.BoundEmail),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Success:   true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})

	redirect := "/dashboard"
	if user.IsAdmin() {
		redirect = "/admin"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func clearEntraFlowCookies(w http.ResponseWriter, secure bool) {
	for _, name := range []string{entraStateCookieName, entraPKCECookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/auth/entra/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})
	}
}

func (s *Server) auditEntraFailure(r *http.Request, email, reason string) {
	database.DB.LogAction(&database.AuditLogEntry{
		UserID:     0,
		UserEmail:  email,
		Action:     "ENTRA_LOGIN_REJECTED",
		EntityType: "Session",
		EntityID:   "",
		Details:    fmt.Sprintf(`{"email":"%s","reason":%q}`, email, reason),
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		Success:    false,
		ErrorMsg:   reason,
	})
	_ = models.UserLevelUser // keep models import alive for future use
}

// renderEntraError shows a branded, friendly error page for any failure in the
// Entra flow. Errors are also logged via log.Printf in the caller.
func (s *Server) renderEntraError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	esc := template.HTMLEscapeString
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Sign-in problem — ` + esc(s.config.CompanyName) + `</title>
    ` + s.getFaviconHTML() + `
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
               background: linear-gradient(135deg, ` + s.getPrimaryColor() + ` 0%, ` + s.getSecondaryColor() + ` 100%);
               min-height:100vh; display:flex; align-items:center; justify-content:center; padding:20px; margin:0; }
        .card { background:white; border-radius:12px; box-shadow:0 20px 60px rgba(0,0,0,0.3);
                padding:40px; max-width:520px; width:100%; }
        h1 { color:` + s.getPrimaryColor() + `; margin:0 0 16px; font-size:22px; }
        p  { color:#444; line-height:1.55; margin:0 0 24px; font-size:14px; }
        a.btn { display:inline-block; padding:12px 24px; background:` + s.getPrimaryColor() + `;
                color:white; text-decoration:none; border-radius:6px; font-weight:600; }
        a.btn:hover { opacity:0.92; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Sign-in problem</h1>
        <p>` + esc(msg) + `</p>
        <a class="btn" href="/login">Back to login</a>
    </div>
</body>
</html>`
	_, _ = w.Write([]byte(html))
}
