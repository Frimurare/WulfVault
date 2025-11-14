# Two-Factor Authentication (2FA) Setup Guide

## Overview

WulfVault 3.2 Beta 1 introduces TOTP-based Two-Factor Authentication (2FA) for enhanced account security. This feature adds an extra layer of protection by requiring a time-based code from an authenticator app in addition to your password.

## Features

- **TOTP Support**: Works with popular authenticator apps (Google Authenticator, Authy, Microsoft Authenticator, 1Password, etc.)
- **QR Code Setup**: Easy setup by scanning a QR code with your phone
- **Backup Codes**: 10 one-time backup codes for account recovery
- **User Control**: Enable/disable 2FA at any time from settings
- **Secure Implementation**: bcrypt-hashed backup codes, HttpOnly cookies, time skew tolerance

## Supported Accounts

✅ **Users** - 2FA available
✅ **Admins** - 2FA available
❌ **Download Accounts** - 2FA not available (by design - simpler auth flow)

---

## For End Users

### Enabling 2FA

1. **Login** to your WulfVault account
2. **Navigate** to Settings (click your name or go to `/settings`)
3. **Click** "Enable 2FA" button
4. **Click** "Generate QR Code"
5. **Scan** the QR code with your authenticator app:
   - Google Authenticator (iOS/Android)
   - Authy (iOS/Android/Desktop)
   - Microsoft Authenticator (iOS/Android)
   - 1Password, Bitwarden, or any TOTP-compatible app
6. **Save** your 10 backup codes in a safe place (print or store securely)
   - ⚠️ These are shown only once!
   - They can be used if you lose access to your authenticator app
7. **Enter** the 6-digit code from your authenticator app to verify
8. **Done!** 2FA is now enabled

### Logging in with 2FA

1. Enter your **email** and **password** as usual
2. You'll be redirected to the **2FA verification page**
3. Open your **authenticator app**
4. Enter the **6-digit code** (or click "Use backup code" if needed)
5. Code auto-submits when 6 digits are entered
6. You're logged in!

### Using Backup Codes

If you lose access to your authenticator app:

1. On the 2FA verification page, click **"Use a backup code instead"**
2. Enter one of your saved **backup codes**
3. The backup code will be **invalidated** after use (one-time use only)
4. You'll have 9 remaining backup codes

⚠️ **Important**: When you're down to your last few backup codes, regenerate new ones!

### Regenerating Backup Codes

1. Go to **Settings**
2. Click **"Regenerate Backup Codes"**
3. Confirm the action
4. **Save** your new 10 backup codes
5. Old backup codes are now invalid

### Disabling 2FA

1. Go to **Settings**
2. Click **"Disable 2FA"**
3. Enter your **password** to confirm
4. 2FA is now disabled

---

## For Administrators

### Deployment Considerations

**Database Migration**: Automatic on startup
- Adds 3 columns to Users table: `TOTPSecret`, `TOTPEnabled`, `BackupCodes`
- Migration runs automatically via `RunMigrations()` in `internal/database/migrations.go`
- No manual database changes needed

**Dependencies**: Two new Go libraries (automatically installed via `go.mod`)
- `github.com/pquerna/otp` v1.5.0 - TOTP generation and validation
- `github.com/skip2/go-qrcode` v0.0.0 - QR code image generation

**Routes Added**:
- `/settings` - User settings page
- `/2fa/setup` - Generate QR code (POST)
- `/2fa/enable` - Verify and enable 2FA (POST)
- `/2fa/disable` - Disable 2FA (POST)
- `/2fa/verify` - 2FA verification during login (GET/POST)
- `/2fa/regenerate-backup-codes` - Create new backup codes (POST)

### Security Features

1. **Secret Storage**: TOTP secrets stored in database, never exposed in JSON/API
2. **Backup Code Hashing**: All backup codes bcrypt-hashed (cost 12) before storage
3. **One-Time Backup Codes**: Automatically removed from database after use
4. **HttpOnly Cookies**: Session cookies not accessible via JavaScript
5. **SameSite Policy**: Strict SameSite policy prevents CSRF attacks
6. **Time Skew Tolerance**: Accepts codes within ±30 seconds window
7. **Session Timeouts**: 5-minute timeout for 2FA setup and verification sessions
8. **Password Required**: Must enter password to disable 2FA

### Known Limitations (Beta)

- **No Rate Limiting**: Not implemented yet (planned for stable release)
  - Recommendation: Use reverse proxy (nginx, Caddy) for rate limiting
- **No Email Notifications**: No emails sent when 2FA is enabled/disabled
  - Planned for stable release
- **No Admin Override**: Admins cannot disable 2FA for other users yet
  - Users must disable their own 2FA or use backup codes

### Monitoring and Support

**User Locked Out?**

If a user loses both their authenticator app AND backup codes:

