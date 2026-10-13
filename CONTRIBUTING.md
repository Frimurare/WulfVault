# Contributing to WulfVault

Thanks for taking the time to contribute. WulfVault is a self-hosted file
sharing platform written in Go, licensed under the AGPL-3.0. This document
explains how to report problems, propose changes and get a pull request merged.

## Ways to contribute

- **Report a bug** — open a [bug report](https://github.com/Frimurare/WulfVault/issues/new?template=bug_report.yml).
- **Suggest a feature** — open a [feature request](https://github.com/Frimurare/WulfVault/issues/new?template=feature_request.yml).
- **Ask a question or share how you use WulfVault** — use [Discussions](https://github.com/Frimurare/WulfVault/discussions).
- **Improve the docs** — typos, unclear steps and missing examples in `README.md`,
  `INSTALLATION.md`, `USER_GUIDE.md` or `docs/` are all fair game.
- **Fix something** — look for issues labelled `good first issue` or `help wanted`.

**Security problems are not reported in public issues.** See [SECURITY.md](SECURITY.md).

## Before you start a larger change

Open an issue (or comment on an existing one) and describe what you want to do
before writing a lot of code. That avoids duplicate work and makes sure the
change fits the project. Small fixes can go straight to a pull request.

## Development setup

Requirements: Go 1.23 or newer, `make` (optional), Docker (optional).

```bash
git clone https://github.com/Frimurare/WulfVault.git
cd WulfVault
go mod download
go run ./cmd/server          # starts on http://localhost:8080
```

On first start WulfVault creates a super-admin from `ADMIN_EMAIL` /
`ADMIN_PASSWORD`; if they are unset the password is printed once to the log.
Runtime data lands in `./data` and `./uploads` (both gitignored).

Useful targets:

| Command | What it does |
|---------|--------------|
| `make build` / `go build -o wulfvault ./cmd/server` | Build the binary |
| `make test` / `go test ./...` | Run the test suite |
| `make docker-build` | Build the Docker image |
| `go vet ./...` | Static checks |
| `gofmt -l .` | List files that are not gofmt-formatted (should print nothing) |

Project layout:

```
cmd/server/      entry point (version string lives here)
internal/        auth · database · models · server
web/             static assets · templates (served from disk)
docs/            configuration, API, SSO and deployment guides
gdpr-compliance/ compliance templates
```

## Coding guidelines

- Run `gofmt` and `go vet` before committing. Keep the code idiomatic Go; no
  new dependencies without a clear need (the binary is deliberately small and
  built with `CGO_ENABLED=0`).
- Add or update tests for behaviour you change. `go test ./...` must pass.
- Database changes go through the existing migration path in `internal/database`
  and must upgrade an existing database without manual steps.
- Anything a user sees (templates, emails, error messages) is written in plain
  English. Keep branding generic — WulfVault is a product, not a customer
  deployment.
- **Never commit secrets, real user data, customer names or production
  configuration.** Test with throwaway values.
- Update `CHANGELOG.md` under an *Unreleased* heading for user-visible changes
  (the file follows [Keep a Changelog](https://keepachangelog.com/)).
- Do not bump the version in `cmd/server/main.go`; the maintainer does that at
  release time.

## Commit messages and pull requests

- One logical change per pull request. Rebase on `main` before opening it.
- Commit subject in the imperative, max ~72 characters, with a body that
  explains *why* when it is not obvious. Reference issues (`Fixes #123`).
- Fill in the pull request template: what changed, how you tested it, and
  whether docs or the changelog were touched.
- A maintainer reviews every pull request. Expect questions; small follow-up
  commits are fine, they are squashed on merge.

## License and contributor agreement

WulfVault is licensed under the **GNU Affero General Public License v3.0**
(see [LICENSE](LICENSE) and [NOTICE.md](NOTICE.md)). By submitting a
contribution you agree that it is your own work (or that you have the right to
submit it) and that it is licensed under the same AGPL-3.0 terms. No separate
CLA is required; please add yourself to [AUTHORS](AUTHORS) in your first pull
request if you like.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). Be kind.

## Questions

Open a discussion or e-mail **ulf@manvarg.se**.
