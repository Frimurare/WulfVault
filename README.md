# Manvarg Sharecare - Secure File Sharing System

A lightweight, self-hosted file sharing system with multi-user support, storage quotas, and detailed download tracking.

**Based on [Gokapi](https://github.com/Forceu/Gokapi)** - See [NOTICE.md](NOTICE.md) for attribution.

## Features

### Core Functionality
- ✅ **Multi-user authentication** (Super Admin, Admin, Regular Users, Download Accounts)
- ✅ **Per-user storage quotas** - Configurable storage limits per user with real-time usage tracking
- ✅ **Isolated file storage** - Each user has their own file list with unique share links
- ✅ **Two download modes:**
  - Authenticated downloads (requires recipient account creation)
  - Direct links (no authentication)
- ✅ **Download tracking** - Know exactly who downloaded what and when with IP addresses
- ✅ **Download history viewer** - See detailed download logs for each file with timestamps and IPs
- ✅ **Expiring file shares** - Auto-delete after X downloads or Y days
- ✅ **Copy-link buttons** for easy sharing
- ✅ **Admin dashboard** with user management and system statistics
- ✅ **User dashboard** with file management and storage usage

### Customization
- ✅ **Configurable branding** - Upload custom logo, set colors, company name
- ✅ **Flexible configuration** - Adjust server URL, port, storage paths, quotas
- ✅ **Multiple admins** - Support for multiple administrators
- ✅ **Trash/Recycle Bin** - Configurable retention period (1-365 days, default 5 days)
- ✅ **Automated cleanup** - Expired files automatically moved to trash, permanent deletion after retention period

### Security
- ✅ **Password hashing** with bcrypt (cost factor 12)
- ✅ **Session management** with automatic expiration (24 hours)
- ✅ **SameSite cookies** for CSRF mitigation
- ✅ **Secure random hash generation** for file links (128-bit entropy)
- ✅ **IP address logging** for all downloads (audit trail)

## User Types & Permissions

Sharecare supports three distinct user types:

### 1. **Admin Users** (Administrators)
- Full system access
- Manage users (create, edit, delete, set quotas)
- View all files in the system
- View detailed download history for any file (who, when, IP address)
- Access trash and restore deleted files
- Configure branding and system settings (including trash retention)
- View download logs and statistics
- Login at: `/admin`

### 2. **Regular Users** (File Uploaders)
- Upload and share files within their storage quota
- Create expiring file shares
- Set download limits and authentication requirements
- View their own files and download statistics
- **View detailed download history** for their own files (who downloaded, when, IP address)
- Delete files (moves to admin trash, configurable retention period)
- **Cannot see other users' files or their download history**
- Login at: `/login` or `/dashboard`

### 3. **Download Accounts** (Recipients)
- Created automatically when downloading authenticated files
- Reusable across multiple file downloads
- No upload permissions
- No dashboard access
- Email + password authentication
- Tracked in download logs

## Authenticated Download Flow

When a user uploads a file with "Require recipient authentication" enabled:

1. **File Upload**
   - User uploads file and checks "🔒 Require recipient authentication"
   - Generates unique download link (e.g., `https://your-domain.com/d/ABC123`)

2. **Recipient Receives Link**
   - Opens link in browser
   - Presented with login/registration form

3. **First-Time Download**
   - Recipient enters email + password
   - **Account created automatically** (Download Account type)
   - Password is hashed with bcrypt
   - Download begins immediately

4. **Subsequent Downloads**
   - If recipient receives another authenticated file
   - Can use same email + password
   - **No need to re-register**
   - System recognizes existing Download Account

5. **Download Tracking**
   - Every download is logged with:
     - Email address (if authenticated download)
     - Timestamp
     - IP address
     - File name and size
     - User agent
   - **Viewable by file owner and admins via "📊 History" button**
   - Shows table with date/time, downloader (email or "Anonymous"), and IP address
   - Authenticated downloads marked with 🔒 badge

### Benefits of Download Accounts

- **Accountability**: Know exactly who downloaded what with email and IP address
- **Audit Trail**: Perfect for compliance and evidence chains - all downloads logged with timestamps
- **Reusability**: Recipients don't need to register multiple times
- **Privacy**: Download accounts only see files explicitly shared with them
- **Security**: Passwords are bcrypt hashed, sessions expire automatically after 24 hours
- **IP Logging**: Every download tracked with source IP address for security and compliance

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
- **Per-User Storage Quotas**: Set custom storage limits for each user individually
- **Trash Retention**: Configure how many days deleted files remain in trash (1-365 days, default 5)
- **Automatic Cleanup**: Expired files moved to trash, permanent deletion after retention period

## Usage

### For Admins

1. **Login** at `https://your-domain.com/admin`
2. **Create users** in the User Management section
3. **Set custom storage quotas** for each user individually (e.g., 5GB for user A, 50GB for user B)
4. **Configure branding** in Settings
5. **View download history** for any file with IP addresses and timestamps
6. **Monitor downloads** and storage usage in Dashboard
7. **Manage trash** and restore accidentally deleted files

### For Users

1. **Login** at `https://your-domain.com`
2. **Drag & drop** files to upload
3. **Set expiration** (downloads and/or time)
4. **Choose link type:**
   - **Authenticated**: Recipient must create download account
   - **Direct**: Anyone with link can download
5. **Copy link** and share via email, Teams, etc.
6. **Track downloads** - Click "📊 History" button to see:
   - Who downloaded (email or Anonymous)
   - When (date and time)
   - From where (IP address)
   - Authentication status (🔒 badge for authenticated downloads)

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

- **Video Surveillance**: Share exported video from Milestone XProtect or OpenEye with audit trail
- **Evidence Chain**: Complete download tracking with IP addresses for legal compliance
- **Document Sharing**: Share system manuals, reports with customers (each user has isolated file space)
- **Large File Transfer**: Alternative to WeTransfer/Sprend with custom quotas per user
- **Customer Service**: Branded file sharing for service agreements
- **Multi-tenant file sharing**: Different storage quotas for different departments or customers

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
