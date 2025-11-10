# Manvarg Sharecare File Sharing System
## Open Source Alternative to WeTransfer

---

## 🎯 PROJECT OVERVIEW

Build a lightweight, self-hosted file sharing system branded as **Manvarg Sharecare**. This is an open-source alternative to WeTransfer/Sprend, inspired by Gokapi but simpler and tailored for surveillance system customers.

**Core Purpose**: Provide Manvarg customers (especially those with Milestone XProtect video surveillance systems) a branded file sharing service as part of their service agreements, since XProtect lacks built-in file sharing like OpenEye systems have.

---

## 🎨 BRANDING REQUIREMENTS

### Visual Identity
- **Brand Name**: Manvarg Sharecare
- **Colors**:
  - Primary: Manvarg blue/security colors (to be extracted from logo)
  - Secondary: Professional dark/light contrast
  - Accent: Trust-inspiring colors
- **Logo Integration**:
  - Manvarg logo must be prominent on all pages
  - Header/navigation bar
  - Login pages
  - Email templates
  - Download pages
  - Admin dashboard

### Design Inspiration
- **WeTransfer**: Clean, minimal interface with focus on the upload/download action
- **Sprend**: Professional, secure feeling
- **Gokapi**: Simple, straightforward file management
- **Goal**: Create a professional, trustworthy interface that screams "security company"

---

## 👥 USER ROLES & AUTHENTICATION

### 1. Admin Users
- Created via admin panel during setup
- Can create/manage regular users
- Can view all files and downloads across system
- Can set global policies
- Can manage storage quotas per user/folder

### 2. Regular Users (Customers)
- Created by admins (no self-registration)
- Must authenticate to upload files
- Can create shareable links with optional download authentication
- Can view their own file statistics
- Have assigned storage quota
- Have assigned folder/directory space

### 3. File Recipients (Download Users)
- **Two modes**:
  1. **Authenticated Downloads**: Recipient must create download account with email + password
  2. **Direct Links**: No authentication required (for quick shares)

---

## 🔐 AUTHENTICATION & TRACKING

### Upload Authentication
- Only authenticated users (created by admin) can upload files
- Session-based login
- JWT tokens for API access

### Download Tracking
**For Authenticated Downloads**:
- Recipient provides email + creates password on first download
- System creates temporary download account
- Tracks:
  - Email address
  - Download timestamps
  - Number of downloads
  - IP address (optional, for security)
  - User agent (optional)

**For Direct Links**:
- No authentication required
- Still track:
  - Number of downloads
  - Download timestamps
  - Basic analytics (IP ranges, not stored permanently)

---

## 📁 FILE SHARING FEATURES

### File Upload & Management
- Drag-and-drop upload interface
- Multiple file uploads
- Progress indicators
- File deduplication (same file = same storage)
- Preview for common file types (images, PDFs)

### Link Generation
- **Random hash-based URLs** (like Gokapi): `https://your-domain.com/d/AbC123XyZ`
- Hash must be cryptographically secure and unpredictable
- Two link types:
  1. **Authenticated Link**: Requires recipient to create download account
  2. **Direct Link**: Open access, no login needed

### File Expiration & Limits
- Set expiration by:
  - Number of downloads (e.g., 5 downloads then delete)
  - Time period (e.g., 7 days then delete)
  - Combination (whichever comes first)
- Manual deletion by uploader
- Admin can delete any file

### Email Integration
- Send download link via email directly from system
- Email template branded with Manvarg Sharecare
- Include:
  - File name and size
  - Expiration info
  - Download instructions
  - Manvarg branding/logo

---

## 📊 STORAGE MANAGEMENT

### Per-User Quotas
- Admin sets maximum storage per user
- User dashboard shows:
  - Current usage
  - Available space
  - Number of active files

### Multi-Directory Support
- Admin can create multiple "storage folders/categories"
- Each folder can have its own quota
- Use case: Different quotas for different customer tiers
- Example:
  - `basic-customers/` - 5GB quota
  - `premium-customers/` - 50GB quota
  - `video-archive/` - 500GB quota

### Storage Optimization
- File deduplication (identical files stored once)
- Automatic cleanup of expired files
- Admin dashboard shows total storage usage

---

## 📈 ANALYTICS & REPORTING

### User Dashboard
- My uploaded files
- Total downloads per file
- Active vs expired files
- Storage usage

### Admin Dashboard
- Total system usage
- Active users
- Total files shared
- Most active users
- Storage trends
- Per-file download statistics with recipient tracking

### Download Tracking Details
For each file, show:
- Total downloads
- List of download accounts (email + timestamps)
- Download count per recipient
- Geographic distribution (optional)
- Time-based download graph

---

## 🛠️ TECHNICAL REQUIREMENTS

### Technology Stack Recommendations

