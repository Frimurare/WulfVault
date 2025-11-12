# Changelog

## [3.3.4] - 2025-11-12 ✨ Welcome Email Feature

### ✨ New Feature

**Welcome Email with Password Setup Link**
- **Feature Added**: Admins can now send welcome emails to new users with a password setup link
- **Use Case**: No need to share passwords manually - users set their own password securely
- **Email Branding**: Includes company logo and name from branding settings
- **User Experience**: Clean, professional welcome email with clear instructions

### 📋 How It Works

**Admin Experience:**
1. Navigate to Admin → Users → Create User
2. Fill in user details (name, email, quota, level)
3. Check "📧 Send welcome email with password setup link" (checked by default)
4. Click Save
5. User receives branded welcome email immediately

**User Experience:**
1. Receives professional welcome email with company branding
2. Email includes their login email and a "Set Password & Login" button
3. Click button to visit secure password setup page
4. Creates their own password
5. Automatically logs in to their account

### 📧 Email Template Features
- Company logo display (from branding settings)
- Personalized with company name
- Secure one-time password setup link (1-hour validity)
- Mobile-friendly responsive design
- Clear instructions and call-to-action button
- Professional gradient design matching Sharecare style

### Technical Details

**Modified Files:**
- `internal/email/templates.go`:
  - Added `SendWelcomeEmail()` function with branding support
  - Accepts company name and logo for customization
  - Generates secure reset token link

- `internal/server/handlers_admin.go` (lines 107-217):
  - Added welcome email checkbox to user creation form
  - Generates temporary password if welcome email is enabled
  - Creates password reset token after user creation
  - Sends branded welcome email via Brevo
  - Logs email sending success/failure

- `cmd/server/main.go` (line 25):
  - Version bumped to 3.3.4

**Security:**
- Uses existing password reset infrastructure (1-hour token validity)
- Temporary password generated and immediately replaced via email
- User must have email access to complete setup
- Failed email sends don't prevent user creation (graceful degradation)

**Configuration:**
- Requires email provider configured in admin settings (Brevo)
- Uses branding settings for company name and logo
- Respects server URL configuration for link generation

---

## [3.3.3] - 2025-11-12 🐛 Critical User Deletion Fix

### 🐛 Bug Fixes

**User Deletion Now Works**
- **Issue Fixed**: Admin user deletion button appeared to do nothing - users weren't deleted
- **Root Cause**: JavaScript `deleteUser()` function reloaded page without validating server response
- **Solution**: Implemented proper async/await pattern with response validation
- **Impact**: Users can now be successfully deleted via admin panel, files properly moved to trash

**Trash Display for Deleted Users**
- **Issue Fixed**: Files from deleted users showed "Unknown" as owner in trash view
- **Solution**: Changed default display text to "Deleted user" for better clarity
- **Impact**: More intuitive trash view when viewing files from deleted accounts

### 📋 Technical Details

**Modified Files:**
- `internal/server/handlers_admin.go` (lines 1393-1412):
  - Converted `deleteUser()` from callback to async/await
  - Added `response.ok` validation before page reload
  - Proper error handling with user-friendly messages
  - Files correctly moved to trash with 5-day retention

- `internal/server/handlers_admin.go` (line 2487):
  - Changed owner display from "Unknown" to "Deleted user"
  - Applied to both trash view instances

- `cmd/server/main.go` (line 25):
  - Version bumped to 3.3.3

- `.gitignore`:
  - Added node_modules/ to ignore list

**User Deletion Flow (Now Fixed):**
1. Admin clicks Delete → Confirmation dialog
2. Server validates and deletes user from database
3. User's files moved to trash (DeletedAt set, 5-day retention)
4. Success: Page reloads, user removed from list
5. Error: Alert shown with specific error message

---

## [3.3.2] - 2025-11-12 🐛 Quick Bugfix - Copy Button

### 🐛 Bug Fix

**Copy URL Button Fixed**
- **Issue Fixed**: "COPY URL" button in admin settings didn't work due to clipboard API limitations on HTTP
- **Solution**: Added fallback to `document.execCommand('copy')` for HTTP contexts
- **Impact**: Copy button now works reliably on both HTTP and HTTPS connections

**Technical Details:**
- `internal/server/handlers_admin.go`: Improved copy function with dual-method approach
  - Primary: Modern clipboard API (for HTTPS)
  - Fallback: execCommand (for HTTP - more compatible)
- Better error handling with user-friendly messages

---

