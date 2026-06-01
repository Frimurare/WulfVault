# WulfVault User Guide v7.1.0 Aurora 🌅

**Complete Guide for Administrators and Users**

---

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Configuration](#configuration)
4. [User Roles & Permissions](#user-roles--permissions)
5. [File Sharing Guide](#file-sharing-guide)
6. [Download Account Guide](#download-account-guide)
7. [Admin Dashboard](#admin-dashboard)
8. [User Management](#user-management)
9. [File Management](#file-management)
10. [Branding & Customization](#branding--customization)
11. [Email Configuration](#email-configuration)
12. [Logging & Monitoring](#logging--monitoring)
13. [Single Sign-On (SSO)](#single-sign-on-sso)
14. [Security Features](#security-features)
15. [File Request Portals](#file-request-portals)
16. [Troubleshooting](#troubleshooting)
17. [Best Practices](#best-practices)

---

## Introduction

### What is WulfVault?

WulfVault is a professional-grade, self-hosted file sharing platform designed for organizations that demand security, accountability, and complete control over their data. Unlike commercial file transfer services, WulfVault gives you:

- **Complete data ownership** - Files stay on your infrastructure
- **No subscription fees** - One-time setup, unlimited transfers
- **Enterprise features** - Multi-user management, audit trails, 2FA
- **Full customization** - Branded pages, custom domains, logo

### Key Features Overview

- ✅ **Large file support** - No practical per-file size limit, fully configurable (chunked uploads)
- ✅ **Multi-user system** - Admins, users, and download accounts
- ✅ **Download tracking** - Know exactly who downloaded what and when
- ✅ **Audit logging** - Comprehensive audit trail for compliance (GDPR, SOC 2)
- ✅ **Two-Factor Authentication** - TOTP-based security
- ✅ **Single Sign-On (SSO)** - Microsoft Entra ID (Azure AD) and Generic OpenID Connect (Google, Okta, Keycloak, Authelia, Auth0, etc.)
- ✅ **Email integration** - SMTP or Brevo for sending links
- ✅ **Branding** - Custom logo, colors, company name
- ✅ **File requests** - Create upload portals for receiving files
- ✅ **Storage quotas** - Per-user limits with usage tracking
- ✅ **File Sharing Wisdom** - 130+ humorous one-liners on dashboards
- ✅ **Team Files Access** - Admins and members can easily view team files
- ✅ **Professional UI** - Clean, modern design with consistent navigation

### Who Should Use This Guide?

- **System Administrators** - Setting up and managing WulfVault
- **Regular Users** - Uploading and sharing files
- **Download Users** - Receiving and accessing files
- **IT Managers** - Understanding security and compliance features

---

## Getting Started

### First-Time Login

1. **Navigate to your WulfVault instance:**
   ```
   http://your-server:8080
   or
   https://files.yourdomain.com
   ```

2. **First Admin Account** (created automatically on first start):
   - On the very first launch (when no users exist) WulfVault creates a single
     super-admin account.
   - The email defaults to `admin@localhost`, or whatever you set via the
     `ADMIN_EMAIL` environment variable.
   - The password is taken from the `ADMIN_PASSWORD` environment variable. If you
     do **not** set `ADMIN_PASSWORD`, a random password is generated and printed
     **once** to the server log/console at startup — copy it immediately.
   - Best practice: set both `ADMIN_EMAIL` and `ADMIN_PASSWORD` before first start
     so you control the credentials.

3. **⚠️ CRITICAL SECURITY STEP:**
   - Immediately go to Settings → Change Password
   - Choose a strong, unique password
   - Never share admin credentials

### Signing In

WulfVault supports two sign-in mechanisms. Local accounts are always available
as the primary mechanism and as a break-glass fallback; Single Sign-On is
optional and configured by an administrator.

#### Local Account (Email + Password)

1. Go to the login page (`/login`)
2. Enter your email and password
3. If Two-Factor Authentication is enabled, enter your 6-digit code (or a backup code)

#### Single Sign-On (Microsoft / OIDC)

If your administrator has enabled SSO, the login page shows an extra
**"Sign in with …"** button (for example "Sign in with Microsoft").

1. Click the **Sign in with …** button on the login page
2. You are redirected to your identity provider (Microsoft Entra ID, Google,
   Okta, Keycloak, etc.) to authenticate
3. After successful authentication you are redirected back to WulfVault and
   signed in automatically — admins land on `/admin`, regular users on `/dashboard`

**Notes for users:**
- SSO uses your existing corporate/organization identity — no separate WulfVault
  password is needed for SSO accounts.
- The first time you sign in via SSO, an account may be created automatically
  (if your admin enabled auto-provisioning and your email domain is allowed).
- If the SSO button is missing, your administrator has either not configured SSO
  or has enforced **local accounts only** (see below).

> **Download recipients do not use SSO.** People who only receive files use a
> separate, password-based **Download Account** that works without any SSO or
> external identity provider. See the [Download Account Guide](#download-account-guide).

### User Interface Overview

#### Admin Panel (`/admin`)
- Dashboard with system statistics
- User management
- File overview across all users
- System settings and branding
- Email configuration

#### User Dashboard (`/dashboard`)
- Upload new files
- View your uploaded files
- Download tracking per file
- Storage quota usage
- Account settings

#### Download Account Portal (`/download/dashboard`)
- View download history
- Change password
- Account settings
- GDPR self-deletion option

---

## Configuration

WulfVault can be configured through environment variables, command-line flags, and the web interface.

### Environment Variables

All configuration can be set via environment variables. These are the primary way to configure WulfVault in Docker deployments.

#### Complete Variable Reference

| Variable | Description | Default Value | Requires Restart |
|----------|-------------|---------------|------------------|
| `SERVER_URL` | Public URL where WulfVault is accessible | `http://localhost:8080` | ❌ No* |
| `PORT` | Port the server listens on | `8080` | ✅ Yes |
| `DATA_DIR` | Directory for database storage | `./data` | ✅ Yes |
| `UPLOADS_DIR` | Directory for uploaded files | `./uploads` | ✅ Yes |
| `MAX_FILE_SIZE_MB` | Maximum file size in megabytes | `5000` (5 GB) | ❌ No* |
| `DEFAULT_QUOTA_MB` | Default storage quota for new users | `5000` (5 GB) | ❌ No* |
| `SESSION_TIMEOUT_HOURS` | Session expiration time in hours | `24` | ✅ Yes |
| `TRASH_RETENTION_DAYS` | Days to keep deleted files in trash | `5` | ❌ No* |

**Note:** Variables marked with * can be changed via Admin Settings in the web interface and take effect immediately. Environment variables override web settings on startup.

### How to Set Environment Variables

#### Method 1: Docker Run Command

When starting WulfVault with `docker run`, use `-e` flags:

```bash
docker run -d \
  --name wulfvault \
  -p 8080:8080 \
  -v ./data:/data \
  -v ./uploads:/uploads \
  -e SERVER_URL=https://files.yourdomain.com \
  -e PORT=8080 \
  -e MAX_FILE_SIZE_MB=5000 \
  -e DEFAULT_QUOTA_MB=10000 \
  -e SESSION_TIMEOUT_HOURS=48 \
  -e TRASH_RETENTION_DAYS=7 \
  wulfvault/wulfvault:latest
```

**Example with all variables:**
```bash
docker run -d \
  --name wulfvault \
  -p 3000:3000 \
  -v /mnt/wulfvault-data:/data \
  -v /mnt/wulfvault-uploads:/uploads \
  -e SERVER_URL=https://files.company.com \
  -e PORT=3000 \
  -e DATA_DIR=/data \
  -e UPLOADS_DIR=/uploads \
  -e MAX_FILE_SIZE_MB=5000 \
  -e DEFAULT_QUOTA_MB=20000 \
  -e SESSION_TIMEOUT_HOURS=24 \
  -e TRASH_RETENTION_DAYS=30 \
  wulfvault/wulfvault:latest
```

#### Method 2: Docker Compose

Create or edit `docker-compose.yml`:

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
      # Required: Set your public URL
      SERVER_URL: https://files.yourdomain.com

      # Optional: Customize these as needed
      PORT: 8080
      DATA_DIR: /data
      UPLOADS_DIR: /data
      MAX_FILE_SIZE_MB: 5000          # 5 GB default (no practical upper limit)
      DEFAULT_QUOTA_MB: 5000          # 5 GB default per user
      SESSION_TIMEOUT_HOURS: 24       # 24 hours default
      TRASH_RETENTION_DAYS: 5         # 5 days default
    restart: unless-stopped
```

**Start with:**
```bash
docker-compose up -d
```

**Restart after changes:**
```bash
docker-compose down
docker-compose up -d
```

#### Method 3: Binary Executable (Command Line)

When running the compiled binary directly, use flags:

```bash
./wulfvault \
  -url=https://files.yourdomain.com \
  -port=8080 \
  -data=./data \
  -uploads=./uploads
```

**Available flags:**
- `-url` → SERVER_URL
- `-port` → PORT
- `-data` → DATA_DIR
- `-uploads` → UPLOADS_DIR

**Note:** File size limits, quotas, session timeout, and trash retention can only be set via environment variables or web interface, not command-line flags.

#### Method 4: Environment File (.env)

Create a `.env` file in your project directory:

```bash
# .env file for WulfVault configuration

# Server Configuration
SERVER_URL=https://files.yourdomain.com
PORT=8080
DATA_DIR=/data
UPLOADS_DIR=/uploads

# File and Storage Limits
MAX_FILE_SIZE_MB=2000
DEFAULT_QUOTA_MB=5000

# Security and Retention
SESSION_TIMEOUT_HOURS=24
TRASH_RETENTION_DAYS=5
```

**Use with Docker Compose:**
```yaml
version: '3.8'
services:
  wulfvault:
    build: .
    image: wulfvault/wulfvault:latest
    env_file:
      - .env
    ports:
      - "${PORT}:${PORT}"
    volumes:
      - ./data:/data
      - ./uploads:/uploads
    restart: unless-stopped
```

### Configuration Priority

Settings are applied in this order (later overrides earlier):

1. **Default values** (hardcoded in application)
2. **Environment variables** (set at container/process start)
3. **Command-line flags** (when using binary)
4. **Web interface settings** (stored in database)

**Example:**
- Default: `MAX_FILE_SIZE_MB=2000`
- Environment variable: `MAX_FILE_SIZE_MB=5000` → Overrides default
- Admin Settings page: Set to 3000 MB → Overrides environment variable

### Runtime vs Restart-Required Changes

#### Can be changed without restart (via Admin Settings):
- ✅ `SERVER_URL` - Change via Admin → Settings
- ✅ `MAX_FILE_SIZE_MB` - Change via Admin → Settings
- ✅ `DEFAULT_QUOTA_MB` - Change via Admin → Settings
- ✅ `TRASH_RETENTION_DAYS` - Change via Admin → Settings
- ✅ `DASHBOARD_STYLE` - Change via Admin → Settings (colorful or plain white)
- ✅ Branding (logo, colors, company name)
- ✅ Per-user storage quotas

#### Requires container/service restart:
- ⚠️ `PORT` - Change requires restart
- ⚠️ `DATA_DIR` - Change requires restart
- ⚠️ `UPLOADS_DIR` - Change requires restart
- ⚠️ `SESSION_TIMEOUT_HOURS` - Change requires restart

### Common Configuration Scenarios

#### Scenario 1: Increase File Size Limit

**Quick (no restart):**
1. Login as admin
2. Go to Admin → Settings
3. Change "Max File Size (MB)" to desired value
4. Click "Save Settings"
5. ✅ Takes effect immediately

**Permanent (with environment variable):**
```bash
docker-compose down
# Edit docker-compose.yml - add or change:
#   MAX_FILE_SIZE_MB: 5000
docker-compose up -d
```

#### Scenario 2: Change Port

**Requires restart:**
```bash
docker-compose down
# Edit docker-compose.yml:
#   ports:
#     - "3000:3000"
#   environment:
#     PORT: 3000
docker-compose up -d
```

#### Scenario 3: Increase User Quotas

**For existing users:**
1. Admin → Users
2. Click edit on user
3. Change "Storage Quota" value
4. Save

**For new users (default):**
1. Admin → Settings
2. Change "Default User Quota (MB)"
3. Save
4. ✅ Applies to users created after this change

#### Scenario 4: Custom Domain Setup

**Step 1: Set environment variable**
```yaml
environment:
  SERVER_URL: https://files.company.com
```

**Step 2: Restart container**
```bash
docker-compose down && docker-compose up -d
```

**Step 3: Verify in Admin → Settings**
- Should show your domain
- Download links will use this URL

#### Scenario 5: Extended Trash Retention

**Option A: Environment variable (permanent)**
```yaml
environment:
  TRASH_RETENTION_DAYS: 30
```

**Option B: Admin Settings (runtime)**
1. Admin → Settings
2. "Trash Retention Period (Days)" → 30
3. Save
4. ✅ Effective immediately

### Verifying Configuration

#### Check Current Settings

**Via Admin Dashboard:**
1. Login as admin
2. Go to Admin → Settings
3. See all current values

**Via Docker logs:**
```bash
docker logs wulfvault

# Output shows:
# WulfVault File Sharing System v7.1.0 Aurora
# Server starting on :8080
# Server URL: https://files.yourdomain.com
```

**Via environment inspection:**
```bash
docker exec wulfvault env | grep -E "SERVER_URL|PORT|MAX_FILE"
```

### Troubleshooting Configuration

#### Problem: Changes not taking effect

**Solution:**
1. Check if change requires restart (see table above)
2. For Docker: `docker-compose down && docker-compose up -d`
3. Check logs: `docker logs wulfvault`
4. Verify no syntax errors in docker-compose.yml

#### Problem: Download links use wrong URL

**Solution:**
1. Set `SERVER_URL` environment variable correctly
2. Do NOT include port if using standard 80/443
3. Do NOT include trailing slash
4. Example: `https://files.company.com` not `https://files.company.com:8080/`

#### Problem: File uploads fail with size error

**Solution:**
1. Check `MAX_FILE_SIZE_MB` setting
2. Ensure reverse proxy (nginx/Caddy) allows large uploads
3. For nginx, set: `client_max_body_size 5000M;`

---

## User Roles & Permissions

### Role Types

#### 1. Super Admin
**Full System Control**
- Manage all users and admins
- Access all files across the system
- Configure branding and settings
- Set up email providers
- View system-wide statistics
- Manage trash and restore files

**Use Case:** IT director, system administrator

#### 2. Admin
**User Management & Oversight**
- Create and manage regular users
- View all files in the system
- Access download tracking
- Cannot modify system settings or branding

**Use Case:** Department manager, team lead

#### 3. Regular User
**File Upload & Sharing**
- Upload files within quota
- Share files with expiration settings
- Track downloads on their files
- Create file request portals
- Manage their own files only

**Use Case:** Employees, team members

#### 4. Download Account
**File Access Only**
- Created automatically for authenticated downloads
- Access files shared with them
- View their download history
- Self-service password change
- GDPR account deletion

**Use Case:** External recipients, clients, partners

### Permission Matrix

| Action | Super Admin | Admin | User | Download Account |
|--------|-------------|-------|------|------------------|
| Upload files | ✅ | ✅ | ✅ | ❌ |
| Share files | ✅ | ✅ | ✅ | ❌ |
| View all files | ✅ | ✅ | ❌ | ❌ |
| Create users | ✅ | ✅ | ❌ | ❌ |
| Modify settings | ✅ | ❌ | ❌ | ❌ |
| Configure branding | ✅ | ❌ | ❌ | ❌ |
| Download shared files | ✅ | ✅ | ✅ | ✅ |
| View own download history | ✅ | ✅ | ✅ | ✅ |

---

## File Sharing Guide

### How to Upload and Share Files

#### Step 1: Upload File

1. **Login** to your dashboard
2. **Drag and drop** a file onto the upload zone, or click to browse
3. **Wait** for upload to complete (progress bar shown)

#### Step 2: Configure Share Settings

**Expiration Options:**
- **Download Limit:** Set how many times file can be downloaded (1-999)
- **Date Expiration:** Set when file expires (days, weeks, months)
- **Both:** File expires when either limit is reached

**Security Options:**
- **Password Protection:** Require password to access file
- **Require Authentication:** Recipient must create download account

**Sharing Options:**
- **Email Integration:** Send download link via email directly
- **Copy Link:** Get shareable URL to send via any channel

#### Step 3: Share the Link

**Option A: Via Email (Integrated)**
1. Click "Email" button
2. Enter recipient email address
3. Add optional message
4. Send - Recipient receives branded email with link

**Option B: Copy & Paste**
1. Click "Copy Link" button
2. Share link via email, chat, SMS, etc.
3. Optionally share password separately (if used)

### Searching and Sorting Files

**NEW in v4.9.7:** Powerful search and sorting capabilities to manage your files efficiently.

#### Search Files

**Location:** Above your file list on the dashboard

**How to Search:**
1. Type in the search bar (🔍 Search files...)
2. Results filter instantly as you type
3. Search works for:
   - **Filename:** Any part of the file name
   - **Extension:** File type (e.g., "pdf", "docx", "jpg")

**Search Examples:**
- Type "report" → Finds "Q4-report.pdf", "Monthly_Report.xlsx"
- Type "pdf" → Shows all PDF files
- Type "2024" → Finds all files with "2024" in the name

**Search Tips:**
- Search is case-insensitive
- Works with all file tabs (All Files, My Files, Team Files)
- Combines with team filter when viewing team files

#### Sort Files

**Location:** Dropdown menu next to search bar

**Sort Options:**
- **📝 Name (A-Z)** - Alphabetical order
- **📝 Name (Z-A)** - Reverse alphabetical
- **📅 Newest First** - Most recently uploaded (default)
- **📅 Oldest First** - Oldest files first
- **📊 Most Downloads** - Most popular files
- **📊 Least Downloads** - Least downloaded
- **📦 Largest First** - Biggest files
- **📦 Smallest First** - Smallest files

**Use Cases:**
- **Find recent uploads:** Use "Newest First" (default)
- **Manage storage:** Use "Largest First" to find big files
- **Find popular content:** Use "Most Downloads"
- **Clean up old files:** Use "Oldest First"

**Combined Search and Sort:**
1. Enter search term (e.g., "invoice")
2. Select sort order (e.g., "Newest First")
3. Results are filtered AND sorted
4. Works seamlessly with file type and team filters

### Understanding Share Modes

#### Direct Download (No Authentication)
```
Recipient clicks link → Downloads immediately
```

**Pros:**
- Fastest for recipient
- No account needed
- Best for trusted recipients

**Cons:**
- Less accountability
- Can't track recipient identity
- Anyone with link can download

**Best For:** Internal sharing, trusted partners

#### Authenticated Download (Recommended)
```
Recipient clicks link → Creates account → Downloads file
```

**Pros:**
- Full accountability - know exactly who downloaded
- Download account for recipient to view history
- More secure
- Recipient can re-download from their portal

**Cons:**
- Recipient must create account first
- Slightly more steps

**Best For:** External sharing, compliance requirements, sensitive files

### Tracking Downloads

**View Download History:**
1. Go to your dashboard
2. Click "History" button on any file
3. See complete download log:
   - Recipient email (for authenticated downloads)
   - Download timestamp
   - IP address
   - User agent (browser/device)

**Export Download Report:**
1. Click "Export CSV" in download history
2. Save report for compliance/audit purposes

---

## Download Account Guide

> **Download Accounts work without SSO or any external identity provider.**
> They are simple, self-service, email + password accounts created for
> recipients of authenticated downloads. A recipient never needs a Microsoft
> account, corporate SSO, or an eID to receive files — they just set a password
> the first time they access a file that requires authentication. This is
> separate from the staff sign-in options (local password or SSO) described in
> [Signing In](#signing-in) and [Single Sign-On (SSO)](#single-sign-on-sso).

### For Recipients: First-Time Download

1. **Receive download link** via email or other channel

2. **Click the link** - You'll see the file splash page

3. **Create download account** (if authenticated download):
   - Enter your email address
   - Create a password (minimum 8 characters)
   - Click "Create Account & Download"

4. **Download the file** - File downloads immediately after account creation

5. **Access your portal** at any time:
   ```
   https://your-wulfvault-instance/download/dashboard
   ```

### Download Account Features

#### Dashboard
- **Download History:** See all files you've downloaded
- **Account Information:** View your email and download count
- **Last Used:** When you last accessed the system

#### Account Settings
- **Change Password:** Update your password anytime
- **GDPR Compliance:**
  - View your data
  - Export your data
  - Delete your account permanently

### Managing Your Download Account

**Change Password:**
1. Login to download portal
2. Click "Change Password"
3. Enter current password
4. Enter new password (minimum 8 characters)
5. Confirm new password
6. Save

**Delete Account (GDPR):**
1. Go to "Account Settings"
2. Scroll to "Delete Account" section
3. Read warnings carefully
4. Type `DELETE` to confirm
5. Account and all data permanently deleted

⚠️ **Warning:** Account deletion is permanent and cannot be undone!

---

## Admin Dashboard

### Overview

The Admin Dashboard provides system-wide visibility and control.

**Access:** `https://your-instance/admin`

### Dashboard Statistics

**Real-time Metrics:**
- 📁 Total files in system
- 👥 Total users (Admin + Regular + Download)
- 📥 Total downloads across all files
- 💾 Total storage used
- 📊 User growth (new users last 30 days)

### Navigation Menu

**Main Sections:**
- **Dashboard** - Overview and statistics
- **Users** - User management
- **Download Accounts** - Recipient account management
- **Files** - All files across system
- **Trash** - Deleted files (recoverable)
- **Branding** - Customize appearance
- **Settings** - System configuration
- **Email Settings** - Configure email providers
- **Identity Providers** - Configure SSO (Microsoft Entra ID / Generic OIDC) or force local-only accounts

### Quick Actions

From dashboard, admins can:
- Create new users
- View recent uploads
- Check system health
- Access file management
- Review trash items

---

## User Management

### Creating Users

#### Create Admin User

1. Go to **Admin → Users**
2. Click **"Create New User"**
3. Fill in details:
   - **Name:** Full name or username
   - **Email:** Must be unique
   - **Password:** Strong password (minimum 8 characters)
   - **User Level:** Select "Admin"
   - **Storage Quota:** Set limit (e.g., 50GB)
   - **Active:** Check to enable immediately
4. Click **"Create User"**

#### Create Regular User

Same process as admin, but:
- **User Level:** Select "User"
- **Storage Quota:** Default is 5GB (5000 MB), adjust as needed

### Editing Users

1. Go to **Admin → Users**
2. Click **pencil icon** next to user
3. Modify:
   - Name, email, user level
   - Storage quota
   - Active/inactive status
4. Click **"Update User"**

**Note:** Cannot change password via edit. Users must change their own password.

### Managing Storage Quotas

**Quota Levels:**
- **Default:** 5000 MB (5 GB) - Applied to new users automatically
- **Custom:** Any size in MB (e.g., 10000 MB = 10GB, 50000 MB = 50GB)
- **Recommended based on use case:**
  - Light users: 5 GB (default)
  - Regular users: 10-20 GB
  - Power users: 50-100 GB
  - Admins: 100+ GB

**Monitoring Usage:**
- Users see quota bar on dashboard
- Shows: Used / Total
- Color coding: Green → Yellow → Red

**What happens when quota is full:**
- User cannot upload new files
- Must delete files to free space
- Trash counts toward quota until permanently deleted

### Deactivating vs Deleting Users

**Deactivate (Recommended):**
- User cannot login
- Files preserved
- Can be reactivated later
- Download accounts can still access files

**Delete:**
- User permanently removed
- Files moved to trash
- After trash retention period, files deleted permanently
- Cannot be undone

### Managing Download Accounts

**View Download Accounts:**
1. Go to **Admin → Users** (Download Accounts tab)
2. See all recipient accounts

**Download Account Actions:**
- **Toggle Active/Inactive:** Prevent login without deleting
- **Create Manually:** Create account for recipient beforehand
- **Edit:** Change email or name
- **Delete:** Permanently remove account

### Pagination & Filtering

**Enterprise Scalability Feature** - For organizations with hundreds or thousands of users.

#### User List Pagination

**Access:** Admin → Users

When viewing users, WulfVault automatically paginates the list to improve performance:

**Pagination Controls:**
- **50 users per page** (default, configurable up to 200)
- **Previous/Next buttons** - Navigate between pages
- **Result counter** - Shows "Showing X-Y of Z users"
- **Independent navigation** - Users and Download Accounts paginate separately

#### Filtering Users

**Available Filters:**

1. **Search by Name or Email**
   - Type in search box to find users instantly
   - Searches both name and email fields
   - Case-insensitive matching

2. **Filter by User Level**
   - **All Users** - Show both regular users and admins
   - **Regular Users** - Show only user level 1
   - **Admins** - Show only admins and super admins (level 2+)

3. **Filter by Status**
   - **All** - Show both active and inactive users
   - **Active** - Show only active users
   - **Inactive** - Show only deactivated users

**How to Use Filters:**
1. Go to **Admin → Users**
2. Use the filter boxes above the user list:
   - Type name/email in **Search** box
   - Select user level from **User Level** dropdown
   - Select status from **Status** dropdown
3. Click **Filter** button to apply
4. Click **Clear** button to reset all filters

**Filter Combinations:**
- Combine multiple filters (e.g., "Active Admins")
- Search + filter (e.g., "John" + "Regular Users")
- Filters persist across pagination (navigate pages without losing filters)

#### Download Account Filtering

**Access:** Admin → Users (scroll to Download Accounts section)

**Available Filters:**
1. **Search by Name or Email** - Find download accounts quickly
2. **Filter by Status** - Active / Inactive / All

**Same Controls:**
- 50 accounts per page (default)
- Previous/Next navigation
- Result counter
- Independent from user pagination

#### Performance Benefits

**For Small Deployments (< 50 users):**
- All users fit on one page
- Filters available for quick searching
- No performance impact

**For Medium Deployments (50-500 users):**
- Pages load 10x faster
- Easy navigation with pagination
- Quick search finds users instantly

**For Large Deployments (500+ users):**
- Essential for usability
- Prevents browser slowdown
- Scales to thousands of users
- Professional enterprise experience

#### Tips for Large User Bases

1. **Use Search** - Fastest way to find specific users
2. **Combine Filters** - Narrow down results efficiently
3. **Adjust Page Size** - Add `?user_limit=100` to URL for larger pages (max 200)
4. **Bookmark Filters** - Save commonly used filter URLs

**Example URLs:**
```
/admin/users?search=john              # Search for "john"
/admin/users?level=2                  # Show only admins
/admin/users?level=1&active=true      # Active regular users
/admin/users?user_offset=50           # Page 2 (users 51-100)
```

---

## File Management

### Viewing All Files (Admins)

**Access:** Admin → Files

**File List Shows:**
- File name and size
- Uploader name
- Upload date
- Expiration status
- Download count
- Actions (view history, delete)

**Search & Filter:**
- Search by filename
- Filter by user
- Sort by date, size, downloads

### File Details

Click on any file to see:
- **Basic Info:** Name, size, type
- **Upload Info:** Who uploaded, when
- **Expiration:** Download limit, date limit
- **Security:** Password protected? Authentication required?
- **Download History:** Full log of all downloads

### Deleting Files

**User Delete:**
1. User goes to their dashboard
2. Clicks delete icon on file
3. File moved to trash (not deleted yet)

**Admin Delete:**
1. Admin → Files
2. Click delete icon
3. File moved to trash

### Trash Management

**Access:** Admin → Trash

**Trash Features:**
- **Retention Period:** Files kept for configured days (default: 5 days, configurable 1-365)
- **Automatic Cleanup:** Items older than retention period are permanently deleted
- **Manual Actions:**
  - **♻️ Restore:** Bring file back with a single click (restores to original uploader with full metadata)
  - **🗑️ Delete Forever:** Permanently delete immediately (cannot be undone, bypasses retention period)
- **Enhanced UI (v4.3+):**
  - Modern gradient-styled action buttons with emoji indicators
  - Visual color coding: Green for restore, Red for permanent delete
  - Hover effects and smooth transitions for better user experience
  - Fully responsive mobile layout with stacked buttons

**View Trash:**
- See deleted files with complete metadata
- Who deleted them (user or system)
- When deleted (precise timestamp)
- Days remaining before permanent deletion
- Original owner and file size information

---

## Branding & Customization

### Customizing Your Instance

**Access:** Admin → Branding

### Logo Upload

1. **Prepare logo:**
   - Format: PNG, JPG, SVG
   - Recommended size: 200x50 pixels
   - Transparent background works best

2. **Upload:**
   - Click "Choose File"
   - Select your logo
   - Click "Upload Logo"

3. **Logo appears:**
   - Login page
   - User dashboard header
   - Admin panel header
   - Download splash pages
   - Email templates

### Color Scheme

**Primary Color:**
- Main brand color
- Used for buttons, links, headers
- Example: `#2563eb` (blue)

**Secondary Color:**
- Accent color for gradients
- Used in headers, highlights
- Example: `#1e40af` (dark blue)

**Changing Colors:**
1. Go to Branding settings
2. Enter hex color code (e.g., #FF5733)
3. Preview changes
4. Click "Save Settings"

### Company Name

**Set Company Name:**
- Appears in page titles
- Shown in headers
- Used in email signatures
- Displayed on download pages

**To Update:**
1. Branding → Company Name
2. Enter your organization name
3. Save

**Example:** "Acme Corporation File Sharing"

### Dashboard Style

**Choose Dashboard Appearance:**

WulfVault offers two dashboard style options configurable via **Admin → Server Settings**:

**Colorful Dashboard (Default):**
- User dashboard: Light gray background (#f5f5f5)
- Admin dashboard: Animated purple/pink gradient with floating effects
- Visual flair and modern aesthetic
- Engaging user experience

**Plain White Dashboard:**
- Both dashboards: Clean white background (#ffffff)
- Minimal, distraction-free interface
- Professional and simple look
- No gradient animations or overlay effects

**To Change Dashboard Style:**
1. Go to **Admin → Server Settings**
2. Check/uncheck **"Use plain white dashboard (instead of colorful)"**
3. Click **"Save Settings"**
4. Changes apply immediately to all users

**File Sharing Wisdom Banner:**
- Automatically uses your branding colors (Primary → Secondary gradient)
- Configured colors in **Admin → Branding** apply to the wisdom banner
- Harmonizes with your overall branding theme

---

## Email Configuration

### Why Configure Email?

Enable email functionality to:
- Send download links directly from WulfVault
- Send password reset emails
- Provide professional branded communications

### Email Provider Options

#### Option 1: SMTP (Self-Hosted Email)

**Best For:** Organizations with existing email server

**Configuration:**
1. Go to **Admin → Email Settings**
2. Select **SMTP Provider**
3. Enter:
   - SMTP Server: `smtp.yourdomain.com`
   - Port: `587` (TLS) or `465` (SSL)
   - Username: Your email address
   - Password: Email password or app-specific password
   - From Address: `noreply@yourdomain.com`
   - From Name: Your organization name
4. Click **"Test Configuration"**
5. Check test email arrives
6. Click **"Save Configuration"**

**Example (Gmail):**
```
Server: smtp.gmail.com
Port: 587
Username: yourname@gmail.com
Password: [App-specific password]
From: yourname@gmail.com
```

#### Option 2: Brevo (SendInBlue)

**Best For:** Organizations without email server, high volume sending

**Setup:**
1. Create account at https://www.brevo.com
2. **Get API key from Brevo dashboard:**
   - Go to **Settings** → **SMTP & API**
   - Click on **"API Keys"** tab (NOT "SMTP")
   - Click **"Generate a new API key"**
   - ⚠️ **IMPORTANT:** Select **"API keys"** - Do NOT create an "SMTP API key"
   - Give it a name (e.g., "WulfVault")
   - Copy the key (starts with `xkeysib-...`)
3. **Verify sender address:**
   - Go to **Settings** → **Senders & IP**
   - Add and verify your sender email address
   - Wait for verification email and confirm
4. In WulfVault:
   - Go to **Admin → Email Settings**
   - Select **Brevo Provider**
   - Enter **API Key** (starts with `xkeysib-...`, NOT `xsmtpsib-...`)
   - Set From Address (must match verified address in Brevo)
   - Set From Name (your organization name)
5. Click **"Test Connection"** to verify
6. Click **"Save"**

**⚠️ CRITICAL: API Key Type**
- **CORRECT:** API Key (REST API) - starts with `xkeysib-...`
- **WRONG:** SMTP API Key - starts with `xsmtpsib-...` ❌
- WulfVault uses Brevo's REST API, NOT SMTP
- Using an SMTP key will result in "401 Unauthorized" errors

**Benefits:**
- Professional email delivery
- Higher deliverability rates
- Free tier available (300 emails/day)
- No email server maintenance required

### Testing Email Configuration

**Always test before going live:**
1. Click "Send Test Email"
2. Enter your email address
3. Check inbox (and spam folder)
4. Verify:
   - Email arrives
   - Branding looks correct
   - Links work properly

### Email Templates

**WulfVault sends emails for:**
- File sharing (when user emails download link)
- Password reset requests
- Account notifications

**Templates include:**
- Your company branding
- Custom logo (if uploaded)
- Professional formatting
- Secure links with proper expiration

---

## Logging & Monitoring

### Overview

WulfVault includes comprehensive logging and monitoring capabilities to help organizations maintain accountability, debug issues, and meet compliance requirements (GDPR, SOC 2, HIPAA, etc.).

**Three Types of Logs:**

1. **Audit Logs** - Compliance-focused: User actions, authentication, file lifecycle (database-stored)
2. **Server Logs** (v6.1.0+) - Operational: HTTP requests, upload events, system operations (file-based, 50MB max)
3. **SysMonitor Logs** (v6.1.0+) - Detailed monitoring: Every chunk upload, system metrics (file-based, 10MB max)

---

### 1. Audit Logs (Compliance)

#### Overview

Enterprise-grade audit logging for compliance requirements and accountability.

### What is Logged?

**User Management:**
- User creation, updates, deletion
- Account activation/deactivation
- Role changes and permission updates
- Quota modifications

**Authentication:**
- Login attempts (successful and failed)
- Logout events
- 2FA enable/disable operations
- Password changes
- Password reset requests and completions

**File Operations:**
- File uploads (who, when, file details)
- File deletions (soft delete to trash)
- File restorations from trash
- Permanent deletions
- File downloads (tracked separately per download)
- File sharing operations

**Team Management:**
- Team creation, updates, deletion
- Member additions and removals
- Role changes within teams
- File sharing to/from teams

**System Configuration:**
- Settings changes (server URL, quotas, retention periods)
- Branding updates (logo, colors, company name)
- Email configuration changes
- Audit log settings modifications

**Download Accounts:**
- Creation, updates, deletion
- Activation/deactivation

**File Requests:**
- Portal creation and deletion
- File uploads through request portals

### Accessing Audit Logs

**For Administrators:**

1. **Navigate to Server Settings:**
   - Login as admin
   - Click **Server** in navigation
   - Scroll down to **Audit Logs** section
   - Click **📊 View Audit Logs**

2. **Alternative route:**
   - Direct URL: `/admin/audit-logs`

### Filtering and Searching

**Filter Options:**
- **User:** See all actions by a specific user
- **Action Type:** Filter by operation (login, file upload, etc.)
- **Entity Type:** Filter by what was affected (User, File, Team, Settings)
- **Date Range:** View logs within specific time period
- **Search:** Free-text search across all log fields

**Example Queries:**
- "Who deleted this file?" → Filter by Entity Type: File, Action: FILE_DELETED
- "All failed logins this week" → Filter by Action: LOGIN_FAILED, Date Range: Last 7 days
- "What did user@example.com do?" → Filter by User

### Exporting Audit Logs

**For Compliance Reports:**

1. Apply desired filters (optional)
2. Click **Export to CSV** button
3. Download file: `audit_logs_YYYY-MM-DD_HH-MM-SS.csv`

**CSV Includes:**
- Log ID
- Timestamp (human-readable)
- User ID and Email
- Action performed
- Entity type and ID
- Detailed context (JSON)
- IP address (if IP logging enabled)
- User agent (browser/client info)
- Success/failure status
- Error messages (if failed)

**Use Cases:**
- Compliance audits (SOC 2, ISO 27001)
- Security investigations
- User activity reports
- Forensic analysis
- External audit tool integration

### Configuring Audit Log Settings

**Retention Policy:**

1. Go to **Server** → **Server Settings**
2. Find **Audit Log Retention (Days)** field
3. Set how many days to keep logs (default: 90)
4. Range: 1-3650 days (1 day to 10 years)
5. Click **Save Settings**

**Size Limit:**

1. In same **Server Settings** page
2. Find **Audit Log Max Size (MB)** field
3. Set maximum database size (default: 100 MB)
4. Range: 10-10000 MB
5. Click **Save Settings**

**How Cleanup Works:**
- Runs automatically every 24 hours
- Deletes logs older than retention period
- If size limit exceeded, deletes oldest logs first
- Maintains both time AND size limits
- Cleanup logged as AUDIT_LOG_CLEANUP action

**Recommended Settings by Organization Size:**

**Small Organizations (< 50 users):**
- Retention: 90 days
- Max Size: 100 MB
- Expected usage: Minimal impact

**Medium Organizations (50-500 users):**
- Retention: 180 days (6 months)
- Max Size: 500 MB
- Expected usage: Moderate log volume

**Large Organizations (500+ users):**
- Retention: 365 days (1 year)
- Max Size: 2000 MB (2 GB)
- Expected usage: High log volume

**Compliance Requirements:**
- **GDPR:** Typically 90-180 days
- **SOC 2:** Typically 365 days
- **HIPAA:** Typically 2555 days (7 years)
- **ISO 27001:** Typically 365-730 days

### Audit Log Security

**Protection Measures:**
- Logs are write-only (cannot be modified)
- Deletion only via automatic cleanup
- Admin access required to view logs
- Timestamped with server time
- IP logging optional (for privacy)

**Data Retention Compliance:**
- Automatic cleanup prevents indefinite retention
- Configurable to meet regulatory requirements
- Export before cleanup for long-term archival

### Troubleshooting Audit Logs

**Logs not appearing:**
- Check if actions are actually being performed
- Verify database migration completed (check server startup logs)
- Confirm admin access to `/admin/audit-logs`

**Export not working:**
- Check browser pop-up blocker
- Verify sufficient disk space
- Try smaller date range if large dataset

**Cleanup not running:**
- Check server logs for cleanup scheduler
- Verify retention days and max size are set
- Confirm server has been running for 24+ hours

---

### 2. Server Logs (Operational)

#### Overview (v6.1.0+)

Server Logs track all HTTP requests, upload events, and system operations. Perfect for debugging, monitoring performance, and tracking file uploads.

**What is Logged:**

1. **HTTP Requests:**
   - All API requests with status codes (200, 404, 500, etc.)
   - Request methods (GET, POST, DELETE, etc.)
   - Request and response sizes
   - Response times and duration
   - IP addresses for security tracking
   - User agents (browser/client info)

2. **Upload Events:**
   - **Upload Started:** Filename, size, upload ID, user, email, IP address
   - **Upload Completed:** Filename, size, duration, average speed, user, email, IP
   - **Upload Progress:** Every 100 chunks with percentage complete
   - **Upload Abandoned:** If upload inactive for >1 hour

3. **System Events:**
   - Server startup and shutdown
   - Configuration changes
   - Error conditions

#### Accessing Server Logs

**Navigate to Server Logs:**
1. Login as admin
2. Click **Server** in navigation
3. Click **Server Logs**
4. Direct URL: `/admin/server-logs`

#### Using Server Logs

**Filter Options:**
- **Search:** Free-text search across all log entries
- **Log Level:** Filter by Success (✅), Warning (⚠️), Error (❌), or Info (📝)
- **Date Range:** View logs within specific time period
- **Items Per Page:** 25, 50, 100, or 200 entries

**View Upload Details:**
- Filter by "UPLOAD" to see all upload-related events
- See when uploads started, completed, or were abandoned
- Track upload performance (speed, duration)
- Identify problematic uploads or network issues

**Export to CSV:**
- Click "Export CSV" button to download logs
- Includes all columns: timestamp, level, status, method, path, duration, sizes, IP
- Perfect for external analysis or reporting

#### Log Rotation

- **Maximum Size:** 50MB
- **Auto-Rotation:** When limit reached, current log renamed to `.old`
- **Location:** `data/server.log`
- **Old Logs:** `data/server.log.old` (overwritten on next rotation)

#### Use Cases

- **Debugging upload issues:** Search for specific filenames to see upload flow
- **Performance monitoring:** Track response times and identify slow endpoints
- **Security auditing:** Review IP addresses and failed requests
- **Capacity planning:** Analyze upload patterns and file sizes

---

### 3. SysMonitor Logs (Detailed Monitoring)

#### Overview (v6.1.0+)

SysMonitor Logs provide extremely detailed system monitoring, tracking every chunk upload with progress percentages. Ideal for system administrators debugging upload issues or monitoring system performance.

**What is Logged:**

1. **Chunk Uploads:**
   - Every single chunk upload (25MB each)
   - Progress percentage (e.g., "📦 Chunk 42 | Upload: 3f2a... | 1050/2100 bytes (50.0%)")
   - Upload ID (first 16 chars for readability)
   - Exact byte counts received vs total

2. **HTTP Chunk Requests:**
   - Successful chunk POST requests with status 200
   - Request sizes (always 25MB for full chunks)
   - Response times
   - IP addresses

**Why Separate from Server Logs?**

Large file uploads can generate thousands of chunk events (e.g., 50GB file = 2,000 chunks). Logging all these to Server Logs would create spam. SysMonitor keeps detailed tracking separate so Server Logs remain clean and readable.

#### Accessing SysMonitor Logs

**Navigate to SysMonitor:**
1. Login as admin
2. Click **Server** in navigation
3. Click **SysMonitor Logs**
4. Direct URL: `/admin/sysmonitor-logs`

#### Using SysMonitor Logs

**Features:**
- **Auto-Refresh:** Updates every 5 seconds automatically
- **Search:** Filter by upload ID, filename, or any keyword
- **Real-Time Monitoring:** Watch uploads progress in real-time
- **Monospace Font:** Easy-to-read technical log format

**Typical Workflow:**
1. User starts large file upload
2. Admin searches for filename or upload ID
3. Watch chunks increment: 0% → 25% → 50% → 75% → 100%
4. Identify stuck uploads (chunks stop incrementing)
5. Check for network issues (long gaps between chunks)

#### Log Rotation

- **Maximum Size:** 10MB (smaller to prevent excessive disk usage)
- **Auto-Rotation:** When limit reached, current log renamed to `.old`
- **Location:** `data/sysmonitor.log`
- **Old Logs:** `data/sysmonitor.log.old` (overwritten on next rotation)

#### Use Cases

- **Debug stuck uploads:** See exactly which chunk failed
- **Monitor large transfers:** Real-time tracking of multi-GB uploads
- **Performance analysis:** Identify slow chunks or network issues
- **Capacity planning:** Understand chunk upload patterns

#### Performance Impact

SysMonitor logging is designed for minimal performance impact:
- Logs written asynchronously (non-blocking)
- 10MB size limit prevents excessive disk usage
- Automatic rotation keeps only recent logs
- No database queries (file-based logging)

---

## Single Sign-On (SSO)

WulfVault can delegate sign-in to an external identity provider so users log in
with their existing organization accounts. Two provider types are supported:

- **Microsoft Entra ID (Azure AD)** — special-cased with multi-tenant aliases
  (`common`, `organizations`, `consumers`) and tenant validation.
- **Generic OpenID Connect (OIDC)** — works with any OIDC-compliant provider,
  including Google Workspace, Okta, Keycloak, Authelia, and Auth0.

The SSO flow uses OAuth 2.0 Authorization Code with PKCE and verifies the ID
token signature against the provider. WulfVault only uses SSO to establish
identity at sign-in; it does not request delegated API access.

> **Local accounts always remain available** as the primary mechanism and as a
> break-glass fallback — enabling SSO does not disable password login.

### Configuring SSO

**Access:** Admin → Settings → **Identity Providers** (`/admin/identity-providers`)

1. **Provider type** — choose *Microsoft Entra ID* or *Generic OpenID Connect*.
   The form shows only the fields relevant to your choice.
2. **Enable SSO** — turn the feature on.
3. **Provider display name** (optional) — text shown on the login button
   ("Sign in with …") and in invite emails. Defaults to "Microsoft" for Entra
   and "SSO" for Generic OIDC.
4. **Redirect URI** — must match exactly what is registered at the provider.
   The default (canonical) path is:
   ```
   https://your-instance/auth/oidc/callback
   ```
   The legacy path `https://your-instance/auth/entra/callback` also works for
   existing Entra installs.
5. **Client ID / Client Secret** — from your provider's app registration. The
   secret is stored encrypted at rest; leave the secret field blank when saving
   to keep the existing value.
6. **Provider-specific field:**
   - **Entra ID:** *Tenant ID* (Directory ID, or `common` / `organizations` /
     `consumers` for multi-tenant apps).
   - **Generic OIDC:** *Issuer URL* (the URL whose
     `/.well-known/openid-configuration` describes the provider, e.g.
     `https://accounts.google.com`). Optionally override *Scopes*
     (default `openid email profile`).
7. **Provisioning:**
   - **Auto-provision new users on first sign-in** — when enabled, a WulfVault
     account is created automatically the first time someone signs in via SSO.
   - **Default role for new users** — *User* or *Admin* (SuperAdmin cannot be
     auto-assigned).
   - **Allowed email domains** — comma-separated allow-list. Leave empty to
     accept any domain returned by the provider.
8. Click **Save Settings**.

#### Microsoft Entra ID — quick app registration

1. In **Entra ID → App registrations**, register a new application.
2. Add a **Web** redirect URI matching the value shown in WulfVault.
3. Under **Certificates & secrets**, create a client secret and copy its
   **Value** (not the secret ID).
4. Note the **Application (client) ID** and **Directory (tenant) ID**.
5. Paste these into the Identity Providers form and save.

### Forcing Local Accounts Only

Administrators can hard-disable SSO with the **"Force local accounts only"**
option on the Identity Providers page.

- It **overrides** the "Enable SSO" checkbox.
- The "Sign in with …" button is **hidden** on the login page.
- Any SSO callback is **rejected** (the `/auth/...` routes return Not Found).

Use this to guarantee that only password-based local accounts can sign in — for
example during incident response, or in environments where external IdP login is
not permitted.

### How SSO Identities Map to Accounts

- An SSO user is stored with an *identity provider* of `entra` (or the configured
  OIDC provider) and a stable external identifier (the Entra OID, or the OIDC
  `sub` claim).
- WulfVault matches the verified email/UPN from the provider to an existing
  account, or — if auto-provisioning is on and the domain is allowed — creates a
  new one with the configured default role.
- SSO sign-in events are recorded in the audit log
  (`ENTRA_LOGIN_SUCCESS`, `ENTRA_USER_PROVISIONED`, `ENTRA_USER_BOUND`,
  `ENTRA_LOGIN_REJECTED`).

> **Download Accounts are unaffected by SSO.** They are always password-based and
> work even when SSO is the only option configured for staff users.

---

## Security Features

### Two-Factor Authentication (2FA)

#### Enabling 2FA (Users)

1. **Go to Settings** (user/admin dashboard)
2. **Find "Two-Factor Authentication"** section
3. **Click "Enable 2FA"**
4. **Scan QR code** with authenticator app:
   - Google Authenticator
   - Authy
   - Microsoft Authenticator
   - Any TOTP app
5. **Save backup codes** (shown once!)
   - Store in password manager
   - Print and store securely
   - Needed if you lose phone
6. **Enter verification code** from app
7. **2FA enabled** ✅

#### Using 2FA at Login

1. Enter email and password normally
2. **2FA prompt appears**
3. Open authenticator app
4. Enter 6-digit code
5. Login completes

#### Backup Codes

**When to use:**
- Lost phone with authenticator app
- Authenticator app not working
- Emergency access needed

**Using backup code:**
1. At 2FA prompt, click "Use Backup Code"
2. Enter one of your backup codes
3. Login successful
4. **Code is consumed** (one-time use only)

**Regenerating codes:**
1. Settings → 2FA section
2. Click "Regenerate Backup Codes"
3. **Old codes invalidated immediately**
4. Save new codes securely

### Password Security Best Practices

**Strong Passwords:**
- Minimum 12 characters recommended
- Mix of uppercase, lowercase, numbers, symbols
- Avoid common words or patterns
- Use password manager

**Password Reset:**
1. **User forgot password:**
   - Click "Forgot Password" on login
   - Enter email address
   - Receive reset email (valid 24 hours)
   - Click link in email
   - Set new password

2. **Admin cannot reset user passwords**
   - Users must use self-service reset
   - Or admin creates new account

### Session Security

**Session Features:**
- **Auto-expiration:** 24 hours (default, configurable via `SESSION_TIMEOUT_HOURS`)
- **Secure cookies:** HttpOnly, SameSite protection
- **CSRF protection:** All forms protected
- **Session cleanup:** Expired sessions automatically removed every hour

**Best Practices:**
- Always logout on shared computers
- Don't share session links
- Use HTTPS in production

### IP Address Logging

**Configuration:** Admin → Settings → Save IP Addresses
**Default:** Disabled (for privacy)

**When enabled:**
- IP addresses logged for all downloads
- Useful for security audits and compliance
- Helps trace unauthorized access
- Required for detailed forensic analysis

**When disabled (default):**
- Privacy-focused mode
- Better GDPR compliance
- Still logs download events (date, time, email)
- No IP addresses stored

---

## File Request Portals

### What are File Requests?

Create a shareable upload link that allows others to upload files TO you.

**Use Cases:**
- Collect files from clients
- Receive project deliverables
- Accept submissions or applications
- Temporary upload portals

### Creating File Requests

1. **Go to Dashboard**
2. **Click "Create File Request"**
3. **Configure:**
   - **Name:** Descriptive name (e.g., "Client Logo Uploads")
   - **Max File Size:** Limit per file
   - **Max Uploads:** Total number of uploads allowed
   - **Expiration Date:** When portal closes
   - **Password:** Optional protection
4. **Click "Create"**
5. **Copy shareable link**

### Sharing File Request Links

**Send link to anyone who should upload:**
```
https://your-instance/upload-request/abc123xyz
```

Recipients can:
- Upload files without account
- See upload confirmation
- No access to previously uploaded files

### Managing File Requests

**View Requests:**
1. Dashboard → "File Requests" tab
2. See all your upload portals

**Request Details:**
- Files received count
- Uploads remaining
- Expiration status
- Link and password

**Received Files:**
- Appear in your regular file list
- Marked as "via file request"
- Can be shared normally afterward

**Delete Request:**
- Stops new uploads
- Previous uploads remain available
- Portal link becomes invalid

---

## Troubleshooting

### Common Issues

#### Cannot Login

**Symptoms:** "Invalid credentials" error

**Solutions:**
1. **Verify credentials:**
   - Email address (case-sensitive)
   - Password (case-sensitive)
   - Check caps lock

2. **Reset password:**
   - Click "Forgot Password"
   - Check email (including spam)
   - Follow reset link

3. **Account inactive:**
   - Contact admin
   - Admin can reactivate account

4. **2FA issues:**
   - Try backup code
   - Ensure phone time is correct (TOTP is time-based)
   - Contact admin if locked out

5. **SSO ("Sign in with …") issues:**
   - If the SSO button is missing, your admin has not enabled SSO or has forced
     local accounts only — use email + password instead.
   - "Sign-in state did not match" / "took too long": start again from the login
     page and make sure cookies are enabled.
   - "Rejected the account": your email domain may not be on the allowed list, or
     auto-provisioning is off and no account exists yet — contact your admin.

#### Upload Fails

**Symptoms:** Upload progress bar stops or errors

**Solutions:**
1. **Check file size:**
   - Must be under configured limit
   - Default: 5 GB, configurable with no practical upper limit

2. **Check quota:**
   - Dashboard shows quota usage
   - Delete old files to free space

3. **Network issues:**
   - Verify internet connection
   - Try again later
   - Use wired connection for large files

4. **Browser issues:**
   - Try different browser
   - Clear cache and cookies
   - Disable browser extensions

#### Email Not Received

**Symptoms:** Password reset or share email not arriving

**Solutions:**
1. **Check spam folder** - Often filtered incorrectly

2. **Verify email address:**
   - Correct spelling
   - No extra spaces

3. **Email configuration:**
   - Admin: Test email settings
   - Verify SMTP/Brevo credentials
   - Check email provider limits

4. **Wait time:**
   - Can take 1-5 minutes
   - Check again before retrying

#### Download Link Doesn't Work

**Symptoms:** "File not found" or "Expired" message

**Solutions:**
1. **Check expiration:**
   - Download limit reached?
   - Date expiration passed?

2. **Verify link:**
   - Full URL copied correctly
   - No line breaks in middle of link

3. **Contact sender:**
   - Ask them to check file status
   - Request new link if expired

4. **Account issues:**
   - If authenticated download, verify account active
   - Try password reset

### Getting Help

**Self-Service Resources:**
- This User Guide
- README.md in repository
- Online documentation

**Contact Admin:**
- For account issues
- Quota increase requests
- Technical problems

**System Logs:**
- Admins can check server logs
- Located in data directory
- Help diagnose technical issues

### Understanding Server Logs

**Log Location:**
- Server logs: `data/logs/server.log`
- Audit logs: Accessible via Admin → Server → View Audit Logs

**HTTP Status Codes:**

Server logs display HTTP status codes to indicate the outcome of requests. Understanding these codes helps diagnose issues quickly:

| Status Code | Color | Meaning | Common Causes |
|-------------|-------|---------|---------------|
| **200** | 🟢 Green | Success | Request completed successfully |
| **201** | 🟢 Green | Created | Resource created successfully (e.g., file uploaded) |
| **204** | 🟢 Green | No Content | Success, no data to return (e.g., delete operation) |
| **301** | 🟡 Yellow | Moved Permanently | Resource moved to new URL |
| **302** | 🟡 Yellow | Found (Redirect) | Temporary redirect to another page |
| **304** | 🟡 Yellow | Not Modified | Cached content still valid |
| **400** | 🔴 Red | Bad Request | Invalid request format or parameters |
| **401** | 🔴 Red | Unauthorized | Authentication required or failed |
| **403** | 🔴 Red | Forbidden | Access denied (insufficient permissions) |
| **404** | 🔴 Red | Not Found | Resource doesn't exist (deleted/expired file) |
| **405** | 🟡 Yellow | Method Not Allowed | Wrong HTTP method (e.g., GET instead of POST) |
| **409** | 🔴 Red | Conflict | Request conflicts with current state |
| **413** | 🔴 Red | Payload Too Large | File size exceeds configured limit |
| **429** | 🟡 Yellow | Too Many Requests | Rate limit exceeded |
| **500** | 🔴 Red | Internal Server Error | Server-side error (check logs for details) |
| **502** | 🔴 Red | Bad Gateway | Proxy/gateway error |
| **503** | 🔴 Red | Service Unavailable | Server temporarily unavailable |

**Status Code Categories:**
- **2xx (Green):** Success - Request processed successfully
- **3xx (Yellow):** Redirection - Client needs to take additional action
- **4xx (Red/Yellow):** Client Error - Problem with the request
- **5xx (Red):** Server Error - Server failed to process valid request

**Common Log Patterns:**

**Normal Operations:**
```
200 GET  /dashboard              # User viewing dashboard
201 POST /upload                 # File uploaded successfully
302 GET  /login                  # Redirect to login page
```

**Authentication Issues:**
```
401 POST /api/login              # Failed login attempt
403 GET  /admin                  # User without admin access
```

**File Problems:**
```
404 GET  /d/abc123xyz            # Expired or deleted file
413 POST /upload                 # File too large
```

**Server Issues:**
```
500 POST /api/user/create        # Server error creating user
503 GET  /                       # Server overloaded
```

**Troubleshooting with Status Codes:**

1. **Frequent 401/403:** Check authentication system and user permissions
2. **Many 404s:** Files expiring too quickly or broken links being shared
3. **413 errors:** Increase `MAX_FILE_SIZE_MB` or use smaller files
4. **500 errors:** Check server logs for stack traces and error details
5. **429 errors:** Implement longer delays between requests

---

## Best Practices

### For Administrators

#### Security
- ✅ Change default admin password immediately
- ✅ Enable 2FA for all admin accounts
- ✅ Use HTTPS in production (SSL certificate)
- ✅ Regular backups of data and uploads directories
- ✅ Keep WulfVault updated
- ✅ Review user accounts quarterly (remove inactive)
- ✅ Monitor storage usage and set appropriate quotas
- ✅ Configure email for password resets

#### User Management
- ✅ Create admin accounts only for trusted personnel
- ✅ Set reasonable storage quotas (can always increase)
- ✅ Deactivate users instead of deleting (preserves data)
- ✅ Regularly review download accounts (cleanup old ones)
- ✅ Document your organization's user policies

#### File Management
- ✅ Configure trash retention (5-7 days recommended)
- ✅ Regularly review and clean trash
- ✅ Monitor system storage capacity
- ✅ Set maximum file size appropriately for your use case
- ✅ Regular database backups

#### Branding
- ✅ Upload professional logo
- ✅ Match colors to corporate branding
- ✅ Use company domain (e.g., files.company.com)
- ✅ Configure email with company branding
- ✅ Test download experience from recipient perspective

### For Users

#### Uploading Files
- ✅ Use descriptive file names (recipients see this)
- ✅ Set appropriate expiration (don't leave files forever)
- ✅ Use password protection for sensitive files
- ✅ Require authentication for external recipients
- ✅ Check quota before large uploads
- ✅ Delete old files you no longer need

#### Sharing Files
- ✅ Use email integration for professional delivery
- ✅ Include context in email message
- ✅ Send password separately (if using password protection)
- ✅ Verify recipient email address before sending
- ✅ Monitor download history
- ✅ Notify recipients when file will expire

#### Security
- ✅ Enable 2FA on your account
- ✅ Use strong, unique passwords
- ✅ Logout on shared/public computers
- ✅ Don't share your login credentials
- ✅ Save backup codes securely
- ✅ Report suspicious activity to admin

### For Download Account Users

#### Best Practices
- ✅ Create strong password when first accessing file
- ✅ Save your download portal link for future reference
- ✅ Download files promptly (before expiration)
- ✅ Delete account when no longer needed (GDPR compliance)
- ✅ Use password manager to store credentials

---

## Appendix

### Keyboard Shortcuts

**Dashboard:**
- `Ctrl/Cmd + U` - Focus upload button
- `Esc` - Close modals

**File Management:**
- `Ctrl/Cmd + F` - Search files
- `Arrow Keys` - Navigate file list

### File Size Limits

**Default Configuration:**
- Maximum file size: **5000 MB (5 GB)** default - Configurable via `MAX_FILE_SIZE_MB` environment variable, no practical upper limit
- Maximum upload size: **5000 MB (5 GB)** default - Configurable with no practical upper limit (chunked uploads, tested with large video surveillance files)
- Default user quota: **5000 MB (5 GB)** - Configurable per user by admin
- Total storage: Based on individual user quotas

**Network Considerations:**
- Large files (>1GB): Use wired connection
- Very large files (>3GB): May take time, be patient
- Upload speed depends on internet connection

### Browser Compatibility

**Fully Supported:**
- ✅ Chrome 90+
- ✅ Firefox 88+
- ✅ Safari 14+
- ✅ Edge 90+

**Mobile:**
- ✅ iOS Safari 14+
- ✅ Android Chrome 90+

**Not Supported:**
- ❌ Internet Explorer (any version)

### System Requirements

**Minimum Server:**
- CPU: 1 core
- RAM: 512 MB
- Storage: 10 GB + file storage space
- OS: Linux (Ubuntu, Debian, RHEL)

**Recommended Server:**
- CPU: 2+ cores
- RAM: 2+ GB
- Storage: 100+ GB (SSD/NVMe recommended for better performance, spinning disks supported)
- OS: Linux with Docker support

### Support & Resources

**Documentation:**
- User Guide (this document)
- README: https://github.com/Frimurare/WulfVault
- Installation Guide: INSTALLATION.md

**Community:**
- GitHub Issues: Report bugs
- GitHub Discussions: Ask questions
- Wiki: Additional documentation

**Updates:**
- Check GitHub for new releases
- Subscribe to release notifications
- Review CHANGELOG.md for changes

---

## Glossary

**2FA (Two-Factor Authentication):** Security feature requiring two forms of verification - password plus authenticator app code.

**Authenticated Download:** File sharing mode requiring recipient to create account before downloading.

**Backup Codes:** One-time use codes for 2FA emergency access if authenticator app unavailable.

**Branding:** Customization of WulfVault appearance with logo, colors, and company name.

**Direct Download:** File sharing mode allowing immediate download without recipient account.

**Download Account:** Automatically created account for recipients of authenticated downloads.

**Expiration:** Automatic file deletion based on download count or date limit.

**File Request:** Upload portal allowing others to upload files to you.

**GDPR:** General Data Protection Regulation - EU privacy law. WulfVault includes compliance features.

**Quota:** Storage limit per user, configurable by administrators.

**Splash Page:** Download page recipients see when clicking file link.

**TOTP (Time-based One-Time Password):** Standard used by authenticator apps for 2FA codes.

**Trash:** Temporary storage for deleted files before permanent deletion.

---

**WulfVault v7.1.0 Aurora**
**Copyright © 2025 Ulf Holmström (Frimurare)**
**Licensed under AGPL-3.0**

---

**End of User Guide**

For the latest version of this guide and additional documentation, visit:
https://github.com/Frimurare/WulfVault

🎉 **Enjoy secure, professional file sharing with complete control!**
