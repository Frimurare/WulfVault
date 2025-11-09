# Sharecare - Installation Guide

Complete guide to installing and deploying Sharecare file sharing system.

## Quick Start (Docker - Recommended)

### Prerequisites
- Docker Engine 20.10+
- Docker Compose V2+
- 2GB+ RAM
- 10GB+ disk space

### Installation Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/Frimurare/Sharecare.git
   cd Sharecare
   ```

2. **Configure environment**
   ```bash
   cp docker-compose.yml docker-compose.local.yml
   ```

   Edit `docker-compose.local.yml` and set your values:
   ```yaml
   environment:
     - SERVER_URL=https://files.yourdomain.com
     - ADMIN_EMAIL=admin@yourdomain.com
     - ADMIN_PASSWORD=your-secure-password-here
     - MAX_FILE_SIZE_MB=5000
     - DEFAULT_QUOTA_MB=10000
   ```

3. **Start the service**
   ```bash
   docker-compose -f docker-compose.local.yml up -d
   ```

4. **Access the application**
   - Open browser: `http://localhost:8080`
   - Login with credentials from step 2
   - Start sharing files!

---

## Proxmox LXC Deployment

Perfect for running in Proxmox LXC containers on Ubuntu/Debian.

### 1. Create LXC Container

In Proxmox:
- Create new LXC container (Ubuntu 22.04 or Debian 12)
- Assign:  2 CPU cores, 2GB RAM, 20GB disk
- Enable nesting: `Options > Features > Nesting: Yes`
- Start container

### 2. Install Docker

SSH into the container:

```bash
# Update system
apt update && apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
apt install docker-compose-plugin -y

# Verify installation
docker --version
docker compose version
```

### 3. Deploy Sharecare

```bash
# Create directory
mkdir -p /opt/sharecare
cd /opt/sharecare

# Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  sharecare:
    image: sharecare/sharecare:latest
    container_name: sharecare
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./uploads:/uploads
    environment:
      - SERVER_URL=http://your-lxc-ip:8080
      - PORT=8080
      - ADMIN_EMAIL=admin@yourdomain.com
      - ADMIN_PASSWORD=ChangeMe123!
      - MAX_FILE_SIZE_MB=5000
      - DEFAULT_QUOTA_MB=10000
    restart: unless-stopped
EOF

# Build image (if not using pre-built)
# git clone https://github.com/Frimurare/Sharecare.git
# cd Sharecare
# docker build -t sharecare/sharecare:latest .

# Start service
docker compose up -d

# Check logs
docker compose logs -f
```

### 4. Configure Reverse Proxy (Optional but Recommended)

For HTTPS access:

**Install Nginx:**
```bash
apt install nginx certbot python3-certbot-nginx -y
```

**Configure Nginx:**
```bash
cat > /etc/nginx/sites-available/sharecare << 'EOF'
server {
    listen 80;
    server_name files.yourdomain.com;

    client_max_body_size 5G;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts for large files
        proxy_connect_timeout 600;
        proxy_send_timeout 600;
        proxy_read_timeout 600;
        send_timeout 600;
    }
}
EOF

ln -s /etc/nginx/sites-available/sharecare /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

**Get SSL Certificate:**
```bash
certbot --nginx -d files.yourdomain.com
```

---

## Manual Installation (Binary)

### Prerequisites
- Go 1.21+
- SQLite3
- Linux/Windows/macOS

### Build from Source

1. **Clone and build**
   ```bash
   git clone https://github.com/Frimurare/Sharecare.git
   cd Sharecare

   # Install dependencies
   go mod download

   # Build
   go build -o sharecare ./cmd/server
   ```

2. **Run**
   ```bash
   # Run setup
   ./sharecare --setup \
     --port 8080 \
     --data ./data \
     --uploads ./uploads \
     --url http://localhost:8080

   # Note the admin password!
   ```

3. **Run as service**

   **Linux (systemd):**
   ```bash
   cat > /etc/systemd/system/sharecare.service << 'EOF'
   [Unit]
   Description=Sharecare File Sharing
   After=network.target

   [Service]
   Type=simple
   User=sharecare
   WorkingDirectory=/opt/sharecare
   ExecStart=/opt/sharecare/sharecare
   Restart=always

   Environment="PORT=8080"
   Environment="DATA_DIR=/opt/sharecare/data"
   Environment="UPLOADS_DIR=/opt/sharecare/uploads"
   Environment="SERVER_URL=http://localhost:8080"

   [Install]
   WantedBy=multi-user.target
   EOF

   systemctl daemon-reload
   systemctl enable sharecare
   systemctl start sharecare
   ```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_URL` | Public URL for the server | `http://localhost:8080` |
