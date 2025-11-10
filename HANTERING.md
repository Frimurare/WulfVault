# Sharecare - Hanteringsguide för användare ulf

## Översikt
Sharecare körs nu som användare `ulf` och du har full kontroll över tjänsten och alla filer.

## Filplacering
- **Program**: `/home/ulf/sharecare/sharecare`
- **Databas**: `/home/ulf/sharecare/data/sharecare.db`
- **Uppladdningar**: `/home/ulf/sharecare/uploads/`
- **Källkod**: `/home/ulf/sharecare/` (Git-repo)

## Hantera tjänsten

### Visa status
```bash
sudo systemctl status sharecare
```

### Starta tjänsten
```bash
sudo systemctl start sharecare
```

### Stoppa tjänsten
```bash
sudo systemctl stop sharecare
```

### Starta om tjänsten
```bash
sudo systemctl restart sharecare
```

### Visa loggar (realtid)
```bash
sudo journalctl -u sharecare -f
```

### Visa loggar (senaste 100 rader)
```bash
sudo journalctl -u sharecare -n 100
```

**OBS**: Alla dessa kommandon kräver INTE lösenord för användare ulf!

## Uppdatera Sharecare

För att uppdatera till senaste versionen från Git och bygga om:

```bash
cd ~/sharecare
./update.sh
```

Scriptet gör följande automatiskt:
1. Stoppar tjänsten
2. Hämtar senaste koden från Git
3. Laddar ner Go-dependencies
4. Bygger om programmet
5. Startar tjänsten igen

## Manuell byggning

Om du behöver bygga manuellt:

```bash
cd ~/sharecare
go build -o sharecare cmd/server/main.go
sudo systemctl restart sharecare
```

## Felsökning

### Tjänsten startar inte
```bash
# Visa detaljerade loggar
sudo journalctl -u sharecare -n 50 --no-pager

# Kontrollera att filer finns
ls -la ~/sharecare/sharecare
ls -la ~/sharecare/data/
```

### Portkonflikter
```bash
# Kontrollera vad som lyssnar på port 8080
sudo ss -tlnp | grep 8080

# Döda process om nödvändigt
sudo lsof -ti:8080 | xargs kill -9
```

### Återställ till fabriksinställningar
```bash
sudo systemctl stop sharecare
rm -rf ~/sharecare/data/*
sudo systemctl start sharecare
```

## Backup

### Backup av databas
```bash
cp ~/sharecare/data/sharecare.db ~/sharecare-backup-$(date +%Y%m%d).db
```

### Backup av uppladdningar
```bash
tar -czf ~/sharecare-uploads-$(date +%Y%m%d).tar.gz ~/sharecare/uploads/
```

## Konfiguration

Tjänstens miljövariabler finns i:
```
/etc/systemd/system/sharecare.service
```

För att ändra konfiguration (kräver root):
```bash
sudo nano /etc/systemd/system/sharecare.service
sudo systemctl daemon-reload
sudo systemctl restart sharecare
```

## Åtkomst

- **Web UI**: http://192.168.86.142:8080
- **Admin login**: admin@sharecare.local
- **Lösenord**: SharecareAdmin2024!

## Support

För frågor eller problem, se:
- GitHub: https://github.com/Frimurare/Sharecare
- Loggar: `sudo journalctl -u sharecare -f`