**Option 1**: Database-level recovery (SuperAdmin only)
```sql
-- Disable 2FA for user (replace with actual user ID)
UPDATE Users SET TOTPEnabled = 0, TOTPSecret = '', BackupCodes = '' WHERE Id = X;
```

**Option 2**: Create new account (if data migration isn't critical)

### Configuration

No configuration needed! 2FA works out-of-the-box with default settings.

**Customizable settings** (future enhancement):
- Backup code count (currently: 10)
- Session timeout (currently: 5 minutes)
- TOTP period (currently: 30 seconds)
- TOTP digits (currently: 6)

---

## Technical Details

### TOTP Implementation

- **Algorithm**: SHA1 (standard for TOTP)
- **Period**: 30 seconds
- **Digits**: 6
- **Time Skew**: ±1 period (30 seconds before/after)

### Database Schema

```sql
-- Columns added to Users table
TOTPSecret TEXT DEFAULT ''        -- Base32-encoded secret (never exposed)
TOTPEnabled INTEGER DEFAULT 0     -- Boolean: 0 = disabled, 1 = enabled
BackupCodes TEXT DEFAULT ''       -- JSON array of bcrypt-hashed codes
```

### Security Best Practices

1. **Secret Storage**: Secrets stored in database, never logged or exposed in API
2. **Backup Codes**: Hashed with bcrypt (cost 12), one-time use only
3. **Session Security**: Temporary cookies with short expiration (5 min)
4. **Password Verification**: Required to disable 2FA
5. **Time Validation**: TOTP validates against current time ±30s

### Code Locations

- **TOTP Logic**: `internal/totp/totp.go`
- **Database Methods**: `internal/database/totp.go`
- **Handlers**: `internal/server/handlers_2fa.go`
- **Settings UI**: `internal/server/handlers_user_settings.go`
- **Login Flow**: `internal/server/handlers_auth.go` (lines 44-64)
- **User Model**: `internal/models/User.go` (lines 28-30)
- **Migration**: `internal/database/migrations.go` (lines 32-41)

---

## Troubleshooting

### "Invalid verification code"

**Causes**:
- Clock drift on server or phone
- Entering code from wrong account in authenticator app
- Code expired (codes change every 30 seconds)

**Solutions**:
- Ensure server time is synchronized (use NTP)
- Check you're using the code for the correct account
- Wait for a fresh code and try again
- Use a backup code if authenticator app is unavailable

### "Setup session not found"

**Cause**: 5-minute setup session expired

**Solution**: Start setup process again (click "Enable 2FA")

### "Session expired" during login

**Cause**: 2FA verification took longer than 5 minutes

**Solution**: Return to login page and login again

### Backup codes not working

**Causes**:
- Code already used (one-time use only)
- Typo in backup code
- Codes were regenerated

**Solutions**:
- Try another backup code
- Ensure code is entered correctly (case-insensitive)
- Contact admin for database-level recovery

---

## FAQ

**Q: Can I use SMS for 2FA instead of an authenticator app?**
A: No, WulfVault only supports TOTP authenticator apps. SMS is less secure and not recommended for enterprise use.

**Q: What happens if I lose my phone?**
A: Use one of your 10 backup codes to login, then disable and re-enable 2FA with your new device.

**Q: Can I use the same authenticator app on multiple devices?**
A: Yes! Most authenticator apps support cloud sync or manual export/import.

**Q: How often do the codes change?**
A: Every 30 seconds.

**Q: Can I disable 2FA if I forget my password?**
A: No. Use the "Forgot Password" flow first to reset your password, then you can disable 2FA.

**Q: Are backup codes case-sensitive?**
A: No, backup codes are case-insensitive.

**Q: Can admins disable 2FA for other users?**
A: Not yet. This is planned for the stable release. For now, users must disable their own 2FA or contact a SuperAdmin for database-level intervention.

**Q: What if I enter the wrong code multiple times?**
A: There's no lockout mechanism yet (beta limitation). However, codes expire every 30 seconds, so an attacker has limited attempts.

---

## Version History

**3.2-beta1** (2025-11-12)
- Initial 2FA implementation
- TOTP support with authenticator apps
- QR code generation for easy setup
- 10 backup codes per user
- Settings page for 2FA management
- Secure login flow with verification page

**Future Enhancements** (Planned for stable release):
- Rate limiting for 2FA verification attempts
- Email notifications for 2FA events
- Admin override to disable 2FA for users
- Recovery email option
- Remember device (30-day trusted devices)
- 2FA enforcement for all admins (optional policy)

---

## Support

For issues or questions:
- **GitHub Issues**: https://github.com/Frimurare/WulfVault/issues
- **Documentation**: See `CHANGELOG.md` for detailed release notes

---

**Version**: 3.2-beta1
**Last Updated**: 2025-11-12
