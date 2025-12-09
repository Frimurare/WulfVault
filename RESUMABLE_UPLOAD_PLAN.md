# Resumable Upload Implementation Plan
## WulfVault - Automatic Retry & Resume för Stora Filer

**Datum:** 2025-12-09
**Version:** 1.0
**Status:** Planning Phase

---

## 📋 Executive Summary

### Problem
- Stora filer (>40GB som Mordnatten20251025.zip) avbryts vid upload på grund av nätverksproblem
- Användare måste starta om från början, vilket slösar tid och bandbredd
- Ingen möjlighet att verifiera filintegritet efter avbrott

### Lösning
Implementera resumable uploads med automatisk retry-mekanism och filintegritetsverifiering.

### Uppskattad Tid
**Total: 40-60 timmar** (1-1.5 veckor för 1 utvecklare)

### Komplexitet
**Medel till Hög** - Kräver omfattande ändringar i både frontend och backend

---

## 🔍 Current Architecture Analysis

### Nuvarande Implementation
```
Client (JavaScript)        Server (Go)
─────────────────────     ─────────────────
1. Välj fil
2. XMLHttpRequest POST    → handleUpload()
3. Skicka hela filen      → ParseMultipartForm()
4. Vänta på svar          → io.Copy() [hela filen]
                          → SaveFile()
                          → Respond 200 OK
```

**Problem med nuvarande approach:**
- ✗ Skickar hela filen i ett request
- ✗ Vid avbrott måste allt startas om från början
- ✗ Ingen möjlighet att resume
- ✗ Ingen chunk-baserad integritetskontroll
- ✗ Timeout för stora filer på dåliga förbindelser

---

## 🎯 Proposed Architecture

### Ny Implementation - Chunked Upload med Resume
```
Client (JavaScript)                    Server (Go)
─────────────────────                 ─────────────────────
1. Välj fil (41.36 GB)
2. Dela upp i chunks (100MB)
3. Initiera upload session            → POST /api/v1/upload/init
   - Filnamn, storlek, SHA256         → CreateUploadSession()
   - Returnerar upload_id             → Respond: {upload_id, chunk_size}

4. Loop för varje chunk:
   - Beräkna chunk SHA256
   - POST /api/v1/upload/chunk        → POST /upload/chunk
   - Upload chunk + metadata          → ValidateChunk()
                                      → SaveChunk()
                                      → UpdateProgress()
   - Vid fel: retry 3 gånger
   - Om retry misslyckas: pausa

5. Om pausa/avbrott:
   - Spara upload_id i localStorage
   - GET /api/v1/upload/status        → GetUploadStatus()
   - Returnerar: completed chunks     → Respond: {uploaded_chunks[]}

6. Resume:
   - Återuppta från senaste chunk
   - Fortsätt från steg 4

7. Finalize upload:
   - POST /api/v1/upload/finalize     → POST /upload/finalize
   - Server kombinerar chunks         → CombineChunks()
   - Verifierar total SHA256          → VerifySHA256()
   - Skapar FileInfo                  → SaveFile()
   - Respond: success                 → Respond: {file_id, sha1}
```

---

## 📦 Required Components

### Backend (Go) - NEW Components

#### 1. Upload Session Manager
```go
// File: internal/upload/session.go (~200 lines)

type UploadSession struct {
    UploadID        string
    FileID          string
    UserID          int
    Filename        string
    TotalSize       int64
    TotalChunks     int
    ChunkSize       int64
    ExpectedSHA256  string
    UploadedChunks  []int
    CreatedAt       int64
    LastUpdatedAt   int64
    Status          string // "pending", "uploading", "paused", "completed", "failed"
    RetryCount      int
    MaxRetries      int
}

func CreateUploadSession(userID int, filename string, size int64, sha256 string) (*UploadSession, error)
func GetUploadSession(uploadID string) (*UploadSession, error)
func UpdateUploadSession(session *UploadSession) error
func DeleteUploadSession(uploadID string) error
```

**Effort:** 4-6 timmar

#### 2. Chunk Handler
```go
// File: internal/server/handlers_upload_chunk.go (~300 lines)

func (s *Server) handleUploadInit(w http.ResponseWriter, r *http.Request)
func (s *Server) handleUploadChunk(w http.ResponseWriter, r *http.Request)
func (s *Server) handleUploadStatus(w http.ResponseWriter, r *http.Request)
func (s *Server) handleUploadFinalize(w http.ResponseWriter, r *http.Request)
func (s *Server) handleUploadCancel(w http.ResponseWriter, r *http.Request)

// Helper functions
func saveChunk(uploadID string, chunkIndex int, data []byte) error
func verifyChunkSHA256(data []byte, expectedSHA256 string) bool
func combineChunks(uploadID string, totalChunks int) (string, error)
func cleanupChunks(uploadID string) error
```

