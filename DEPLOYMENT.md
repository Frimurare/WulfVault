# Sharecare - Deployment Guide

## Manual Start/Stop

### Start Servern Manuellt
```bash
cd ~/sharecare
nohup ./sharecare > server.log 2>&1 &
```

### Stoppa Servern
```bash
pkill sharecare
```

### Kolla Status
```bash
pgrep -f sharecare  # Visa PID
curl http://localhost:8080/health  # Kolla version
```

---

## Autostart med Systemd (Rekommenderat)

### 1. Installera Service-filen
```bash
sudo cp /tmp/sharecare.service /etc/systemd/system/
sudo systemctl daemon-reload
```

### 2. Aktivera Autostart
```bash
sudo systemctl enable sharecare
```

### 3. Starta Tjänsten
```bash
sudo systemctl start sharecare
```

### 4. Kolla Status
```bash
sudo systemctl status sharecare
sudo journalctl -u sharecare -f  # Live logs
```

### Starta om Servern
```bash
sudo systemctl restart sharecare
```

**OBS:** När systemd är aktiverad fungerar "Restart Server"-knappen i Admin UI automatiskt!

---

## Autostart vid Container Reboot

När systemd service är aktiverad (`systemctl enable sharecare`) startar servern automatiskt när containern bootas om.

### Testa Autostart
```bash
sudo reboot
# Efter reboot:
systemctl status sharecare  # Ska vara "active (running)"
```

---

## Troubleshooting

### Servern startar inte
```bash
# Kolla logs
sudo journalctl -u sharecare -n 50

# Kolla permissions
ls -l ~/sharecare/sharecare
ls -ld ~/sharecare/data ~/sharecare/uploads
```

### Portar upptagna
```bash
sudo lsof -i :8080
# Döda processen:
sudo kill -9 <PID>
```

### Bygg om efter uppdateringar
```bash
cd ~/sharecare
git pull
go build -o sharecare cmd/server/main.go
sudo systemctl restart sharecare
```

---

## Produktionsinställningar

### Kör på annan port (t.ex. 443 för HTTPS)
```bash
# Ändra i systemd service:
sudo nano /etc/systemd/system/sharecare.service

# Lägg till environment variabel:
[Service]
Environment="PORT=443"
Environment="SERVER_URL=https://yourdomain.com"

sudo systemctl daemon-reload
sudo systemctl restart sharecare
```

### Reverse Proxy (Nginx/Caddy)
För produktionsmiljö, använd en reverse proxy framför Sharecare:
- Hanterar SSL/TLS certificates
- Rate limiting
- DDoS-skydd
- Static asset caching

Exempel Nginx config:
```nginx
location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

---

**Version:** 3.0.1
**Support:** https://github.com/Frimurare/Sharecare/issues
