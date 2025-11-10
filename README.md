# Manvarg Sharecare - Secure File Sharing System

A lightweight, self-hosted file sharing system with multi-user support, storage quotas, and detailed download tracking.

**Based on [Gokapi](https://github.com/Forceu/Gokapi)** - See [NOTICE.md](NOTICE.md) for attribution.

## Features

### Core Functionality
- ✅ **Multi-user authentication** (Super Admin, Admin, Regular Users, Download Accounts)
- ✅ **Storage quotas** per user with usage tracking
- ✅ **File deduplication** (identical files stored once)
- ✅ **Two download modes:**
  - Authenticated downloads (requires recipient account creation)
  - Direct links (no authentication)
- ✅ **Download tracking** - Know exactly who downloaded what and when
- ✅ **Expiring file shares** - Auto-delete after X downloads or Y days
- ✅ **Copy-link buttons** for easy sharing
- ✅ **Admin dashboard** with user management and system statistics
- ✅ **User dashboard** with file management and storage usage

### Customization
- ✅ **Configurable branding** - Upload custom logo, set colors, company name
- ✅ **Flexible configuration** - Adjust server URL, port, storage paths, quotas
- ✅ **Multiple admins** - Support for multiple administrators

### Security
- ✅ **Password hashing** with bcrypt
- ✅ **Session management** with automatic expiration
- ✅ **CSRF protection**
- ✅ **Secure random hash generation** for file links
- ✅ **Optional IP tracking** for downloads

## Quick Start

### Docker (Recommended for Proxmox LXC)

```bash
docker run -d \
  -p 8080:8080 \
  -v ./data:/data \
  -v ./uploads:/uploads \
  -e SERVER_URL=https://files.yourdomain.com \
  -e ADMIN_EMAIL=admin@yourdomain.com \
  sharecare/sharecare:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  sharecare:
    image: sharecare/sharecare:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./uploads:/uploads
    environment:
      - SERVER_URL=https://files.yourdomain.com
      - ADMIN_EMAIL=admin@yourdomain.com
      - ADMIN_PASSWORD=changeme
      - MAX_FILE_SIZE_MB=5000
      - DEFAULT_QUOTA_MB=10000
    restart: unless-stopped
```

### Manual Installation

1. Download the binary for your platform:
   ```bash
   wget https://github.com/Frimurare/Sharecare/releases/latest/download/sharecare-linux-amd64
   chmod +x sharecare-linux-amd64
   ```

2. Create a configuration file:
   ```bash
   ./sharecare-linux-amd64 --setup
   ```

3. Run the server:
   ```bash
   ./sharecare-linux-amd64 --config config.yaml
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_URL` | Public URL of the server | `http://localhost:8080` |
| `PORT` | Server port | `8080` |
| `DATA_DIR` | Data directory for database and config | `./data` |
| `UPLOADS_DIR` | Directory for uploaded files | `./uploads` |
| `ADMIN_EMAIL` | Initial admin email | `admin@localhost` |
| `ADMIN_PASSWORD` | Initial admin password | Random (printed on first run) |
| `MAX_FILE_SIZE_MB` | Maximum file size in MB | `2000` |
| `DEFAULT_QUOTA_MB` | Default storage quota per user | `5000` |

### Admin Settings (Configurable in Web UI)

- **Branding**: Company name, logo, primary/secondary colors
- **File Expiration**: Default expiration policies
- **Download Authentication**: Require auth by default or allow direct links
- **Storage Quotas**: Set different quotas for different users
- **IP Tracking**: Enable/disable IP address logging

## Usage

### For Admins

1. **Login** at `https://your-domain.com/admin`
2. **Create users** in the User Management section
3. **Set quotas** for each user
4. **Configure branding** in Settings
5. **Monitor downloads** and storage usage in Dashboard

### For Users

1. **Login** at `https://your-domain.com`
2. **Drag & drop** files to upload
3. **Set expiration** (downloads and/or time)
4. **Choose link type:**
   - **Authenticated**: Recipient must create download account
   - **Direct**: Anyone with link can download
5. **Copy link** and share via email, Teams, etc.
6. **Track downloads** in your dashboard

### For Download Recipients (Authenticated Mode)

1. **Click download link**
2. **Create account** with email + password
3. **Download file**
4. Account can be reused for future downloads

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/Frimurare/Sharecare.git
cd Sharecare

# Install dependencies
go mod download

# Build
go build -o sharecare cmd/server/main.go

# Run
./sharecare
```

### Running Tests

```bash
go test ./...
```

## Deployment on Proxmox LXC

1. Create Ubuntu/Debian LXC container
2. Install Docker:
   ```bash
   apt update && apt install -y docker.io docker-compose
   ```
3. Deploy using Docker Compose (see above)
4. Configure reverse proxy (nginx/Caddy) for HTTPS

## Use Cases

- **Video Surveillance**: Share exported video from Milestone XProtect or OpenEye
- **Evidence Chain**: Trackable downloads for legal purposes
- **Document Sharing**: Share system manuals, reports with customers
- **Large File Transfer**: Alternative to WeTransfer/Sprend
- **Customer Service**: Branded file sharing for service agreements

## API

REST API available for automation. See [API.md](docs/API.md) for details.

Endpoints:
- `/api/v1/upload` - Upload file
- `/api/v1/files` - List files
- `/api/v1/download/:id` - Download file
- `/api/v1/users` - Manage users (admin only)

## License

This project is licensed under the **AGPL-3.0** license, same as Gokapi.

See [LICENSE](LICENSE) for the full license text.

## Attribution

Based on **Gokapi** by Forceu - https://github.com/Forceu/Gokapi

See [NOTICE.md](NOTICE.md) for full attribution.

## Support

- **Issues**: https://github.com/Frimurare/Sharecare/issues
- **Documentation**: https://github.com/Frimurare/Sharecare/wiki

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Security

Found a security vulnerability? Please email ulf@manvarg.se instead of creating a public issue.

---

**Made with ❤️ for surveillance system customers and privacy-conscious file sharing**
