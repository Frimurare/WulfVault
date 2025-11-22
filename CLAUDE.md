# WulfVault Development Notes

**Last Updated:** 2025-11-22
**Current Version:** v4.7.8 Shotgun
**Current Branch:** wolfface → main (ready for release)

---

## Project Overview

WulfVault is a self-hosted secure file sharing system written in Go with a web UI. It's designed for small to medium organizations needing secure file sharing without cloud dependencies.

**Repository:** https://github.com/Frimurare/WulfVault
**Server runs on:** http://localhost:8080 (local development)

---

## Current Development Session (2025-11-22)

### What We Accomplished Today

#### v4.7.8 Shotgun - Resend Email Provider (Recommended) + DNS Verification

**1. Resend Integration**
- **User Request:** Integrate Resend.com as recommended email provider
- **Why Recommended:** Built on AWS SES with industry-leading deliverability and inbox placement
- **Solution:**
  - Created Resend provider in `internal/email/resend.go` (154 lines)
  - Uses simple REST API with Bearer token authentication
  - Endpoint: `https://api.resend.com/emails`
  - Requires only: API key, from email/name (simpler than Mailgun)
  - Returns 200 OK on success with email ID
- **API Key:** `re_eJthnzJ8_4bcb9nTdLFJBcBRYZAUmAaxG`
- **Testing:** Successfully sent test email to `uffe.holmstrom@gmail.com`
- **Email ID:** `516c879c-3a12-4d98-ae8b-7d64c19627d0`

**2. DNS Domain Verification (wulfvault.se)**
- **Challenge:** Resend requires domain verification to send to anyone (not just test account)
- **Initial Errors:**
  - `403: The gmail.com domain is not verified` - trying to send from @gmail.com
  - `403: You can only send testing emails to your own email address` - test mode limitation
- **Solution Process:**
  1. Added `wulfvault.se` domain to Resend (EU region: eu-west-1)
  2. Received 4 DNS records from Resend (DKIM, SPF MX, SPF TXT, DMARC)
  3. **User confusion:** Initially added records under wrong sections (*, @, www) in Loopia
  4. **Guided user:** Explained Loopia DNS editor requires separate sections for each subdomain
  5. **Correct setup:** Added 4 records in their own sections:
     - `resend._domainkey` (TXT) - DKIM verification
     - `send` (MX + TXT) - SPF authorization
     - `_dmarc` (TXT) - DMARC policy
  6. **Wait time:** 10-15 minutes for DNS propagation
  7. **Verification:** Clicked "Verify" in Resend control panel
  8. **Success:** Domain verified! ✅
  9. **Update WulfVault:** Changed From Email from `onboarding@resend.dev` to `noreply@wulfvault.se`
  10. **Test:** Successfully sent emails to anyone! 🎉

**3. DNS Documentation Created**
- Created comprehensive `docs/LOOPIA.md` guide
- Synced with existing `resend_loopia_setup.md` from GitHub
- Documented common mistakes (adding to wrong sections, full domain names, etc.)
- Included success story with timeline (12:35-12:59, 24 minutes total)
- Added troubleshooting section for DNS propagation issues

**4. UI/UX Prioritization**
- **Resend marked as "(recommended)"** with green badge in UI
- **Tab order changed:** Resend (first), Brevo, Mailgun, SendGrid, SMTP
- **Helpful text:** "Recommended provider: Built on AWS SES with excellent deliverability"
- Resend configuration form identical to SendGrid/Brevo (API key only)

**5. Email Provider Ecosystem - Now 5 Providers**
1. **Resend (recommended)** - AWS SES-based, best deliverability, **VERIFIED & WORKING** ✅
2. **Brevo** - API-based (formerly Sendinblue) - user's account closed by Brevo
3. **Mailgun** - API-based with domain/region - requires sandbox activation
4. **SendGrid** - API-based, simple setup
5. **SMTP** - Plain SMTP with/without TLS (tested with MailHog)

**Files Modified:**
- `internal/email/resend.go` - New Resend provider (154 lines)
- `internal/email/email.go` - Added Resend case to GetActiveProvider()
- `internal/server/handlers_email.go` - Added Resend UI, endpoints, handlers
- `cmd/server/main.go` - Version 4.7.7 → 4.7.8
- `CHANGELOG.md` - Documented Resend support + DNS verification
- `README.md` - Updated email providers section with all 5 providers
- `docs/LOOPIA.md` - New comprehensive DNS verification guide
- `CLAUDE.md` - This file

