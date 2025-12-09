# WulfVault Update History

This document tracks major updates and feature releases for WulfVault.

---

## Version 6.0.0 BloodMoon Beta 2 (2025-12-09)

### Major Update: Custom Chunked Upload System

**This is a significant architectural change** - WulfVault now features a custom-built chunked upload system replacing the previous tus.io integration.

### 🚀 New Features

#### Custom Chunked Upload Implementation
- **Automatic file chunking:** Files are automatically split into 5MB chunks for reliable transmission
- **Per-chunk retry logic:** Each chunk has independent retry capability (up to 10 attempts)
- **Exponential backoff:** Intelligent retry delays (1s, 2s, 4s, 8s, up to 10s max)
- **Network interruption recovery:** Upload continues from last successful chunk after connection loss
- **Full-screen progress overlay:** Large visual feedback with real-time statistics:
  - 72px red "UPLOADING - X%" text that turns green at 100%
  - Progress bar with smooth animations
  - Speed calculation (MB/s) and ETA display
  - Retry indicator showing connection interruption attempts
  - File info: name, size, and bytes uploaded
- **Success overlay with back button:** Large green "PRESS HERE TO GO BACK" button for users who leave uploads running for hours
- **Enhanced error handling:** Persistent error display with manual dismissal and retry count
- **Optimized for unstable connections:** Perfect for overnight uploads on unreliable broadband

### 🔧 Technical Changes

#### Backend
- **New file:** `internal/server/handlers_chunked_upload.go` (~350 lines)
  - `ChunkedUpload` struct with session management
  - `handleChunkedUploadInit()` - Initialize upload session
  - `handleChunkedUploadChunk()` - Receive and write chunk data
  - `handleChunkedUploadComplete()` - Finalize upload, calculate SHA1, create DB entry
  - `cleanupStaleUploads()` - Goroutine to remove abandoned uploads after 1 hour
- **Removed:** `internal/server/handlers_resumable_upload.go` (tus.io integration)
- **Updated:** `internal/server/server.go` - New API routes:
  - `POST /api/upload/init` - Initialize chunked upload
  - `POST /api/upload/chunk?upload_id=X&chunk_index=Y` - Upload chunk
  - `POST /api/upload/complete?upload_id=X` - Complete upload

#### Frontend
- **Updated:** `web/static/js/dashboard.js`
  - Removed tus-js-client dependency
  - New `uploadFileInChunks()` function with retry logic
  - `showUploadProgressOverlay()` - Full-screen visual feedback
  - `updateUploadProgress()` - Real-time progress updates with speed/ETA
  - `showUploadSuccess()` - Green success animation with back button
  - `showUploadError()` - Persistent error display with retry count
  - `showRetryIndicator()` - Visual feedback for retry attempts
- **Updated:** `internal/server/handlers_user.go` - UI improvements:
  - Green gradient Upload button (large, prominent)
  - Red gradient Cancel button with proper spacing (15px margin)

### 📊 Upload Flow

1. User selects file → frontend calls `/api/upload/init` with metadata
2. Backend creates temp file in `.chunks/` directory, returns upload_id
3. Frontend splits file into 5MB chunks
4. Each chunk POSTed to `/api/upload/chunk` with upload_id and chunk_index
5. On chunk failure: automatic retry with exponential backoff (up to 10 attempts)
6. After all chunks: POST to `/api/upload/complete`
7. Backend moves file from `.chunks/` to final location
8. Calculate SHA1 hash for integrity
9. Create database entry and update user storage quota
10. Send email notification if file >5GB
11. Show success overlay with "PRESS HERE TO GO BACK" button
12. Auto-reload after 3 seconds (or manual with button)

### 🎯 Performance & Reliability

- **Chunk size:** 5MB (optimal balance between speed and reliability)
- **Max retries:** 10 attempts per chunk (up from initial 5)
- **Retry backoff:** 1s, 2s, 4s, 8s, 10s, 10s, 10s, 10s, 10s, 10s
- **Session timeout:** 1 hour of inactivity before cleanup
- **Concurrent safety:** Mutex locks on upload sessions
- **Audit logging:** `FILE_UPLOADED_CHUNKED` action type for tracking

### 📝 Audit Trail

Chunked uploads are logged with:
- Upload initialization: `✅ Chunked upload initialized: <upload_id> (<filename>, <bytes>) by user <id>`
- Each chunk received: `📦 Chunk X received for upload <upload_id> (<bytes_received>/<total_size> bytes)`
- Upload completion: `✅ Chunked upload completed: <filename> (<size>) by user <id>`

### 🔍 Verified in Production

Successfully tested with:
- **Anchoragev3.mp4** - 231.6 MB uploaded in 47 chunks
- **Äggdårar 2021-10-06.zip** - 662.9 MB uploaded in 127 chunks
- All chunks logged and verified in server logs

### 🎨 UI Improvements

- Large green Upload button with gradient and hover effects
- Red Cancel button with proper 15px spacing
- Full-screen upload overlay with:
  - Large 72px status text
  - Smooth progress bar animations
  - Real-time speed and ETA calculations
  - Green success animation with pulsing effect
  - Red error state with manual dismissal
  - Large green "PRESS HERE TO GO BACK" button on success

### 🚨 Breaking Changes

- **Changed upload API:** Previous `/files/` endpoint replaced with `/api/upload/*` endpoints
- **Session cookies:** Upload sessions managed server-side with cleanup

### 📚 Documentation Updates

- Updated README.md to v6.0.0 BloodMoon Beta 2
- Added chunked upload feature documentation
- Created UPDATE_HISTORY.md for tracking major releases

### 🎉 Benefits

- **Reliability:** Network interruptions no longer cause full upload failure
- **Visibility:** Users see exact progress and retry attempts
- **User Experience:** Clear success indicator for long-running uploads
- **Accountability:** Complete audit trail of all upload operations
- **Ownership:** Full control of upload code, no external dependencies
- **GitHub-friendly:** All code is custom and shareable

---

## Previous Versions

### Version 5.0.3 FullMoon
- Upload UX improvements
- Enhanced file upload interface

### Version 4.9.5 Silverbullet
- Comprehensive feature set with Teams, 2FA, Audit Logs
- GDPR compliance features
- Email integrations (5 providers)

### Version 4.7.9
- Authenticated downloads as default
- Modern glassmorphic admin dashboard
- Twemoji integration

### Version 4.7
- File comments/descriptions
- Enhanced email templates
- Improved admin file management

### Version 4.5.13
- Enterprise pagination & filtering
- User search functionality

### Version 4.3
- Trash system with enhanced UI
- Restore functionality

### Version 4.2.2
- Smart expiry management
- Live countdown timers

### Version 4.2
- Team collaboration features
- Multi-team file sharing

---

**For detailed feature lists, see README.md**
