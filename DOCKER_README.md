# WulfVault - Secure File Transfer System

![Docker Image Version](https://img.shields.io/docker/v/frimurare/wulfvault?sort=semver)
![Docker Image Size](https://img.shields.io/docker/image-size/frimurare/wulfvault/latest)
![Docker Pulls](https://img.shields.io/docker/pulls/frimurare/wulfvault)

WulfVault is a secure, self-hosted file transfer system built with Go. Perfect for organizations that need a private, GDPR-compliant file sharing solution with advanced features like team management, two-factor authentication, single sign-on, and comprehensive audit logging.

## 🌟 Key Features

- **Secure File Sharing** - Upload and share files securely with granular permissions
- **Team Management** - Organize users into teams with dedicated file spaces
- **Two-Factor Authentication** - Enhanced security with TOTP-based 2FA
- **Single Sign-On** - Microsoft Entra ID and Generic OIDC, configured in the admin UI
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

A `docker-compose.yml` is included in the repository:

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
      - SERVER_URL=http://localhost:8080
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
docker compose up -d
```

To build the image locally instead of pulling, replace the `image:` line with
`build: .` (the repository includes a `Dockerfile`).

## 🔧 Configuration

### Environment Variables

These are read at process startup (see `cmd/server/main.go`).

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_URL` | `http://localhost:8080` | Public URL of your WulfVault instance |
| `PORT` | `8080` | Port to listen on |
| `DATA_DIR` | `/data` | Directory for the SQLite database and master key (set by the image) |
| `UPLOADS_DIR` | `/uploads` | Directory for uploaded files (set by the image) |
| `ADMIN_EMAIL` | `admin@localhost` | Initial admin email (first run only) |
| `ADMIN_PASSWORD` | _(random, shown in logs)_ | Initial admin password (first run only) |

> SSO (Entra ID / OIDC), SMTP, branding, quotas and other application settings
> are **not** environment variables — they are configured in the admin UI under
> **Settings** and stored in the database. See
> [docs/CONFIGURATION.md](docs/CONFIGURATION.md).

### Database & Secrets

The SQLite database is always created as **`wulfvault.db`** inside `DATA_DIR`
(e.g. `/data/wulfvault.db`). There is no variable to override the filename.

Secrets such as the OIDC client secret and email API keys are encrypted at rest
with **AES-256-GCM**. The master key is generated automatically on first run and
stored inside the database (`secret_master_key` configuration row) — it lives in
the same `wulfvault.db` file, so backing up the `/data` volume preserves it.

### Volumes

- `/data` - Contains the SQLite database (`wulfvault.db`) and its WAL files
- `/uploads` - Stores all uploaded files

**Important:** Always mount these as volumes to persist data across container restarts.

## 📊 First Run

On first startup, WulfVault creates an admin user:

```
Username: admin
Email:    admin@localhost (or $ADMIN_EMAIL)
Password: <randomly generated unless $ADMIN_PASSWORD is set>
```

If you did not set `ADMIN_PASSWORD`, the generated password is printed to the
container logs once. Retrieve it with:

```bash
docker logs wulfvault | grep "Admin Password"
```

**Important:** Change this password immediately after first login!

## 🔐 Security Features

- **Password Hashing** - bcrypt (cost factor 12) for secure password storage
- **Session Management** - Secure session handling with a 24-hour session lifetime and a 10-minute inactivity timeout
- **2FA Support** - TOTP-based two-factor authentication
- **SSO** - Entra ID / OIDC with PKCE; client secret encrypted at rest
- **Audit Logging** - Complete audit trail of all actions
- **HTTPS Support** - Use behind a reverse proxy for HTTPS

## 🌐 HTTPS Setup

WulfVault runs on HTTP by default (port 8080). For HTTPS access, use a reverse proxy like Nginx or Caddy. This is the recommended approach for production deployments.

### Option 1: Caddy (Easiest - Auto SSL)

```bash
sudo apt install caddy

cat > /etc/caddy/Caddyfile <<EOF
files.example.com {
    reverse_proxy localhost:8080
}
EOF

sudo systemctl restart caddy
```

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

> When running behind an HTTPS reverse proxy, set `SERVER_URL` to your public
> `https://` URL so generated links and the SSO redirect URI are correct.

## 🩺 Health Check

`GET /health` returns `200 OK` when the server is running.

## 📦 Image Variants

### Tags

- `latest` - Latest stable release (currently `v7.1.0 Aurora`)
- `vX.Y.Z` — pin to a specific semver release

### Architecture

Currently supports `amd64` (x86_64) architecture.

## 📚 Documentation

- [Installation Guide](INSTALLATION.md)
- [Configuration Reference](docs/CONFIGURATION.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Entra ID / OIDC SSO Setup](docs/ENTRA_ID_SSO_SETUP.md)
- [Changelog](https://github.com/Frimurare/WulfVault/blob/main/CHANGELOG.md)

## 🐛 Issues & Support

Report issues on [GitHub Issues](https://github.com/Frimurare/WulfVault/issues),
or contact ulf@manvarg.se.

## 📜 License

Licensed under GNU Affero General Public License v3.0 (AGPL-3.0).

## 👤 Author

Ulf Holmström (Frimurare)

---

**Latest Version:** v7.1.0 Aurora
