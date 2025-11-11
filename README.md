# Sharecare - Enterprise File Sharing Platform

**Version 3.0.0** | **Production Ready** | **Self-Hosted** | **Open Source**

Sharecare is a professional-grade, self-hosted file sharing platform designed for organizations that demand security, accountability, and complete control over their data. Built with Go for exceptional performance and reliability, Sharecare eliminates the need for expensive cloud-based file transfer services while providing enterprise features like multi-user management, granular storage quotas, comprehensive audit trails, and automatic password reset functionality.

**Perfect for:** Law enforcement agencies, healthcare providers, legal firms, creative agencies, and any organization handling sensitive or large files that require detailed download tracking and GDPR compliance.

**Based on [Gokapi](https://github.com/Forceu/Gokapi)** - See [NOTICE.md](NOTICE.md) for attribution.

---

## Why Sharecare?

Many organizations need to share large files regularly but face challenges:
- Commercial file transfer services charge high fees per user or transfer
- Large video files (surveillance footage, recordings) exceed typical email limits
- Need to know exactly who downloaded what and when for compliance
- Want complete control over data security and retention

Sharecare solves this by providing:
- Self-hosted solution - your data stays on your infrastructure
- No per-transfer costs or user limits
- Complete download tracking with email addresses and timestamps
- Customizable storage quotas per user
- Support for files up to 5GB+ (configurable)

---

## Key Features

### File Sharing
- Drag-and-drop upload interface
- Files up to 5GB+ (configurable, tested with large video files)
- Two link types:
  - **Authenticated downloads** - Recipient creates account (email + password)
  - **Direct links** - No authentication required
- Password protection for sensitive files
- Expiring shares - auto-delete after X downloads or Y days
- Upload requests - Allow others to upload files to you

### User Management
- **Admin users** - Full system access, manage users, view all files
- **Regular users** - Upload and share files within their quota
- **Download accounts** - Automatically created for authenticated downloads
- Per-user storage quotas (configurable individually)
- Active/inactive user status

### Download Tracking & Accountability
- Know exactly **who** downloaded files (email address for authenticated downloads)
- See **when** files were downloaded (timestamps)
- Track **from where** downloads originated (IP addresses)
- Complete audit trail for compliance and evidence chains
- Downloadable download history per file

### Security & Privacy
- bcrypt password hashing (cost factor 12)
- Session management with automatic expiration (24 hours)
- Secure random hash generation for download links
- Optional password protection per file
- IP address logging for audit trails
- SameSite cookies for CSRF protection

### Customization
- Custom branding - Upload logo, set colors and company name
- Configurable trash retention (1-365 days)
- Automated cleanup of expired files
- Flexible file size limits and storage quotas

---

## Quick Start

### First-Time Setup

1. **Download and start Sharecare** (see installation methods below)

2. **Initial admin credentials:**
   - **Email:** `admin@sharecare.local`
   - **Password:** `SharecareAdmin2024!`

   **⚠️ IMPORTANT:** Change the admin password immediately after first login!

3. **Login to admin panel:**
   - Navigate to `http://your-server:8080/admin`
   - Use the default credentials above
   - Go to Settings and change your password

4. **Configure your instance:**
   - Set your server URL (Admin > Settings)
   - Customize branding (Admin > Branding)
   - Create regular users (Admin > Users)
   - Set storage quotas per user

5. **Start sharing files:**
   - Users login at `http://your-server:8080`
   - Drag and drop files to upload
   - Copy share links and send to recipients

---

## Installation

### Docker (Recommended)

```bash
docker run -d \
  --name sharecare \
  -p 8080:8080 \
  -v ./data:/data \
  -v ./uploads:/uploads \
  -e SERVER_URL=https://files.yourdomain.com \
  frimurare/sharecare:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  sharecare:
    image: frimurare/sharecare:latest
    container_name: sharecare
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./uploads:/uploads
    environment:
      - SERVER_URL=https://files.yourdomain.com
      - MAX_FILE_SIZE_MB=5000
      - DEFAULT_QUOTA_MB=10000
    restart: unless-stopped
```

Save as `docker-compose.yml` and run:
```bash
docker-compose up -d
```

### Build from Source

**Prerequisites:** Go 1.21+

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

**Default credentials on first run:**
- Email: `admin@sharecare.local`
- Password: `SharecareAdmin2024!`

See [INSTALLATION.md](INSTALLATION.md) for detailed deployment guides including Proxmox LXC, reverse proxy configuration, and SSL setup.

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_URL` | Public URL of the server | `http://localhost:8080` |
| `PORT` | Server port | `8080` |
| `DATA_DIR` | Data directory for database | `./data` |
| `UPLOADS_DIR` | Directory for uploaded files | `./uploads` |
| `MAX_FILE_SIZE_MB` | Maximum file size in MB | `2000` |
| `DEFAULT_QUOTA_MB` | Default storage quota per user (MB) | `5000` |

### Admin Settings (Web UI)

After logging in as admin, configure:
- **Branding** - Logo, colors, company name
- **Storage Quotas** - Set custom limits per user
- **Trash Retention** - How long deleted files are kept (default: 5 days)
- **File Size Limits** - Maximum upload size

---

## Use Cases

### Large Video File Sharing
Organizations with surveillance systems or video production need to share large video exports:
- Export video footage (often 1GB+ files)
- Upload to Sharecare with expiration and authentication
- Share link with investigators, management, or customers
- Track exactly who downloaded the footage and when
- Maintain complete audit trail for legal compliance

### Secure Document Distribution
Share sensitive documents with accountability:
- Service agreements and contracts
- System documentation and manuals
- Evidence files requiring chain of custody
- Financial reports with download tracking

### Team File Collaboration
Internal file sharing for distributed teams:
- Large design files and CAD drawings
- Project deliverables too large for email
- Marketing materials and video content
- Backup distribution to remote locations

### Requesting Files from Others
Create upload request links for:
- Collecting files from customers or contractors
- Receiving large files without email attachments
- Temporary upload portals with size and time limits
- Anonymous file submission with tracking

---

## User Workflows

### Uploading and Sharing Files

1. **Login** at `http://your-server/dashboard`
2. **Drag and drop** file (or click to browse)
3. **Set options:**
   - Expiration: Downloads limit or time limit or both
   - Authentication: Require recipient to create account (optional)
   - Password: Protect file with password (optional)
4. **Upload** and get shareable link
5. **Share link** via email, chat, or other channels
6. **Track downloads** - Click "History" button to see who downloaded

### Receiving Authenticated Files

1. **Click download link** received from sender
2. **Create download account** with email and password (first time only)
3. **Login** and download file
4. **Reuse account** for future authenticated downloads

### Admin User Management

1. **Login** to admin panel at `http://your-server/admin`
2. **Create users:**
   - Set email, password, user level
   - Assign storage quota (e.g., 5GB, 50GB, or custom)
   - Set active/inactive status
3. **Monitor usage** in dashboard
4. **View all files** and download history across system
5. **Manage trash** and restore accidentally deleted files

---

## API

Sharecare provides a REST API for automation and integrations.

**Basic endpoints:**
- `/api/upload` - Upload files programmatically
- `/api/files` - List user's files
- `/api/download/:id` - Download file by ID

**Authentication:** API requests require session cookies or token-based auth.

See full API documentation in [API.md](docs/API.md) (coming soon).

---

## Security

### Default Security Features

- Passwords hashed with bcrypt (cost factor 12)
- Secure random hash generation for download links (128-bit entropy)
- Session tokens with automatic expiration (24 hours)
- CSRF protection via SameSite cookies
- Files stored outside web root with access control
- IP address logging for all downloads
- No directory listing or file enumeration

### Recommended Production Setup

1. **Change default admin password immediately**
2. **Use HTTPS** - Deploy behind reverse proxy (nginx/Caddy) with SSL
3. **Enable firewall** - Only expose ports 80/443
4. **Regular backups** - Backup `./data` and `./uploads` directories
5. **Monitor logs** - Watch for suspicious download patterns
6. **Update regularly** - Keep Sharecare up to date
7. **Strong passwords** - Enforce password policies for all users

---

## Troubleshooting

### Can't login with default credentials

Make sure you're using:
- Email: `admin@sharecare.local`
- Password: `SharecareAdmin2024!`

If it still doesn't work, check the server logs for initialization errors.

### Files not uploading

- Check `MAX_FILE_SIZE_MB` environment variable
- Verify user has available storage quota
- Check disk space on server
- Review browser console for JavaScript errors

### Download links not working

- Verify `SERVER_URL` is set correctly in environment
- Check that files haven't expired
- Ensure file wasn't deleted or moved to trash
- Check server logs for errors

### More help

- Check [INSTALLATION.md](INSTALLATION.md) for detailed setup
- Review logs: `docker-compose logs -f sharecare`
- Open issue on GitHub: https://github.com/Frimurare/Sharecare/issues

---

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
Sharecare/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── auth/           # Authentication and sessions
│   ├── database/       # SQLite database operations
│   ├── models/         # Data models
│   └── server/         # HTTP handlers and routing
├── web/
│   ├── static/         # CSS, JavaScript, images
│   └── templates/      # HTML templates
├── INSTALLATION.md     # Detailed installation guide
├── LICENSE            # AGPL-3.0 license
└── README.md          # This file
```

### Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

---

## License

This project is licensed under the **AGPL-3.0** license - see [LICENSE](LICENSE) for details.

Based on **Gokapi** by Forceu - https://github.com/Forceu/Gokapi

See [NOTICE.md](NOTICE.md) for full attribution and license information.

---

## Support

- **Issues:** https://github.com/Frimurare/Sharecare/issues
- **Discussions:** https://github.com/Frimurare/Sharecare/discussions
- **Documentation:** https://github.com/Frimurare/Sharecare/wiki

---

**Made for organizations that need secure, accountable file sharing for large files**
