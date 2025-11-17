# WulfVault Development Notes

**Last Updated:** 2025-11-17
**Current Version:** 4.6.0 Champagne  
**Developer:** Ulf Holmström (Frimurare)  
**Contact:** ulf@manvarg.se

---

## 📋 Project Overview

**WulfVault** is an enterprise-grade self-hosted file sharing platform built in Go, inspired by Gokapi but significantly enhanced with enterprise features.

### Repository Information
- **GitHub:** https://github.com/Frimurare/WulfVault
- **Docker Hub:** https://hub.docker.com/r/frimurare/wulfvault
- **License:** AGPL-3.0
- **Git Author:** Frimurare <ulf@manvarg.se>

### Development Environment
- **Language:** Go 1.23
- **Database:** SQLite
- **Deployment:** LXC Container on Proxmox
- **Location:** `/home/ulf/WulfVault/`
- **Data Directory:** `/home/ulf/WulfVault/data/`
- **Uploads Directory:** `/home/ulf/WulfVault/uploads/`

---

## 🎯 Current Status (v4.6.0 Champagne - GDPR Compliance)

### Latest Release Features
**Enterprise User Management with Pagination & Filtering**

#### Key Features Implemented:
1. **User Management Pagination**
   - 50 users per page (default, configurable up to 200)
   - Previous/Next navigation buttons
   - Result counter showing "Showing X-Y of Z users"
   - Database-optimized queries with LIMIT/OFFSET

2. **Advanced Filtering**
   - **Search:** Real-time search by name or email (debounced 500ms)
   - **User Level Filter:** All Users / Regular Users Only / Admins Only
   - **Status Filter:** All / Active / Inactive
   - Filter state preserved across pagination

3. **Download Account Management**
   - Independent pagination system (separate from user pagination)
   - Same filtering capabilities as regular users
   - Search by name or email

4. **Performance**
   - Scales to thousands of users without performance degradation
   - Mobile-responsive UI
   - Proper SQL parameterization for security

#### Technical Implementation:
```
Files Changed (642 lines of production code):
- internal/database/users.go (+132 lines)
  - UserFilter struct with pagination/filtering options
  - GetUsersWithPagination(filter *UserFilter)
  - GetUsersCount(filter *UserFilter)
  
- internal/database/downloads.go (+122 lines)
  - DownloadAccountFilter struct
  - GetDownloadAccountsWithPagination(filter *DownloadAccountFilter)
  - GetDownloadAccountsCount(filter *DownloadAccountFilter)
  
- internal/server/handlers_admin.go (+388 lines)
  - handleAdminUsers() - Enhanced with pagination UI
  - renderAdminUsers() - Filter controls and pagination buttons
  - JavaScript for AJAX loading and filter state management
```

### Documentation Status
✅ All documentation updated to v4.5.13:
- `README.md` - Main project documentation
- `USER_GUIDE.md` - Complete user and admin guide
- `CHANGELOG.md` - Comprehensive release notes
- `DEPLOYMENT.md` - Installation and deployment instructions

---

## 🚀 Build & Deployment Process

### Local Development Build
```bash
cd /home/ulf/WulfVault
go build -o wulfvault ./cmd/server
./wulfvault
```

### Running Server
```bash
# Start server
./wulfvault

# With custom port
./wulfvault -port 8080

# With custom data directory
./wulfvault -data ./data -uploads ./uploads

# Check running process
ps aux | grep wulfvault | grep -v grep

# Stop server
kill <PID>
```

### Docker Deployment (Automated)
**GitHub Actions automatically builds and publishes to Docker Hub on every push to main.**

Workflow file: `.github/workflows/docker-publish.yml`

Published tags:
- `frimurare/wulfvault:latest` (always points to main branch)
- `frimurare/wulfvault:4.5.13 Gold` (version-specific tag)
- `frimurare/wulfvault:main-<git-sha>` (commit-specific tag)

Platforms built:
- `linux/amd64`
- `linux/arm64`

Build time: ~5-10 minutes after push

### Version Updates
When releasing a new version:

1. Update version in `cmd/server/main.go`:
   ```go
   const (
       Version = "4.5.XX Gold"
   )
   ```

2. Update `CHANGELOG.md` with release notes

3. Update `README.md` and `USER_GUIDE.md` if needed

4. Commit and push to main - Docker Hub builds automatically

---

## 📁 Project Structure

