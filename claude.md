# Claude Code - WulfVault Development Guide

Denna fil innehåller viktig information för att fortsätta utveckla WulfVault med Claude Code i framtida sessioner.

---

## Repository Information

**GitHub Repository:** https://github.com/Frimurare/WulfVault
**Current Version:** v6.0.1 BloodMoon 🌙
**Main Branch:** `main` (skyddad - kräver PRs)
**Latest Release:** https://github.com/Frimurare/WulfVault/releases/tag/v6.0.1

### Repository Structure
```
WulfVault/
├── cmd/server/          # Main application entry point
│   └── main.go         # Version constant här
├── internal/
│   ├── server/         # HTTP handlers och routing
│   ├── auth/           # Autentisering och sessions
│   ├── database/       # SQLite databas operations
│   ├── models/         # Data models
│   └── config/         # Konfiguration
├── web/static/         # Frontend assets (JS, CSS)
├── data/               # SQLite databas och logs
├── uploads/            # Uppladdade filer
├── USER_GUIDE.md       # Användardokumentation
├── README.md           # Projektbeskrivning
└── UPDATE_HISTORY.md   # Changelog
```

---

## Credentials & Access

### GitHub Token
**Det är OK att fråga Ulf efter GitHub token när du behöver det.**

Token format: `ghp_...` (classic personal access token)

Exempel på hur token används:
```bash
export GH_TOKEN="<fråga-ulf-efter-token>"
```

Token används för:
- `gh pr create` - Skapa pull requests
- `gh pr merge` - Merga pull requests
- `gh release create` - Skapa releases
- `git push origin --delete <branch>` - Radera branches

### Sudo Access
**Det är OK att fråga Ulf efter sudo-lösenord när systemd-operationer behövs.**

Systemd service: `/etc/systemd/system/wulfvault.service`
Service user: `ulf`
Log fil: `/var/log/wulfvault.log`

**OBS:** I de flesta fall behövs INTE sudo - vi kan starta/stoppa processen manuellt.

---

## Development Workflow

### 1. Branch Protection Rules
Main branch är skyddad och kräver pull requests. **Pusha ALDRIG direkt till main.**

**Standard workflow:**
```bash
# 1. Gör ändringar på main (lokalt)
git add -A
git commit -m "Commit message"

# 2. Skapa feature branch
git checkout -b feature/beskrivning

# 3. Pusha branch
git push -u origin feature/beskrivning

# 4. Skapa och merga PR
export GH_TOKEN="<token>"
gh pr create --title "Titel" --body "Beskrivning" --base main
gh pr merge --squash --delete-branch

# 5. Uppdatera lokal main
git checkout main
git pull origin main
```

### 2. Version Bumping

Versioner definieras i `cmd/server/main.go`:
```go
const (
    Version = "6.0.1 BloodMoon 🌙"
)
```

**Versionsschema:**
- **Major (X.0.0)**: Stora breaking changes
- **Minor (X.Y.0)**: Nya features, nya endpoints, stora UI-ändringar
- **Patch (X.Y.Z)**: Bugfixar, små förbättringar, dokumentationsuppdateringar

**Kodnamn:** Använd kreativa kodnamn (BloodMoon, FullMoon, Silverbullet, etc.)
**Emoji i version:** OK att använda relevanta emoji (🌙 för BloodMoon, etc.)

### 3. Build & Deploy Process

**Steg 1: Bygg**
```bash
go build -o wulfvault ./cmd/server
```

**Steg 2: Stoppa gammal process**
```bash
pkill -f "WulfVault/wulfvault"
# eller
ps aux | grep wulfvault  # hitta PID
kill -9 <PID>
```

**Steg 3: Starta ny process**
```bash
nohup ./wulfvault >> /var/log/wulfvault.log 2>&1 &
```

**Steg 4: Verifiera**
```bash
ps aux | grep wulfvault | grep -v grep
tail -20 /var/log/wulfvault.log
```

