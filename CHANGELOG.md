# Changelog

All notable changes to WulfVault will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [6.1.4] - BloodMoon 🌙 - 2025-12-10

### Changed
- **Upload Retry Timeout Extended**: Increased from 30 to 50 retry attempts
  - Total retry time increased from ~3 minutes to ~7.5 minutes
  - Better tolerance for router restarts and network interruptions
  - Exponential backoff with 10-second maximum delay per retry
  - Updated all user-facing messages and documentation

### Technical
- Modified `web/static/js/dashboard.js`: MAX_RETRIES 30→50
- Updated retry messaging in upload UI

## [6.1.3] - BloodMoon 🌙 - 2025-12-10

### Changed
- **Complete Email Translation**: All remaining emails translated from Swedish to English
  - Download notification emails (when files are downloaded)
  - Splash link sharing emails ("Someone Shared a File with You")
  - Account deletion confirmation emails (GDPR compliance)
  - Helper function messages (e.g., getRandomQuote)

### Technical
- Updated `internal/email/templates.go`: All email templates now in English

## [6.1.2] - BloodMoon 🌙 - 2025-12-10

### Changed
- **Password Reset Translation**: Complete translation from Swedish to English
  - Password reset request page
  - Password reset success page
  - Password reset email template
  - All user-facing text in password recovery flow

### Technical
- Modified `internal/email/templates.go`: SendPasswordResetEmail function
- Updated `internal/server/handlers_password_reset.go`: All render functions

## [6.1.1] - BloodMoon 🌙 - 2025-12-10

### Fixed
- **Chunk Size Display**: Fixed dashboard showing incorrect 5MB chunk size
  - Added cache busting parameter to dashboard.js (?v=6.1.1)
  - Ensures browsers load updated 25MB chunk size setting

### Technical
- Modified `internal/server/handlers_dashboard.go`: Added version query parameter

## [6.1.0] - BloodMoon 🌙 - 2025-12-10

### Added
- **SysMonitor Logs**: New detailed system monitoring log system
  - Separate log file (`data/sysmonitor.log`) for detailed chunk upload tracking
  - 10MB maximum size with automatic rotation
  - Accessible via Admin Panel: Server > SysMonitor Logs
  - Tracks every chunk upload with progress percentage
  - Auto-refresh every 5 seconds in UI
  - Search functionality for filtering logs

### Changed
- **Server Logs Enhancement**: Upload events now visible in Server Logs
  - Upload start logs show: filename, size, upload ID, user, email, IP address
  - Upload complete logs show: filename, size, duration, average speed, user, email, IP
  - Upload progress logs (every 100 chunks) show: filename, progress percentage
  - Upload abandoned logs show: filename, progress, inactive time
  - System events now display full log message in UI instead of empty columns
- **Chunk Upload Size**: Increased from 5MB to 25MB per chunk
  - Improved upload performance for stable network connections
  - Reduced HTTP request overhead (80% fewer requests)
  - Better throughput for large file transfers

### Technical
- Added `internal/server/sysmonitor.go` for dedicated monitoring logs
- Modified `internal/server/handlers_chunked_upload.go` to log detailed chunk progress
- Updated `internal/server/handlers_server_logs.go` parser to include upload events
- Enhanced `internal/server/middleware.go` to route chunk logs to SysMonitor
- Created `internal/server/handlers_sysmonitor_logs.go` for admin UI
- Updated Server Logs UI to display full system event messages

## [6.0.2] - BloodMoon 🌙 - 2025-12-09

### Fixed
- Improved UI spacing for action buttons across admin pages
  - Added 40px padding-top to container elements
  - Better visual separation between navigation header and page content
  - Affects Users, Teams, Trash, and Download Accounts pages
  - Creates more breathing room for "Empty All Trash", "+ Create User", "+ Create Team", and "+ Create Download Account" buttons

### Changed
- Removed claude.md from repository (moved to local development environment)

## [6.0.1] - BloodMoon 🌙 - 2025-12-07

### Added
- "Keep Me Logged In" feature for persistent login sessions
- Enhanced user convenience with remember-me functionality

## [6.0.0] - BloodMoon 🌙 - 2025-11-18

### Added
- Verified uploads and history tracking
- Major feature updates and improvements

### Breaking Changes
- Updated history tracking system

## Previous Versions

For historical versions prior to 6.0.0, please see git commit history.
