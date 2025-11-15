# Bug Fix: Localhost URL in File Links

**Date**: 2025-11-12
**Version**: v3.2-RC2
**Severity**: High - Affects all file download links
**Status**: ✅ RESOLVED

## Problem Description

All file download links displayed "localhost" in the URL instead of the correct external IP address configured by the admin. For example:
- Incorrect: `http://localhost:8080/s/73faedb917fcd595d60823850bc96370`
- Correct: `http://90.229.210.13:8080/s/73faedb917fcd595d60823850bc96370`

### User Experience Impact
1. Admin configured the correct server URL (`90.229.210.13:8080`) via the admin settings
2. File links were generated with `localhost` instead
3. When server restarted, the problem persisted
4. External users could not access shared files (localhost resolves to their own machine)

## Root Cause Analysis

### Configuration System
WulfVault uses a two-tier configuration system:
1. **config.json file** (`data/config.json`) - Persistent configuration storage
2. **Command-line flags** - Override config file when explicitly set

### Code Analysis

**File**: `cmd/server/main.go`

**Lines 84-88**:
```go
// Check if server URL was explicitly provided
serverURLFromEnv := getEnv("SERVER_URL", "")
if serverURLFromEnv != "" || isFlagPassed("url") {
    cfg.ServerURL = *serverURL
}
```

This code shows that if the `-url` flag is explicitly passed, it **overrides** the config file.

### The Bug
The server was being started with:
```bash
./wulfvault -port 8080 -url http://localhost:8080
```

This meant that:
1. Admin sets correct URL via web UI → Saved to `data/config.json` ✅
2. Server restart → Started with `-url http://localhost:8080` ❌
3. Command-line flag overrides saved config → localhost wins 💥

### Evidence

**Config file** (`data/config.json`):
```json
{
  "serverUrl": "http://90.229.210.13:8080",
  ...
}
```

**Server startup log** (before fix):
```
2025/11/12 11:40:59 Server configuration:
2025/11/12 11:40:59   - URL: http://localhost:8080  ❌
```

**Process command** (before fix):
```bash
./wulfvault -port 8080 -url http://localhost:8080  ❌
```

## Solution

### Fix Implementation
Remove the `-url` flag from the server startup command. Let the server read the URL from `config.json` instead.

**Before**:
```bash
./wulfvault -port 8080 -url http://localhost:8080
```

**After**:
```bash
./wulfvault -port 8080
```

### How It Works Now
1. Server starts without explicit `-url` flag
2. `config.LoadOrCreate()` reads `data/config.json`
3. Config file contains `"serverUrl": "http://90.229.210.13:8080"`
4. Server uses this URL for all file links
5. Admin can change URL via web UI, and it persists across restarts ✅

### Verification

**Server startup log** (after fix):
```
2025/11/12 12:00:29 Server configuration:
2025/11/12 12:00:29   - URL: http://90.229.210.13:8080  ✅
2025/11/12 12:00:29 📍 Server URL: http://90.229.210.13:8080  ✅
```

**Process command** (after fix):
```bash
ulf  2060  ./wulfvault -port 8080  ✅
```

## Lessons Learned

1. **Command-line flags override config files** - This is by design for flexibility, but must be used carefully
2. **Startup scripts should use minimal flags** - Only override when necessary (e.g., port for different environments)
3. **Configuration priority order**:
   - Environment variables (highest priority)
   - Command-line flags
   - Config file
   - Default values (lowest priority)

## Testing Recommendations

To verify the fix:
1. ✅ Check server startup log shows correct URL
2. ✅ Navigate to file list in admin panel
3. ✅ Verify file download links contain correct IP/domain
4. ✅ Test that links work from external machine
5. ✅ Change URL via admin settings
6. ✅ Restart server (without `-url` flag)
7. ✅ Verify new URL persists

## Future Prevention

### Recommended Startup Command
```bash
# Production - use config file for URL
./wulfvault -port 8080

# Development - override URL for testing
./wulfvault -port 8080 -url http://localhost:8080

# Docker/Production with environment variables
SERVER_URL=https://wulfvault.example.com PORT=443 ./wulfvault
```

### systemd Service File
If using systemd, ensure the service file does NOT include `-url`:
```ini
[Service]
ExecStart=/path/to/wulfvault -port 8080
# DO NOT: ExecStart=/path/to/wulfvault -port 8080 -url http://localhost:8080
```

## Related Files
- `cmd/server/main.go` - Flag parsing and config override logic (lines 84-88)
- `internal/config/config.go` - Config file loading and saving
- `data/config.json` - Persistent configuration storage

## Resolution Status
✅ **FIXED** - Server now correctly uses config file URL
✅ **TESTED** - File links show correct IP address
✅ **DOCUMENTED** - This document serves as permanent record