**OBS:** Binary heter `wulfvault` (ingen .exe eller liknande på Linux)

### 4. Commit Message Format

Använd detaljerade commit messages med följande struktur:

```
Title på en rad (kort sammanfattning)

Längre beskrivning av ändringar och varför de gjordes.

## Key Features/Changes
- Bullet point 1
- Bullet point 2

## Technical Details
- Implementation notes
- Files changed

## Benefits/Bug Fixes
- Vad användaren får ut av detta

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### 5. Release Creation

När en version är klar:
```bash
export GH_TOKEN="<token>"
gh release create v6.0.1 --title "WulfVault v6.0.1 BloodMoon 🌙" --notes "
# WulfVault v6.0.1 BloodMoon 🌙

Beskrivning av release...

## New Features
...

## Bug Fixes
...

## Upgrading
...
"
```

---

## Key Files to Know

### Backend (Go)
- `cmd/server/main.go` - Entry point, version, startup logic
- `internal/server/server.go` - Router setup, middleware
- `internal/server/handlers_auth.go` - Login, logout, session handling
- `internal/server/handlers_2fa.go` - Two-factor authentication
- `internal/server/handlers_admin.go` - Admin panel handlers
- `internal/server/handlers_teams.go` - Team file management
- `internal/database/database.go` - Database operations
- `internal/auth/auth.go` - Authentication logic

### Frontend (JavaScript)
- `web/static/js/dashboard.js` - Main dashboard logic, upload handling
- `web/static/css/style.css` - Global styles

### Documentation
- `USER_GUIDE.md` - Omfattande användarguide (1800+ rader)
- `README.md` - Projektöversikt och features
- `UPDATE_HISTORY.md` - Changelog för alla versioner

---

## Common Tasks

### Lägga till en ny feature

1. **Planera** - Diskutera med Ulf vad som ska implementeras
2. **Använd TodoWrite** - Skapa todo-lista för alla steg
3. **Implementera** - Gör ändringar i relevanta filer
4. **Testa** - Bygg och kör lokalt, verifiera funktionalitet
5. **Dokumentera** - Uppdatera USER_GUIDE.md om relevant
6. **Version bump** - Öka version i main.go
7. **Commit & PR** - Följ workflow ovan
8. **Release** - Skapa GitHub release

### Fixa en bugg

1. **Identifiera** - Hitta orsaken (använd Read, Grep)
2. **Fixa** - Gör minimal ändring som löser problemet
3. **Version bump** - Patch version (X.Y.Z+1)
4. **Deploy** - Bygg, stoppa, starta, verifiera
5. **Commit & PR** - Med "Bug fix:" i titel

### Uppdatera dokumentation

Dokumentation är viktig! Uppdatera alltid:
- USER_GUIDE.md när features ändras
- README.md om stora features läggs till
- claude.md (denna fil) när workflow ändras

---

## Database & Storage

### Database Location
SQLite databas: `/home/ulf/WulfVault/data/wulfvault.db`

**Tabeller:**
- `users` - Användare och admins
- `sessions` - Aktiva sessions
- `files` - Uppladdade filer
- `download_accounts` - Externa mottagare
- `teams` - Team management
- `audit_logs` - Alla system-events
- `file_requests` - Upload portaler

### File Storage
Uppladdade filer: `/home/ulf/WulfVault/uploads/`
Filnamn = File ID (UUID)

### Logs
- Server log: `/var/log/wulfvault.log` (använd `tail -f` för live)
- Audit logs: I databas, åtkomst via Admin → Server → View Audit Logs

---

## Environment Variables

Standard miljövariabler (från systemd service):
```bash
SERVER_URL=http://sharecare.dyndns.org:8080  # eller wulfvault.dyndns.org
PORT=8080
DATA_DIR=/home/ulf/WulfVault/data
UPLOADS_DIR=/home/ulf/WulfVault/uploads
MAX_FILE_SIZE_MB=5000
DEFAULT_QUOTA_MB=10000
```

**OBS:** Dessa kan också konfigureras via Admin Settings i web-gränssnittet.

---

## Testing & Verification

### Efter deployment, verifiera:

1. **Process körs:**
   ```bash
   ps aux | grep wulfvault | grep -v grep
   ```

2. **Version är korrekt:**
   ```bash
   tail /var/log/wulfvault.log | grep -i "wulfvault\|version"
   ```

3. **Web UI svarar:**
   ```bash
   curl -I http://localhost:8080/login
   ```

4. **Logga in manuellt:**
   - Öppna http://wulfvault.dyndns.org:8080
   - Verifiera att nya features fungerar
   - Testa både regular user och admin

---

## Recent Features (v6.0.x BloodMoon)

### v6.0.1 BloodMoon 🌙
- **Keep Me Logged In**: Checkbox på login för 30-dagars sessions
- Fungerar med 2FA
- Fungerar för download accounts
- Session-tid: 24h default, 30 dagar om ikryssad

### v6.0.0 BloodMoon
- **Team Files Sorting**: 8 sorteringsalternativ (datum, namn, storlek, ägare)
- **Admin Delete**: Ta bort filer direkt från team files-vyn
- **Empty All Trash**: Töm hela papperskorgen med ett klick
- **150 Upload One-Liners**: Underhållande meddelanden under uppladdning med 💾 emoji
- **Extended Retry**: 30 retries (~5 min) för upload chunks
- **HTTP Status Codes**: Tabell i USER_GUIDE.md med förklaring av status-koder

---

## Communication Style

**Ulf föredrar:**
- Svenska i konversation
- Direkta svar utan överdrivet många emojis
- Tekniska detaljer när det är relevant
- Fråga om credentials när du behöver (GitHub token, sudo-lösen)
- Todo-listor för att hålla koll på uppgifter

**Bra att veta:**
- Ulf är teknisk och förstår kod
- OK att visa kod-snippets och tekniska detaljer
- Förklara "varför" när du gör design-val
- Använd TodoWrite för att tracka progress på större tasks

---

## Troubleshooting

### Build errors
```bash
# Kontrollera Go version
go version  # Bör vara 1.21+