## [3.3.1] - 2025-11-12 🔧 Critical Configuration Fix

### 🐛 Critical Bug Fixes

**Server URL Configuration Priority Fixed**
- **Issue Fixed**: Environment variable `SERVER_URL` was overriding admin panel settings, causing link generation issues
- **Solution**: Database settings (from admin panel) now have highest priority over environment variables
- **Impact**: Admin-configured URLs persist across server restarts, fixing incorrect link generation

### ✨ UI Improvements

**Public URL Display in Admin Settings**
- Added prominent "Current Public URL" display box at top of settings page
- Shows the exact URL that users should use to access the system
- One-click "COPY URL" button for easy sharing
- Visual feedback when URL is copied to clipboard
- Highlighted in yellow with red text for high visibility

**Configuration Priority (Fixed):**
1. **Database (Admin Panel Settings)** - Highest priority ✅
2. **Environment Variables** - Second priority
3. **Config.json** - Fallback default

**Benefits:**
- ✅ Settings configured in admin panel persist across restarts
- ✅ No need to edit systemd service files for URL changes
- ✅ Clear visibility of public URL for easy user communication
- ✅ One-click URL copying for administrators

**Technical Details:**
- `cmd/server/main.go`: Fixed configuration priority loading (lines 82-97)
- `internal/server/handlers_admin.go`: Added public URL display and copy functionality

---

## [3.3.0] - 2025-11-12 🔧 Critical Bugfix Release

### 🐛 Critical Bug Fixes

**File Orphaning Prevention**
- **Issue Fixed**: When admins deleted users, their uploaded files remained in the system without an owner, consuming storage indefinitely
- **Solution**: All user files are now automatically moved to trash (soft-deleted) when user is deleted
- **Impact**: Prevents storage waste, maintains data integrity, enables file recovery

### ✨ Improvements

**Enhanced User Deletion Workflow**
- Added `SoftDeleteUserFiles()` database function to handle file cleanup
- Updated `DeleteUser()` to accept `deletedBy` parameter for audit trail
- Modified admin handler to capture admin ID during deletion
- Improved confirmation dialog with detailed information:
  - Warns admin that files will be moved to trash
  - Explains 5-day retention period
  - Clarifies files can be recovered or permanently deleted

**Benefits:**
- ✅ No orphaned files consuming storage
- ✅ Files recoverable from trash for 5 days
- ✅ Complete audit trail (who deleted files)
- ✅ Admin informed before destructive actions
- ✅ Consistent with existing trash workflow

**Technical Details:**
- `internal/database/files.go`: Added `SoftDeleteUserFiles(userId, deletedBy)`
- `internal/database/users.go`: Updated `DeleteUser(id, deletedBy)` signature
- `internal/server/handlers_admin.go`: Enhanced confirmation message and admin ID tracking

---

## [3.2.3] - 2025-11-12 🏆 Golden Release

### 🎉 GOLDEN RELEASE - Production Ready

This is the first stable production release of Sharecare, marking a complete rewrite (~95% new code) architecturally inspired by Gokapi.

### 📚 New Documentation
- **Comprehensive User Guide**: 76-page complete manual covering all features
  - Administrator guide with setup and configuration
  - User workflows for file sharing and management
  - Download account portal documentation
  - Security best practices
  - Troubleshooting section
  - Available as `USER_GUIDE.md` (convertible to PDF)

### ⚖️ Attribution & Licensing Updates
- **Copyright Headers**: Added to all 55 .go files, .js, and .css files
- **Meta Tags**: `<meta name="author" content="Ulf Holmström">` in all 29 HTML pages
- **Attribution Footer**: "Powered by Sharecare © Ulf Holmström – AGPL-3.0" in Settings pages
- **Project Files**:
  - `NOTICE` - Copyright and attribution requirements
  - `AUTHORS` - Project contributors
  - `CODEOWNERS` - Code ownership (@Frimurare)
- **Watermark Constant**: `SharecareSignature` in config.go
- **License**: Updated to AGPL-3.0 with network copyleft protection
- **Clarity**: Updated attribution from "Based on Gokapi" to "Architecturally inspired by Gokapi — Complete rewrite (~95% new code)"

### 📝 Enhanced README
- **Feature Categories**: Organized into 8 comprehensive sections:
  - 🚀 File Sharing & Transfer (10 features)
  - 👥 User Management & Access Control (6 features)
  - 📊 Download Tracking & Accountability (6 features)
  - 🔐 Security & Authentication (5 categories)
  - 🎨 Branding & Customization (3 categories)
  - 🌐 Email & Notifications (3 categories)
  - 📁 File Request System (2 categories)
  - 🔧 Administration & Management (4 categories)