**User Feedback:**
- "Wulfvault fungerar nu med resend API ... snyggt jobbat" (It works now with Resend API ... well done)
- "Jag tycker du gjort ett fantastiskt jobb idag, och Brevo kan dra åt skogen" (I think you did a fantastic job today, and Brevo can go to hell)

---

## Previous Session (2025-11-22 earlier)

#### v4.7.7 Shotgun - Mailgun & SendGrid Email Providers

**1. Mailgun Integration**
- **User Need:** Brevo closed user's API key, needed alternative email provider
- **Solution:**
  - Created complete Mailgun provider in `internal/email/mailgun.go`
  - Uses multipart form-data (not JSON) for API requests
  - Supports US and EU regions with different API endpoints
  - Requires: API key, domain, region, from email/name
  - Added database columns: `MailgunDomain`, `MailgunRegion`
  - Full UI with test connection before activation
- **Testing:** Configured with sandbox domain, waiting for user to authorize recipients
- **API Endpoint:** `https://api.mailgun.net/v3/{domain}/messages` (US) or `.eu.` (EU)

**2. SendGrid Integration**
- **User Need:** Wanted multiple email provider options for reliability
- **Solution:**
  - Created SendGrid provider in `internal/email/sendgrid.go`
  - Simpler than Mailgun - only needs API key
  - Uses Bearer token authentication with API v3
  - JSON-based API (similar to Brevo)
  - Requires: API key, from email/name
  - Accepts both 200 OK and 202 Accepted responses
- **API Endpoint:** `https://api.sendgrid.com/v3/mail/send`

**3. Email Provider Architecture**
- **Now supports 4 providers total:**
  1. Brevo - API-based (original)
  2. Mailgun - API-based with domain/region
  3. SendGrid - API-based (simplest)
  4. SMTP - Plain SMTP with/without TLS
- **All providers:**
  - Implement same `EmailProvider` interface
  - Support all 5 email functions (splash link, upload/download notifications, GDPR deletion, etc.)
  - Encrypted credentials (AES-256-GCM)
  - Test connection before activation
  - "Make Active" button to switch providers
- **UI:** Email Settings now has 4 tabs, consistent UX across all providers

**4. Database Schema Updates**
- Added columns to `EmailProviderConfig` table:
  - `MailgunDomain TEXT`
  - `MailgunRegion TEXT DEFAULT 'us'`
- No changes needed for SendGrid (uses existing `ApiKeyEncrypted` column)

**Files Modified:**
- `internal/email/mailgun.go` - New Mailgun provider
- `internal/email/sendgrid.go` - New SendGrid provider
- `internal/email/email.go` - Updated `GetActiveProvider()` for both new providers
- `internal/server/handlers_email.go` - Added UI and endpoints for both providers (1429 lines → ~2000 lines)
- `cmd/server/main.go` - Version 4.7.6 → 4.7.7 Shotgun
- `CHANGELOG.md` - Documented new providers
- `CLAUDE.md` - This file

**Testing Notes:**
- Mailgun requires account activation and authorized recipients for sandbox
- SendGrid ready to test with valid API key
- Both providers follow same pattern as Brevo for consistency

---

## Previous Session (2025-11-22 earlier)

#### v4.7.6 Galadriel - Email Provider Activation & Plain SMTP Support

**1. Email Provider Activation Controls**
- **Problem:** User had both Brevo and SMTP configured, but no way to choose which was active
- **Root Cause:** Saving settings was supposed to activate provider, but wasn't reliable
- **Solution:**
  - Created `/api/email/activate` endpoint in `internal/server/handlers_email.go`
  - Added red "🚀 Make Active" buttons for both Brevo and SMTP in UI
  - Buttons only show for configured but inactive providers
  - Deactivates all providers, then activates the selected one
  - Audit logging for activation events
  - Page auto-reloads to show updated status
- **Result:** Users can now configure multiple providers and explicitly choose which one to use

**2. Fixed SMTP Settings UI Bugs**
- **Problem 1:** SMTP settings disappeared after page refresh
- **Root Cause:** `getSMTPHost()`, `getSMTPPort()`, `getSMTPUsername()` returned empty strings (were TODOs)
- **Solution:** Implemented database queries in all three functions
- **Problem 2:** TLS checkbox always checked after refresh, even when unchecked and saved
- **Root Cause:** Checkbox had hardcoded `checked` attribute in HTML
- **Solution:** Made checkbox read from database dynamically using inline function
- **Problem 3:** Port reverting to 587 instead of saved value
- **Root Cause:** `getSMTPPort()` returned hardcoded "587" as default
- **Solution:** Return database value, only use 587 if no value exists
- **Result:** All SMTP settings now properly persist and display correctly

