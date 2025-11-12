# Sharecare User Guide v3.2.3

**Complete Guide for Administrators and Users**

---

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [User Roles & Permissions](#user-roles--permissions)
4. [File Sharing Guide](#file-sharing-guide)
5. [Download Account Guide](#download-account-guide)
6. [Admin Dashboard](#admin-dashboard)
7. [User Management](#user-management)
8. [File Management](#file-management)
9. [Branding & Customization](#branding--customization)
10. [Email Configuration](#email-configuration)
11. [Security Features](#security-features)
12. [File Request Portals](#file-request-portals)
13. [Troubleshooting](#troubleshooting)
14. [Best Practices](#best-practices)

---

## Introduction

### What is Sharecare?

Sharecare is a professional-grade, self-hosted file sharing platform designed for organizations that demand security, accountability, and complete control over their data. Unlike commercial services like WeTransfer or Sprend, Sharecare gives you:

- **Complete data ownership** - Files stay on your infrastructure
- **No subscription fees** - One-time setup, unlimited transfers
- **Enterprise features** - Multi-user management, audit trails, 2FA
- **Full customization** - Branded pages, custom domains, logo

### Key Features Overview

- ✅ **Large file support** - Up to 5GB+ per file (configurable)
- ✅ **Multi-user system** - Admins, users, and download accounts
- ✅ **Download tracking** - Know exactly who downloaded what and when
- ✅ **Two-Factor Authentication** - TOTP-based security
- ✅ **Email integration** - SMTP or Brevo for sending links
- ✅ **Branding** - Custom logo, colors, company name
- ✅ **File requests** - Create upload portals for receiving files
- ✅ **Storage quotas** - Per-user limits with usage tracking

### Who Should Use This Guide?

- **System Administrators** - Setting up and managing Sharecare
- **Regular Users** - Uploading and sharing files
- **Download Users** - Receiving and accessing files
- **IT Managers** - Understanding security and compliance features

---

## Getting Started

### First-Time Login

1. **Navigate to your Sharecare instance:**
   ```
   http://your-server:8080
   or
   https://files.yourdomain.com
   ```

2. **Default Admin Credentials** (first-time setup only):
   - Email: `admin@sharecare.local`
   - Password: `SharecareAdmin2024!`

3. **⚠️ CRITICAL SECURITY STEP:**
   - Immediately go to Settings → Change Password
   - Choose a strong, unique password
   - Never share admin credentials

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
   https://your-sharecare-instance/download/dashboard
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
- **Storage Quota:** Typically lower (e.g., 10GB)

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
- **Custom:** Any size in MB (e.g., 5000 MB = 5GB)
- **Recommended:**
  - Regular users: 5-10 GB
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
- **Retention Period:** Files kept for configured days (default: 5)
- **Automatic Cleanup:** Old items permanently deleted
- **Manual Actions:**
  - **Restore:** Bring file back (restores to original uploader)
  - **Permanent Delete:** Delete immediately (cannot be undone)

**View Trash:**
- See deleted files
- Who deleted them
- When deleted
- Days remaining before permanent deletion

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

---

## Email Configuration

### Why Configure Email?

Enable email functionality to:
- Send download links directly from Sharecare
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
2. Get API key from Brevo dashboard
3. In Sharecare:
   - Select **Brevo Provider**
   - Enter API Key
   - Set From Address (must be verified in Brevo)
   - Set From Name
4. Test configuration
5. Save

**Benefits:**
- Professional email delivery
- Higher deliverability rates
- Free tier available (300 emails/day)

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

**Sharecare sends emails for:**
- File sharing (when user emails download link)
- Password reset requests
- Account notifications

**Templates include:**
- Your company branding
- Custom logo (if uploaded)
- Professional formatting
- Secure links with proper expiration

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
- **Auto-expiration:** 24 hours (configurable)
- **Secure cookies:** HttpOnly, SameSite protection
- **CSRF protection:** All forms protected

**Best Practices:**
- Always logout on shared computers
- Don't share session links
- Use HTTPS in production

### IP Address Logging

**Configuration:** Admin → Settings → Save IP Addresses

**When enabled:**
- IP addresses logged for all downloads
- Useful for security audits
- Helps trace unauthorized access

**When disabled:**
- Privacy-focused mode
- GDPR compliance in some regions
- Still logs download events (without IP)

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

#### Upload Fails

**Symptoms:** Upload progress bar stops or errors

**Solutions:**
1. **Check file size:**
   - Must be under configured limit
   - Default: 2GB, configurable to 5GB+

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

---

## Best Practices

### For Administrators

#### Security
- ✅ Change default admin password immediately
- ✅ Enable 2FA for all admin accounts
- ✅ Use HTTPS in production (SSL certificate)
- ✅ Regular backups of data and uploads directories
- ✅ Keep Sharecare updated
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
- Maximum file size: 2GB
- Configurable up to 5GB+ (tested with large video files)
- Total storage: Based on quota per user

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
- Storage: 10 GB (plus file storage)
- OS: Linux (Ubuntu, Debian, RHEL)

**Recommended Server:**
- CPU: 2+ cores
- RAM: 2+ GB
- Storage: 100+ GB SSD
- OS: Linux with Docker support

### Support & Resources

**Documentation:**
- User Guide (this document)
- README: https://github.com/Frimurare/Sharecare
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

**Branding:** Customization of Sharecare appearance with logo, colors, and company name.

**Direct Download:** File sharing mode allowing immediate download without recipient account.

**Download Account:** Automatically created account for recipients of authenticated downloads.

**Expiration:** Automatic file deletion based on download count or date limit.

**File Request:** Upload portal allowing others to upload files to you.

**GDPR:** General Data Protection Regulation - EU privacy law. Sharecare includes compliance features.

**Quota:** Storage limit per user, configurable by administrators.

**Splash Page:** Download page recipients see when clicking file link.

**TOTP (Time-based One-Time Password):** Standard used by authenticator apps for 2FA codes.

**Trash:** Temporary storage for deleted files before permanent deletion.

---

**Sharecare v3.2.3 Golden Release**
**Copyright © 2025 Ulf Holmström (Frimurare)**
**Licensed under GPL-3.0**

*Architecturally inspired by Gokapi - Complete rewrite with enterprise features*

---

**End of User Guide**

For the latest version of this guide and additional documentation, visit:
https://github.com/Frimurare/Sharecare

🎉 **Enjoy secure, professional file sharing with complete control!**
