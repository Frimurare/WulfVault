# 📋 WulfVault Audit Log - Komplett Testguide

## 🚀 Förberedelser

1. **Bygg om servern:**
   ```bash
   cd /path/to/WulfVault
   go build -o wulfvault ./cmd/server
   ```

2. **Starta om tjänsten**
   - Servern måste köras med den nya versionen (4.5.11 Gold)

3. **Öppna Audit Logs:**
   - Logga in som Admin
   - Gå till **Admin → Audit Logs**

## ✅ Vad Ska Loggas och Hur Man Testar

### 🔐 AUTHENTICATION (Autentisering)

| Action | Vad Gör Du | Vad Syns i Loggen |
|--------|------------|-------------------|
| **LOGIN_SUCCESS** | Logga in med korrekt lösenord | `{"email":"din@email.se","success":true}` |
| **LOGIN_FAILED** | Försök logga in med fel lösenord | `{"email":"din@email.se","success":false,"reason":"invalid_credentials"}` |
| **LOGOUT** | Klicka "Logout" | `{"email":"din@email.se"}` |
| **DOWNLOAD_ACCOUNT_LOGIN_SUCCESS** | Logga in som download-konto | `{"email":"download@email.se","success":true,"account_type":"download"}` |

### 📁 FILE OPERATIONS (Filoperationer)

| Action | Vad Gör Du | Vad Syns i Loggen |
|--------|------------|-------------------|
| **FILE_UPLOADED** | Ladda upp en fil | `{"filename":"test.pdf","size":"1048576","requires_auth":"true"}` |
| **FILE_DOWNLOADED** | Ladda ner en fil (autentiserad eller anonym) | `{"filename":"test.pdf","size":"1048576"}` |
| **FILE_DELETED** | Ta bort en fil (flyttas till trash) | `{"filename":"test.pdf","size":"1048576"}` |
| **FILE_RESTORED** | **Admin → Trash** → Klicka "Restore" | `{"filename":"test.pdf","size":"1048576"}` |
| **FILE_PERMANENTLY_DELETED** | **Admin → Trash** → Klicka "Delete Forever" | `{"filename":"test.pdf","size":"1048576"}` |

**Test FILE_PERMANENTLY_DELETED:**
1. Ladda upp en fil
2. Ta bort filen (hamnar i trash)
3. Gå till **Admin → Trash**
4. Klicka **"🗑️ Delete Forever"** på filen
5. Bekräfta varningen
6. ✅ Kolla Audit Logs → Du ska se **FILE_PERMANENTLY_DELETED**

**Test FILE_RESTORED:**
1. Gå till **Admin → Trash**
2. Klicka **"♻️ Restore"** på en fil
3. ✅ Kolla Audit Logs → Du ska se **FILE_RESTORED**

### 👤 USER MANAGEMENT (Användarhantering)

| Action | Vad Gör Du | Vad Syns i Loggen |
|--------|------------|-------------------|
| **USER_CREATED** | **Admin → Dashboard** → Klicka "Create User" → Fyll i formulär → Spara | `{"email":"ny@email.se","name":"Namn","user_level":1}` |
| **USER_UPDATED** | **Admin → Dashboard** → Klicka ✏️ på en användare → Ändra något → Spara | `{"email":"user@email.se","name":"Nytt Namn","user_level":2}` |
| **USER_DELETED** | **Admin → Dashboard** → Klicka 🗑️ på en användare → Bekräfta | `{"email":"user@email.se","name":"Namn"}` |

**Test USER_CREATED:**
1. Gå till **Admin Dashboard**
2. Klicka **"+ Create User"**
3. Fyll i: Email, Name, Password, User Level
4. Klicka **"Create"**
5. ✅ Kolla Audit Logs → Du ska se **USER_CREATED**

**Test USER_DELETED:**
1. Gå till **Admin Dashboard**
2. Leta upp en testanvändare
3. Klicka **🗑️ Delete**
4. Bekräfta borttagning
5. ✅ Kolla Audit Logs → Du ska se **USER_DELETED**

### 👥 TEAM OPERATIONS (Teamhantering)

| Action | Vad Gör Du | Vad Syns i Loggen |
|--------|------------|-------------------|
| **TEAM_CREATED** | **Teams** → "Create Team" → Fyll i namn → Spara | `{"team_name":"Team Alpha","storage_quota":"5000"}` |
| **TEAM_UPDATED** | **Teams** → Klicka ✏️ → Ändra namn/quota → Spara | `{"team_name":"Team Beta","storage_quota":"10000"}` |
| **TEAM_DELETED** | **Teams** → Klicka 🗑️ → Bekräfta | `{"team_name":"Team Alpha"}` |
| **TEAM_MEMBER_ADDED** | **Teams** → Klicka på team → "Add Member" → Välj user → Spara | `{"team_id":"1","user_email":"user@email.se","role":"member"}` |
| **TEAM_MEMBER_REMOVED** | **Teams** → Klicka på team → Klicka 🗑️ på medlem → Bekräfta | `{"team_id":"1","user_email":"user@email.se"}` |

### ⚙️ SETTINGS (Inställningar)

| Action | Vad Gör Du | Vad Syns i Loggen |
|--------|------------|-------------------|
| **SETTINGS_UPDATED** | **Admin → Settings** → Ändra Server URL/Port → Spara | `{"server_url":"http://nya-url.se","port_changed":false}` |
| **BRANDING_UPDATED** | **Admin → Settings** → Ändra Company Name/Logo → Spara | `{"company_name":"Nytt Namn","logo_updated":true}` |
| **EMAIL_SETTINGS_UPDATED** | **Admin → Email Settings** → Konfigurera SMTP → Spara | `{"provider":"smtp","from_email":"no-reply@firma.se"}` |