| `PORT` | Server port | `8080` |
| `DATA_DIR` | Database and config directory | `./data` |
| `UPLOADS_DIR` | File storage directory | `./uploads` |
| `ADMIN_EMAIL` | Initial admin email | `admin@localhost` |
| `ADMIN_PASSWORD` | Initial admin password | Random (shown at setup) |
| `MAX_FILE_SIZE_MB` | Maximum upload size (MB) | `2000` |
| `DEFAULT_QUOTA_MB` | Default user quota (MB) | `5000` |

### Configuration File

Located at `data/config.json`:

```json
{
  "serverUrl": "http://localhost:8080",
  "port": "8080",
  "dataDir": "./data",
  "uploadsDir": "./uploads",
  "maxFileSizeMB": 2000,
  "defaultQuotaMB": 5000,
  "saveIp": false,
  "branding": {
    "companyName": "Sharecare",
    "primaryColor": "#0066CC",
    "secondaryColor": "#333333",
    "logoPath": "",
    "logoBase64": "",
    "faviconPath": "",
    "footerText": "Secure File Sharing",
    "welcomeMessage": "Welcome to Sharecare - Secure File Sharing",
    "customCSS": ""
  }
}
```

---

## Upgrading

### Docker

```bash
cd /opt/sharecare
docker compose pull
docker compose up -d
```

### Binary

```bash
# Backup data
cp -r data data.backup
cp -r uploads uploads.backup

# Download new version
wget https://github.com/Frimurare/Sharecare/releases/latest/download/sharecare-linux-amd64
chmod +x sharecare-linux-amd64
mv sharecare-linux-amd64 sharecare

# Restart
systemctl restart sharecare
```

---

## Troubleshooting

### Permission Errors

```bash
# Fix ownership
chown -R 1000:1000 data uploads

# Fix permissions
chmod -R 755 data uploads
```

### Database Locked

```bash
# Stop service
docker compose down  # or systemctl stop sharecare

# Check for stuck processes
lsof data/sharecare.db

# Restart
docker compose up -d
```

### Out of Disk Space

```bash
# Check usage
df -h
du -sh uploads/*

# Clean expired files (automatic, but manual if needed)
# Login as admin > Settings > Clean Expired Files
```

### Network Issues

```bash
# Check if port is available
netstat -tlnp | grep 8080

# Check firewall
ufw status
ufw allow 8080/tcp
```

---

## Security Recommendations

1. **Change default admin password immediately**
2. **Use HTTPS in production** (Nginx + Let's Encrypt)
3. **Enable firewall** (only allow 80/443)
4. **Regular backups** of `data/` and `uploads/`
5. **Monitor logs** for suspicious activity
6. **Limit file sizes** based on your needs
7. **Set storage quotas** per user
8. **Use strong passwords** for all accounts

---

## Backup & Restore

### Backup

```bash
# Stop service
docker compose down

# Backup
tar -czf sharecare-backup-$(date +%Y%m%d).tar.gz data uploads

# Restart
docker compose up -d
```

### Restore

```bash
# Stop service
docker compose down

# Restore
tar -xzf sharecare-backup-YYYYMMDD.tar.gz

# Restart
docker compose up -d
```

---

## Support

- **GitHub Issues**: https://github.com/Frimurare/Sharecare/issues
- **Documentation**: https://github.com/Frimurare/Sharecare/wiki

---

## License

AGPL-3.0 - See [LICENSE](LICENSE) for details.

Based on [Gokapi](https://github.com/Forceu/Gokapi) by Forceu.
