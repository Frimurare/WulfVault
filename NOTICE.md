# Attribution and Acknowledgments

## Architecturally Inspired by Gokapi

WulfVault is architecturally inspired by **Gokapi** by Forceu, but represents a complete rewrite (~95% new code).

- **Original Project:** https://github.com/Forceu/Gokapi
- **License:** AGPL-3.0
- **Copyright:** Forceu and contributors

We thank the Gokapi team for their excellent work that inspired the foundational architecture of temporary file sharing with expiration.

## WulfVault Enhancements (Complete Rewrite)

WulfVault is a complete rewrite that adds extensive enterprise features:

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

**Code Statistics:**
- Total: 33,455 lines of Go code (as of v6.1.8)
- Gokapi imports in production code: 0
- Conceptual similarity: ~10% (basic data models, database schema foundation)
- New code: ~90% (all HTTP handlers, database layer, email, 2FA, admin system, teams, pagination)

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