- **Clear Positioning**: Emphasizes Sharecare as complete alternative to commercial file transfer services
- **Target Audience**: Expanded to include government, education, healthcare sectors

### 🔍 Code Analysis & Documentation
- **Total Codebase**: 18,016 lines of Go code + 733 lines of tests
- **Gokapi Code Usage**: 0 production imports (only 5 test utility imports)
- **Original Features**: ~80% completely new code
  - Multi-user system: ~11,000 lines
  - Email integration: 1,042 lines
  - 2FA implementation: 118 lines
  - Download accounts: ~1,500 lines
  - Admin dashboards: ~2,000 lines
- **Conceptual Similarity**: ~15% (basic models, database schema foundation)

### 🛠️ Technical Improvements
- **Version Management**: Consistent v3.2.3 across all files and frontend
- **Documentation Structure**: Clear separation of user, admin, and developer docs
- **License Compliance**: Full AGPL-3.0 compliance with proper notices and network copyleft

### 📊 Statistics

**Code Distribution:**
- internal/server: 10,973 lines (58.5%) - HTTP handlers, routing
- internal/database: 2,654 lines (14.1%) - Custom SQLite layer
- internal/models: 2,502 lines (13.3%) - Data structures
- internal/email: 1,042 lines (5.6%) - Email integration
- internal/auth: 263 lines (1.4%) - Authentication
- internal/totp: 118 lines (0.6%) - Two-Factor Auth

**Features NOT in Gokapi (100% Sharecare):**
- Multi-user authentication system
- Role-based access control (4 user types)
- Email integration (SMTP/Brevo)
- Two-Factor Authentication
- Download account portal
- File request upload portals
- Comprehensive audit logging
- Branding system
- Storage quota management
- Self-service password reset
- GDPR compliance features

### 🎯 Production Readiness

This release represents a stable, feature-complete, production-ready file sharing platform with:
- ✅ Complete documentation for all user types
- ✅ Proper licensing and attribution
- ✅ Comprehensive feature set
- ✅ Security best practices implemented
- ✅ GDPR compliance built-in
- ✅ Professional branding capabilities
- ✅ Enterprise-grade audit trails

### 🍾 Milestone

**Golden Release 3.2.3** marks the transition from beta/RC to stable production software. Ready for deployment in enterprise environments requiring:
- Secure file transfer with accountability
- Multi-tenant user management
- Compliance and audit requirements
- Custom branding and white-labeling
- Complete data sovereignty

---

## [3.2.2-RC3] - 2025-11-12 🔧 Critical Bug Fix

### Bug Fixes
- **🐛 CRITICAL: Localhost URL Override**: Fixed server URL defaulting to localhost when port ≠ 8080
  - Problem: When using custom port (e.g., 3000), SERVER_URL was being overridden with "http://localhost:3000"
  - Impact: Download links broke when using custom domains with non-standard ports
  - Solution: Removed automatic localhost fallback in `getPublicURL()` function
  - Now properly uses configured SERVER_URL regardless of port

### Details
- Modified `internal/server/server.go:getPublicURL()` to trust SERVER_URL environment variable
- Only port is appended if SERVER_URL doesn't already contain it
- Prevents production issues when running on non-standard ports with custom domains

---

## [3.2-RC2] - 2025-11-12 🚀 Comprehensive Analytics Dashboard

### New Features - Extended Dashboard Analytics
This release adds extensive new analytics capabilities to the admin dashboard, providing deep insights into usage patterns, security posture, file statistics, and trends.

#### 📈 Usage Statistics
- **Active Files**: Track files downloaded in last 7 and 30 days
- **Average File Size**: Monitor typical file sizes across the platform
- **Average Downloads per File**: Understand file popularity and sharing patterns

#### 🔐 Security Overview
- **2FA Adoption Rate**: Track percentage of Users/Admins with Two-Factor Authentication enabled
- **Average Backup Codes Remaining**: Monitor backup code usage to identify users who may need to regenerate

#### 📁 File Statistics
- **Largest File**: Display the biggest file currently stored with size
- **Most Active User**: Highlight the user who has uploaded the most files

#### ⚡ Trend Data
- **Top File Types**: Show the 3 most common file extensions with counts
- **Most Active Weekday**: Identify which day of the week has the most download activity
- **Storage Trend**: Display storage growth over the last 30 days with percentage change

