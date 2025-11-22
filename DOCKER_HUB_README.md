# 🐺 WulfVault — Self-Hosted Enterprise File Sharing

**Secure. Audited. Brandable. Mobile-ready. AGPL-3.0 • Built in Go**

WulfVault is a high-performance, self-hosted file transfer platform for organizations who need security, auditability, large file support and full control over their data. Not Dropbox. Not sync. This is a secure large-file delivery system — closer to a self-hosted Sprend/WeTransfer with enterprise-grade features.

## 🚀 Latest Release: v4.7.6 Galadriel

### What's New in v4.7.6

**Email Provider Activation Controls**
- Explicit "Make Active" buttons for switching between Brevo and SMTP
- Configure multiple email providers and choose which one to use
- Clear visual feedback and audit logging

**Plain SMTP Support**
- Works with MailHog and test SMTP servers without TLS
- Custom plain SMTP implementation for development/testing
- Full MIME multipart support (text + HTML emails)

**SMTP Settings Fixes**
- Fixed settings disappearing after page refresh
- Fixed TLS checkbox not reflecting saved state
- Fixed port reverting to default instead of saved value
- All SMTP settings now properly persist

[View Full Changelog](https://github.com/Frimurare/WulfVault/blob/main/CHANGELOG.md)

---

## 🔑 Key Features

### File Sharing
- Upload large files (5GB+)
- Direct or authenticated downloads
- Password-protected links
- Time-based or download-count expiration
- File Request portals (receive uploads)
- SMTP/Brevo email notifications
- Full metadata & per-file download history

### 📱 Mobile-Optimized UI
- 100% responsive design
- Works flawlessly on phones/tablets
- Full admin + user dashboards available on mobile

### 👥 Users, Teams & Quotas
- Super Admin / Admin / Regular User roles
- Teams with shared access & permissions
- Per-user and per-team storage quotas
- Download-only accounts for external users

### 📊 Audit Logging
- Complete audit trail (uploads, downloads, logins, deletes)
- CSV export for compliance
- Configurable retention (90 days - 10 years)
- Optional GDPR-aware IP logging

### 🔐 Security
- TOTP Two-Factor Authentication (2FA)
- bcrypt password hashing
- Randomized download tokens (128-bit entropy)
- No directory listing or enumeration
- GDPR-compliant user self-deletion & data export
- TLS/HTTPS support via reverse proxy

### 🎨 Branding
- Custom logo upload
- Custom primary and secondary colors
- Branded download pages for recipients
- Custom company name throughout interface

---

## 🐋 Quick Start

### Pull Image

```bash
docker pull frimurare/wulfvault:latest
```

### Run with Docker

```bash
docker run -d \
  --name wulfvault \
  -p 8080:8080 \
  -v ./data:/data \
  -v ./uploads:/uploads \
  -e SERVER_URL=https://files.yourdomain.com \
  frimurare/wulfvault:latest
```

### Run with Docker Compose

```yaml
version: '3.8'
services:
  wulfvault:
    image: frimurare/wulfvault:latest
    container_name: wulfvault
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./uploads:/uploads
    environment:
      - SERVER_URL=https://files.yourdomain.com
      - MAX_FILE_SIZE_MB=5000
      - DEFAULT_QUOTA_MB=10000
    restart: unless-stopped
```

Then start:
```bash
docker compose up -d
```

### First Login

Server runs at: `http://localhost:8080`

**Initial admin credentials:**
- Email: `admin@wulfvault.local`
- Password: `WulfVaultAdmin2024!`

⚠️ **Change the password immediately after first login!**

---

## 📋 Available Tags

- `latest` - Latest stable release (currently v4.7.6 Galadriel)
- `4.7.6-Galadriel` - Specific version tag
- `main-<hash>` - Latest commit from main branch

---

## 🌍 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_URL` | Public URL of the server | `http://localhost:8080` |
| `PORT` | Server port | `8080` |
| `DATA_DIR` | Data directory for database | `/data` |
| `UPLOADS_DIR` | Directory for uploaded files | `/uploads` |
| `MAX_FILE_SIZE_MB` | Maximum file size in MB | `2000` (2 GB) |
| `DEFAULT_QUOTA_MB` | Default storage quota per user | `5000` (5 GB) |
| `SESSION_TIMEOUT_HOURS` | Session expiration time | `24` |
| `TRASH_RETENTION_DAYS` | Days to keep deleted files | `5` |

---

## 💡 Why WulfVault?

✅ **Self-Hosted** - Your data stays on your infrastructure
✅ **No Subscriptions** - No per-user or per-transfer costs
✅ **Complete Audit Trail** - Know exactly who downloaded what and when
✅ **GDPR Compliant** - Built-in compliance features (Grade: A-, 94%)
✅ **Enterprise Features** - Teams, quotas, branding, 2FA, audit logs
✅ **Large File Support** - Files up to 5GB+ (configurable)
✅ **Mobile Ready** - Fully responsive UI works on all devices

**Perfect for:** Law enforcement agencies, healthcare providers, legal firms, creative agencies, government departments, educational institutions, and any organization handling sensitive or large files.

---

## 📚 Documentation

- **GitHub Repository:** https://github.com/Frimurare/WulfVault
- **User Guide:** https://github.com/Frimurare/WulfVault/blob/main/USER_GUIDE.md
- **Installation Guide:** https://github.com/Frimurare/WulfVault/blob/main/INSTALLATION.md
- **API Documentation:** https://github.com/Frimurare/WulfVault/blob/main/docs/API.md
- **GDPR Compliance:** https://github.com/Frimurare/WulfVault/tree/main/gdpr-compliance

---

## 🔒 Production Deployment

For production use, we recommend:

1. **Use HTTPS** - Deploy behind reverse proxy (nginx/Caddy) with SSL
2. **Set strong passwords** - Change default admin password immediately
3. **Enable 2FA** - Configure two-factor authentication for admins
4. **Configure backups** - Regularly backup `./data` and `./uploads`
5. **Use firewall** - Only expose ports 80/443
6. **Monitor logs** - Watch for suspicious activity

See [INSTALLATION.md](https://github.com/Frimurare/WulfVault/blob/main/INSTALLATION.md) for detailed deployment guides.

---

## 📊 GDPR Compliance

WulfVault is **GDPR-compliant** with privacy-by-design features:

✅ Right of Access - Users can export all their data
✅ Right to Erasure - Account deletion with soft delete
✅ Right to Rectification - Users can update their profile
✅ Right to Data Portability - JSON export of personal data
✅ Audit Logging - Complete activity tracking
✅ Data Minimization - Only necessary data collected
✅ Encryption - TLS/HTTPS for all connections

**Compliance Grade: A- (94%)**

---

## 🛠️ Support & Contributing

- **Issues:** https://github.com/Frimurare/WulfVault/issues
- **Discussions:** https://github.com/Frimurare/WulfVault/discussions
- **Contributing:** Pull requests welcome!

---

## 📄 License

**AGPL-3.0** - This ensures that if anyone uses WulfVault to provide a service over a network (like SaaS), they must share their modifications with the community.

Architecturally inspired by [Gokapi](https://github.com/Forceu/Gokapi) by Forceu.

---

## ⭐ Like it?

Give the project a star on GitHub — it helps visibility and keeps development moving forward!

**[★ Star on GitHub](https://github.com/Frimurare/WulfVault)**
