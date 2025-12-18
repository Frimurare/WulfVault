# Changelog

All notable changes to WulfVault will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [6.2.0] - BloodMoon 🌙 - 2025-12-18

### Added
- **Duplicate Files Detection System**: Comprehensive duplicate file management across admin interface
  - **Admin Dashboard Widget**: New "Duplicate Files" section at bottom of dashboard
    - Shows count of duplicate groups and total duplicate files
    - Lists files with identical name AND size combinations
    - Displays file IDs for easy identification
    - Orange color scheme for visual distinction
  - **Dedicated Duplicate Files View**: New page accessible from Files menu
    - Full pagination support (10, 25, 50, 100, 200 files per page)
    - Shows all duplicate files with complete metadata
    - Displays file descriptions/notes for each duplicate
    - Color-coded badges: 🔍 DUPLICATE (orange), Active/Expired status, Auth status
    - Action buttons: View History, Copy Link, Delete File
    - Statistics bar showing: Duplicate Groups, Total Duplicate Files, Currently Showing
    - Mobile-responsive layout with stacked buttons on small screens
  - **Smart Duplicate Detection**:
    - Matches files by exact filename AND exact size in bytes
    - Automatically skips files pending deletion
    - Efficient in-memory grouping algorithm
    - Real-time detection on page load
  - **Navigation Menu**: Added "Duplicate Files" option in Files dropdown menu
    - Located between "All Files" and "Trash" for logical workflow
    - Accessible at `/admin/duplicates`

### Technical
- Modified `internal/server/handlers_admin.go`:
  - Added `DuplicateFile` and `DuplicateFileDetail` structs
  - Added `findDuplicateFiles()` for dashboard widget
  - Added `findDuplicateFilesDetailed()` for dedicated page
  - Added `handleAdminDuplicates()` handler with pagination support
  - Added `renderAdminDuplicates()` with full UI rendering
  - Added `selected()` helper function for dropdown options
- Modified `internal/server/server.go`:
  - Added `/admin/duplicates` route with admin authentication
- Modified `internal/server/header.go`:
  - Added "Duplicate Files" link to Files dropdown menu
- Modified `cmd/server/main.go`:
  - Updated version to 6.2.0 BloodMoon 🌙

### User Experience
- Administrators can now easily identify and manage duplicate files
- Visual orange highlighting makes duplicates stand out
- Full file information (notes, owner, downloads, expiry) helps decide which duplicate to keep
- Pagination prevents performance issues with large numbers of duplicates
- Dashboard widget provides quick overview of duplicate file situation

## [6.1.9] - BloodMoon 🌙 - 2025-12-14

### Added
- **Comprehensive REST API Documentation**: Complete documentation for all REST API endpoints in `docs/API.md`
  - **Audit & Logging API**: Documented audit logs, server logs, and system monitor endpoints
    - GET `/api/v1/admin/audit-logs` - Retrieve audit logs with pagination and filtering
    - GET `/api/v1/admin/audit-logs/export` - Export audit logs to CSV
    - GET `/api/v1/admin/server-logs` - Retrieve server logs with line limits
    - GET `/api/v1/admin/server-logs/export` - Export server logs
    - GET `/api/v1/admin/sysmonitor-logs` - Retrieve system monitor logs
  - **GDPR Compliance API**: Documented user data export endpoint
    - GET `/api/v1/user/export-data` - Export user's personal data (GDPR Right to Data Portability)
  - **Pagination Support**: Documented query parameters for paginated endpoints
    - Query parameters: `page`, `per_page`, `sort_by`, `sort_order`
    - Examples with curl commands and response formats
  - **File Comments/Descriptions**: Documented `comment` field in file API responses
- **API Test Report**: Created comprehensive `API_TEST_REPORT.md` documenting REST API testing results
  - All major endpoints tested and verified (Authentication, Users, Files, Teams, Admin Stats)
  - API Health Score: A (95%)
  - Status: APPROVED FOR PRODUCTION USE ✅

### Changed
- **API Documentation Version**: Updated `docs/API.md` from v4.7.4 to v6.1.9
- **Documentation Cleanup**: Removed all legacy version markers from documentation files
  - Removed references to v4.5.x, v4.6.x, v4.7.x versions
  - Removed obsolete codenames (Gold, Champagne) from feature descriptions
  - Updated all version references to v6.1.9 BloodMoon 🌙

### Technical
- Modified `docs/API.md`: Added 6 new sections with 200+ lines of endpoint documentation
- Modified `cmd/server/main.go`: Updated version to 6.1.9 BloodMoon 🌙
- Modified `README.md`, `DOCKER_README.md`, `USER_GUIDE.md`, `GDPR_COMPLIANCE_SUMMARY.md`: Updated to v6.1.9 BloodMoon 🌙
- Created `API_TEST_REPORT.md`: Comprehensive REST API testing documentation

## [6.1.8] - BloodMoon 🌙 - 2025-12-12