### Bug Fixes
- **🐛 Critical: Historical Data Accuracy**: Fixed bug where deleting files would retroactively remove them from historical statistics
  - Previously, when a file was deleted, ALL statistics (uploaded/downloaded data for all periods) would decrease
  - Now statistics correctly reflect historical data - if a file was uploaded in January, it remains in the year's upload statistics even after deletion
  - Affects: `GetBytesUploadedToday/Week/Month/Year()` - removed `AND DeletedAt = 0` clause
  - Download statistics were already correct (using DownloadLogs) but added clarifying comments

### Implementation Details
- **New Database Methods** (15 new methods in `internal/database/`):
  - Usage: `GetActiveFilesLast7Days()`, `GetActiveFilesLast30Days()`, `GetAverageFileSize()`, `GetAverageDownloadsPerFile()`
  - Security: `Get2FAAdoptionRate()`, `GetAverageBackupCodesRemaining()`
  - Files: `GetLargestFile()`, `GetMostActiveUser()`
  - Trends: `GetTopFileTypes()`, `GetMostActiveWeekday()`, `GetStorageTrendLastMonth()`
- **Dashboard UI**: 5 new sections with 14 additional stat cards
- **Performance**: All queries optimized with proper aggregation and indexing

### Benefits
- **Complete Visibility**: Admins now have comprehensive insights into platform usage
- **Proactive Security**: Monitor 2FA adoption and identify users needing attention
- **Trend Analysis**: Understand usage patterns to inform capacity planning
- **Historical Accuracy**: Statistics now correctly represent historical data regardless of deletions

---

## [3.2-beta4] - 2025-11-12 📊 Data Transfer Statistics Enhancement

### New Features
- **📊 Enhanced Data Transfer Dashboard**: Split statistics into Downloaded and Uploaded data
  - **📥 Downloaded Data**: Displays data transferred to users (4 cards: Today, This Week, This Month, This Year)
  - **📤 Uploaded Data**: Displays data uploaded by users (4 cards: Today, This Week, This Month, This Year)
  - Beautiful gradient designs for each card to distinguish downloaded vs uploaded stats
  - Real-time tracking of both upload and download bandwidth usage

### Implementation Details
- **Database Methods**: 4 new methods added to track uploads separately
  - `GetBytesUploadedToday()` - Tracks uploads from start of day
  - `GetBytesUploadedThisWeek()` - Tracks uploads from start of week (Monday)
  - `GetBytesUploadedThisMonth()` - Tracks uploads from start of month
  - `GetBytesUploadedThisYear()` - Tracks uploads from start of year
- **Query Logic**: Upload stats query Files table by UploadDate (excludes soft-deleted files)
- **Download Stats**: Existing methods query DownloadLogs with Files join for accurate size tracking
- **Dashboard UI**: Two separate sections with 8 total cards (4 downloaded + 4 uploaded)

### Benefits
- Better visibility into upload vs download bandwidth consumption
- Helps identify patterns in user behavior
- Useful for capacity planning and quota management
- Separate tracking enables more granular analytics

---

## [3.2-beta2] - 2025-11-12 🔑 Password Management Update

### New Features
- **🔑 Self-Service Password Change**: Users and admins can now change their own passwords
  - Accessible from `/settings` page under "Security Settings"
  - Requires current password verification for security
  - Minimum 8 characters for new password
  - Client-side and server-side validation
  - Must be different from current password
  - Instant feedback on success or errors

### Implementation Details
- **Route**: `/change-password` (POST, requires authentication)
- **Handler**: `handleChangePassword` in `handlers_user_settings.go`
- **Database**: New `UpdateUserPassword` method in `database/users.go`
- **Security**: Current password must be verified before change
- **Validation**:
  - All fields required
  - Min 8 characters
  - New password ≠ current password
  - Passwords must match (confirmation)

### User Experience
- Modal dialog for password change
- Clear error messages for validation failures
- Success message with auto-close (2 seconds)
- Accessible to both users and admins from Settings page

---

## [3.2-beta1] - 2025-11-12 🔐 Two-Factor Authentication (2FA) Beta Release

