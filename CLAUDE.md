# WulfVault Development Notes

**Last Updated:** 2025-11-20
**Current Version:** v4.7.5 Galadriel
**Current Branch:** wolfface

---

## Project Overview

WulfVault is a self-hosted secure file sharing system written in Go with a web UI. It's designed for small to medium organizations needing secure file sharing without cloud dependencies.

**Repository:** https://github.com/Frimurare/WulfVault
**Server runs on:** http://localhost:8080 (local development)

---

## Current Development Session (2025-11-20)

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
