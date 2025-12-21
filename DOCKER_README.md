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

- `latest` - Latest stable release (currently v6.1.9)
- `v6.1.9` - Specific version tag
- `v6.x.x` - Major.minor.patch versions

### Architecture

Currently supports `amd64` (x86_64) architecture.

## 📝 Version 6.1.9 BloodMoon 🌙

Latest release includes:

- **Advanced Pagination System** - File counter with "Showing X of Y files", configurable items per page (5-250)
- **Team File Enhancements** - File descriptions visible in team files view with real-time search
- **Login Improvements** - Fixed critical double-login bug where users had to login twice
- **Keep Me Logged In Enhancement** - 30-day sessions exempt from inactivity timeout
- **UI Fixes** - Red delete buttons in Admin Files, fixed button layout with long file notes
- **Hourly Chunk Cleanup** - Automated cleanup of orphaned chunks every hour
- **Extended Retry Logic** - 50 retry attempts (~7.5 minutes) for better upload reliability

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

**Latest Version:** v6.0.2 BloodMoon 🌙
**Last Updated:** December 2025
**Image Size:** ~15MB compressed