**3. Plain SMTP Support (for MailHog and test servers)**
- **Problem:** Even with TLS unchecked, got "unencrypted connection" error
- **Root Cause:** gomail library REQUIRES STARTTLS/TLS, refuses plain SMTP connections
- **Solution:**
  - Created custom `sendPlainSMTP()` function in `internal/email/smtp.go:82-151`
  - Uses Go's standard `net/smtp` library for plain SMTP
  - Modified `SendEmail()` to route to plain SMTP when `useTLS = false`
  - Full MIME multipart/alternative support (text + HTML parts)
  - Raw SMTP protocol: MAIL FROM, RCPT TO, DATA commands
- **Testing:** Successfully tested with MailHog at 192.168.86.142:1025
- **Result:** WulfVault now works with MailHog, test servers, and any SMTP without TLS

**4. MailHog Installation**
- Installed MailHog for testing: `go install github.com/mailhog/MailHog@latest`
- SMTP: localhost:1025, Web UI: localhost:8025
- Configured WulfVault database to use MailHog
- MailHog kept installed on container for future testing

**Files Modified:**
- `internal/email/smtp.go` - Added plain SMTP implementation
- `internal/server/handlers_email.go` - Activation endpoint, UI fixes
- `internal/server/server.go` - Added activation route
- `CHANGELOG.md` - Documented all changes
- `cmd/server/main.go` - Version already at 4.7.6

---

## Previous Development Session (2025-11-20)

### What We Accomplished Today

#### v4.7.5 Galadriel - SMTP Security Fix

**1. Wolf Favicon Fix**
- **Problem:** Wolf emoji favicon was added in v4.7.2 but never displayed in browser tabs
- **Root Cause:** Favicon HTML was in `<body>` but browsers only recognize it in `<head>`
- **Solution:** Created `getFaviconHTML()` helper in `header.go:13-16`, updated 31 locations across 13 handler files
- **Result:** Wolf emoji 🐺 now displays correctly in all browser tabs

**2. SMTP Security Fix (CRITICAL)**
- **Problem:** SMTP implementation had security vulnerability when TLS disabled
- **Root Cause:** Code set `InsecureSkipVerify=true` when `useTLS=false`, allowing Man-in-the-Middle attacks
- **Impact:** Attackers could intercept password resets, file sharing links, GDPR data
- **Solution:**
  - Removed dangerous `InsecureSkipVerify=true` setting in `internal/email/smtp.go:61-68`
  - Now safely delegates to gomail's default behavior when TLS disabled
  - Added comprehensive logging (8 new log statements with emojis)
  - Better error messages with host:port context
- **Files updated:**
  - `internal/email/smtp.go` - Security fix + logging
  - `CHANGELOG.md` - Documented security fix
- **Result:** SMTP now secure, works with Gmail, Outlook, SendGrid, Mailgun, custom servers

**3. Email System Architecture Verification**
- **Verified:** All 10 email functions use `EmailProvider` interface
- **Verified:** NO direct calls to Brevo - all via `GetActiveProvider()`
- **Verified:** Provider-switching works seamlessly between Brevo and SMTP
- **Functions tested:**
  1. SendSplashLinkEmail
  2. SendFileDownloadNotification
  3. SendFileUploadNotification
  4. SendAccountDeletionConfirmation
  5. SendWelcomeEmail
  6. SendTeamInvitationEmail
  7. SendPasswordResetEmail
  8-10. Provider-specific implementations (via interface)

---

## Branch Strategy

### main
- Production-ready code
- Currently at v4.7.3 Galadriel

### wolfface (NEW)
- Active development branch
- Created 2025-11-20
- Currently at v4.7.4 Galadriel
- Use this branch for future development
- Will merge back to main when ready

### Legacy branches
- v4.7.1-Galadriel
- v4.7.2-Galadriel

---

## Local Development Environment

### Current State
- **Local version:** v4.7.4 Galadriel (binary rebuilt)
- **Running on:** http://localhost:8080
- **Process ID:** Check with `pgrep -f wulfvault`
- **Branch:** wolfface
- **Data directory:** ./data
- **Upload directory:** ./uploads
- **Logs:** /tmp/wulfvault.log

### Server Management Commands

**Build:**
```bash
cd /home/ulf/WulfVault
go build -o wulfvault cmd/server/main.go
```

**Start server:**
```bash
./wulfvault
# Or in background with logging:
nohup ./wulfvault > /tmp/wulfvault.log 2>&1 &
```

