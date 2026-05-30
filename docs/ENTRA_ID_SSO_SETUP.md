# Entra ID (Azure AD) SSO Setup

Configure Microsoft Entra ID single sign-on for **WulfVault** v7.1.0 "Aurora".

WulfVault's identity provider is configured entirely from the **Settings →
Identity Providers** page in the admin UI — there are no environment variables
to set. The values are stored in the database, and the client secret is
encrypted at rest with AES-256-GCM.

> WulfVault also supports a **Generic OIDC** provider (Google, Okta, Keycloak,
> etc.) selected via the "Provider type" field. This guide covers Microsoft
> Entra ID; the Generic OIDC flow is the same except you supply an **Issuer
> URL** instead of a Tenant ID.

---

## 1. Register an App in Azure

1. Go to **Azure Portal → Microsoft Entra ID → App registrations → New registration**.
2. Name it `WulfVault SSO`.
3. Set the redirect URI (platform: **Web**) to your WulfVault host plus the
   callback path:
   `https://files.example.com/auth/oidc/callback`
4. Click **Register**.

> The canonical callback path is `/auth/oidc/callback`. The legacy
> `/auth/entra/callback` path also still works (both map to the same handler),
> but new installs should use `/auth/oidc/callback`. Use your real public
> hostname in place of `files.example.com`, and it must match `SERVER_URL`.

---

## 2. Create a Client Secret

1. Go to **Certificates & secrets → New client secret**.
2. Copy the secret **value** (not the secret ID) — you will paste it into WulfVault.

---

## 3. Note IDs

From the app's **Overview** page, copy:

- **Application (client) ID**
- **Directory (tenant) ID**

---

## 4. Configure WulfVault

In WulfVault, go to **Settings → Identity Providers**, select provider type
**Microsoft Entra ID**, and fill in:

| Field | Value |
|-------|-------|
| Enabled | On |
| Provider display name | optional, e.g. `Microsoft` (shown on the login button) |
| Client ID | `YOUR_CLIENT_ID` (Application/client ID) |
| Client secret | the secret **value** from step 2 (stored encrypted; leave blank when editing to keep the existing one) |
| Tenant ID | `YOUR_TENANT_ID` (Directory/tenant ID, or `common` / `organizations` / `consumers`) |
| Redirect URI | `https://files.example.com/auth/oidc/callback` |
| Auto-provision | optional — create accounts on first SSO login |
| Default role | role for auto-provisioned users |
| Allowed domains | optional, comma-separated, e.g. `example.com` |
| Force local accounts only | optional — hides the SSO button and rejects callbacks |

For Entra, the OIDC issuer is derived automatically from the tenant ID
(`https://login.microsoftonline.com/YOUR_TENANT_ID/v2.0`); the Issuer URL field
is only used for the Generic OIDC provider type.

Save the settings, then use the SSO button on the login page.

---

## Notes

- The redirect URI you enter in WulfVault must exactly match the redirect URI
  registered in Azure, including the `/auth/oidc/callback` path.
- When **Allowed domains** is set, only users whose email domain is in the list
  may sign in.
- **Force local accounts only** overrides **Enabled**: it hides the SSO button
  and rejects SSO callbacks, leaving only username/password login.
- The client secret is stored encrypted at rest (AES-256-GCM); it is never kept
  in plaintext on disk or in environment variables.