### Added
- **Advanced Pagination System**: Major upgrade to file list management across the application
  - **My Files Dashboard**:
    - File counter showing "Showing X of Y files" (updates dynamically based on filters and search)
    - Configurable items per page: 5, 25, 50, 100, 200, 250 files (default: 25)
    - Previous/Next page navigation with visual feedback
    - Page indicator showing current page and total pages
    - Fully integrated with existing filters (All Files, My Files, Team Files)
    - Works seamlessly with team filtering and search functionality
  - **Team Shared Files**:
    - Same pagination controls as My Files
    - File counter with real-time updates
    - Integrates with file search and sorting features
  - **Technical Implementation**:
    - Dual-attribute filtering system (`data-filter-hidden` and `data-search-hidden`)
    - Separate state management for tab filters, team filters, search, and pagination
    - Efficient DOM manipulation with proper state isolation
    - No page reload required - all updates happen client-side

### Fixed
- **Pagination Logic**: Fixed multiple issues in initial pagination implementation
  - Corrected visible item counting that was causing incorrect totals
  - Fixed page navigation that wasn't working when changing items per page
  - Resolved filter state conflicts between search and tab/team filters
  - Fixed pagination not updating correctly after filter changes

### Technical
- Modified `internal/server/handlers_user.go`: Added complete pagination system to user dashboard
- Modified `internal/server/handlers_teams.go`: Added pagination to team files view
- Modified `cmd/server/main.go`: Updated version to 6.1.8

## [6.1.7] - BloodMoon 🌙 - 2025-12-12

### Fixed
- **Double Login Bug**: Fixed critical issue where users had to log in twice before accessing the system
  - Root cause: `CreateSession()` was not updating `Users.LastOnline` timestamp
  - First login would create valid session but fail inactivity check immediately (LastOnline was 30+ minutes old)
  - Second login would succeed because LastOnline was updated as side effect
  - Now properly updates `LastOnline` when creating session in `internal/auth/auth.go`
  - Ensures users can access dashboard on first login attempt

### Added
- **Team Files Enhancements**:
  - File descriptions/comments now visible in team files view
  - Added search field to filter team files by filename, owner, or description
  - Search updates in real-time as you type
  - Better organization and discoverability of shared team files

### Changed
- **Code Cleanup**: Removed temporary debug logging added during troubleshooting
  - Removed debug statements from `handlers_auth.go`, `server.go`, `handlers_user.go`, `handlers_admin.go`
  - Kept essential audit logging for security and monitoring

### Technical
- Modified `internal/auth/auth.go`: Added `LastOnline` update in `CreateSession()` function
- Modified `internal/server/handlers_teams.go`: Added file description display and search functionality
- Modified `cmd/server/main.go`: Updated version to 6.1.7

## [6.1.6] - BloodMoon 🌙 - 2025-12-11

### Fixed
- **Double Login Issue**: Fixed issue where users had to log in twice
  - Removed SameSite cookie attribute for HTTP connections
  - Session cookies now work correctly on first login attempt
  - Affects both regular login and 2FA login flows

- **Delete Button Styling**: Fixed grey delete buttons in Admin Files view
  - Delete buttons now properly display in red (#dc3545)
  - Added missing `.btn-danger` CSS definition
  - Consistent styling across all file management views

- **File List Layout**: Fixed button wrapping with long file notes
  - Action buttons (History, Copy, Delete) now stay on same row
  - Long file descriptions/notes no longer push buttons to next line
  - Added `flex-shrink: 0` and `min-width: 340px` to `.file-actions`
  - Mobile-responsive: buttons stack vertically on small screens (<768px)

### Added
- **"Keep Me Logged In" Enhancement**: Inactivity timeout now respects "Remember Me" sessions
  - Sessions with >2 days validity exempt from 10-minute inactivity timeout
  - 30-day sessions (Remember Me checked) won't auto-logout after 10 minutes
  - New `IsLongSession()` function to detect long-duration sessions
  - Only regular 24-hour sessions subject to inactivity timeout

- **Hourly Chunk Cleanup**: Automated cleanup of orphaned upload chunks
  - Runs every hour to remove abandoned chunks older than 2 hours
  - Reduces disk space usage from failed/interrupted uploads
  - Complements existing server startup cleanup

### Changed
- **Cache Busting**: Updated dashboard.js version to 6.1.6
  - Forces browser to reload latest JavaScript with all fixes

### Technical
- Modified `internal/server/handlers_auth.go`: Removed SameSite attribute from session cookies
- Modified `internal/server/handlers_2fa.go`: Removed SameSite attribute for 2FA session cookies
- Modified `internal/server/handlers_admin.go`: Added `.btn-danger` CSS, fixed `.file-actions` layout
- Modified `internal/server/server.go`: Added `IsLongSession` check in `requireAuth` and `requireAdmin`
- Modified `internal/auth/auth.go`: Added `IsLongSession()` function
- Modified `cmd/server/main.go`: Added hourly chunk cleanup scheduler, updated version to 6.1.6
- Modified `internal/server/handlers_user.go`: Added inline red styling to delete button, updated cache busting
- Modified `.gitignore`: Added FUTURE_FEATURES.md to exclusion list

## [6.1.5] - BloodMoon 🌙 - 2025-12-11

### Added
- **Retry Count Enhancement**: Increased upload retry attempts from 30 to 50
  - Provides ~7.5 minutes total retry time (up from ~5 minutes)
  - Better handling of router restarts and network interruptions
  - Updated UI messages to reflect new retry count

### Changed
- **Version Update**: Bumped to 6.1.5 for retry enhancement release

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