```
WulfVault/
├── cmd/server/              # Main application entry point
│   └── main.go             # Version defined here
├── internal/
│   ├── auth/               # Authentication & sessions
│   ├── cleanup/            # File expiration & trash cleanup
│   ├── config/             # Configuration management
│   ├── database/           # SQLite database layer
│   │   ├── users.go        # User CRUD + pagination
│   │   ├── downloads.go    # Download accounts + pagination
│   │   ├── files.go        # File management
│   │   └── audit.go        # Audit logging
│   ├── email/              # Email sending (Brevo/SMTP)
│   ├── models/             # Data structures
│   └── server/             # HTTP server & handlers
│       ├── server.go       # Routes and middleware
│       ├── handlers_admin.go    # Admin panel handlers
│       ├── handlers_auth.go     # Login/logout
│       ├── handlers_files.go    # File upload/download
│       └── handlers_rest_api.go # REST API endpoints
├── web/static/             # CSS, JS, images
├── docs/                   # Additional documentation
├── .github/workflows/      # GitHub Actions (Docker builds)
└── data/                   # SQLite database (gitignored)
```

### Important Files
- `cmd/server/main.go` - Version number, startup logic, cleanup schedulers
- `internal/server/handlers_admin.go` - Admin panel logic including pagination
- `internal/database/users.go` - User management with filtering
- `CHANGELOG.md` - Complete release history
- `README.md` - Project overview and features
- `USER_GUIDE.md` - End-user documentation

---

## 🔧 Common Development Tasks

### Adding a New Feature
1. Create feature branch: `git checkout -b feature-name`
2. Implement changes
3. Update version in `cmd/server/main.go` if releasing
4. Update `CHANGELOG.md` with changes
5. Test locally: `go build && ./wulfvault`
6. Commit with descriptive message (author: Frimurare <ulf@manvarg.se>)
7. Push to GitHub
8. Merge to main when ready

### Database Changes
Database schema is in `internal/database/database.go`

To add new tables/columns:
1. Update schema in `Initialize()` function
2. Add migration logic if needed
3. Update corresponding CRUD functions
4. Test with fresh database and existing database

### Git Workflow
```bash
# Check current branch
git branch

# Pull latest changes
git pull origin main

# Create feature branch
git checkout -b feature-name

# Stage changes
git add <files>

# Commit (ensure author is correct)
git config user.name "Frimurare"
git config user.email "ulf@manvarg.se"
git commit -m "Description"

# Push to GitHub
git push origin feature-name

# Merge to main (when ready)
git checkout main
git merge --no-ff feature-name
git push origin main
```

### Creating Backups
```bash
cd /home/ulf
tar -czf WulfVault-backup-$(date +%Y%m%d-%H%M%S).tar.gz WulfVault/
```

Latest backup: `/home/ulf/WulfVault-backup-20251117-103848.tar.gz` (589MB)

### Restoring from Backup
```bash
cd /home/ulf
tar -xzf WulfVault-backup-YYYYMMDD-HHMMSS.tar.gz
cd WulfVault
go build -o wulfvault ./cmd/server
```

---

## 🐛 Debugging & Troubleshooting

### Common Issues

#### 1. "Loading users..." stuck on admin page
**Cause:** JavaScript fetch() not sending session cookies  
**Solution:** Ensure `credentials: 'include'` in fetch options  
**File:** `internal/server/handlers_admin.go` (JavaScript section)

#### 2. Session timeout during file uploads
**Handled:** Active transfer detection prevents timeout  
**Code:** `hasActiveTransfer()` in middleware checks for ongoing uploads/downloads

#### 3. Docker build fails on GitHub Actions
**Check:** 
- GitHub Secrets: `DOCKER_HUB_USERNAME` and `DOCKER_HUB_TOKEN` must be set
- Workflow file: `.github/workflows/docker-publish.yml`
- Build logs: https://github.com/Frimurare/WulfVault/actions

#### 4. Database locked errors
**Cause:** SQLite doesn't handle concurrent writes well  
**Solution:** Transactions are used for writes, reads are concurrent-safe  
**Mitigation:** Keep write operations short

### Logging
Server logs to stdout. To capture logs:
```bash
./wulfvault > wulfvault.log 2>&1 &
tail -f wulfvault.log
```

### Database Inspection
```bash
sqlite3 /home/ulf/WulfVault/data/wulfvault.db

# Useful queries:
SELECT COUNT(*) FROM Users;
SELECT Id, Name, Email, Userlevel FROM Users;
SELECT COUNT(*) FROM Files;
SELECT * FROM AuditLog ORDER BY Timestamp DESC LIMIT 10;
```

---

## 📝 Code Style & Conventions

### Naming Conventions
- **Go files:** lowercase with underscores (e.g., `handlers_admin.go`)
- **Functions:** CamelCase (e.g., `GetUsersWithPagination`)
- **Structs:** PascalCase (e.g., `UserFilter`)
- **Database fields:** PascalCase in Go, same in SQLite (e.g., `UserLevel`)
- **JSON fields:** camelCase (e.g., `userLevel` in JSON responses)