**Test SETTINGS_UPDATED:**
1. Gå till **Admin → Settings**
2. Ändra **Server URL** (t.ex. `http://wulfvault.dyndns.org`)
3. Klicka **"Save"**
4. ✅ Kolla Audit Logs → Du ska se **SETTINGS_UPDATED**

### 📥 DOWNLOAD ACCOUNTS (Nedladdningskonton)

| Action | Vad Gör Du | Vad Syns i Loggen |
|--------|------------|-------------------|
| **DOWNLOAD_ACCOUNT_CREATED** | **Admin → Download Accounts** → "Create" → Fyll i → Spara | `{"email":"download@firma.se","name":"Download User"}` |
| **DOWNLOAD_ACCOUNT_CREATED** (self-registration) | Användare skapar eget konto via fil-länk | `{"email":"user@email.se","name":"Namn","self_registered":true}` |
| **DOWNLOAD_ACCOUNT_DELETED** | **Admin → Download Accounts** → 🗑️ → Bekräfta | `{"email":"download@firma.se","name":"Namn","soft_delete":true,"admin_deleted":true}` |
| **DOWNLOAD_ACCOUNT_DELETED** (self-delete) | Download-användare raderar sitt eget konto | `{"email":"download@firma.se","soft_delete":true,"admin_deleted":false}` |

## 🔍 Testa Details Viewer (NYTT!)

### Hover Tooltip:
1. Gå till **Audit Logs**
2. **Hovra** med musen över en cell i **Details**-kolumnen
3. ✅ Du ska se hela JSON-strängen i en tooltip

### Modal Popup:
1. **Klicka** på en cell i **Details**-kolumnen
2. ✅ En modal öppnas med formaterad, läsbar JSON
3. Klicka **✕** eller utanför modalen för att stänga

## 🎯 Vad Som Fortfarande INTE Loggas

**OBS! Dessa operationer loggas INTE ännu:**
- ❌ Password reset/ändringar (PASSWORD_CHANGED)
- ❌ 2FA enable/disable (2FA_ENABLED, 2FA_DISABLED)
- ❌ User activation/deactivation (USER_ACTIVATED, USER_DEACTIVATED)

Detta är funktioner som kanske inte finns implementerade än, eller så saknas audit logging för dem.

## 📊 Pagination Tester

1. Gå till **Audit Logs**
2. I **Filters**, ändra **"Items Per Page"** dropdown:
   - Välj **20** → Visar max 20 entries
   - Välj **50** → Visar max 50 entries
   - Välj **100** → Visar max 100 entries
   - Välj **200** → Visar max 200 entries

3. Testa **Previous** / **Next** knappar:
   - Om du har fler än 20 entries, ska **Next** vara aktiverad
   - Klicka **Next** → Sidan går till nästa 20 entries
   - Klicka **Previous** → Tillbaka till föregående sida

## 🏁 Komplett Checklista

- [ ] **LOGIN_SUCCESS** - Lyckad inloggning
- [ ] **LOGIN_FAILED** - Misslyckad inloggning
- [ ] **LOGOUT** - Utloggning
- [ ] **FILE_UPLOADED** - Fil uppladdad
- [ ] **FILE_DOWNLOADED** - Fil nedladdad
- [ ] **FILE_DELETED** - Fil borttagen (till trash)
- [ ] **FILE_RESTORED** - Fil återställd från trash ⭐ NY!
- [ ] **FILE_PERMANENTLY_DELETED** - Fil permanent raderad ⭐ NY!
- [ ] **USER_CREATED** - Användare skapad
- [ ] **USER_UPDATED** - Användare uppdaterad
- [ ] **USER_DELETED** - Användare raderad
- [ ] **TEAM_CREATED** - Team skapat
- [ ] **TEAM_UPDATED** - Team uppdaterat
- [ ] **TEAM_DELETED** - Team raderat
- [ ] **TEAM_MEMBER_ADDED** - Medlem tillagd i team
- [ ] **TEAM_MEMBER_REMOVED** - Medlem borttagen från team
- [ ] **SETTINGS_UPDATED** - Systeminställningar ändrade
- [ ] **BRANDING_UPDATED** - Branding uppdaterat
- [ ] **EMAIL_SETTINGS_UPDATED** - Email-inställningar ändrade
- [ ] **DOWNLOAD_ACCOUNT_CREATED** - Download-konto skapat
- [ ] **DOWNLOAD_ACCOUNT_DELETED** - Download-konto raderat
- [ ] **DOWNLOAD_ACCOUNT_LOGIN_SUCCESS** - Download-konto inloggning

## 🐛 Om Något Saknas

Om du utför en operation och den **inte** syns i Audit Logs:

1. **Refresh** sidan (F5)
2. Kontrollera att **Items Per Page** är tillräckligt hög (t.ex. 200)
3. Kontrollera att inga **filters** är aktiva (klicka "Reset")
4. Kolla server-loggen för errors:
   ```bash
   # Om du kör servern manuellt, se terminal output
   # Eller kolla log-filen om du kör som systemd service
   ```

5. Rapportera vilken operation som saknas!

## 📈 Success Metrics

Efter alla tester ska du ha minst:
- ✅ 20+ olika audit log entries
- ✅ Minst 10 olika action types
- ✅ Details visas korrekt i både tooltip och modal
- ✅ Pagination fungerar smidigt
- ✅ Alla file operations (upload, download, delete, restore, permanent delete) loggas

---

**Version:** 4.5.11 Gold
**Datum:** 2025-11-17
**Testad av:** ___________________
**Resultat:** ✅ / ❌
