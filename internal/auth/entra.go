// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// entra.go — Microsoft Entra ID OpenID Connect provider and core flow primitives.
//
// Flow: Authorization Code + PKCE, per RFC 7636 / OAuth2 + OIDC.
//   1) /auth/entra/login    → generate state + PKCE verifier, store in short-lived cookies,
//                              redirect user-agent to Entra authorize endpoint.
//   2) /auth/entra/callback → validate state, exchange code (verifier) for tokens,
//                              verify ID-token signature against tenant JWKS, extract claims.
//
// The provider is rebuilt on each request rather than process-cached: config can change at
// runtime via admin UI and we don't want stale provider metadata. OIDC discovery is one HTTP
// call to the tenant's well-known endpoint and is cached internally by coreos/go-oidc.

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// EntraProvider bundles everything needed to drive an OIDC flow against one tenant.
// Returned by BuildEntraProvider; callers must check Enabled before using it.
type EntraProvider struct {
	Enabled  bool
	Config   *EntraConfig
	OAuth2   *oauth2.Config
	Verifier *oidc.IDTokenVerifier
	Provider *oidc.Provider
}

// EntraClaims is the subset of ID-token / userinfo claims we care about.
// Field names mirror Microsoft's standard OIDC claims.
type EntraClaims struct {
	OID               string   `json:"oid"`                // immutable tenant-scoped user identifier — use as ExternalID
	Subject           string   `json:"sub"`                // OIDC sub; fallback when oid missing
	Email             string   `json:"email"`              // verified email (sometimes empty)
	PreferredUsername string   `json:"preferred_username"` // UPN; often the real email
	Name              string   `json:"name"`               // display name
	TenantID          string   `json:"tid"`                // tenant the user belongs to
	Groups            []string `json:"groups"`             // group IDs, only if app registration requests them
}

// PrimaryEmail returns the most reliable email available from the claims.
// Order: verified email → preferred_username (if it looks like an email).
func (c *EntraClaims) PrimaryEmail() string {
	if e := strings.TrimSpace(c.Email); e != "" {
		return strings.ToLower(e)
	}
	if u := strings.TrimSpace(c.PreferredUsername); strings.Contains(u, "@") {
		return strings.ToLower(u)
	}
	return ""
}

// ExternalIdentifier returns the stable per-user ID we persist as User.ExternalID.
// Prefer the immutable Entra OID; fall back to OIDC sub for resilience.
func (c *EntraClaims) ExternalIdentifier() string {
	if s := strings.TrimSpace(c.OID); s != "" {
		return s
	}
	return strings.TrimSpace(c.Subject)
}

// BuildEntraProvider constructs an EntraProvider from the current configuration.
// Returns Enabled=false if SSO is disabled, force-local is set, or required fields
// are missing — caller can short-circuit to a 404 without leaking config state.
//
// ctx should be a request-scoped context; OIDC discovery is performed synchronously.
func BuildEntraProvider(ctx context.Context, cfg *EntraConfig) (*EntraProvider, error) {
	if cfg == nil || !cfg.ShouldShowSSOButton() {
		return &EntraProvider{Enabled: false}, nil
	}
	if cfg.TenantID == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		return &EntraProvider{Enabled: false}, errors.New("entra: tenant_id, client_id and client_secret all required")
	}

	issuer := "https://login.microsoftonline.com/" + cfg.TenantID + "/v2.0"

	// Multi-tenant app registrations point at /common, /organizations or
	// /consumers. Microsoft returns the literal {tenantid} placeholder in
	// the discovery document's `issuer` field for those endpoints, which
	// trips go-oidc's strict issuer match. Use InsecureIssuerURLContext so
	// discovery succeeds; we still verify the *actual* tenant in ExchangeCode
	// by checking the `tid` claim against cfg.TenantID (and against the user's
	// expected tenant for single-tenant deployments).
	discoveryCtx := ctx
	t := strings.ToLower(cfg.TenantID)
	if t == "common" || t == "organizations" || t == "consumers" {
		discoveryCtx = oidc.InsecureIssuerURLContext(ctx, issuer)
	}

	// Discovery: fetches /.well-known/openid-configuration and JWKS.
	// go-oidc caches keys internally with auto-refresh on signature failures.
	provider, err := oidc.NewProvider(discoveryCtx, issuer)
	if err != nil {
		return &EntraProvider{Enabled: false}, fmt.Errorf("entra: oidc discovery failed for tenant %s: %w", cfg.TenantID, err)
	}

	verifierCfg := &oidc.Config{ClientID: cfg.ClientID}
	// For multi-tenant aliases the per-token issuer differs from the
	// discovery issuer — skip the library check and rely on our explicit
	// `tid` validation downstream.
	if t == "common" || t == "organizations" || t == "consumers" {
		verifierCfg.SkipIssuerCheck = true
	}
	verifier := provider.Verifier(verifierCfg)

	oauth2Cfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURI,
		// openid + profile + email is the minimum for identity resolution.
		// Add "offline_access" if/when we want refresh tokens; we don't yet
		// because Entra is consulted only at login, not on every request.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &EntraProvider{
		Enabled:  true,
		Config:   cfg,
		OAuth2:   oauth2Cfg,
		Verifier: verifier,
		Provider: provider,
	}, nil
}

