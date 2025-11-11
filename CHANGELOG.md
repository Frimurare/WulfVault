# Changelog

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