**Effort:** 8-10 timmar

#### 3. Database Schema för Upload Sessions
```sql
-- File: internal/database/migrations/add_upload_sessions.sql

CREATE TABLE IF NOT EXISTS UploadSessions (
    UploadID TEXT PRIMARY KEY,
    FileID TEXT,
    UserID INTEGER,
    Filename TEXT,
    TotalSize INTEGER,
    TotalChunks INTEGER,
    ChunkSize INTEGER,
    ExpectedSHA256 TEXT,
    UploadedChunks TEXT, -- JSON array [1,2,3,4,5]
    CreatedAt INTEGER,
    LastUpdatedAt INTEGER,
    Status TEXT,
    RetryCount INTEGER,
    MaxRetries INTEGER
);

CREATE INDEX idx_upload_sessions_user ON UploadSessions(UserID);
CREATE INDEX idx_upload_sessions_status ON UploadSessions(Status);
```

**Effort:** 2-3 timmar

#### 4. Cleanup Scheduler
```go
// File: internal/cleanup/upload_sessions.go (~100 lines)

// Radera gamla/övergivna upload sessions (>7 dagar gamla)
func CleanupStaleUploadSessions() error
func StartUploadSessionCleanupScheduler()
```

**Effort:** 2-3 timmar

### Frontend (JavaScript) - NEW Components

#### 1. Resumable Upload Client
```javascript
// File: web/static/js/resumable-upload.js (~500 lines)

class ResumableUpload {
    constructor(file, uploadEndpoint, options = {}) {
        this.file = file;
        this.chunkSize = options.chunkSize || 100 * 1024 * 1024; // 100MB default
        this.maxRetries = options.maxRetries || 3;
        this.retryDelay = options.retryDelay || 5000; // 5 seconds
        this.uploadID = null;
        this.uploadedChunks = [];
        this.totalChunks = Math.ceil(file.size / this.chunkSize);
        this.currentChunk = 0;
        this.isPaused = false;
        this.isCancelled = false;
    }

    async calculateSHA256() { /* ... */ }
    async initUpload() { /* ... */ }
    async uploadChunk(chunkIndex) { /* ... */ }
    async verifyChunkIntegrity(chunk, expectedHash) { /* ... */ }
    async retryChunk(chunkIndex, attempt) { /* ... */ }
    async pauseUpload() { /* ... */ }
    async resumeUpload() { /* ... */ }
    async finalizeUpload() { /* ... */ }
    async getUploadStatus() { /* ... */ }
    onProgress(callback) { /* ... */ }
    onError(callback) { /* ... */ }
    onComplete(callback) { /* ... */ }
    onRetry(callback) { /* ... */ }
}
```

**Effort:** 10-12 timmar

#### 2. Integration med Dashboard
```javascript
// File: web/static/js/dashboard.js (modifications ~150 lines)

// Replace simple upload with resumable upload
async function handleResumableUpload(file, options) {
    const upload = new ResumableUpload(file, '/api/v1/upload', {
        chunkSize: 100 * 1024 * 1024, // 100MB
        maxRetries: 3,
        retryDelay: 5000
    });

    // Progress tracking
    upload.onProgress((progress) => {
        updateProgressBar(progress.percent);
        updateUploadStatus(progress.message);
    });

    // Retry handling
    upload.onRetry((info) => {
        showRetryNotification(info.attempt, info.maxRetries);
    });

    // Error handling
    upload.onError((error) => {
        if (error.retriesExhausted) {
            showFinalError(error.message);
        }
    });

    // Success handling
    upload.onComplete((result) => {
        if (result.integrityVerified) {
            showSuccessWithIntegrityCheck();
        }
    });

    await upload.start();
}
```

**Effort:** 4-5 timmar

#### 3. UI Components
```html
<!-- Upload Progress UI with Retry Info -->
<div class="upload-progress">
    <div class="progress-bar">
        <div class="progress-fill" style="width: 45%"></div>
    </div>
    <div class="upload-stats">
        <span class="chunk-info">Chunk 45/400 (11.25 GB / 41.36 GB)</span>
        <span class="speed">Speed: 5.2 MB/s</span>
        <span class="eta">ETA: 1h 23m</span>
    </div>
    <div class="retry-info" style="display: none;">
        <span class="retry-message">⚠️ Network error, retrying... (Attempt 2/3)</span>
    </div>
    <div class="actions">
        <button id="pauseBtn">⏸ Pause</button>
        <button id="cancelBtn">❌ Cancel</button>
    </div>
</div>
```

**Effort:** 3-4 timmar

### Configuration & Settings

