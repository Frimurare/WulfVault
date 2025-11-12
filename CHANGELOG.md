# Changelog

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
