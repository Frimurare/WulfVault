# WulfVault - Deployment Guide

## Manual Start/Stop

### Start Servern Manuellt
```bash
cd ~/wulfvault
nohup ./wulfvault > server.log 2>&1 &
```

### Stoppa Servern
```bash
pkill wulfvault
```

### Kolla Status
```bash
pgrep -f wulfvault  # Visa PID
curl http://localhost:8080/health  # Kolla version
```

---

## Autostart med Systemd (Rekommenderat)

### 1. Installera Service-filen
```bash
sudo cp /tmp/wulfvault.service /etc/systemd/system/
sudo systemctl daemon-reload
```

### 2. Aktivera Autostart
```bash
sudo systemctl enable wulfvault
```

### 3. Starta Tjänsten
```bash
sudo systemctl start wulfvault
```

### 4. Kolla Status
```bash
sudo systemctl status wulfvault
sudo journalctl -u wulfvault -f  # Live logs
```

### Starta om Servern
```bash
sudo systemctl restart wulfvault
```

**OBS:** När systemd är aktiverad fungerar "Restart Server"-knappen i Admin UI automatiskt!

---

## Autostart vid Container Reboot

När systemd service är aktiverad (`systemctl enable wulfvault`) startar servern automatiskt när containern bootas om.

### Testa Autostart
```bash
sudo reboot
# Efter reboot:
systemctl status wulfvault  # Ska vara "active (running)"
```

---

## Troubleshooting

### Servern startar inte
```bash
# Kolla logs
sudo journalctl -u wulfvault -n 50

# Kolla permissions
ls -l ~/wulfvault/wulfvault
ls -ld ~/wulfvault/data ~/wulfvault/uploads
```

### Portar upptagna
```bash
sudo lsof -i :8080
# Döda processen:
sudo kill -9 <PID>
```

### Bygg om efter uppdateringar
```bash
cd ~/wulfvault
git pull
go build -o wulfvault cmd/server/main.go
sudo systemctl restart wulfvault
```

---

## Produktionsinställningar

### Kör på annan port (t.ex. 443 för HTTPS)
```bash
# Ändra i systemd service:
sudo nano /etc/systemd/system/wulfvault.service

# Lägg till environment variabel:
[Service]
Environment="PORT=443"
Environment="SERVER_URL=https://yourdomain.com"

sudo systemctl daemon-reload
sudo systemctl restart wulfvault
```

### Reverse Proxy (Nginx/Caddy)
För produktionsmiljö, använd en reverse proxy framför WulfVault:
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

**Version:** 3.2-beta2
**Support:** https://github.com/Frimurare/WulfVault/issues