### New Features
- **🔐 TOTP Two-Factor Authentication (2FA)**: Enterprise-grade security for user and admin accounts
  - TOTP support using authenticator apps (Google Authenticator, Authy, Microsoft Authenticator, etc.)
  - QR code generation for easy setup - just scan and verify
  - 10 backup codes per user for account recovery
  - One-time use backup codes (automatically removed after use)
  - User-friendly settings page at `/settings` for managing 2FA
  - Enable/disable 2FA with password confirmation for security
  - Regenerate backup codes with tracking of remaining codes
  - Secure login flow with 2FA verification page
  - 5-minute session timeout for 2FA setup and verification
  - **Note**: 2FA is available for Users and Admins only (not download accounts)

### Implementation Details
- **Database**: New columns added to Users table (TOTPSecret, TOTPEnabled, BackupCodes)
- **Security**: bcrypt hashing for backup codes, HttpOnly cookies, strict SameSite policy
- **Libraries**: pquerna/otp for TOTP generation, skip2/go-qrcode for QR code images
- **Migration**: Automatic database migration on startup (no manual steps required)
- **Login Flow**: Seamlessly redirects to 2FA verification when enabled
- **Time Skew**: Accepts codes within ±30 seconds window for reliability

### New Routes
- `/settings` - User settings page with 2FA controls
- `/2fa/setup` - Generate QR code and backup codes
- `/2fa/enable` - Verify TOTP code and activate 2FA
- `/2fa/disable` - Disable 2FA (requires password)
- `/2fa/verify` - TOTP verification during login
- `/2fa/regenerate-backup-codes` - Create new backup codes

### Security Features
- HttpOnly cookies prevent XSS attacks
- Backup codes are bcrypt-hashed (cost 12)
- One-time use backup codes enhance security
- Password required to disable 2FA
- Time-limited setup sessions (5 minutes)
- TOTP time skew tolerance (±1 period = 30 seconds)

### User Experience
- Clean, modern UI for 2FA management
- Visual status badges (ENABLED/DISABLED)
- Inline QR code display for easy scanning
- Backup codes shown with warning to save them
- Counter for remaining backup codes
- Automatic form submission when 6-digit code is entered
- Alternative backup code input option

### Known Limitations (Beta)
- Download accounts do not support 2FA (by design - simpler auth flow)
- No email notifications for 2FA events yet (planned for stable release)
- Rate limiting not implemented yet (planned for stable release)

---

## [3.1.2] - 2025-11-11 🔧 Critical Upload Timeout Fix

### Bug Fixes
- **Critical Upload Timeout Fix**: Fixed 60-second timeout causing large file uploads to fail at ~60%
  - Changed `ReadTimeout` to `ReadHeaderTimeout` - timeout now only applies to headers, not upload body
  - This allows uploads to take the full 8 hours as intended (not just 60 seconds)
  - Fixes "Upload Failed - Network Error" for files taking longer than 60 seconds to upload
  - Critical fix for 1TB+ file uploads or slower network connections

### Technical Details
- ReadHeaderTimeout: 60 seconds - for request headers only (not body)
- WriteTimeout: 8 hours - full time available for uploads
- No more 60-second upload body timeout

## [3.1.1] - 2025-11-11 🔧 Critical Bug Fix

### Bug Fixes
- **Upload Timeout Issue**: Fixed network error when uploading large files (500MB+)
  - Increased WriteTimeout from 15 seconds to 8 hours for very large file uploads on slow connections
  - Increased ReadTimeout from 15 to 60 seconds for better header processing
  - This allows users to upload multi-gigabyte files even on slow internet connections
  - Server now supports uploads taking up to 8 hours to complete
- **Maximum File Size Limit**: Increased from 5GB to 150GB
  - UI now reflects the new 150GB maximum file size limit
  - Suitable for large video files, backups, and forensic data

### Technical Details
- WriteTimeout: 8 hours (28,800 seconds) - for large file transfers
- ReadTimeout: 60 seconds - for request header processing
- IdleTimeout: 120 seconds - for keep-alive connections
- Maximum file size: 150 GB

---

## [3.1.0] - 2025-11-11 🎊 Dashboard Enhancement Release

### New Features
- **Enhanced Admin Dashboard**: Comprehensive statistics and metrics for better oversight
  - **Data Transfer Analytics**: Track bandwidth usage with granular time periods
    - Total bytes sent today, this week, this month, and this year
    - Human-readable size formatting (MB, GB, TB)
  - **User Growth Metrics**: Monitor user base expansion
    - Users added this month (regular + download accounts)
    - Users removed this month
    - Monthly growth percentage calculation
  - **Fun Facts Section**: Engaging statistics about system usage
    - Most downloaded file with download count
  - **Improved Layout**: Quick Actions menu moved to top for better navigation
  - **Visual Enhancements**: Gradient backgrounds and colored borders for statistics cards