**Backend Options** (choose most appropriate):
1. **Go** (like Gokapi) - Fast, single binary, cross-platform
2. **Python + Flask/FastAPI** - Easy to maintain, good libraries
3. **Node.js + Express** - Modern, good ecosystem

**Frontend**:
- Modern HTML5 + CSS3 + JavaScript
- Bootstrap 5 or Tailwind CSS for responsive design
- Vanilla JS or lightweight framework (Alpine.js, Htmx)
- Progressive enhancement approach

**Database**:
- SQLite for simplicity (embedded)
- Optional: PostgreSQL for larger deployments
- Schema must support:
  - Users (admin + regular)
  - Files (metadata, hashes, quotas)
  - Download accounts (email, password hash, timestamps)
  - Download logs (tracking table)
  - Storage folders/categories

**Storage**:
- Local filesystem by default
- Files stored with hash-based naming
- Metadata in database

### Cross-Platform Support
- **Must run on**:
  - Linux (primary)
  - Windows (for potential XProtect server deployment)
  - macOS (development)
- Single binary deployment preferred (if using Go)
- Docker container support mandatory
- Docker Compose with volumes for data persistence

### Configuration
- Environment variables or config file for:
  - Server address/hostname (must be configurable for each installation)
  - Port
  - Database path
  - Storage path
  - Email SMTP settings
  - Admin credentials (initial setup)
  - Session secrets
  - File size limits
  - Default expiration policies

### Security Requirements
- HTTPS support (reverse proxy recommended)
- Password hashing (bcrypt/argon2)
- CSRF protection
- Rate limiting on uploads/downloads
- File type validation
- Virus scanning integration option (ClamAV)
- Session management
- Secure random hash generation for file links

---

## 📦 DEPLOYMENT

### Installation Methods

