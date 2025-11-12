# Sharecare - Enterprise File Sharing Platform

**Version 3.3.0** | **Self-Hosted** | **Open Source** | **AGPL-3.0**

Sharecare is a professional-grade, self-hosted file sharing platform designed for organizations that demand security, accountability, and complete control over their data. Built with Go for exceptional performance and reliability, Sharecare provides a complete alternative to commercial file transfer services, eliminating subscription costs while offering superior features: multi-user management with role-based access, per-user storage quotas, comprehensive audit trails with email tracking, branded download pages, two-factor authentication, self-service password management, file request portals, and GDPR-compliant account deletion.

**Perfect for:** Law enforcement agencies, healthcare providers, legal firms, creative agencies, government departments, educational institutions, and any organization handling sensitive or large files that require detailed download tracking, compliance documentation, and enterprise-grade security.

**Architecturally inspired by [Gokapi](https://github.com/Forceu/Gokapi)** — Sharecare is a complete rewrite (~95% new code) adding multi-user management, audit logging, email integration, 2FA, branding, and enterprise features.

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

### 🚀 File Sharing & Transfer
- **Drag-and-drop upload interface** - Modern, intuitive file upload experience
- **Large file support** - Files up to 5GB+ (configurable, tested with video surveillance footage)
- **Two sharing modes:**
  - **Authenticated downloads** - Recipients create secure download accounts (email + password)
  - **Direct download links** - No authentication required for quick sharing
- **Password-protected files** - Add extra security layer with password protection per file
- **Expiring shares** - Auto-delete after X downloads or Y days (or both)
- **Custom expiration settings** - Flexible download limits (1-999) and date-based expiration
- **Upload request portals** - Create shareable links for others to upload files to you
- **Email integration** - Send download links directly via email with customizable templates
- **File preview & metadata** - View file details, size, upload date, and download statistics
- **Trash system** - Deleted files kept for configurable retention period (1-365 days) with restore capability

### 👥 User Management & Access Control
- **Role-based access:**
  - **Super Admin** - Full system control, user management, branding, settings
  - **Admin users** - Manage users and view all files across the system
  - **Regular users** - Upload and share files within their storage quota
  - **Download accounts** - Automatically created for authenticated downloads with self-service portal
- **Per-user storage quotas** - Individually configurable storage limits (MB to TB)
- **User dashboard** - Real-time quota usage, file management, and download statistics
- **Active/inactive status** - Temporarily disable users without deletion
- **Bulk user operations** - Efficient management of multiple users
- **Download account portal** - Recipients can view their download history and manage their accounts

### 📊 Download Tracking & Accountability
- **Complete audit trail:**
  - Track exactly **who** downloaded files (email addresses for authenticated downloads)
  - Record **when** downloads occurred (precise timestamps)
  - Log **from where** downloads originated (IP addresses with configurable privacy controls)
- **Per-file download history** - View detailed download logs for each file
- **Exportable reports** - Download tracking data in CSV format for compliance
- **Real-time statistics** - Dashboard shows total files, downloads, and storage usage
- **Download count limits** - Automatically expire files after reaching download threshold
- **Email notifications** - Optional notifications when files are downloaded (configurable)

### 🔐 Security & Authentication
- **Two-Factor Authentication (2FA):**
  - TOTP-based (compatible with Google Authenticator, Authy, etc.)
  - Backup codes for account recovery
  - Regenerable backup codes with old code invalidation
  - Per-user 2FA enrollment
- **Password security:**
  - bcrypt hashing with cost factor 12
  - Self-service password change for all user types
  - Password reset via email with secure tokens (24-hour expiration)
  - Minimum password length enforcement (8 characters)
- **Session management:**
  - Secure session cookies with automatic expiration (24 hours configurable)
  - SameSite cookies for CSRF protection
  - Secure logout with session invalidation
- **File access control:**
  - Secure random hash generation for download links (128-bit entropy)
  - Optional password protection per file
  - Automatic link expiration
  - No file enumeration or directory listing
- **Privacy controls:**
  - Optional IP address logging (GDPR-configurable)
  - GDPR-compliant download account self-deletion
  - Self-service data export for download accounts

### 🎨 Branding & Customization
- **Full branding control:**
  - Upload custom logo (replaces default Sharecare branding)
  - Custom primary and secondary colors for entire interface
  - Custom company name displayed throughout system
  - Branded download pages shown to all recipients
- **Configurable system settings:**
  - Trash retention period (1-365 days)
  - Default storage quota for new users
  - Maximum file size limits
  - Server URL and port configuration
- **Automated maintenance:**
  - Scheduled cleanup of expired files
  - Automatic trash purging based on retention policy
  - Database optimization and maintenance

### 🌐 Email & Notifications
- **Multiple email providers:**
  - SMTP configuration for self-hosted email
  - Brevo (SendInBlue) API integration for transactional email
  - Encrypted credential storage
  - Test email functionality before deployment
- **Email templates:**
  - Password reset emails with secure tokens
  - File sharing notifications with download links
  - Custom branding in all email communications
  - Professional HTML email templates
- **Email tracking:**
  - Log all sent emails with timestamps
  - Track email delivery status
  - Audit trail for compliance

### 📁 File Request System
- **Inbound file collection:**
  - Create upload request links for receiving files from others
  - Customizable upload limits (file size and count)
  - Expiration dates for upload requests
  - Password protection for upload portals
- **Use cases:**
  - Collect files from customers or contractors
  - Receive large files without email attachments
  - Temporary upload portals with time limits
  - Anonymous file submission with accountability

### 🔧 Administration & Management
- **Comprehensive admin dashboard:**
  - System-wide statistics (total files, downloads, users, storage)
  - Recent activity monitoring
  - User growth analytics
  - Quick access to all management functions
- **File management:**
  - View all files across all users
  - Search and filter capabilities
  - Delete files with trash safety net
  - Restore accidentally deleted files
  - Permanent deletion from trash
- **User administration:**
  - Create, edit, and delete users
  - Manage download accounts
  - Adjust quotas on the fly
  - Toggle user active/inactive status
- **System settings:**
  - Configure server URL and port
  - Set system-wide defaults
  - Manage trash retention
  - Control privacy and logging settings

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
| `MAX_FILE_SIZE_MB` | Maximum file size in MB | `2000` (2 GB) |
| `DEFAULT_QUOTA_MB` | Default storage quota per user (MB) | `5000` (5 GB) |
| `SESSION_TIMEOUT_HOURS` | Session expiration time | `24` |
| `TRASH_RETENTION_DAYS` | Days to keep deleted files | `5` |

### Admin Settings (Web UI)

After logging in as admin, configure:
- **Branding** - Logo, colors, company name
- **Storage Quotas** - Set custom limits per user (default: 5 GB per user)
- **Trash Retention** - How long deleted files are kept (default: 5 days, range: 1-365 days)
- **File Size Limits** - Maximum upload size (default: 2 GB, configurable up to 5GB+)
- **Session Timeout** - Login session duration (default: 24 hours)
- **IP Logging** - Enable/disable IP address tracking (default: disabled)

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

## Server Restart Feature

The Admin Settings page includes a **"Restart Server"** button that is **currently disabled** until systemd service is installed.

### Why is it disabled?

The restart button requires a process manager (systemd, supervisor, etc.) to automatically restart the server after shutdown. Without this, clicking the button will stop the server without restarting it.

### How to enable it

1. **Install systemd service** (requires sudo):
   ```bash
   sudo cp /tmp/sharecare.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable sharecare
   sudo systemctl start sharecare
   ```

2. **Uncomment the restart button** in the code:
   - Open `internal/server/handlers_admin.go`
   - Find the section marked `<!-- RESTART SERVER BUTTON - DISABLED`
   - Remove the `<!--` and `-->` comment markers
   - Also uncomment the JavaScript function at the bottom
   - Rebuild: `go build -o sharecare cmd/server/main.go`
   - Restart the service: `sudo systemctl restart sharecare`

3. **The button will now work!** It will use `systemctl restart sharecare` to gracefully restart the server.

See [DEPLOYMENT.md](DEPLOYMENT.md) for complete deployment and autostart instructions.

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

**Why AGPL-3.0?** This license ensures that if anyone uses Sharecare to provide a service over a network (like SaaS), they must share their modifications with the community. This prevents companies from taking the code, making improvements, and keeping them proprietary. It protects the open-source nature of the project while requiring attribution and source disclosure for all network use.

**Architecturally inspired by Gokapi** by Forceu - https://github.com/Forceu/Gokapi

See [NOTICE.md](NOTICE.md) for full attribution and license information.

---

## Support

- **Issues:** https://github.com/Frimurare/Sharecare/issues
- **Discussions:** https://github.com/Frimurare/Sharecare/discussions
- **Documentation:** https://github.com/Frimurare/Sharecare/wiki

---

**Made for organizations that need secure, accountable file sharing for large files**
