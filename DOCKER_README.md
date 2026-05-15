# WulfVault - Secure File Transfer System

![Docker Image Version](https://img.shields.io/docker/v/frimurare/wulfvault?sort=semver)
![Docker Image Size](https://img.shields.io/docker/image-size/frimurare/wulfvault/latest)
![Docker Pulls](https://img.shields.io/docker/pulls/frimurare/wulfvault)

WulfVault is a secure, self-hosted file transfer system built with Go. Perfect for organizations that need a private, GDPR-compliant file sharing solution with advanced features like team management, two-factor authentication, and comprehensive audit logging.

## 🌟 Key Features

- **Secure File Sharing** - Upload and share files securely with granular permissions
- **Team Management** - Organize users into teams with dedicated file spaces
- **Two-Factor Authentication** - Enhanced security with TOTP-based 2FA
- **Chunked Uploads** - Reliable upload of large files with automatic resume
- **Audit Logging** - Complete audit trail of all system activities
- **Download Accounts** - Create temporary accounts for external file recipients
- **File Request Portals** - Allow external users to upload files securely
- **Auto-Cleanup** - Automatic deletion of old files with trash recovery
- **Responsive UI** - Modern web interface that works on desktop and mobile
- **Branding Support** - Customize logo and company name

## 🚀 Quick Start

### Using Docker Run

```bash
docker run -d \
  --name wulfvault \
  -p 8080:8080 \
  -v wulfvault-data:/data \
  -v wulfvault-uploads:/uploads \
  -e SERVER_URL=http://your-domain.com:8080 \
  frimurare/wulfvault:latest
```

### Using Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  wulfvault:
    image: frimurare/wulfvault:latest
    container_name: wulfvault
    ports:
      - "8080:8080"
    volumes:
      - wulfvault-data:/data
      - wulfvault-uploads:/uploads
    environment:
      - SERVER_URL=http://your-domain.com:8080
      - PORT=8080
      - MAX_FILE_SIZE_MB=5000
      - DEFAULT_QUOTA_MB=10000
    restart: unless-stopped

volumes:
  wulfvault-data:
  wulfvault-uploads:
```

Then run:

```bash
docker-compose up -d
```

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_URL` | `http://localhost:8080` | Public URL of your WulfVault instance |
| `PORT` | `8080` | Port to listen on |
| `DATA_DIR` | `/data` | Directory for database and configuration |
| `UPLOADS_DIR` | `/uploads` | Directory for uploaded files |
| `MAX_FILE_SIZE_MB` | `5000` | Maximum file size in MB |
| `DEFAULT_QUOTA_MB` | `10000` | Default storage quota per user in MB |

### Volumes

- `/data` - Contains the SQLite database and configuration files
- `/uploads` - Stores all uploaded files

**Important:** Always mount these as volumes to persist data across container restarts.

## 📊 First Run

On first startup, WulfVault creates an admin user:

```
Username: admin
Password: <randomly generated>
```

The password is shown in the container logs. Retrieve it with:

```bash
docker logs wulfvault | grep "Admin Password"
```

**Important:** Change this password immediately after first login!

## 🔐 Security Features

- **Password Hashing** - Argon2id for secure password storage
- **Session Management** - Secure session handling with configurable timeouts
- **2FA Support** - TOTP-based two-factor authentication
- **Audit Logging** - Complete audit trail of all actions
- **File Encryption** - Optional encryption for stored files
- **CORS Protection** - Configurable CORS policies
- **HTTPS Support** - Use behind reverse proxy for HTTPS

## 🌐 HTTPS Setup

WulfVault runs on HTTP by default (port 8080). For HTTPS access, use a reverse proxy like Nginx or Caddy. This is the recommended approach for production deployments.

**Why use a reverse proxy?**
- Automatic SSL certificate management (Let's Encrypt)
- Professional SSL/TLS handling
- Load balancing support
- Better performance with caching

### Option 1: Caddy (Easiest - Auto SSL)

Caddy automatically obtains and renews SSL certificates from Let's Encrypt:

```bash
# Install Caddy
sudo apt install caddy

# Create Caddyfile
cat > /etc/caddy/Caddyfile <<EOF
files.example.com {
    reverse_proxy localhost:8080
}
EOF

# Start Caddy
sudo systemctl restart caddy
```

That's it! Caddy handles SSL automatically.

### Option 2: Nginx (Manual SSL)

```nginx
server {
    listen 443 ssl http2;
    server_name files.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    client_max_body_size 5000M;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # For chunked uploads
        proxy_request_buffering off;
        proxy_http_version 1.1;
    }
}
```

### Traefik Example

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.wulfvault.rule=Host(`files.example.com`)"
  - "traefik.http.routers.wulfvault.entrypoints=websecure"
  - "traefik.http.routers.wulfvault.tls.certresolver=letsencrypt"
  - "traefik.http.services.wulfvault.loadbalancer.server.port=8080"
```

## 📦 Image Variants

### Tags

- `latest` - Latest stable release (currently `v6.2.9 BloodMoon 🌙`)
- `v6.2.9`, `v6.2.7`, `v6.1.9`, ... — pin to a specific version
- `vX.Y.Z` — semver tags

### Architecture

Currently supports `amd64` (x86_64) architecture.

## 📝 What's new in v6.2.x BloodMoon 🌙

The v6.2 series shipped a series of email-pipeline fixes and added a
dedicated session-verification endpoint for API integrations such as
Prudencia Evidence Courier.

### v6.2.9 — Mailgun column migration on fresh installs

Fresh installs without an existing database failed to create the
Mailgun-related columns during schema bootstrap. The migration is now
idempotent and runs cleanly on both new and existing databases.

### v6.2.8 — 2FA double-submit race

Closed a race condition where the 2FA verification form could be
submitted twice in rapid succession and accept the second submission
against an already-consumed TOTP code. The handler now locks the code
the moment it is first verified.

### v6.2.7 — `GET /api/whoami` endpoint

Dedicated JSON endpoint to verify a session cookie without side
effects:

```json
200 OK
{
  "authenticated": true,
  "id": 123,
  "email": "user@example.com",
  "name": "User Name",
  "role": "user",
  "storage_used_mb": 42,
  "storage_quota_mb": 1000,
  "server_version": "6.2.7 BloodMoon 🌙",
  "two_factor_enabled": false
}

401 Unauthorized
{ "authenticated": false, "error": "Not authenticated" }
```

`Cache-Control: no-store` ensures fresh auth checks. Replaces the
previous practice of probing `/login` or `/dashboard` and HTML-scraping
the response. **Required by Prudencia Evidence Courier v1.0.6+.**

### v6.2.6 — CRITICAL: web-UI uploads dropped notification emails for 4 months

Files uploaded via the web dashboard between **v6.0.0 (December 2025)
and v6.2.5** silently dropped recipient notification emails. Three
issues conspired: `dashboard.js` never extracted `send_to_email` into
upload metadata, the chunked-upload handler never had any
email-sending code, and the SMTP provider required a password (which
blocked dev servers like MailHog and IP-relay production setups). The
legacy `POST /upload` endpoint was unaffected, which is why the bug
remained hidden. **All web-UI installs running v6.0.0 - v6.2.5 should
upgrade immediately.**

### v6.2.5 — Expiration reminder emails

Automatic reminders to file owners when their shares are about to
expire: a halfway reminder (~2-3 days out on a 5-day share) and an
urgent reminder 1 day before expiration with red urgency styling.
Includes file name, size, current download count, and a direct
download link. Runs every 6 hours via the cleanup scheduler.

### v6.2.4 — Email sender info, comments, English templates

Sender attribution and uploader comments now appear in download
notification emails. English email templates added alongside the
Swedish ones.

### v6.2.0 → v6.2.3 — BloodMoon foundation

The initial v6.2 line: hardened email pipeline, signed splash-link
URLs for share previews, and admin-settings refinements.

### v6.1.9 — Pagination + login fixes (December 2025)

Advanced pagination (file counter with "Showing X of Y", 5-250 per
page), team-file descriptions visible with real-time search, fix for
the double-login bug, 30-day "keep me logged in" sessions exempt from
inactivity timeout, hourly orphan-chunk cleanup, extended upload
retry logic (50 attempts ≈ 7.5 minutes).

For per-release detail see
[CHANGELOG.md](https://github.com/Frimurare/WulfVault/blob/main/CHANGELOG.md).

## 🔍 System Requirements

### Minimum

- 256MB RAM
- 500MB disk space
- Single CPU core

### Recommended

- 512MB RAM
- 1GB+ disk space (depending on usage)
- 2+ CPU cores for multiple concurrent uploads

## 📚 Documentation

- [GitHub Repository](https://github.com/Frimurare/WulfVault)
- [User Guide](https://github.com/Frimurare/WulfVault/blob/main/USER_GUIDE.md)
- [Changelog](https://github.com/Frimurare/WulfVault/blob/main/CHANGELOG.md)

## 🐛 Issues & Support

Report issues on [GitHub Issues](https://github.com/Frimurare/WulfVault/issues)

## 📜 License

Licensed under GNU Affero General Public License v3.0 (AGPL-3.0)

## 👤 Author

Ulf Holmström (Frimurare)

---

**Latest Version:** v6.2.9 BloodMoon 🌙
**Last Updated:** 2026-05-11
**Image Size:** ~14.5 MB compressed