### Database Field Naming
- SQLite columns: PascalCase (e.g., `Userlevel`, `StorageQuotaMB`)
- Go structs: Match SQLite exactly
- JSON tags: camelCase for API responses

### Git Commit Messages
Format:
```
[Type] Short description (50 chars max)

Detailed explanation of what changed and why.
Include technical details, affected files, and impact.

Author: Frimurare <ulf@manvarg.se>
```

Types: `Fix`, `Feature`, `Release`, `Update`, `Refactor`, `Docs`

### Code Comments
- Add comments for complex logic
- Document exported functions with GoDoc format
- Include copyright header in new files

---

## 🔐 Security Considerations

### Authentication
- Password hashing: bcrypt (cost factor 14)
- Session cookies: HttpOnly, Secure (in production)
- Session timeout: 10 minutes of inactivity (extended during active transfers)
- 2FA: TOTP support for users and admins

### SQL Injection Prevention
- **ALWAYS** use parameterized queries
- Example:
  ```go
  db.Query("SELECT * FROM Users WHERE Email = ?", email)
  ```
- Never concatenate user input into SQL strings

### File Upload Security
- Size limits enforced
- Quota checks before upload
- Virus scanning: NOT implemented (future enhancement)
- File path traversal prevented

### Audit Logging
- All admin actions logged
- User login/logout logged
- File operations logged
- API access logged
- Retention: 90 days (configurable)

---

## 📊 Database Schema Highlights

### Users Table
```sql
CREATE TABLE Users (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Name TEXT,
    Email TEXT UNIQUE,
    Password TEXT,
    Userlevel INTEGER,        -- 0=SuperAdmin, 1=Admin, 2=User
    Permissions INTEGER,
    StorageQuotaMB INTEGER,
    StorageUsedMB INTEGER,
    IsActive INTEGER,         -- 1=Active, 0=Inactive
    CreatedAt INTEGER,
    LastOnline INTEGER,
    ResetPassword INTEGER
)
```

### DownloadAccounts Table
```sql
CREATE TABLE DownloadAccounts (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Name TEXT,
    Email TEXT UNIQUE,
    Password TEXT,
    CreatedAt INTEGER,
    LastUsed INTEGER,
    DownloadCount INTEGER,
    IsActive INTEGER,
    DeletedAt INTEGER,
    DeletedBy TEXT,
    OriginalEmail TEXT
)
```

### Files Table (excerpts)
- File metadata, encryption keys, expiration, download limits
- Soft delete with trash retention (5 days default)

### AuditLog Table
- Timestamp, UserId, Action, Details, IPAddress, UserAgent
- Retention: 90 days, max size: 100MB

---

## 🚀 Future Enhancements (Ideas)

### Short-term
- [ ] Virus scanning integration (ClamAV)
- [ ] Email notifications for file expiration
- [ ] Advanced user permissions (granular)
- [ ] API rate limiting

### Medium-term
- [ ] File versioning
- [ ] Public sharing links with password protection
- [ ] WebDAV support
- [ ] LDAP/Active Directory integration

### Long-term
- [ ] Multi-tenancy support
- [ ] S3-compatible storage backend
- [ ] Real-time collaboration features
- [ ] Mobile app (iOS/Android)

---

## 🎓 Learning Resources

### Go Development
- Official Go documentation: https://golang.org/doc/
- Effective Go: https://golang.org/doc/effective_go
- SQLite in Go: https://github.com/mattn/go-sqlite3

### Project Inspiration
- Gokapi: https://github.com/Forceu/Gokapi (architectural inspiration)

---

## 📞 Contact & Support

**Developer:** Ulf Holmström (Frimurare)  
**Email:** ulf@manvarg.se  
**GitHub:** https://github.com/Frimurare/WulfVault

---

## 🎉 Recent Milestones

### v4.5.13 Gold (2025-11-17)
✅ Enterprise user management pagination  
✅ Advanced filtering and search  
✅ Complete documentation overhaul  
✅ Repository cleanup (removed dev notes)  
✅ Git author corrections  
✅ Docker Hub auto-publishing  

### v4.5.12 Gold (2025-11-16)
✅ Complete audit logging system  
✅ Admin UI for audit logs  
✅ Pagination controls for audit logs  

### v4.5.x Series
✅ 2FA/TOTP support  
✅ Email integration (Brevo/SMTP)  
✅ Teams feature  
✅ REST API  
✅ Download accounts  
✅ Trash system  

---

**Remember:** Always commit with correct author (`Frimurare <ulf@manvarg.se>`) and write clear commit messages. Test locally before pushing to main. GitHub Actions handles Docker builds automatically.

