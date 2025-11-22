# WulfVault - Enterprise File Sharing Platform

**Version 4.7.9 Shotgun** | **Self-Hosted** | **Open Source** | **AGPL-3.0**

WulfVault is a professional-grade, self-hosted file sharing platform designed for organizations that demand security, accountability, and complete control over their data. Built with Go for exceptional performance and reliability, WulfVault provides a complete alternative to commercial file transfer services, eliminating subscription costs while offering superior features: multi-user management with role-based access, per-user storage quotas, enterprise-grade audit logging for compliance (GDPR, SOC 2, HIPAA), comprehensive download tracking, branded download pages, two-factor authentication, self-service password management, file request portals, and GDPR-compliant account deletion.

**Perfect for:** Law enforcement agencies, healthcare providers, legal firms, creative agencies, government departments, educational institutions, and any organization handling sensitive or large files that require detailed download tracking, compliance documentation, and enterprise-grade security.

**Architecturally inspired by [Gokapi](https://github.com/Forceu/Gokapi)** — WulfVault is a complete rewrite (~95% new code) adding multi-user management, audit logging, email integration, 2FA, branding, and enterprise features.

---

## Why WulfVault?

Many organizations need to share large files regularly but face challenges:
- Commercial file transfer services charge high fees per user or transfer
- Large video files (surveillance footage, recordings) exceed typical email limits
- Need to know exactly who downloaded what and when for compliance
- Want complete control over data security and retention

WulfVault solves this by providing:
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
  - **Authenticated downloads (v4.7.9+: DEFAULT)** - Recipients create secure download accounts (email + password) - **Now checked by default for enhanced security**
  - **Direct download links** - Optional: uncheck RequireAuth for quick sharing without authentication
- **Password-protected files** - Add extra security layer with password protection per file
- **Expiring shares** - Auto-delete after X downloads or Y days (or both)
- **Custom expiration settings** - Flexible download limits (1-999) and date-based expiration
- **Upload request portals** - Create shareable links for others to upload files to you
- **Email integration** - Send download links directly via email with customizable templates
- **File preview & metadata** - View file details, size, upload date, and download statistics
- **File comments/descriptions (v4.7+):**
  - Add notes and context to shared files
  - Comments visible in file details and admin views
  - Searchable in admin file management
  - Perfect for adding context like "Q3 Financial Report - Final Version"
- **Trash system with enhanced UI (v4.3+):**
  - Deleted files kept for configurable retention period (1-365 days) with restore capability
  - Modern gradient-styled action buttons: ♻️ Restore (green) and 🗑️ Delete Forever (red)
  - One-click restore or permanent deletion with visual feedback and hover effects
  - Complete audit trail: see who deleted files, when, and days remaining before auto-deletion
  - Fully responsive mobile layout with optimized touch targets
- **File Sharing Wisdom** - 180+ humorous one-liners on dashboards reminding users why email attachments fail

### 👥 User Management & Access Control
- **Role-based access:**
  - **Super Admin** - Full system control, user management, branding, settings
  - **Admin users** - Manage users and view all files across the system
  - **Regular users** - Upload and share files within their storage quota
  - **Download accounts** - Automatically created for authenticated downloads with self-service portal
- **Team collaboration (v4.2+):**
  - **Create teams** - Organize users into teams for shared file access
  - **Multi-team file sharing** - Share files with multiple teams simultaneously
  - **Team management UI** - Add/remove team members with visual badges
  - **Team roles** - Owner, Admin, and Member permissions
  - **Team storage quotas** - Per-team storage limits and usage tracking
  - **Smart team badges** - Files show team names or count with hover tooltips
  - **Real-time team sync** - Instant updates when files are shared/unshared
  - **Team filter dropdown (v4.7.3+)** - Filter Team Files by specific team for easy navigation when in multiple teams
- **Per-user storage quotas** - Individually configurable storage limits (MB to TB)
- **User dashboard** - Real-time quota usage, file management, and download statistics
- **Active/inactive status** - Temporarily disable users without deletion
- **Bulk user operations** - Efficient management of multiple users
- **Download account portal** - Recipients can view their download history and manage their accounts

### 📊 Download Tracking & Accountability
- **Modern Glassmorphic Admin Dashboard (v4.7.9+):**
  - **2025 modern design** - Animated gradient backgrounds with glassmorphism effects
  - **Real-time statistics** - Total users, active users, downloads, storage trends
  - **Comprehensive metrics** - Download/upload data (today, week, month, year)
  - **User growth tracking** - Monthly user additions/removals with growth percentages
  - **Security overview** - 2FA adoption rates, backup code status
  - **File statistics** - Largest files, most active users, top file types
  - **Trend analysis** - Storage trends, most active days, download patterns
  - **Twemoji integration** - Colorful emojis across all platforms (Linux, Windows, macOS)
  - **Responsive design** - Mobile-first with smooth animations and transitions
- **Complete audit trail:**
  - Track exactly **who** downloaded files (email addresses for authenticated downloads)
  - Record **when** downloads occurred (precise timestamps)
  - Log **from where** downloads originated (IP addresses with configurable privacy controls)
- **Per-file download history** - View detailed download logs for each file
- **Exportable reports** - Download tracking data in CSV format for compliance
- **Download count limits** - Automatically expire files after reaching download threshold
- **Email notifications** - Optional notifications when files are downloaded (configurable)

### 📋 Enterprise Audit Logging
- **Comprehensive audit trail:**
  - All user actions logged (logins, file operations, settings changes)
  - Detailed tracking of authentication events (2FA, password changes, failed logins)
  - Complete file lifecycle logging (upload, download, delete, restore, permanent deletion)
  - Team management operations (member changes, role updates, file sharing)
  - System configuration changes (settings, branding, quotas, audit policy)
  - Email sends logged with recipient and file details (v4.7+)
  - File request uploads logged with uploader IP and file info (v4.7+)
- **Compliance-ready:**
  - Meets GDPR, SOC 2, HIPAA, and ISO 27001 audit requirements
  - Configurable retention periods (1 day to 10 years)
  - Automatic cleanup based on time and size limits
  - Immutable write-only log entries
- **Advanced filtering and search:**
  - Filter by user, action type, entity type, date range
  - Full-text search across all log fields
  - Pagination for large datasets (50 logs per page)
  - Real-time statistics (total logs, recent activity, failed actions)
- **Export capabilities:**
  - CSV export for external audit tools
  - Timestamped export files
  - Complete log data including IP addresses, user agents, detailed context
  - Filterable exports for targeted reporting
- **Configurable retention:**
  - Set retention period: 90 days (default) up to 10 years
  - Size-based cleanup: 100 MB (default) up to 10 GB
  - Automated daily cleanup scheduler
  - Settings accessible via Server Settings page
- **Admin access:**
  - Accessible via Server → Audit Logs
  - Direct URL: `/admin/audit-logs`
  - Admin-only access with secure authentication

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
  - Upload custom logo (replaces default WulfVault branding)
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
- **5 Email Providers Supported:**
  - **Resend (recommended)** - Built on AWS SES with best-in-class deliverability
  - **SendGrid** - Industry-leading email API with simple setup
  - **Mailgun** - Powerful API with domain/region configuration
  - **Brevo** - API-based transactional email (formerly SendInBlue)
  - **SMTP** - Classic SMTP with/without TLS for self-hosted servers
- **Security & Management:**
  - Encrypted credential storage (AES-256-GCM)
  - Test email functionality before activation
  - Switch between providers with one click
  - Complete DNS verification guides (Loopia, generic DNS)
- **Email templates:**
  - Password reset emails with secure tokens
  - File sharing notifications with download links
  - Custom branding in all email communications
  - Professional HTML email templates
- **Redesigned email templates (v4.7+):**
  - Professional table-based layout for all email clients (including Outlook dark mode)
  - Large, prominent action buttons with clear CTAs
  - "What is this?" explanations for non-technical recipients
  - Clear expiration warnings with date/time
  - Upload request, download/share, and download notification emails
- **Email tracking:**
  - Log all sent emails with timestamps
  - Track email delivery status
  - Audit trail for compliance

### 📁 File Request System
- **Inbound file collection:**
  - Create upload request links for receiving files from others
  - Customizable upload limits (file size and count)
  - 24-hour link expiration for security (with clear countdown timers)
  - Password protection for upload portals
- **Smart expiry management (v4.2.2+):**
  - **Live countdown timers** - "Expires in 23 hours", "Expires in 5 hours" with color-coded urgency
  - **Grace period display** - After expiry: "EXPIRED - Auto-removal in 5 days" countdown
  - **Automatic cleanup** - Expired requests removed after 5 days to keep dashboard clean
  - **Visual feedback** - Green (active), orange (urgent), red (expired) status indicators
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
  - View all files across all users with complete metadata
  - Advanced search and filter capabilities
  - Delete files with trash safety net (configurable retention period)
  - One-click restore for accidentally deleted files with full metadata preservation
  - Permanent deletion from trash with confirmation dialogs
  - Detailed trash view: who deleted, when, days remaining, original owner
  - Modern, responsive UI with gradient buttons and emoji indicators
  - **Improved All Files view (v4.7.3+)** - Card-based layout with clear file separation, grouped file+note display, and better visual hierarchy
- **User administration:**
  - Create, edit, and delete users with full audit trail
  - Manage download accounts with comprehensive controls
  - Adjust quotas on the fly
  - Toggle user active/inactive status
  - **Enterprise pagination & filtering (v4.5.13+):**
    - Search users by name or email instantly
    - Filter by user level (Regular Users / Admins)
    - Filter by status (Active / Inactive)
    - 50 users per page (configurable up to 200)
    - Previous/Next navigation with result counters
    - Independent pagination for users and download accounts
    - Mobile-responsive filter UI
    - Scales to thousands of users without performance degradation
- **System settings:**
  - Configure server URL and port
  - Set system-wide defaults
  - Manage trash retention
  - Control privacy and logging settings

---

## Quick Start

### First-Time Setup

1. **Download and start WulfVault** (see installation methods below)

2. **Initial admin credentials:**
   - **Email:** `admin@wulfvault.local`
   - **Password:** `WulfVaultAdmin2024!`

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
# Clone repository first
git clone https://github.com/Frimurare/WulfVault.git
cd WulfVault

# Build and run with Docker
docker build -t wulfvault/wulfvault:latest .
docker run -d \
  --name wulfvault \
  -p 8080:8080 \
  -v ./data:/data \
  -v ./uploads:/uploads \
  -e SERVER_URL=https://files.yourdomain.com \
  wulfvault/wulfvault:latest
```

### Docker Compose

```bash
# Clone repository
git clone https://github.com/Frimurare/WulfVault.git
cd WulfVault

# Start with Docker Compose (uses docker-compose.yml in repo)
docker compose up -d --build
```

Or create custom `docker-compose.yml`:
```yaml
version: '3.8'
services:
  wulfvault:
    build: .
    image: wulfvault/wulfvault:latest
    container_name: wulfvault
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

### Build from Source

**Prerequisites:** Go 1.21+

```bash
# Clone repository
git clone https://github.com/Frimurare/WulfVault.git
cd WulfVault

# Install dependencies
go mod download

# Build
go build -o wulfvault cmd/server/main.go

# Run
./wulfvault
```

**Default credentials on first run:**
- Email: `admin@wulfvault.local`
- Password: `WulfVaultAdmin2024!`

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
- Upload to WulfVault with expiration and authentication
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
   - Set email, password, user level (Super Admin, Admin, or Regular User)
   - Assign storage quota (e.g., 5GB, 50GB, or custom)
   - Set active/inactive status with toggle
   - Organize users into teams for shared file access
3. **Monitor usage** in dashboard
   - Real-time system statistics (files, downloads, storage, users)
   - User growth analytics (new users last 30 days)
   - Storage usage per user with quota visualization
4. **View all files** and download history across system
   - Complete download audit trail with email, timestamp, IP (if enabled)
   - Export download reports in CSV format for compliance
5. **Manage trash** with enhanced UI and restore accidentally deleted files
   - Modern gradient-styled buttons: ♻️ Restore (green) and 🗑️ Delete Forever (red)
   - Complete audit trail: who deleted, when, days remaining
   - One-click restore or permanent deletion with confirmation dialogs

---

## API

WulfVault provides a **complete REST API** for automation, integrations, and third-party applications.

**Available APIs:**
- **User Management** - Create, read, update, delete users; manage storage quotas
- **File Management** - Upload, download, delete files; manage metadata and passwords
- **Download Accounts** - Manage download-only user accounts
- **File Requests** - Create and manage upload request portals
- **Trash Management** - List, restore, and permanently delete files
- **Teams** - Manage teams, members, and file sharing
- **Email** - Configure email settings and send file links
- **Admin/System** - System statistics, branding, and settings

**Example API calls:**

```bash
# List all users (admin only)
curl -b cookies.txt http://localhost:4949/api/v1/users

# Upload a file
curl -b cookies.txt -F "file=@document.pdf" \
  http://localhost:4949/api/v1/upload

# Get system statistics
curl -b cookies.txt http://localhost:4949/api/v1/admin/stats
```

**Authentication:** API requests use session-based authentication via cookies.

**Full documentation:** See [API.md](docs/API.md) for complete endpoint reference, request/response examples, and code samples in Python, JavaScript, and cURL.

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
6. **Update regularly** - Keep WulfVault up to date
7. **Strong passwords** - Enforce password policies for all users

---

## GDPR Compliance

**Status:** ✅ **WulfVault is GDPR-Compliant** (Grade: A-, 94%)

WulfVault is designed with **privacy-by-design** and **privacy-by-default** principles, making it suitable for organizations handling personal data under GDPR and other data protection regulations.

### Built-in GDPR Features

#### User Rights Implementation
- ✅ **Right of Access (Art. 15)** - Users can export all their data via `/api/v1/user/export-data`
- ✅ **Right to Rectification (Art. 16)** - Users can update their profile and password via settings
- ✅ **Right to Erasure (Art. 17)** - Account deletion with GDPR-compliant soft deletion at `/settings/delete-account`
- ✅ **Right to Data Portability (Art. 20)** - JSON export of all personal data
- ✅ **Right to Be Informed (Art. 13/14)** - Privacy Policy templates provided

#### Technical Measures (Art. 32)
- ✅ **Audit Logging** - Comprehensive activity tracking with configurable retention (1-3650 days)
- ✅ **Encryption in Transit** - TLS/HTTPS for all connections (TLS 1.2+ required)
- ⚠️ **Encryption at Rest** - Not built-in; use OS-level disk encryption (LUKS, BitLocker, FileVault)
- ✅ **Password Security** - bcrypt hashing (cost factor 12, never plaintext)
- ✅ **2FA Support** - TOTP-based two-factor authentication
- ✅ **Session Security** - HttpOnly, Secure, SameSite cookies with 24-hour timeout
- ✅ **Data Minimization** - Only necessary data collected, no tracking or analytics
- ✅ **IP Logging** - Optional (disabled by default for privacy)

#### Organizational Measures
- ✅ **Data Retention Policies** - Configurable audit log retention with automatic cleanup
- ✅ **Soft Deletion** - Deleted accounts anonymized (`deleted-user-XXX@deleted.local`) with audit trail preserved
- ✅ **Breach Notification Procedure** - Complete incident response guide provided
- ✅ **Records of Processing Activities (ROPA)** - Template for Art. 30 compliance

### GDPR Compliance Documentation

Complete GDPR compliance package available in the `/gdpr-compliance/` directory:

| Document | Purpose | Required Action |
|----------|---------|-----------------|
| **README.md** | Complete GDPR compliance guide | Read and follow |
| **PRIVACY_POLICY_TEMPLATE.md** | Privacy Policy for users | Customize & publish |
| **COOKIE_POLICY_TEMPLATE.md** | Cookie usage transparency | Customize & publish |
| **DATA_PROCESSING_AGREEMENT_TEMPLATE.md** | B2B processor contracts (Art. 28) | Customize if B2B |
| **BREACH_NOTIFICATION_PROCEDURE.md** | Incident response plan (Art. 33/34) | Review & follow |
| **DEPLOYMENT_CHECKLIST.md** | Pre-launch compliance verification | Complete all items |
| **RECORDS_OF_PROCESSING_ACTIVITIES.md** | Art. 30 documentation | Maintain & update |
| **COOKIE_CONSENT_BANNER.html** | Cookie consent template (optional) | Use only if adding analytics |

### Quick Compliance Setup

1. **Customize Templates** (1-2 hours)
   ```bash
   cd gdpr-compliance/
   # Edit all *_TEMPLATE.md files, replace [PLACEHOLDERS]
   ```

2. **Publish Required Policies**
   - Privacy Policy → Must be accessible at `/privacy-policy`
   - Cookie Policy → Must be accessible at `/cookie-policy`

3. **Complete Deployment Checklist**
   ```bash
   # Review and complete all sections
   cat gdpr-compliance/DEPLOYMENT_CHECKLIST.md
   ```

4. **Enable GDPR Features**
   - Users can export data: `/settings/account` → "Export My Data"
   - Users can delete accounts: `/settings/delete-account`
   - Admins can export audit logs: Admin → Audit Logs → Export CSV

### Compliance Scorecard

| Feature | Status | Grade |
|---------|--------|-------|
| Data Collection & Minimization | ✅ | A+ |
| Audit Logging | ✅ | A |
| User Right: Delete | ✅ | A+ |
| User Right: Access | ✅ | A |
| User Right: Portability | ✅ | A |
| User Right: Rectification | ✅ | A |
| Authentication & Security | ✅ | A+ |
| Role-Based Access Control | ✅ | A+ |
| Encryption (Transit) | ✅ | A |
| Encryption (At Rest) | ⚠️ Use OS-level | B |
| Data Retention Policies | ✅ | A |
| Privacy Documentation | ⚠️ Templates Provided | A |
| Cookie Consent | ✅ N/A (Essential cookies only) | A |
| Breach Notification | ⚠️ Procedure Provided | A |
| **OVERALL COMPLIANCE** | **✅** | **A- (94%)** |

### Cookie Compliance Note

**WulfVault does NOT require cookie consent banner** because it only uses:
- ✅ **Session cookies** - Strictly necessary for authentication (exempt from consent)
- ❌ **No analytics cookies** - No Google Analytics, Facebook Pixel, or tracking
- ❌ **No marketing cookies** - No third-party advertising or retargeting

**When you WOULD need the cookie banner:**
- If you add Google Analytics or similar analytics tools
- If you add marketing/advertising tracking
- If you add third-party tracking scripts

The `COOKIE_CONSENT_BANNER.html` template is provided for organizations that add analytics later.

### Regulatory Standards Supported

- ✅ **GDPR** (EU General Data Protection Regulation)
- ✅ **UK GDPR** (United Kingdom)
- ✅ **ePrivacy Directive** (Cookie Law) - Compliant without banner
- ✅ **SOC 2** (Audit logging and access controls)
- ⚠️ **HIPAA** (Healthcare - requires OS-level disk encryption)
- ✅ **ISO 27001** (Information security management)

### Data Processing Summary

**Personal Data Collected:**
- User accounts: Name, email, password (hashed), role, creation date
- Authentication: 2FA secrets (encrypted), backup codes (hashed)
- Activity data: Login timestamps, file actions, IP addresses (optional)
- Files: Metadata (filename, size, MIME type) and contents

**Legal Basis (GDPR Art. 6):**
- Contractual necessity (6(1)(b)) - Service provision
- Legitimate interest (6(1)(f)) - Security, fraud prevention
- Legal obligation (6(1)(c)) - Audit compliance

**Data Retention:**
- User accounts: Until deletion (soft delete with 30-day grace period)
- Audit logs: Configurable (90 days default, 1-3650 days available)
- Deleted files: 5 days in trash (configurable)
- Backups: [Configure based on your policy]

**Data Transfers:**
- Default: All data stays on your server (location: [SPECIFY])
- If using cloud hosting: Document location and safeguards (SCCs if outside EU)

### For Different Organization Types

#### Small Organizations (<250 employees)
- ✅ Basic compliance: Privacy Policy, data export, account deletion
- ⚠️ ROPA: Only required if processing is not occasional or involves special data
- ⚠️ DPO: Not required unless large-scale monitoring or special category data

#### Medium/Large Organizations (250+ employees)
- ✅ Full GDPR compliance required
- ✅ ROPA (Records of Processing Activities) mandatory
- ⚠️ DPO: Required if public authority or large-scale systematic monitoring

#### B2B SaaS Providers
- ✅ Data Processing Agreement (DPA) required for each customer
- ✅ Sub-processor notification procedures
- ✅ Breach notification within 24 hours to customers

#### Healthcare / Finance / Government
- ✅ Enable OS-level disk encryption (LUKS, BitLocker, FileVault)
- ✅ Conduct Data Protection Impact Assessment (DPIA)
- ✅ Enhanced audit logging and retention
- ✅ Annual security assessments

### Implementation Time

| Task | Effort | Priority |
|------|--------|----------|
| Customize Privacy Policy | 1-2 hours | CRITICAL |
| Complete Deployment Checklist | 2-4 hours | CRITICAL |
| Review Breach Notification Procedure | 1 hour | CRITICAL |
| Customize DPA (if B2B) | 2-3 hours | HIGH |
| Test user data export | 30 min | HIGH |
| Test account deletion | 30 min | HIGH |
| Add cookie consent banner (optional, only if using analytics) | 1 hour | LOW |
| Configure audit retention | 15 min | MEDIUM |
| **Total Estimated Time** | **10-15 hours** | |

### Support Resources

- **Full Documentation:** `/gdpr-compliance/README.md`
- **Compliance Report:** See root directory for detailed analysis
- **EU GDPR Portal:** https://gdpr.eu/
- **ICO Guidance (UK):** https://ico.org.uk/
- **EDPB Guidelines:** https://edpb.europa.eu/

### Important Notes

⚠️ **Action Required:**
- WulfVault provides GDPR-compliant technical features
- **You must customize templates** with your organization details before use
- **Consult legal counsel** for compliance verification in your jurisdiction
- **Keep documentation current** as data processing changes

✅ **Best Practices:**
- Enable HTTPS/TLS in production (required)
- Configure audit log retention per your jurisdiction
- Appoint Data Protection Officer (DPO) if required
- Conduct annual GDPR compliance reviews
- Train staff on data protection procedures

---

## Server Restart Feature

The Admin Settings page includes a **"Restart Server"** button that is **currently disabled** until systemd service is installed.

### Why is it disabled?

The restart button requires a process manager (systemd, supervisor, etc.) to automatically restart the server after shutdown. Without this, clicking the button will stop the server without restarting it.

### How to enable it

1. **Install systemd service** (requires sudo):
   ```bash
   sudo cp /tmp/wulfvault.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable wulfvault
   sudo systemctl start wulfvault
   ```

2. **Uncomment the restart button** in the code:
   - Open `internal/server/handlers_admin.go`
   - Find the section marked `<!-- RESTART SERVER BUTTON - DISABLED`
   - Remove the `<!--` and `-->` comment markers
   - Also uncomment the JavaScript function at the bottom
   - Rebuild: `go build -o wulfvault cmd/server/main.go`
   - Restart the service: `sudo systemctl restart wulfvault`

3. **The button will now work!** It will use `systemctl restart wulfvault` to gracefully restart the server.

See [DEPLOYMENT.md](DEPLOYMENT.md) for complete deployment and autostart instructions.

---

## Troubleshooting

### Can't login with default credentials

Make sure you're using:
- Email: `admin@wulfvault.local`
- Password: `WulfVaultAdmin2024!`

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
- Review logs: `docker compose logs -f wulfvault`
- Open issue on GitHub: https://github.com/Frimurare/WulfVault/issues

---

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
WulfVault/
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

**Why AGPL-3.0?** This license ensures that if anyone uses WulfVault to provide a service over a network (like SaaS), they must share their modifications with the community. This prevents companies from taking the code, making improvements, and keeping them proprietary. It protects the open-source nature of the project while requiring attribution and source disclosure for all network use.

**Architecturally inspired by Gokapi** by Forceu - https://github.com/Forceu/Gokapi

See [NOTICE.md](NOTICE.md) for full attribution and license information.

---

## Support

- **Issues:** https://github.com/Frimurare/WulfVault/issues
- **Discussions:** https://github.com/Frimurare/WulfVault/discussions
- **Documentation:** https://github.com/Frimurare/WulfVault/wiki

---

**Made for organizations that need secure, accountable file sharing for large files**