#### Admin Settings Panel
```html
<!-- File: internal/server/handlers_admin.go - Add new settings -->

Resumable Upload Settings:
- Chunk Size (default: 100MB, range: 10MB - 500MB)
- Max Retries (default: 3, range: 0-10)
- Retry Delay (default: 5 seconds, range: 1-60s)
- Auto-resume on page reload (checkbox)
- Keep upload sessions (default: 7 days)
```

**Effort:** 2-3 timmar

---

## 🔧 Implementation Breakdown

### Phase 1: Backend Foundation (12-16 hours)
1. ✅ Create UploadSession model and database schema (3h)
2. ✅ Implement session manager (CRUD operations) (4h)
3. ✅ Create chunk storage system (filesystem structure) (3h)
4. ✅ Implement cleanup scheduler (2h)

### Phase 2: Backend API Endpoints (10-12 hours)
1. ✅ POST /api/v1/upload/init - Initialize upload (2h)
2. ✅ POST /api/v1/upload/chunk - Upload chunk (4h)
3. ✅ GET /api/v1/upload/status - Get upload status (1h)
4. ✅ POST /api/v1/upload/finalize - Finalize and combine (4h)
5. ✅ DELETE /api/v1/upload/cancel - Cancel upload (1h)

### Phase 3: Frontend Core (10-12 hours)
1. ✅ Implement ResumableUpload class (6h)
2. ✅ SHA256 calculation (client-side) (2h)
3. ✅ Chunk upload with retry logic (4h)

### Phase 4: Frontend Integration (6-8 hours)
1. ✅ Integrate with dashboard.js (3h)
2. ✅ Build progress UI components (3h)
3. ✅ localStorage persistence for resume (2h)

### Phase 5: Testing & Refinement (6-8 hours)
1. ✅ Test with small files (1h)
2. ✅ Test with large files (40GB+) (2h)
3. ✅ Test network interruption scenarios (2h)
4. ✅ Test retry logic (1h)
5. ✅ Test integrity verification (2h)

### Phase 6: Documentation & Admin UI (4-6 hours)
1. ✅ Update CHANGELOG.md (1h)
2. ✅ Create user guide (2h)
3. ✅ Admin settings panel (3h)

---

## 📊 Effort Estimates by Role

| Component | Hours | Complexity |
|-----------|-------|------------|
| **Backend Development** | 22-28h | High |
| - Database schema & migrations | 2-3h | Low |
| - Upload session manager | 4-6h | Medium |
| - Chunk handlers & API | 10-12h | High |
| - Cleanup scheduler | 2-3h | Low |
| - Testing & debugging | 4-6h | Medium |
| **Frontend Development** | 14-18h | High |
| - ResumableUpload class | 10-12h | High |
| - Dashboard integration | 4-5h | Medium |
| - UI components | 3-4h | Low |
| **Configuration & Admin** | 2-3h | Low |
| **Testing & QA** | 6-8h | Medium |
| **Documentation** | 2-3h | Low |
| **TOTAL** | **46-60h** | **Medium-High** |

---

## 🚨 Technical Challenges & Risks

### 1. Storage Management
**Challenge:** Storing chunks for partial uploads consumes disk space
**Risk:** Medium
**Mitigation:**
- Implement aggressive cleanup (delete stale uploads >7 days)
- Set upload session limits per user
- Monitor disk usage

### 2. Concurrent Uploads
**Challenge:** Multiple users uploading large files simultaneously
**Risk:** High
**Mitigation:**
- Rate limiting per user
- Queue system for large uploads
- Monitor server load

### 3. SHA256 Calculation (Client-Side)
**Challenge:** Calculating SHA256 for 41GB file in browser is slow
**Risk:** Medium
**Mitigation:**
- Calculate incrementally per chunk
- Show progress: "Preparing upload... calculating hash"
- Use Web Workers to avoid blocking UI

### 4. Browser Compatibility
**Challenge:** File API and crypto.subtle not supported in old browsers
**Risk:** Low
**Mitigation:**
- Feature detection with fallback to normal upload
- Display warning if browser doesn't support resumable uploads

### 5. Network Timing
**Challenge:** Detecting network failure vs slow connection
**Risk:** Medium
**Mitigation:**
- Smart timeout calculation based on chunk size
- Exponential backoff for retries
- User can manually pause/resume

---

## 💡 Alternative Approaches

### Option A: Current Plan - Full Implementation
**Pros:** Complete control, perfect UX, production-ready
**Cons:** 46-60 hours development time
**Recommendation:** ⭐ Best for long-term solution

### Option B: Simplified Retry (No Chunking)
**Approach:** Keep current upload, just add retry on failure
**Pros:** Quick implementation (8-10 hours)
**Cons:** Still uploads entire file each retry, wastes bandwidth
**Recommendation:** ❌ Not recommended for 40GB+ files