# Rensa moduler och bygg om
go clean -modcache
go mod tidy
go build -o wulfvault ./cmd/server
```

### Process startar inte
```bash
# Kolla loggen för errors
tail -50 /var/log/wulfvault.log

# Kolla om port 8080 är upptagen
ss -tulpn | grep 8080

# Testa att köra direkt (för att se errors)
./wulfvault
```

### Git push rejected
```bash
# Main är skyddad - skapa PR istället
git checkout -b feature/branch-namn
git push -u origin feature/branch-namn
gh pr create --title "..." --body "..." --base main
```

### GitHub token expired
**Fråga Ulf efter ny token!** Det är helt OK.

---

## Next Session Checklist

När du startar en ny session:

1. **Läs denna fil** för att få kontext
2. **Kolla senaste commit**: `git log --oneline -5`
3. **Kolla current version**: `grep Version cmd/server/main.go`
4. **Kolla branches**: `git branch -a`
5. **Kolla om server körs**: `ps aux | grep wulfvault`
6. **Fråga Ulf** vad som ska göras

---

## Important Notes

- **Main branch är skyddad** - använd alltid PRs
- **Fråga om credentials** - GitHub token, sudo-lösen, etc.
- **Version bumps** - Uppdatera version i main.go för alla releases
- **Test before merge** - Bygg och testa lokalt först
- **Dokumentera** - Uppdatera USER_GUIDE.md för användar-synliga features
- **Commit messages** - Använd detaljerade meddelanden med struktur
- **Todo-listor** - Använd TodoWrite för större tasks

---

**Skapad:** 2025-12-09
**Senaste uppdatering:** v6.0.1 BloodMoon 🌙
**Författare:** Claude Code + Ulf Holmström

---

*Denna fil är levande dokumentation - uppdatera den när workflows, struktur eller viktiga detaljer ändras.*