1. **Docker** (Primary Method)
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v ./data:/data \
     -v ./uploads:/uploads \
     -e SERVER_URL=https://files.manvarg.se \
     -e ADMIN_EMAIL=ulf@manvarg.se \
     manvarg/sharecare:latest
   ```

2. **Docker Compose** (Recommended)
   ```yaml
   version: '3.8'
   services:
     manvarg-sharecare:
       image: manvarg/sharecare:latest
       ports:
         - "8080:8080"
       volumes:
         - ./data:/data
         - ./uploads:/uploads
       environment:
         - SERVER_URL=https://files.manvarg.se
         - ADMIN_EMAIL=ulf@manvarg.se
   ```

3. **Binary Installation** (Linux/Windows)
   - Download binary for platform
   - Create config file
   - Run as system service

### Multi-Instance Support
- Each installation is independent
- Server URL must be configurable per instance
- Allows deployment on:
  - Manvarg's own servers (main service)
  - Customer's XProtect servers (future option)
  - Customer's own infrastructure

---

## 🚀 FEATURES BREAKDOWN

### Phase 1 (MVP - Must Have)
- [ ] Admin login and user management
- [ ] User login and authentication
- [ ] File upload with drag-and-drop
- [ ] Random hash-based link generation
- [ ] Two link types: authenticated + direct
- [ ] Download with email + password creation (authenticated mode)
- [ ] Download without auth (direct mode)
- [ ] Basic expiration (downloads count OR time)
- [ ] Download tracking (who, when, how many)
- [ ] Email sending (with link)
- [ ] User dashboard (my files, usage)
- [ ] Admin dashboard (all files, all users)
- [ ] Storage quota per user
- [ ] Manvarg Sharecare branding throughout
- [ ] Docker deployment
- [ ] Basic documentation

### Phase 2 (Enhanced - Should Have)
- [ ] Multi-folder/category support with quotas
- [ ] File deduplication
- [ ] Advanced analytics dashboard
- [ ] Export reports (CSV/PDF)
- [ ] File preview (images, PDFs)
- [ ] Bulk operations
- [ ] API for programmatic uploads
- [ ] Webhook notifications
- [ ] Custom email templates
- [ ] Two-factor authentication for admins

### Phase 3 (Advanced - Nice to Have)
- [ ] S3-compatible storage backend option
- [ ] Virus scanning integration
- [ ] End-to-end encryption option
- [ ] Mobile app
- [ ] LDAP/Active Directory integration
- [ ] Advanced access controls (groups, permissions)
- [ ] Audit logs
- [ ] Retention policies

---

## 📋 USER WORKFLOWS

### Workflow 1: Upload and Share with Authentication
1. User logs into Manvarg Sharecare
2. Drags and drops video file from XProtect export
3. Sets expiration: 5 downloads OR 14 days
4. Chooses "Require download authentication"
5. Enters recipient email
6. System sends email with link
7. Recipient clicks link, creates download account (email + password)
8. Downloads file
9. User can see in dashboard: "Downloaded by john@customer.com at 2024-11-08 14:30"

### Workflow 2: Quick Share (Direct Link)
1. User logs in
2. Uploads file
3. Sets expiration: 3 days
4. Chooses "Direct link (no authentication)"
5. Copies link, shares via Teams/email manually
6. Recipient downloads immediately
7. User sees download count in dashboard

### Workflow 3: Admin Creates Customer Account
1. Admin logs into admin panel
2. Creates new user: "Camera Operator at Company X"
3. Sets quota: 10GB
4. Assigns to folder: "premium-customers"
5. System generates random password, emails to customer
6. Customer logs in, changes password, starts uploading

---

## 🎨 UI/UX GUIDELINES

### Design Principles
- **Simplicity First**: Like WeTransfer, minimal clicks to share
- **Professional**: Enterprise-grade, not playful
- **Trust Signals**: Security badges, encryption mentions, Prudencia branding
- **Mobile Responsive**: Works on tablets/phones
- **Accessibility**: WCAG 2.1 AA compliance

### Key Pages

1. **Landing/Login Page**
   - Manvarg logo prominent
   - Clean login form
   - Optional: Marketing copy about secure file sharing

2. **User Dashboard**
   - Upload area (drag-drop zone)
   - List of active files
   - Storage usage indicator
   - Quick stats

3. **Admin Dashboard**
   - User management
   - System stats
   - All files view
   - Storage overview

4. **Download Page**
   - Manvarg branding
   - File info (name, size)
   - For authenticated: Login/create account form
   - Download button
   - Expiration notice

5. **Email Template**
   - Manvarg Sharecare header
   - Clear call-to-action
   - File details
   - Professional footer

---

## 📚 DOCUMENTATION REQUIREMENTS

### README.md
- Project description
- Features overview
- Quick start guide
- Docker installation
- Configuration options
- Screenshots

### INSTALLATION.md
- Detailed setup instructions
- System requirements
- Docker/Docker Compose
- Binary installation
- Reverse proxy configuration (nginx/Apache examples)
- SSL/TLS setup

### CONFIGURATION.md
- All configuration options
- Environment variables
- Config file format
- Email setup (SMTP)
- Storage backends

### API.md
- REST API documentation
- Authentication
- Endpoints
- Request/response examples
- Error codes

### USER_GUIDE.md
- How to upload files
- How to share links
- Understanding expiration
- Managing downloads
- FAQ

### ADMIN_GUIDE.md
- User management
- Quota management
- Monitoring usage
- Troubleshooting

---

## 🔒 SECURITY CONSIDERATIONS

### File Security
- Files stored outside web root
- Access only through application logic
- No directory listing
- Hash-based naming prevents enumeration

### Authentication Security
- Passwords hashed with bcrypt (cost factor 12+)
- Session tokens with CSRF protection
- Rate limiting on login attempts
- Optional 2FA for admins

### Data Privacy
- GDPR compliance considerations
- Email addresses stored encrypted (optional)
- Download logs retention policy (configurable)
- Right to deletion

### Infrastructure Security
- Run as non-root user
- Minimal container image
- Regular security updates
- Input validation
- Output encoding

---

## 🎁 DIFFERENTIATORS FROM GOKAPI

### What We Keep from Gokapi
- Random hash-based download links
- Expiration by downloads and/or time
- Simple, clean interface
- Self-hosted
- Docker deployment

### What We Change/Add
1. **Download Authentication**: Recipients can create accounts to download
2. **Download Tracking**: Know exactly who downloaded what and when
3. **Multi-tenant Folders**: Different quotas for different customer tiers
4. **Complete Branding**: Manvarg Sharecare throughout
5. **Simpler**: Focus on core file sharing, remove advanced features
6. **Customer Focus**: Built specifically for surveillance system customers
7. **Email Integration**: Built-in, not optional
8. **Two Link Modes**: Authenticated vs direct

---

## 💼 BUSINESS CONTEXT

### Target Users
- Manvarg customers with video surveillance service agreements
- Primarily Milestone XProtect users (Windows-based)
- May expand to OpenEye customers
- Internal Manvarg staff

### Use Cases
1. **Video Export Sharing**: Export from XProtect, upload to Manvarg Sharecare, share with authorities/management
2. **Documentation Sharing**: Share system manuals, reports with customers
3. **Evidence Chain**: Trackable video downloads for legal purposes
4. **Customer Convenience**: No need for external services like WeTransfer
5. **Brand Experience**: Reinforce Manvarg as full-service security provider

### Success Metrics
- Number of active customer accounts
- Files shared per month
- Storage utilization
- User satisfaction (from feedback)
- Reduction in external file sharing service usage

---

## 🛣️ ROADMAP

### Version 1.0 (Launch - 3 months)
- All Phase 1 features
- Docker deployment
- Basic documentation
- Manvarg.se deployment

### Version 1.5 (6 months)
- Phase 2 features
- API for integrations
- Enhanced analytics
- Mobile-optimized interface

### Version 2.0 (12 months)
- Phase 3 features
- XProtect plugin (if feasible)
- Enterprise features (LDAP, SSO)
- Multi-language support (Swedish + English)

---

## 🤝 CONTRIBUTION GUIDELINES

### Code Style
- Follow language-specific conventions
- Clear variable names
- Comments for complex logic
- README per module

### Git Workflow
- Main branch: stable releases
- Develop branch: active development
- Feature branches: new features
- Semantic versioning (MAJOR.MINOR.PATCH)

### Testing
- Unit tests for core logic
- Integration tests for API
- Manual testing checklist
- Docker testing

---

## 📄 LICENSE

**Open Source License**: Choose one of:
- MIT License (most permissive)
- Apache 2.0 (patent protection)
- AGPL v3 (if you want derivative works to remain open source)

**Recommendation**: MIT or Apache 2.0 for maximum adoption while keeping it open source.

---

## 🚦 GETTING STARTED (For Claude Code)

### Build Order
1. Set up project structure and configuration system
2. Implement authentication system (admin + user + download accounts)
3. Build file upload and storage system
4. Create hash-based link generator
5. Implement download tracking
6. Build user dashboard
7. Build admin dashboard
8. Implement email sending
9. Add quota management
10. Apply Prudencia Security branding
11. Create Docker deployment
12. Write documentation
13. Add tests

### Technology Recommendation
**Use Go** (like Gokapi) for:
- Single binary deployment (easy for Windows)
- Excellent performance
- Strong standard library
- Cross-platform
- Active ecosystem

Alternative: **Python + Flask** if easier to maintain

### File Structure (Example for Go)
```
manvarg-sharecare/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── storage/
│   ├── database/
│   ├── email/
│   ├── handlers/
│   └── models/
├── web/
│   ├── templates/
│   ├── static/
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│   │       └── manvarg-logo.svg
├── config/
│   └── config.example.yaml
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── docs/
│   ├── README.md
│   ├── INSTALLATION.md
│   ├── CONFIGURATION.md
│   └── API.md
├── scripts/
│   └── setup.sh
├── tests/
├── .gitignore
├── LICENSE
├── Makefile
└── go.mod
```

---

## 📞 SUPPORT & CONTACT

**Project Maintainer**: Ulf Holmström @ Manvarg
**Contact**: ulf@manvarg.se
**Repository**: https://github.com/Frimurare/Sharecare
**Documentation**: https://github.com/Frimurare/Sharecare/wiki

---

## ✅ DEFINITION OF DONE

The project is considered "launch ready" when:

- [ ] Admin can create users with quotas
- [ ] Users can upload files and create two types of links
- [ ] Download authentication works (email + password)
- [ ] Download tracking shows who downloaded what
- [ ] Email sending works with Manvarg branding
- [ ] Manvarg Sharecare logo and colors throughout
- [ ] Expiration (downloads + time) works
- [ ] User dashboard shows storage usage
- [ ] Admin dashboard shows system overview
- [ ] Docker deployment tested and working
- [ ] Documentation complete (README, INSTALLATION, USER_GUIDE)
- [ ] Server URL is configurable per installation
- [ ] Works on Linux and Windows
- [ ] Security audit passed (basic)
- [ ] Performance tested (100+ concurrent downloads)
- [ ] Mobile responsive
- [ ] No critical bugs

---

## 💡 NOTES FOR CLAUDE CODE

- **Start Simple**: Build MVP first, then add features
- **Security First**: This handles potentially sensitive video evidence
- **Test on Windows**: XProtect runs on Windows, test there
- **Email Early**: Don't leave email integration for last
- **Branding Everywhere**: Every page should scream "Manvarg Sharecare"
- **Think Multi-Instance**: Each deployment is independent
- **Document as You Go**: Don't leave docs for the end

**Most Important**: This needs to be production-ready, not a prototype. It will handle real customer files and potentially legal evidence from surveillance systems.

---

## 🎯 SUCCESS VISION

Imagine a Manvarg customer who just exported a video from their XProtect system showing a break-in. They:

1. Log into `files.manvarg.se` with their customer account
2. Drag the 2GB video file into the browser
3. Set it to expire after 5 downloads or 30 days
4. Enter the police officer's email
5. Click "Send"
6. The officer receives a professional email with Manvarg branding
7. Opens the link, creates a download account (email + password)
8. Downloads the video
9. Customer sees: "Downloaded by officer.svensson@police.se on 2024-11-08"
10. Customer feels confident using Manvarg's complete security solution

**That's the goal.**

---

**END OF PROJECT BRIEF**
