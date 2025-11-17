# Attribution and Acknowledgments

## Architecturally Inspired by Gokapi

WulfVault is architecturally inspired by **Gokapi** by Forceu, but represents a complete rewrite (~95% new code).

- **Original Project:** https://github.com/Forceu/Gokapi
- **License:** AGPL-3.0
- **Copyright:** Forceu and contributors

We thank the Gokapi team for their excellent work that inspired the foundational architecture of temporary file sharing with expiration.

## WulfVault Enhancements (Complete Rewrite)

WulfVault is a complete rewrite that adds extensive enterprise features:

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

This project is licensed under the **AGPL-3.0** license.

**Why AGPL-3.0?**
The GNU Affero General Public License (AGPL-3.0) is specifically designed to prevent proprietary use of open-source software in network services. Key protections:

- **Network copyleft:** Companies running WulfVault as a SaaS must share their source code
- **Attribution protection:** All modifications must credit the original author
- **Community benefit:** Improvements must be contributed back to the community
- **Anti-exploitation:** Prevents "taking without giving back" in cloud deployments

This ensures that WulfVault remains free and open-source, even when used to provide commercial services.

See [LICENSE](LICENSE) for the full license text.

