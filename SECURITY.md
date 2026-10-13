# Security Policy

WulfVault handles other people's files, so security reports are taken
seriously and handled with priority.

## Supported versions

| Version | Supported |
|---------|-----------|
| 7.2.x   | Yes — security fixes are released as patch versions |
| < 7.2   | No — please upgrade; see the upgrade notes in [CHANGELOG.md](CHANGELOG.md) |

Only the latest minor release receives fixes. Docker images on Docker Hub
(`frimurare/wulfvault`) follow the same versions.

## Reporting a vulnerability

**Please do not open a public issue for security problems.**

Report privately, either way works:

1. **GitHub private vulnerability reporting** (preferred):
   https://github.com/Frimurare/WulfVault/security/advisories/new
2. **E-mail:** ulf@manvarg.se — put `WulfVault security` in the subject.

Include what you can of: affected version, deployment type (binary / Docker),
steps to reproduce, impact, and a proof of concept if you have one. You do not
need to prove exploitability; a plausible description is enough to start.

## What to expect

- Acknowledgement within **3 working days**.
- An initial assessment and a plan (or a request for more detail) within
  **10 working days**.
- A fix for confirmed issues as soon as practical, normally as a patch release
  with a changelog entry and a GitHub security advisory. You will be credited
  unless you prefer not to be.
- Please give us a reasonable window (90 days is the default) before public
  disclosure, and we will keep you informed of progress.

## Scope

In scope: the WulfVault server (`cmd/`, `internal/`, `web/`), the official
Dockerfile and docker-compose setup, and the documentation where it gives
insecure advice.

Out of scope: vulnerabilities in third-party services you integrate (your
identity provider, SMTP server, reverse proxy), issues that require an already
compromised administrator account or host, and missing hardening that the
documentation already tells you to do (TLS termination, firewalling, backups).

## Hardening reminders for operators

Change the first-run admin password, terminate TLS at a reverse proxy, expose
only ports 80/443, keep `./data` and `./uploads` outside the web root and
backed up, enable 2FA for administrators, and keep WulfVault updated. See the
Security section of the [README](README.md#security).
