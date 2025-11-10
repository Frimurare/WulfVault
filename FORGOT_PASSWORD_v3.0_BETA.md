# Forgot Password Feature - Planned for v3.0 Beta

**Status:** Planerad
**Target Release:** Sharecare v3.0 Beta
**Förutsättning:** E-postintegration (v3.0 Alpha 1) måste vara implementerad och testad

---

## Översikt

"Forgot Password"-funktionaliteten gör det möjligt för användare att återställa sitt lösenord via e-post om de glömt det. Detta är en kritisk funktion för användarvänlighet och minskar adminbelastning.

## Funktionella Krav

### 1. Återställningsflöde

#### Steg 1: Begär återställning
- Användaren klickar på "Glömt lösenord?" på inloggningssidan
- Anger sin e-postadress
- Systemet:
  - Verifierar att e-postadressen finns i systemet
  - Genererar en unik återställningstoken (kryptografiskt säker, 32 bytes)
  - Skapar en återställningsförfrågan med 1 timmars utgångsdatum
  - Skickar e-post med återställningslänk
  - **Visar alltid samma meddelande** oavsett om e-posten finns (säkerhet)

#### Steg 2: E-postnotifiering
E-postmeddelandet innehåller:
- Återställningslänk: `https://sharecare.example.com/reset-password/{TOKEN}`
- Utgångstid (1 timme)
- Instruktioner
- Säkerhetsvarning om användaren inte begärt återställning

#### Steg 3: Återställ lösenord
- Användaren klickar på länken i e-postmeddelandet
- Token valideras:
  - Kontrollera att token finns
  - Kontrollera att token inte har utgått
  - Kontrollera att token inte redan har använts
- Om giltig: visa formulär för att ange nytt lösenord
- Om ogiltig: visa felmeddelande med länk till "Begär ny återställning"

#### Steg 4: Bekräfta nytt lösenord
- Användaren anger nytt lösenord (min 8 tecken)
- Användaren bekräftar lösenordet
- Systemet:
  - Validerar lösenordsstyrka
  - Hashar lösenordet med bcrypt
  - Uppdaterar användarkontot
  - Markerar återställningstoken som använd
  - Raderar alla aktiva sessioner för användaren
  - Skickar bekräftelsemail
  - Omdirigerar till inloggningssidan med framgångsmeddelande

### 2. Säkerhetskrav

#### Token-säkerhet
- **Token-generering:** 32 bytes kryptografiskt säker slumpdata (hex-enkodad = 64 tecken)
- **Token-lagring:** Hashas med bcrypt innan lagring (som lösenord)
- **Utgångstid:** 1 timme från skapande
- **Single-use:** Token kan endast användas en gång
- **Rate limiting:** Max 3 återställningsförfrågningar per e-postadress per timme

#### Lösenordskrav
- **Minsta längd:** 8 tecken
- **Rekommenderat:** 12+ tecken med mix av stora/små bokstäver, siffror och specialtecken
- **Validering:** Kolla mot vanliga lösenord (top 1000)
- **Historik:** Förhindra återanvändning av de 3 senaste lösenorden

#### Anti-brute force
- **Rate limiting på återställningssidan:** Max 5 försök per IP per timme
- **Account lockout:** Efter 5 misslyckade försök, lås kontot i 15 minuter
- **Logging:** Logga alla återställningsförsök (lyckade och misslyckade)

#### Information disclosure prevention
- Visa alltid samma meddelande oavsett om e-postadressen finns
- Logga inte om e-post finns/inte finns i offentliga loggar
- Dölj tokendetaljer i felmeddelanden

## Teknisk Implementation

### Databas-ändringar

#### Ny tabell: PasswordResets

```sql
CREATE TABLE IF NOT EXISTS PasswordResets (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    UserId INTEGER NOT NULL,
    TokenHash TEXT NOT NULL UNIQUE,
    CreatedAt INTEGER NOT NULL,
    ExpiresAt INTEGER NOT NULL,
    IsUsed INTEGER DEFAULT 0,
    UsedAt INTEGER DEFAULT 0,
    IpAddress TEXT,
    UserAgent TEXT,
    FOREIGN KEY (UserId) REFERENCES Users(Id)
);

CREATE INDEX IF NOT EXISTS idx_passwordresets_tokenhash ON PasswordResets(TokenHash);
CREATE INDEX IF NOT EXISTS idx_passwordresets_expiresat ON PasswordResets(ExpiresAt);
```

#### Uppdatera Users-tabellen

```sql
-- Lägg till kolumner för lösenordshistorik och account lockout
ALTER TABLE Users ADD COLUMN PasswordHistory TEXT; -- JSON array med hashar
ALTER TABLE Users ADD COLUMN AccountLockedUntil INTEGER DEFAULT 0;
ALTER TABLE Users ADD COLUMN FailedLoginAttempts INTEGER DEFAULT 0;
```

### API-endpoints

#### POST /api/auth/forgot-password
Request:
```json
{
    "email": "user@example.com"
}
```

Response (alltid 200 OK):
```json
{
    "message": "Om e-postadressen finns i systemet har ett återställningsmail skickats."
}
```

