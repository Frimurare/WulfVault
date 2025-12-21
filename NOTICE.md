# Attribution and Acknowledgments

## WulfVault - Enterprise File Sharing Platform

WulfVault is a professional-grade, self-hosted file sharing platform built from the ground up with enterprise features and security in mind.

## Key Features

- **Multi-user system** - Role-based access (Super Admin, Admin, Users, Download Accounts)
- **Email integration** - SMTP/Brevo support, email sharing, audit logs
- **Two-Factor Authentication** - TOTP with backup codes and recovery system
- **Download account system** - Separate authentication for recipients with self-service portal
- **File request portals** - Upload request links for collecting files
- **Comprehensive audit system** - Download logs, email logs, IP tracking, action logging
- **Branding system** - Custom logos, colors, company name, CSS customization
- **Storage quota management** - Per-user and per-team quotas with usage tracking
- **Password management** - Self-service reset via email with secure tokens
- **Admin dashboards** - System-wide analytics, glassmorphic design, real-time statistics
- **Soft deletion** - Trash system with configurable retention (1-365 days)
- **Team collaboration** - Multi-team file sharing with role-based permissions
- **Advanced pagination** - File list management with configurable items per page (5-250)
- **File descriptions** - Comments and notes on shared files with search integration
- **Chunked uploads** - Large file support with automatic retry and progress tracking
- **Duplicate detection** - Identify and manage duplicate files across the system
- **GDPR compliance** - Data export, account deletion, audit trails

## Code Statistics

- **Total:** 33,455+ lines of Go code (as of v6.2.3)
- **Architecture:** Clean, modular design with clear separation of concerns
- **Database:** SQLite with comprehensive schema and migrations
- **Frontend:** Modern JavaScript with responsive design
- **Backend:** Go 1.23+ with efficient handlers and middleware

## License

This project is licensed under the **AGPL-3.0** license.

**Why AGPL-3.0?**
The GNU Affero General Public License (AGPL-3.0) is specifically designed to prevent proprietary use of open-source software in network services. Key protections:

- **Network copyleft:** Companies running WulfVault as a SaaS must share their source code
- **Attribution protection:** All modifications must credit the original author
- **Community benefit:** Improvements must be contributed back to the community
- **Anti-exploitation:** Prevents "taking without giving back" in cloud deployments

This ensures that WulfVault remains free and open-source, even when used to provide commercial services.

See [LICENSE](LICENSE) for the full license text.

## Author

**Copyright © 2025 Ulf Holmström (Frimurare)**

WulfVault is developed and maintained with a focus on security, compliance, and user experience.