### Option C: Use Third-Party Library
**Library:** tus.io (resumable upload protocol)
**Pros:** Battle-tested, 20-30 hours implementation
**Cons:** External dependency, less control
**Recommendation:** ⚠️ Good middle-ground if time is critical

### Option D: Hybrid Approach
**Approach:** Implement basic chunking, add advanced features later
**Phase 1:** Basic chunking + resume (30-35 hours)
**Phase 2:** Add retry, integrity checks later (15-20 hours)
**Recommendation:** ✅ Good if you want to iterate

---

## 📝 Success Messages

### Successful Upload After Retries
```
✅ Upload Successful!

Your upload was interrupted 2 times due to network errors, but we managed
to upload it to WulfVault anyway.

File: Mordnatten20251025.zip (41.36 GB)
SHA256: abc123def456... ✓ VERIFIED

The integrity check shows that the file has been uploaded successfully.
We recommend that the receiver test the file to be 100% sure it is correct
before deleting the original!
```

### Successful Resume After Manual Pause
```
✅ Upload Resumed & Completed!

You paused the upload at 67% (27.7 GB) and resumed it successfully.

File: Mordnatten20251025.zip (41.36 GB)
SHA256: abc123def456... ✓ VERIFIED
Chunks: 414/414 uploaded successfully
Retries: 0

The file has been uploaded and verified successfully!
```

### Failed After Max Retries
```
❌ Upload Failed After 3 Retries

We tried to upload your file but encountered persistent network problems:

File: Mordnatten20251025.zip (41.36 GB)
Progress: 89% (36.8 GB uploaded)
Status: Connection timeout on chunk 367/414

What went wrong:
• Lost internet connection
• Weak or unstable network
• Server timeout

You can resume this upload later by clicking "Resume Upload" in the
dashboard. Your progress has been saved.
```

---

## 🎯 Recommended Approach

### My Recommendation: **Option D - Hybrid Approach**

**Phase 1: Core Functionality (30-35 hours)**
- Implement chunked uploads
- Basic resume capability
- Simple retry (1-2 attempts)
- localStorage persistence
- Basic integrity check

**Phase 2: Enhanced Features (15-20 hours)**
- Advanced retry logic (exponential backoff)
- Detailed progress tracking
- Pause/Resume UI
- Admin settings panel
- SHA256 verification

**Why This Approach?**
1. ✅ Get core functionality working quickly
2. ✅ Test with real users before adding complexity
3. ✅ Can deploy Phase 1 and gather feedback
4. ✅ Easier to debug issues in smaller increments
5. ✅ Less risk of over-engineering

---

## 📅 Suggested Timeline

### Sprint 1 (Week 1): Phase 1 - Core
- Day 1-2: Backend session manager + database
- Day 3-4: Backend chunk handlers
- Day 5-6: Frontend ResumableUpload class
- Day 7: Integration + basic testing

### Sprint 2 (Week 2): Phase 2 - Enhancement
- Day 1-2: Advanced retry logic
- Day 3-4: Progress UI + pause/resume
- Day 5: Admin settings
- Day 6-7: Testing + documentation

---

## 🔍 Final Assessment

### Complexity: **7/10** (Medium-High)
- Requires solid understanding of chunked uploads
- Network timing is tricky
- Must handle edge cases (partial uploads, corrupted chunks)

### Time: **46-60 hours** (Full implementation)
- Option D (Hybrid): 30-35h Phase 1, 15-20h Phase 2

### Value: **9/10** (Very High)
- Solves critical problem for large files
- Significantly improves UX
- Reduces support burden
- Production-ready solution

### Recommendation: **✅ PROCEED with Hybrid Approach**

Start with Phase 1 (core functionality) and evaluate before Phase 2.
This allows you to:
1. Test with real users (like Mordnatten20251025.zip uploads)
2. Gather feedback on UX
3. Identify edge cases
4. Decide if Phase 2 features are needed

---

## 📚 Resources & References

### Resumable Upload Protocols
- [tus.io](https://tus.io/) - Open protocol for resumable uploads
- [Google Cloud Resumable Uploads](https://cloud.google.com/storage/docs/resumable-uploads)
- [Uppy File Uploader](https://uppy.io/) - Reference implementation

### Browser APIs
- [File API](https://developer.mozilla.org/en-US/docs/Web/API/File)
- [SubtleCrypto (SHA256)](https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto)
- [Web Workers](https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API)

### Go Libraries
- [golang.org/x/crypto/sha256](https://pkg.go.dev/crypto/sha256)
- File chunk handling patterns

---

**Created by:** Claude Code
**Date:** 2025-12-09
**For:** WulfVault v5.0.1+
