# Attribution and Acknowledgments

## Architecturally Inspired by Gokapi

Sharecare is architecturally inspired by **Gokapi** by Forceu, but represents a complete rewrite (~95% new code).

- **Original Project:** https://github.com/Forceu/Gokapi
- **License:** AGPL-3.0
- **Copyright:** Forceu and contributors

We thank the Gokapi team for their excellent work that inspired the foundational architecture of temporary file sharing with expiration.

## Sharecare Enhancements (Complete Rewrite)

Sharecare is a complete rewrite that adds extensive enterprise features:

- **Multi-user system** (~11,000 lines) - Role-based access (Super Admin, Admin, Users, Download Accounts)
- **Email integration** (1,042 lines) - SMTP/Brevo support, email sharing, audit logs
- **Two-Factor Authentication** (118 lines) - TOTP with backup codes
- **Download account system** - Separate authentication for recipients with self-service portal
- **File request portals** - Upload request links for collecting files
- **Comprehensive audit system** - Download logs, email logs, IP tracking
- **Branding system** - Custom logos, colors, company name
- **Storage quota management** - Per-user quotas with usage tracking
- **Password management** - Self-service reset via email
- **Admin dashboards** - System-wide analytics and management
- **Soft deletion** - Trash system with configurable retention (1-365 days)

**Code Statistics:**
- Total: 18,016 lines of Go code
- Gokapi imports in production code: 0
- Conceptual similarity: ~15% (basic data models, database schema foundation)
- New code: ~80% (all HTTP handlers, database layer, email, 2FA, admin system)

## License

This project is licensed under the GPL-3.0 license.

See [LICENSE](LICENSE) for the full license text.