// PKCEPair is a fresh code_verifier + code_challenge for one auth attempt.
type PKCEPair struct {
	Verifier  string // raw verifier; store in cookie, send with token exchange
	Challenge string // S256 hash of verifier; sent in auth request
}

// GeneratePKCE creates a PKCE pair per RFC 7636 §4.
// Verifier: 43-char base64url of 32 random bytes. Challenge: base64url(SHA256(verifier)).
func GeneratePKCE() (*PKCEPair, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(raw)

	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	return &PKCEPair{Verifier: verifier, Challenge: challenge}, nil
}

// GenerateState produces a CSRF-protection state token (32 random bytes, base64url).
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// AuthCodeURL returns the URL to redirect the user-agent to. Includes state and
// the S256-hashed PKCE challenge per Microsoft's recommendations.
func (p *EntraProvider) AuthCodeURL(state string, pkce *PKCEPair) string {
	return p.OAuth2.AuthCodeURL(state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("code_challenge", pkce.Challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		// "select_account" forces the account picker; otherwise Entra silently
		// reuses the last sign-in if the user is already authenticated in the
		// browser, which surprises customers who share machines.
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
}

// ExchangeCode swaps the authorization code (plus PKCE verifier) for tokens,
// verifies the ID-token signature against the tenant JWKS, and returns the
// parsed claims. The raw access_token is intentionally discarded — we only
// need identity at sign-in time, not delegated API access.
func (p *EntraProvider) ExchangeCode(ctx context.Context, code, codeVerifier string) (*EntraClaims, error) {
	if !p.Enabled {
		return nil, errors.New("entra: provider disabled")
	}

	// Apply a hard timeout so a slow Entra response can't hang the request.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	token, err := p.OAuth2.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("entra: code exchange failed: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("entra: id_token missing from token response")
	}

	idToken, err := p.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("entra: id_token verification failed: %w", err)
	}

	var claims EntraClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("entra: failed to parse id_token claims: %w", err)
	}

	// Defense-in-depth: verify the token's tenant matches the configured tenant.
	// Microsoft already does this via the issuer check inside Verifier, but
	// belt-and-braces protects against a future config/provider mix-up.
	if p.Config.TenantID != "" && claims.TenantID != "" && claims.TenantID != p.Config.TenantID {
		return nil, fmt.Errorf("entra: tenant mismatch: got %s, expected %s", claims.TenantID, p.Config.TenantID)
	}

	if claims.PrimaryEmail() == "" {
		return nil, errors.New("entra: id_token has no email or preferred_username claim")
	}
	if claims.ExternalIdentifier() == "" {
		return nil, errors.New("entra: id_token has neither oid nor sub claim")
	}

	return &claims, nil
}

// DomainAllowed reports whether the email's domain is on the allow-list.
// Empty allow-list = any domain accepted.
func (p *EntraProvider) DomainAllowed(email string) bool {
	if p == nil || p.Config == nil || len(p.Config.AllowedDomains) == 0 {
		return true
	}
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return false
	}
	domain := strings.ToLower(email[at+1:])
	for _, d := range p.Config.AllowedDomains {
		if domain == d {
			return true
		}
	}
	return false
}