#### GET /reset-password/{token}
- Renderar återställningsformulär om token är giltig
- Visar felmeddelande om token är ogiltig/utgången

#### POST /api/auth/reset-password
Request:
```json
{
    "token": "abc123...",
    "newPassword": "newSecurePassword123",
    "confirmPassword": "newSecurePassword123"
}
```

Response:
```json
{
    "success": true,
    "message": "Lösenordet har återställts. Du kan nu logga in."
}
```

### E-postmallar

#### Återställningsmail
**Ämne:** "Återställ ditt Sharecare-lösenord"

**HTML-version:**
```html
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button {
            display: inline-block; padding: 12px 24px;
            background: #2563eb; color: white;
            text-decoration: none; border-radius: 5px;
        }
        .warning { background: #fff3cd; padding: 15px; border-left: 4px solid #ffc107; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Återställ ditt lösenord</h2>
        <p>Vi har fått en begäran om att återställa lösenordet för ditt Sharecare-konto.</p>

        <p>Klicka på knappen nedan för att skapa ett nytt lösenord:</p>

        <p>
            <a href="{{.ResetLink}}" class="button">Återställ lösenord</a>
        </p>

        <p>Eller kopiera och klistra in denna länk i din webbläsare:<br/>
        <code>{{.ResetLink}}</code></p>

        <p><strong>Denna länk går ut om 1 timme.</strong></p>

        <div class="warning">
            <p><strong>⚠️ Observera:</strong></p>
            <ul>
                <li>Om du inte begärt denna återställning, ignorera detta mail</li>
                <li>Ditt nuvarande lösenord fungerar fortfarande</li>
                <li>Dela aldrig denna länk med någon</li>
            </ul>
        </div>
    </div>
</body>
</html>
```

#### Bekräftelsemail (efter återställning)
**Ämne:** "Ditt Sharecare-lösenord har återställts"

**HTML-version:**
```html
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h2>✓ Lösenord återställt</h2>
        <p>Ditt Sharecare-lösenord har återställts.</p>

        <p><strong>Tidpunkt:</strong> {{.Timestamp}}<br/>
        <strong>IP-adress:</strong> {{.IpAddress}}</p>

        <p>Om du inte utförde denna åtgärd, kontakta omedelbart en administratör.</p>

        <p><a href="{{.LoginUrl}}">Logga in nu</a></p>
    </div>
</body>
</html>
```

### Frontend-sidor

#### Glömt lösenord-sida
- Minimalistisk design
- Ett e-postfält
- "Skicka återställningslänk"-knapp
- Länk tillbaka till inloggning

#### Återställningsformulär
- Lösenordsfält med styrkeindikator
- Bekräftelselösenordsfält
- Realtidsvalidering
- "Återställ lösenord"-knapp

### Cleanup-jobb

Skapa automatiskt jobb som körs varje timme:
```go
// Radera utgångna återställningstokens
func CleanupExpiredPasswordResets() {
    now := time.Now().Unix()
    database.DB.Exec(`
        DELETE FROM PasswordResets
        WHERE ExpiresAt < ? OR
              (IsUsed = 1 AND UsedAt < ?)
    `, now, now - (7 * 24 * 60 * 60)) // Radera använda tokens äldre än 7 dagar
}
```

## Testplan

### Enhetstester
- [ ] Token-generering och hashning
- [ ] Token-validering (giltig, utgången, använd)
- [ ] Lösenordsstyrkevalidering
- [ ] Lösenordshistorikskontroll
- [ ] Rate limiting-logik

### Integrationstester
- [ ] Fullständigt återställningsflöde
- [ ] E-postutskick
- [ ] Session-invalidering efter återställning
- [ ] Account lockout efter misslyckade försök
- [ ] Cleanup-jobb

### Säkerhetstester
- [ ] Brute force-försök på återställningslänkar
- [ ] Email enumeration prevention
- [ ] Token reuse prevention
- [ ] XSS och CSRF-skydd
- [ ] SQL injection-skydd

### Användartester
- [ ] Användarvänlighet av formulär
- [ ] Tydlighet i felmeddelanden
- [ ] E-postmallarnas läsbarhet
- [ ] Mobilvänlighet

## Dokumentation

### Användarmanual
- Steg-för-steg-guide för lösenordsåterställning
- Vanliga problem och lösningar
- Säkerhetstips

### Administratörsmanual
- Hur man manuellt återställer användarlösenord
- Hur man kontrollerar återställningsloggar
- Hur man justerar säkerhetsinställningar

## Framtida Förbättringar (v3.1+)

- [ ] 2FA-stöd vid återställning
- [ ] SMS-baserad återställning som alternativ
- [ ] Säkerhetsfrågor som extra verifiering
- [ ] Proaktiva varningar om svaga lösenord
- [ ] Integrering med Have I Been Pwned API

---

## Slutsats

Forgot Password-funktionaliteten är kritisk för v3.0 Beta och bör implementeras omedelbart efter att v3.0 Alpha 1 (e-postintegration) har testats och verifierats i produktion.

Uppskattad tid: 16-20 timmar
Prioritet: Hög
Beroenden: E-postintegration (v3.0 Alpha 1)
