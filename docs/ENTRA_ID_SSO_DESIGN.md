# Entra ID SSO — Design Sketch

**Status:** Draft (2026-05-30)
**Tracks:** [#29](https://github.com/Frimurare/WulfVault/issues/29)
**Target version:** 6.3.0

## Goal

Add Microsoft Entra ID (formerly Azure AD) single sign-on to WulfVault. Required by Prudencia Security and similar customers pursuing ISO 27001 — they need all app logins to flow through their Entra tenant so MFA, conditional access, and audit live in one place.

## Non-goals

- **LDAP.** Issue #29 asks for "OpenID or LDAP". Skip LDAP — no customer demand, legacy protocol, adds attack surface.
- **Replacing local auth.** Local accounts stay as the primary mechanism. SSO is additive.
- **Touching download accounts.** `DownloadAccount` (internal/models/DownloadAccount.go) and the `/d/{fileId}` route are completely separate from user auth and remain so. Anonymous and email-based download links keep working unchanged.
- **Breaking Evidence Courier integration.** Prudencia Evidence Courier (`Frimurare/Prudencia-Evidence-Courier`) consumes WulfVault via `POST /login`, `GET /api/whoami`, `POST /upload`, `POST /file/delete` using a dedicated service account. None of those endpoints are touched. Pattern stays the same in the OIDC PR — service accounts remain `IdentityProvider="local"` so password-form login keeps working; only end-user accounts get promoted to Entra when the customer enables SSO. Verified on v6.2.9 (2026-05-30): all four endpoints respond as before, and critically `/api/whoami` returns 401 (not 404) so EC's "old WulfVault version" detection does not trip.

## Hard requirements

1. Local accounts (username + password + optional TOTP) **always** work — break-glass for admins, fallback if Entra tenant is down.
2. Admin can flip a setting **"Force local accounts only"** post-install. When set, the Entra button disappears from `/login` and the callback rejects all Entra logins. Useful when a customer wants to disable SSO temporarily without uninstalling.
3. Download accounts are untouched — `/d/{fileId}` does not gain any Entra dependency.
4. Default after upgrade: Entra is **disabled** (no tenant configured yet → admin enables it from settings when ready).

## Settings (new rows in `Configuration` table)

| Key | Type | Default | Notes |
|---|---|---|---|
| `entra_enabled` | bool | `0` | Master toggle. When `0`, no Entra UI shown, callback returns 404. |
| `entra_force_local_only` | bool | `0` | Overrides `entra_enabled` for login UI. Admin escape hatch. |
| `entra_tenant_id` | string | `""` | GUID of customer's Entra tenant. |
| `entra_client_id` | string | `""` | App registration client ID. |
| `entra_client_secret` | string | `""` | App registration secret (stored encrypted — reuse same approach as SMTP password). |
| `entra_redirect_uri` | string | auto | Defaults to `https://<server_url>/auth/entra/callback`. |
| `entra_auto_provision` | bool | `1` | If `1`, first sign-in creates a User automatically with `entra_default_role`. If `0`, admin must pre-create the user (email-matched). |
| `entra_default_role` | int | `2` | UserLevel for auto-provisioned users (2 = regular User). |
| `entra_allowed_domains` | string | `""` | Comma-separated email domains. Empty = any tenant member allowed. |

All settable from a new **Settings → Identity Providers** page in the admin UI.

## User model changes

Add two fields to `internal/models/User.go`:

```go
IdentityProvider string // "local" (default) or "entra"
ExternalID       string // Entra OID claim; unique index when non-empty
```

Migration: existing users get `IdentityProvider = "local"`, `ExternalID = ""`. No data loss.

Behavior:
- Local users still authenticate via password as today.
- Entra users have `Password = ""` and cannot log in via the password form (rejected with "This account uses Microsoft sign-in").
- An admin can convert a local user → Entra user from the user-edit page (sets `IdentityProvider="entra"`, clears `Password`, sets `ExternalID` next time they sign in). One-way for now; reversal is a manual DB edit.

## Auth flow

```
User visits /login
  │
  ├─ Sees password form (always)
  └─ Sees "Sign in with Microsoft" button IF entra_enabled && !entra_force_local_only
       │
       ▼
  GET /auth/entra/login
       │  generates state + PKCE verifier, stores in short-lived cookie
       │  redirects to https://login.microsoftonline.com/{tenant}/oauth2/v2.0/authorize
       ▼
  User authenticates with Microsoft (MFA enforced by Entra, not us)
       │
       ▼
  GET /auth/entra/callback?code=...&state=...
       │  validate state cookie
       │  exchange code → id_token + access_token (PKCE)
       │  verify id_token signature against tenant JWKS
       │  extract claims: oid, email, name, preferred_username
       │  check entra_allowed_domains
       │  resolve identity:
       │    1. lookup user by ExternalID = oid → found? sign in
       │    2. lookup user by Email (case-insensitive) → found && IdentityProvider="local"?
       │       → reject with "Email already used by local account; admin must convert it"
       │    3. lookup user by Email → found && IdentityProvider="entra" && ExternalID=""?
       │       → bind: set ExternalID = oid, sign in
       │    4. not found && entra_auto_provision=1 → create user, sign in
       │    5. not found && entra_auto_provision=0 → reject with "Account not provisioned"
       ▼
  Create session (reuse auth.CreateSession), redirect to /
```

TOTP/2FA: skipped for Entra users — Entra handles MFA. The post-login `/2fa/verify` redirect only triggers for `IdentityProvider="local"` users.

## Library choice

- `github.com/coreos/go-oidc/v3/oidc` — handles discovery, JWKS, ID-token validation. ~1k stars, maintained, MIT.
- `golang.org/x/oauth2` — for the code exchange / PKCE flow. Standard Go ecosystem.

Both already common in Go OIDC implementations. No CGo, no vendor weirdness.

## Files touched

| File | Change |
|---|---|
| `internal/models/User.go` | Add `IdentityProvider`, `ExternalID` fields |
| `internal/database/users.go` | Migration (ALTER TABLE), `GetUserByExternalID()`, `GetUserByEmail()` helpers if missing |
| `internal/auth/entra.go` | **new** — OIDC provider init, login/callback handlers, identity resolver |
| `internal/server/server.go` | Register `/auth/entra/login` and `/auth/entra/callback` routes (public, no `requireAuth`) |
| `internal/server/handlers_auth.go` | Conditionally render SSO button in `renderLoginPage()`; in `handleLogin()` reject `IdentityProvider="entra"` accounts with a clear message |
| `internal/server/handlers_settings.go` (or wherever admin settings live) | New Identity Providers settings page + POST handler |
| `internal/server/handlers_users.go` | Admin "convert to Entra" action on user-edit page |
| `docs/2FA_SETUP.md` → add `docs/ENTRA_SETUP.md` | Customer-facing setup guide (app registration walkthrough, redirect URI, required claims) |
| `CHANGELOG.md` | 6.3.0 entry |

Estimated effort: ~2 days for the happy path + 1 day for admin UI + 1 day for setup-guide docs and customer testing with Prudencia's tenant. Call it a week including review.

## Open questions

1. **Group → role mapping?** Entra can send group claims (e.g. "WulfVault Admins" Entra group → UserLevel 1). Worth adding in v1, or defer to v6.3.1? Prudencia probably wants it; ISO auditors love role-driven access.
2. **Single sign-out?** Front-channel logout via Entra is fiddly. Default: local session destroy + redirect to Entra logout URL. Acceptable for v1.
3. **Multi-tenant?** v1 = single tenant per WulfVault install (matches the deployment model — one tenant per customer). Don't over-engineer.

## Out of scope for this issue

- Google Workspace SSO (separate issue if requested)
- SAML 2.0 (skip — OIDC covers modern needs)
- LDAP (rejected above)
