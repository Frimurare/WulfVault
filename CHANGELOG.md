# Changelog

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