**Stop server:**
```bash
pkill wulfvault
```

**Restart server:**
```bash
pkill wulfvault && sleep 1 && nohup ./wulfvault > /tmp/wulfvault.log 2>&1 &
```

**Check logs:**
```bash
tail -f /tmp/wulfvault.log
```

**Check version:**
```bash
grep "WulfVault File Sharing System" /tmp/wulfvault.log | tail -1
```

---

## Key Architecture Notes

### Directory Structure
- `cmd/server/main.go` - Entry point, version constant defined here
- `internal/server/` - HTTP handlers and server logic
  - `header.go` - Header HTML generation, favicon, navigation
  - `handlers_user.go` - User management, GDPR export/deletion
  - `handlers_gdpr.go` - GDPR-specific endpoints
  - `handlers_files.go` - File operations
  - `handlers_admin.go` - Admin panel
  - `handlers_*.go` - Various specialized handlers
- `internal/database/` - SQLite database operations
- `web/static/` - CSS, JS, images (if applicable)
- `gdpr-compliance/` - GDPR templates and procedures
- `docs/` - API documentation

### Important Technical Details
- **Database:** SQLite (NOT SQLCipher - encryption is NOT built-in)
- **Password hashing:** bcrypt
- **2FA:** TOTP
- **Sessions:** Secure cookie-based
- **File storage:** Local filesystem with configurable path
- **HTML Generation:** Server-side rendered HTML strings in Go handlers

### What's NOT Implemented (don't document as features)
- SQLCipher database encryption
- Rate limiting (documented as "not implemented" - this is honest)
- Granular per-user/per-group audit logging
- Large file chunked upload optimization

---

## Working Style with User (Ulf)

### Communication
- User communicates in Swedish
- I respond in Swedish unless code/docs require English
- User prefers direct, professional responses
- User values honesty above all - never document features that don't exist

### Code Quality Standards
- **Documentation must accurately reflect actual features**
- No fictional/aspirational features in docs
- Test features before documenting them
- Keep version numbers synchronized across all docs
- When fixing bugs, understand root cause before implementing solution

### Git Practices
- **Commit messages:** English, conventional commit format (fix:, feat:, docs:)
- **Always include:** Claude Code attribution in commits
- **Branch strategy:** Use `wolfface` for development, merge to `main` when stable
- **Push carefully:** Verify changes before pushing

### Development Workflow
1. Understand the problem thoroughly
2. Find root cause (don't guess)
3. Make surgical changes (minimum necessary)
4. Test/verify the fix works
5. Update version numbers if releasing
6. Update CHANGELOG.md with user-friendly description
7. Commit with clear message
8. Build and restart local server
9. Push to appropriate branch

---

## Version History Quick Reference

### v4.7.5 Galadriel (2025-11-20) - wolfface branch
- **CRITICAL:** Fixed SMTP security vulnerability (InsecureSkipVerify)
- Added comprehensive SMTP logging
- Verified email system architecture (10 functions, provider-switching)

### v4.7.4 Galadriel (2025-11-20) - wolfface branch
- Fixed wolf favicon not displaying (moved to `<head>`)

### v4.7.3 Galadriel (2025-11-18) - main branch
- Team Filter Dropdown
- Improved All Files View with better visual grouping

### v4.7.2 Galadriel (2025-11-18)
- Added wolf emoji favicon (but it didn't work until v4.7.4)
- Email notification improvements
- Various bug fixes

---

## Important Reminders

1. **Always verify features exist in code before documenting them**
2. **SQLCipher is NOT implemented** - recommend OS-level encryption (LUKS, BitLocker, FileVault)
3. **Rate limiting is NOT implemented** - recommend reverse proxy if needed
4. **When making HTML changes** - remember there are 31+ locations where HTML is generated
5. **Favicon must be in `<head>`** - browsers ignore it in `<body>`
6. **Version constant** is in `cmd/server/main.go:26`
7. **User's local server** runs on http://localhost:8080 but external URL is http://wulfvault.dyndns.org:8080
8. **Current branch is wolfface** - continue development here

---

## Next Steps / Future Work

Potential improvements discussed:
1. ✅ Wolf favicon now works!
2. Consider implementing rate limiting (currently documented as "not implemented")
3. Add granular audit logging per user/group
4. Consider large file upload optimization
5. Keep documentation synchronized with actual features

---

*This file serves as context for future Claude Code sessions working on WulfVault.*
*Last session: Fixed favicon display issue by moving HTML to proper `<head>` location.*
