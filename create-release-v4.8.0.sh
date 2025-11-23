#!/bin/bash

# WulfVault v4.8.0 Release Script
# Run this after authenticating with: gh auth login

gh release create v4.8.0 --draft --title "v4.8.0 Shotgun - UI Improvements + Download Time Tracking + Unified Authentication" --notes "$(cat <<'EOF'
# v4.8.0 Shotgun 🎨🔐📊

## Major Features

### 🎨 UI/UX Improvements
- **Fixed long filename display** - Filenames now truncated to 600px with hover tooltip showing full name
- **Fixed button alignment** - History, Email, Edit, Delete buttons consistently placed using flexbox
- **5 Most Active Users** - Dashboard now shows top 5 users instead of just one
- **Edit dialog** - Confirmed full feature parity with upload form (password, auth, teams, etc.)

### 🔐 Unified Authentication System
- Regular users and admins can now use **existing accounts** to download password-protected files
- No more duplicate "download user" accounts for existing users
- System checks for existing user account BEFORE creating download account
- Prevents confusion and duplicate account creation

### 📊 Download Time Tracking
- All file downloads now tracked with **duration measurement**
- Download time stored in audit log (e.g., "took 2.45 seconds")
- Helps detect:
  - 🤖 Bot activity (0-second "downloads" without actual file transfer)
  - 🐌 Network performance issues (slow downloads)
  - ✅ Complete vs. incomplete downloads

## Technical Details

**Frontend Changes:**
- Updated filename HTML with \`<span>\` wrapper for max-width control
- File-actions container uses flexbox with \`gap: 8px\`
- Buttons have \`flex: 0 0 auto\` to prevent resizing

**Backend Changes:**
- Added \`GetTop5ActiveUsers()\` database function
- Enhanced \`handleAuthenticatedDownload()\` to check existing user sessions
- Modified \`handleDownloadAccountCreation()\` to authenticate existing users
- Download timing measured around \`http.ServeFile()\` call
- Audit log includes \`download_time_seconds\` in JSON details

## Files Changed
- \`internal/server/handlers_user.go\` - Filename display, button alignment
- \`internal/server/handlers_admin.go\` - Top 5 active users
- \`internal/server/handlers_files.go\` - Download timing, unified auth
- \`internal/database/downloads.go\` - New GetTop5ActiveUsers() function
- \`CHANGELOG.md\` - Full changelog

## Upgrade Notes
- No database migrations required
- No breaking changes
- Existing download accounts remain functional
- New authentication flow applies to future downloads

---

**Full Changelog**: https://github.com/Frimurare/WulfVault/blob/main/CHANGELOG.md
EOF
)"

echo "Release draft v4.8.0 created successfully!"
echo "Visit https://github.com/Frimurare/WulfVault/releases to publish it."
