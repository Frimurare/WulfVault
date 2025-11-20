# WulfVault Development Notes

**Last Updated:** 2025-11-20
**Current Version:** v4.7.4 Galadriel

---

## Project Overview

WulfVault is a self-hosted secure file sharing system written in Go with a web UI. It's designed for small to medium organizations needing secure file sharing without cloud dependencies.

**Repository:** https://github.com/Frimurare/WulfVault
**Server runs on:** http://localhost:8080

---

## Recent Development Session (2025-11-17 to 2025-11-19)

### What We Accomplished

#### 1. Version 4.7.3 Galadriel Features
- **Team Filter Dropdown** - Users can now filter files by specific team in the file list
- **Improved All Files View** - Files and their attached notes are now grouped together visually

#### 2. Documentation Cleanup (Critical)
**Problem:** Documentation contained fictional features that were never implemented (unprofessional)

**Removed all SQLCipher references** - This database encryption was NEVER implemented
- Updated 7 markdown files to recommend OS-level disk encryption (LUKS, BitLocker, FileVault) instead
- Files affected:
  - README.md
  - GDPR_COMPLIANCE_SUMMARY.md
  - gdpr-compliance/DATA_PROCESSING_AGREEMENT_TEMPLATE.md
  - gdpr-compliance/DEPLOYMENT_CHECKLIST.md
  - gdpr-compliance/PRIVACY_POLICY_TEMPLATE.md
  - gdpr-compliance/README.md
  - gdpr-compliance/RECORDS_OF_PROCESSING_ACTIVITIES.md

**Rate Limiting** - API docs were already honest (says "not implemented"), no changes needed

#### 3. Git Workflow
- Updated README.md from v4.7.2 to v4.7.3
- Updated docs/API.md from v4.1.0 to v4.7.3
- All changes pushed to GitHub main branch

---

## Key Architecture Notes

### Directory Structure
- `cmd/server/main.go` - Entry point
- `internal/server/` - HTTP handlers and server logic
  - `handlers_user.go` - User management, GDPR export/deletion
  - `handlers_gdpr.go` - GDPR-specific endpoints
  - `handlers_files.go` - File operations
  - `handlers_admin.go` - Admin panel
- `internal/database/` - SQLite database operations
- `web/templates/` - HTML templates
- `web/static/` - CSS, JS, images
- `gdpr-compliance/` - GDPR templates and procedures

### Important Technical Details
- **Database:** SQLite (NOT SQLCipher - encryption is NOT built-in)
- **Password hashing:** bcrypt
- **2FA:** TOTP
- **Sessions:** Secure cookie-based
- **File storage:** Local filesystem with configurable path

### What's NOT Implemented (don't add to docs)
- SQLCipher database encryption
- Rate limiting
- Granular per-user/per-group audit logging
- Large file chunked upload optimization

---

## Working Style with User (Ulf)

### Communication
- User communicates in Swedish, I respond in Swedish or English as appropriate
- User prefers direct, professional responses
- User values honesty - never document features that don't exist

### Code Quality Standards
- Documentation must accurately reflect actual features
- No fictional/aspirational features in docs
- Test features before documenting them
- Keep version numbers synchronized across all docs

### Git Practices
- Commit messages in English
- Use conventional commit format (docs:, feat:, fix:)
- Always include the Claude Code attribution in commits
- Push to main branch after verification

---

## Server Management

### Starting the Server
```bash
cd /home/ulf/WulfVault
./wulfvault
# Or with logging:
nohup ./wulfvault > /tmp/wulfvault.log 2>&1 &
```

### Building
```bash
go build -o wulfvault cmd/server/main.go
```

### Checking Logs
```bash
tail -f /tmp/wulfvault.log
```

---

## Next Steps / Future Work

Potential improvements discussed:
1. Implement rate limiting (currently documented as "not implemented")
2. Add granular audit logging per user/group
3. Consider large file upload optimization
4. Keep documentation synchronized with actual features

---

## Important Reminders

1. **Always verify features exist in code before documenting them**
2. **SQLCipher is NOT implemented** - use OS-level encryption
3. **Rate limiting is NOT implemented** - use reverse proxy if needed
4. **Check all .md files when making documentation changes** - there are many GDPR templates
5. **User's server typically runs at localhost:8080**

---

*This file serves as context for future Claude Code sessions working on WulfVault.*
