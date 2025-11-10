# Sharecare v2.0.0 Testing Checklist

## Fixade Problem

### 1. Password Protection för Filer ✅
**Problem**: Filer utan lösenord sparades med tomma strängar (`""`) istället för NULL, vilket gjorde att systemet trodde att alla filer krävde lösenord.

**Lösning**:
- Ändrade `SaveFile()` i `/internal/database/files.go` att spara NULL istället för tomma strängar
- Fixade 6 existerande filer i databasen som hade tomma lösenord
- Testfil skapad med lösenord "TestPass123"

**Test URL för lösenordsskyddad fil**:
```
http://192.168.86.142:8080/d/ad368e9436342df5ecafcc9110a8bad9
Password: TestPass123
```

### 2. File Request Credentials ✅
**Problem**: JavaScript fetch() skickade inte session-cookies, vilket gjorde att alla API-anrop redirectades till login.

**Lösning**:
- Lagt till `credentials: 'same-origin'` i alla tre fetch-anrop:
  - `/file-request/create`
  - `/file-request/list`
  - `/file-request/delete`

---

## Manual Testning Krävs

### Test 1: Password Protection
1. Ladda upp en fil med password protection:
   - Logga in på: http://192.168.86.142:8080/dashboard
   - Välj en fil att ladda upp
   - Kryssa i "🔐 Password protect this file"
   - Ange ett lösenord (t.ex. "MySecret123")
   - Klicka "Upload File"

2. Verifiera att lösenordet krävs:
   - Kopiera nedladdningslänken från filen
   - Öppna länken i ett inkognito-fönster
   - Du ska se en "Password Required"-sida
   - Testa med fel lösenord → ska ge felmeddelande
   - Testa med rätt lösenord → ska ladda ner filen

3. Verifiera att filer UTAN lösenord fungerar normalt:
   - Ladda upp en fil UTAN att kryssa i password-skydd
   - Nedladdningslänken ska fungera direkt utan lösenord

### Test 2: File Request (Upload Request)
1. Skapa en upload request:
   - Logga in på: http://192.168.86.142:8080/dashboard
   - Scrolla ner till "📥 Request Files from Others"
   - Klicka "➕ Create Upload Request"
   - Fyll i:
     - Title: "Test Upload"
     - Message: "Send your files"
     - Expires in days: 7
     - Max size: 100 MB
   - Klicka OK

2. Verifiera att requesten skapades:
   - Du ska få en popup med en upload-länk
   - Länken ska visas under "Request Files from Others"
   - Du ska kunna kopiera länken

3. Testa upload-länken (från annan dator om möjligt):
   - Öppna länken i inkognito-fönster
   - Du ska se en upload-sida
   - Ladda upp en testfil
   - Filen ska synas i din dashboard

### Test 3: Kombinerad Password + Auth
1. Ladda upp en fil med BÅDE password och "Require authentication":
   - Kryssa i både "🔐 Password protect" OCH "🔒 Require authentication"
   - Ange lösenord
   - Ladda upp

2. Verifiera dual-protection:
   - Öppna nedladdningslänken
   - Ska först fråga efter lösenord
   - Efter korrekt lösenord → ska fråga efter email/password för mottagaren
   - Efter registrering → ska ladda ner filen

---

## Kända Problem (Om Något Inte Fungerar)

### Om "Create Upload Request" fortfarande inte fungerar:
1. Öppna webbläsarens Developer Tools (F12)
2. Gå till Console-fliken
3. Tryck på "Create Upload Request"
4. Leta efter felmeddelanden i konsolen
5. Skicka mig felmeddelandena

### Om password-skydd inte fungerar:
1. Kolla serverns loggar: `tail -50 /home/ulf/sharecare/server.log`
2. Leta efter felmeddelanden
3. Verifiera att filen har lösenord i databasen:
   ```bash
   go run /tmp/test_password.go | grep "filnamn"
   ```

---

## Att Testa Efter Framgång
- [ ] Password-skyddade filer
- [ ] Filer utan lösenord fungerar normalt
- [ ] File Request skapande
- [ ] Upload via File Request länk
- [ ] Kombinerad Password + Auth
- [ ] Delete File Request
- [ ] File Request expiration

## Serverstatus
- Server URL: `http://192.168.86.142:8080`
- Startas med: `./start.sh`
- Stoppas med: `pkill -9 sharecare`
- Loggar: `tail -f server.log`