- **Discrete Branding**: Professional footer with Manvarg attribution and GitHub link

### Technical Improvements
- 7 new database query functions for real-time statistics
- Optimized SQL queries with proper JOINs and aggregations
- Time-based calculations for day/week/month/year periods
- Consistent DeletedAt handling across all queries

### Bug Fixes
- Fixed byte calculation to use SizeBytes field instead of formatted Size string
- Corrected Most Downloaded File query to properly filter deleted files

### Community
- We welcome ideas and suggestions for dashboard improvements! Please open an issue on GitHub.

---

## [3.0.1] - 2025-11-11

### Bug Fixes
- **File Edit Form Submission**: Fixed "Missing file_id" error when editing file settings
  - Added multipart/form-data parsing support for JavaScript FormData API
  - Client-side form data is now correctly parsed by the server
  - Added validation to prevent empty fileId submission

---

## [3.0.0] - 2025-11-11 🎉 GOLD RELEASE

### New Features
- **Password Reset System**: Complete "Forgot Password" functionality for all user types (admin, regular users, download accounts)
  - Secure token-based reset links with 1-hour expiration
  - Humorous yet professional email notifications
  - Password visibility toggle (hold to view)
  - Dual-field password confirmation with validation
- **Server Restart Control**: Admin can restart server from Settings interface
  - Red warning-styled button with confirmation dialog
  - Graceful shutdown support
  - Compatible with process managers (systemd, supervisor)
- **Download Account Auto-Login**: New users are automatically logged into their dashboard after registration
- **Account Settings Navigation**: Fixed cancel button to return to dashboard instead of logging out

### Improvements
- Enhanced security with one-time use reset tokens
- Professional email templates with branding
- Mobile-responsive password reset pages
- Comprehensive audit logging for all security events
- Improved user experience across all account types

### Production Readiness
- All core features tested and stable
- GDPR-compliant data handling
- Secure password hashing with bcrypt
- Complete audit trail for compliance
- Professional documentation

---

## [3.0.0-rc2] - 2025-11-11

### New Features
- **Download Account Auto-Login**: New users creating accounts during file download are automatically logged in and redirected to their dashboard
- **Account Settings Cancel Button**: Fixed cancel button to redirect to dashboard instead of logging out

### Improvements
- Enhanced user experience for first-time download account creation
- Better session management for download accounts
- Improved redirect flow after account creation

### Planned for v3.1
- Password reset functionality for all user types (admin, users, download accounts)
- Two-factor authentication via email or authenticator app

---

## [3.0.0-rc1] - 2025-01-11

### New Features
- **Email Integration** (Server-side untested)
  - Optional email field when uploading files - sends download link to recipient
  - Optional email field in Upload Requests - sends upload invitation to recipient
  - Professional HTML and plain text email templates
  - Email history tracking in database (EmailLogs table)
  - Configurable SMTP/Brevo email providers
  - Test email functionality in admin settings

- **Upload Request System**
  - Create shareable upload links that allow others to upload files to you
  - Set maximum file size limits per request
  - 24-hour expiration for security (shows "expired" message for 10 days before deletion)
  - Auto-cleanup of expired requests

- **Enhanced Download Tracking**
  - Download history per file with timestamps
  - IP address logging for accountability
  - Email addresses captured for authenticated downloads

### Improvements
- Fixed duplicate modal code causing email field to be hidden in Upload Requests
- Removed redundant JavaScript implementations
- Updated version display consistency (RC1)
- Improved form validation and error handling
- Better mobile responsiveness

### Security
- File request links expire after 24 hours
- Automatic cleanup of expired file requests
- Enhanced password protection for shared files

### Technical
- Async email sending using goroutines
- Email provider abstraction layer
- Database schema updates for email logging
- Cleanup schedulers for expired files and requests

### Known Issues / Notes
- **IMPORTANT**: Email server functionality is untested - SMTP/Brevo configuration needs verification in production environment
- File request upload notification to requester - not yet implemented (planned for v3.1)

## [3.0.0-beta.5] - 2025-01-11
- Fixed modal duplication bug
- Version number consistency fixes

## [3.0.0-beta.4] - 2025-01-10
- Initial email functionality
- User management improvements

## [3.0.0-beta.3] - 2025-01-09
- Upload request feature
- Download tracking

## Earlier Versions
See git history for complete changelog.
