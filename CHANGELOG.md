# Changelog

All notable changes to WulfVault will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [6.2.9] - BloodMoon 🌙 - 2026-05-11

### Fixed

- **SMTP/email-provider test fails on fresh installs with
  `SQL logic error: no such column: MailgunDomain`** (#30)

  The Mailgun-related columns `MailgunDomain` and `MailgunRegion`
  are referenced in `internal/email/email.go` and
  `internal/server/handlers_email.go`, but were missing from the
  `EmailProviderConfig` table in fresh installs — they weren't in
  `internal/database/schema.go`'s `CREATE TABLE` statement and no
  `addColumnIfNotExists` migration added them on upgrade.

  Symptoms: on a fresh Docker install, configuring any email
  provider (SMTP, Mailgun, etc.) and clicking *Test connection*
  returned `400 Bad Request` with the log line
  `GetActiveProvider scan error: SQL logic error: no such column:
  MailgunDomain (1)`.

  Note: `MailgunKey` is **not** a separate column — the Mailgun
  API key is stored in the generic `ApiKeyEncrypted` field that
  the rest of the codebase already uses for every provider.

  Fix:
  - `internal/database/schema.go` — `MailgunDomain TEXT` and
    `MailgunRegion TEXT DEFAULT 'us'` added to the
    `EmailProviderConfig` `CREATE TABLE`.
  - `internal/database/migrations.go` — two
    `addColumnIfNotExists` calls so existing installs get the
    columns added on next start without losing data.

  Thanks to @axedoardo for the detailed report and the manual
  workaround.

## [6.2.8] - BloodMoon 🌙 - 2026-05-11

### Fixed

- **2FA verify bounces user back to /login on first attempt**

  The TOTP verify form auto-submitted 100 ms after the 6th digit
  was entered (`input` event) **and** also on button click /
  Enter. Pasting a 6-digit code from an authenticator app or
  pressing Enter quickly produced two near-simultaneous POSTs to
  `/2fa/verify`. The first cleared the `totp_pending` cookie,
  created a session, and `303`'d to `/admin`; the second arrived
  without `totp_pending` and `303`'d to `/login`. The browser
  followed the second redirect — and the user ended up on the
  login page even though a valid session had been created in the
  database.

  This is the bug behind *"I have to log in 2-3 times before I
  get in"*.

  Fix:
  - **Client** — a submit-once guard locks both `totp-form` and
    `backup-form` after the first submit and disables the verify
    button, so neither the timer-driven submit nor a manual click
    can race.
  - **Server** — if `/2fa/verify` is hit without `totp_pending`
    but the user already holds a valid session cookie (concurrent
    winner case), redirect to `/admin` or `/dashboard` instead of
    `/login`.

## [6.2.7] - BloodMoon 🌙 - 2026-04-11

### Added

- **`GET /api/whoami` — dedicated session verification endpoint**
  New JSON endpoint for API clients and integrations (notably
  Prudencia Evidence Courier) to reliably verify that a session
  cookie is valid, without any side effects.

  Before this endpoint existed, clients had to probe `/login` or
  `/dashboard` and inspect HTML response bodies to determine auth
  state — unreliable, slow, and prone to false positives when the
  browser had cookie handling quirks.

  Response format:
  ```json
  200 OK
  {
    "authenticated": true,
    "id": 123,
    "email": "user@example.com",
    "name": "User Name",
    "role": "user",
    "storage_used_mb": 42,
    "storage_quota_mb": 1000,
    "server_version": "6.2.7 BloodMoon 🌙",
    "two_factor_enabled": false
  }

  401 Unauthorized
  { "authenticated": false, "error": "Not authenticated" }
  ```

  The `Cache-Control: no-store` header is set so clients always get
  a fresh auth check.

### Technical

- New handler in `internal/server/handlers_user.go`: `handleWhoAmI()`
- New route registered in `internal/server/server.go`: `/api/whoami`
  under `requireAuth` middleware
- Version bumped to `6.2.7 BloodMoon 🌙` in `cmd/server/main.go`
  (single source of truth)

## [6.2.6] - BloodMoon 🌙 - 2026-04-10

### 🚨 CRITICAL BUGFIX — Web UI uploads dropped notification emails for 4 months

For ~4 months (since v6.0.0 in December 2025), files uploaded via the WulfVault **web UI dashboard** never triggered notification emails to recipients. The legacy `POST /upload` endpoint (used by API clients like Prudencia Evidence Courier) continued to work correctly, which is why the bug remained hidden.

**Three separate issues conspired to cause this:**

1. **Frontend** (`web/static/js/dashboard.js`): The `metadata` object passed to chunked upload was missing the `send_to_email` field. Even though the upload form had `<input name="send_to_email">`, the JavaScript never extracted its value into the metadata sent to `/api/upload/init`. The field was lost before reaching the server.

2. **Backend** (`internal/server/handlers_chunked_upload.go`): The chunked upload handler had no email-sending code at all. When the chunked upload system was created in v6.0.0, the email block from `handlers_files.go` was never copied over. Even if the frontend had sent `send_to_email`, the server would have ignored it.

3. **SMTP provider initialization** (`internal/email/email.go`): The SMTP provider in `GetActiveProvider()` required a non-empty password. This blocked any test/dev SMTP server like MailHog (which accepts mail without auth) and could break production SMTP servers configured for IP-based relay.

This was NOT a regression from v6.2.4 or v6.2.5. Both of those releases touched email code (sender info, splash link signature, expiration reminders) but did not modify the chunked upload handler or the dashboard.js metadata. The bug existed from the day chunked upload was introduced.

**Impact:** All file sharing via the web UI in v6.0.0 through v6.2.5 silently dropped notification emails. Recipients never received the file link via email — only the uploader saw the success message in the browser. Affected ALL email providers (SMTP, Resend, Brevo, SendGrid, Mailgun) because none of them were ever called from the chunked upload code path.

### Fixed

- **`web/static/js/dashboard.js`**: Added `send_to_email: formData.get('send_to_email') || ''` to the metadata object sent to `/api/upload/init`. The form field is now correctly forwarded to the backend.
- **`internal/server/handlers_chunked_upload.go`**: Added complete email-sending block after upload completion. Mirrors the implementation in `handlers_files.go`. Reads `send_to_email` from upload metadata, builds the same HTML/text email with sender info, comment, file size, expiration, and download links, then calls the active email provider asynchronously. New log marker `(chunked upload)` for traceability. Imports added: `html` for HTML escaping, `internal/email` for `GetActiveProvider()`.
- **`internal/email/email.go`**: SMTP provider no longer requires a password to initialize. Empty password is allowed for dev/test servers (like MailHog) and IP-relay production setups. The check is now: SMTP host must be set; password is optional. The actual SMTP connection (in `smtp.go`) already handled passwordless auth correctly via `sendPlainSMTP()` when TLS is disabled.

### Verified end-to-end on lab CT 122

Tested with both providers:

**With MailHog (SMTP, no auth):**
```
✅ UPLOAD COMPLETED: 'mh-test2.txt'
GetActiveProvider found: provider=smtp
SMTP provider loaded (host=127.0.0.1:1025, user=, tls=false, hasPassword=false)
📧 Sending email via SMTP to recipient@example.com through 127.0.0.1:1025
✓ Email sent successfully via plain SMTP
File download link email sent to recipient@example.com (chunked upload)
```

MailHog API confirmed all 3 test mails received (chunked + legacy + Swedish chars).

**With Resend (after restore):**
```
✅ UPLOAD COMPLETED: 'final.txt'
📧 Sending email via Resend to uffe.holmstrom@gmail.com
📩 Resend Response Status: 200 OK
✓ Email sent successfully via Resend
File download link email sent to uffe.holmstrom@gmail.com (chunked upload)
```

### Migration Notes

- No database changes
- No config changes
- No breaking API changes
- Drop-in binary replacement (plus updated `web/static/js/dashboard.js`)
- **Recommended: rebuild and redeploy immediately** — every WulfVault install since v6.0.0 has this bug

## [6.2.5] - BloodMoon 🌙 - 2026-04-08

### Added
- **Expiration reminder emails**: Automatic reminders sent to file owners when shared files are about to expire
  - Halfway reminder (~2-3 days before expiration for 5-day files)
  - Urgent reminder (1 day before expiration) with red urgency styling
  - Includes file name, size, download count, and download link
  - Runs every 6 hours as part of the cleanup scheduler
- **Database**: `GetFilesExpiringInDays(days)` query for finding soon-to-expire files

### Technical
- New file: `internal/cleanup/reminders.go` — reminder email logic
- Modified `internal/cleanup/cleanup.go` — scheduler now accepts optional `serverURL` for reminder links
- Modified `internal/database/files.go` — added `GetFilesExpiringInDays` query
- Modified `cmd/server/main.go` — passes `ServerURL` config to cleanup scheduler

## [6.2.4] - BloodMoon 🌙 - 2026-04-08

### Changed
- **Email notifications now include sender information**: Share emails show who sent the file (sender email), making it clear who the file comes from. Critical for police/security evidence sharing where anonymous emails get discarded.
- **File comments included in share emails**: The comment/description text is now shown in the email body, providing context about the shared file.
- **Download confirmation includes downloader identity**: When require-auth is enabled, the download notification email to the sender now includes who downloaded the file.
- **Email language changed to English**: Share emails now use English (was Swedish) for international compatibility.
- **Updated email footer**: "WulfVault Secure File Transfer" branding.

### Technical
- Modified `internal/email/templates.go`: Added `senderEmail` parameter to `GenerateSplashLinkHTML/Text`, new `getSenderHTML` helper
- Modified `internal/email/email.go`: Updated `EmailProvider` interface + `SendSplashLinkEmail` signature
- Modified all 5 email providers (smtp, resend, sendgrid, brevo, mailgun): Updated `SendSplashLinkEmail` signatures
- Modified `internal/server/handlers_files.go`: Upload share email now includes sender info + file comment
- Modified `internal/server/handlers_email.go`: Splash link email now includes sender info

## [6.2.3] - BloodMoon 🌙 - 2025-12-21

### Changed
- **Complete Removal of Legacy References**: Removed all references to legacy codebase
  - Removed all legacy attributions and acknowledgments from documentation
  - Updated startup message to reflect WulfVault as standalone enterprise platform
  - Cleaned up all code comments removing legacy references
  - Deleted old test files that referenced legacy imports
  - Updated NOTICE.md to focus purely on WulfVault features and licensing
  - WulfVault is now presented as a fully independent enterprise file sharing platform

### Technical
- Modified `cmd/server/main.go`:
  - Updated startup message: "Enterprise File Sharing | Self-Hosted | Open Source (AGPL-3.0)"
  - Updated version to 6.2.3 BloodMoon 🌙
- Modified `internal/database/schema.go`:
  - Removed legacy references from SQL schema comments
- Modified `internal/models/Authentication.go`:
  - Updated comments to reference WulfVault instead of legacy code
- Deleted legacy test files:
  - Removed `internal/models/*_test.go` files (6 files)
- Updated documentation files:
  - `README.md` - Removed legacy attribution sections
  - `DOCKER_README.md` - Removed acknowledgments section
  - `INSTALLATION.md` - Removed legacy references
  - `USER_GUIDE.md` - Removed legacy attributions
  - `NOTICE.md` - Complete rewrite focusing on WulfVault as independent platform

### User Experience
- Cleaner, more professional presentation of WulfVault as an independent enterprise platform
- Documentation now focuses entirely on WulfVault's unique features and capabilities
- No confusion about project origins or dependencies

## [6.2.2] - BloodMoon 🌙 - 2025-12-21

### Fixed
- **Chunked Upload Team Assignment**: Fixed critical bug where team sharing didn't work during chunked uploads
  - Files uploaded with team checkboxes selected were not being shared to those teams
  - Only affected chunked uploads (large files), regular uploads worked correctly
  - Root cause: Team IDs were not being passed in upload metadata to backend
  - JavaScript now correctly includes `team_ids` in metadata as comma-separated string
  - Backend now parses team IDs from metadata and shares files to selected teams
  - Team membership verification and proper logging now working for chunked uploads

### Technical
- Modified `web/static/js/dashboard.js`:
  - Added team IDs extraction from form data (`formData.getAll('team_ids[]')`)
  - Added `team_ids` to upload metadata as comma-separated string
- Modified `internal/server/handlers_chunked_upload.go`:
  - Added `strings` import for parsing team IDs
  - Added team assignment logic in `handleChunkedUploadComplete()`
  - Parses comma-separated team IDs from metadata
  - Verifies team membership before sharing
  - Calls `ShareFileToTeam()` for each valid team
  - Logs successful team shares
- Modified `cmd/server/main.go`:
  - Updated version to 6.2.2 BloodMoon 🌙

### User Experience
- Users can now reliably share uploaded files to teams regardless of file size
- Team sharing works consistently for both small files (regular upload) and large files (chunked upload)
- Proper logging ensures team shares are tracked in audit logs

## [6.2.1] - BloodMoon 🌙 - 2025-12-18

### Improved
- **Enhanced Search Functionality**: Search now includes file descriptions/notes across all file views
  - **My Files Dashboard**: Search box now searches in filename, extension, AND file description/note
  - **Admin All Files**: Search box now searches in filename, extension, username, AND file description/note
  - **Teams Shared Files**: Search box now searches in filename, owner, AND file description/note
  - More efficient search using data attributes instead of full text content
  - Helps users find files based on what the file contains, not just the filename

### Fixed
- **Teams Shared Files Delete**: Fixed "Failed to delete" error message appearing despite successful deletion
  - Delete button now correctly checks HTTP response status instead of expecting JSON `success` field
  - File deletion works correctly and shows proper success message

### Technical
- Modified `internal/server/handlers_user.go`:
  - Added `data-comment` attribute to file items in My Files view
  - Updated `searchAndSortFiles()` to include comment in search filter
- Modified `internal/server/handlers_admin.go`:
  - Added `data-comment` attribute to file items in All Files view
  - Updated `searchAndSortFiles()` to include comment in search filter
- Modified `internal/server/handlers_teams.go`:
  - Added `data-comment` attribute to file items in Teams Shared Files view
  - Updated `filterAndPaginate()` to search by comment instead of full text content
  - Fixed delete button response handling to check `res.ok` instead of `data.success`
- Modified `cmd/server/main.go`:
  - Updated version to 6.2.1 BloodMoon 🌙

### User Experience
- Users can now find files by searching for words in their file descriptions
- Example: Searching "invoice" will find all files with "invoice" in the description, even if filename is "document_2024.pdf"
- More intuitive and powerful file discovery across all file management views

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
    - **Upload timestamp** showing when each file was uploaded (format: "HH:MM on YYYY-MM-DD")
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

### Fixed
- **Duplicate Files Page**: Delete button now works correctly
  - Fixed endpoint from `/admin/files/delete` to `/file/delete`
  - Improved error handling with async/await pattern
  - Shows detailed error messages on delete failure
  - Enhanced confirmation dialog with "cannot be undone" warning

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
