# WulfVault Configuration Reference

Configuration for **WulfVault** v7.1.0 "Aurora".

WulfVault has two layers of configuration:

1. **Process configuration** — command-line flags / environment variables read
   at startup that control how the server boots (port, directories, public URL,
   initial admin).
2. **Application settings** — everything else (SSO, SMTP, branding, quotas,
   retention, etc.) is configured by an administrator through the **Settings**
   UI and persisted in the database. These are *not* environment variables.

---

## Process Configuration

Defined in `cmd/server/main.go`. Each flag falls back to an environment
variable, which falls back to a default.

| Flag | Environment variable | Default | Description |
|------|----------------------|---------|-------------|
| `-port` | `PORT` | `8080` | HTTP listen port |
| `-data` | `DATA_DIR` | `./data` | Directory for the SQLite database and master key |
| `-uploads` | `UPLOADS_DIR` | `./uploads` | Directory for uploaded files |
| `-url` | `SERVER_URL` | `http://localhost:8080` | Public base URL (used in generated links and the SSO redirect hint) |
| `-setup` | — | `false` | Force the initial-setup path |

Initial admin (first run only, when no users exist):

| Environment variable | Default | Description |
|----------------------|---------|-------------|
| `ADMIN_EMAIL` | `admin@localhost` | Email of the first super-admin |
| `ADMIN_PASSWORD` | _(random, printed to logs once)_ | Password of the first super-admin |

> A `server_url` / `port` value stored in the database (set via the admin UI)
> takes precedence over the environment/flag values at runtime.

The SQLite database is always created as **`wulfvault.db`** inside the data
directory (`internal/database/database.go`). There is no setting to override the
database filename.

---

## Identity Provider / SSO

> SSO is **not** configured via environment variables. An administrator
> configures it in the web UI under **Settings → Identity Providers**. Settings
> are stored as individual rows in the Configuration table (keys keep an
> `entra_` prefix for install continuity), and the client secret is encrypted
> at rest with AES-256-GCM (see [Secrets](#secrets)).

WulfVault v7.1 supports one identity provider per install: **Microsoft Entra ID**
or a **Generic OIDC** provider. The UI fields and their Configuration keys:

| Field | Stored key | Description |
|-------|------------|-------------|
| Enabled | `entra_enabled` | Turns the SSO login button on/off |
| Force local accounts only | `entra_force_local_only` | Hides the SSO button and rejects callbacks; overrides Enabled |
| Provider type | `entra_provider_type` | `entra` or `generic_oidc` |
| Provider display name | `entra_provider_display_name` | Label on the login button (default "Microsoft" / "SSO") |
| Client ID | `entra_client_id` | Provider "Application (client) ID" |
| Client secret | `entra_client_secret_enc` | Client secret value, encrypted at rest |
| Tenant ID | `entra_tenant_id` | Azure "Directory (tenant) ID" (or `common`/`organizations`/`consumers`); Entra only |
| Issuer URL | `entra_generic_issuer_url` | OIDC issuer; Generic OIDC only |
| Scopes | `entra_generic_scopes` | Space-separated; default `openid email profile`; Generic OIDC only |
| Redirect URI | `entra_redirect_uri` | Full callback URL registered at the provider (canonical path `/auth/oidc/callback`; legacy `/auth/entra/callback` also works) |
| Auto-provision | `entra_auto_provision` | Create users on first SSO login |
| Default role | `entra_default_role` | Role assigned to auto-provisioned users |
| Allowed domains | `entra_allowed_domains` | Comma-separated email domains permitted to sign in (empty = any) |

For Entra, the OIDC issuer is derived from the tenant ID. The
`entra_generic_issuer_url` field is used only for Generic OIDC providers.

See [ENTRA_ID_SSO_SETUP.md](ENTRA_ID_SSO_SETUP.md) for the Azure
app-registration walkthrough.

---

## Secrets

Secrets such as the OIDC client secret and email API keys are encrypted at rest
using **AES-256-GCM** (`internal/secrets/secrets.go`). One master key per
install is generated on first use and stored in the database (the
`secret_master_key` configuration row, hex-encoded; a legacy
`email_encryption_key` row is honored first for older installs). Because the key
lives in `wulfvault.db`, backing up the data directory preserves it — there is
no separate key file or key environment variable.

---

## Storage

Uploaded files are stored under the uploads directory (`UPLOADS_DIR`,
default `./uploads`).

---

For SSO provider setup, see [ENTRA_ID_SSO_SETUP.md](ENTRA_ID_SSO_SETUP.md).
