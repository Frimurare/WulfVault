# WulfVault Deployment Guide

Production deployment for **WulfVault** v7.1.0 "Aurora".

---

## Build

```bash
go build -o wulfvault ./cmd/server
```

Run it with dedicated data and uploads directories:

```bash
./wulfvault \
  -port 8080 \
  -data /var/lib/wulfvault/data \
  -uploads /var/lib/wulfvault/uploads \
  -url https://files.example.com
```

The equivalent environment variables are `PORT`, `DATA_DIR`, `UPLOADS_DIR`, and
`SERVER_URL`. Print the version banner by running the binary (it logs
`WulfVault File Sharing System v7.1.0 Aurora` on startup).

---

## systemd

```ini
[Unit]
Description=WulfVault file sharing
After=network.target

[Service]
Type=simple
User=wulfvault
WorkingDirectory=/opt/wulfvault
ExecStart=/opt/wulfvault/wulfvault
Restart=always
Environment=PORT=8080
Environment=DATA_DIR=/var/lib/wulfvault/data
Environment=UPLOADS_DIR=/var/lib/wulfvault/uploads
Environment=SERVER_URL=https://files.example.com

[Install]
WantedBy=multi-user.target
```

---

## Reverse Proxy (nginx)

```nginx
server {
    listen 443 ssl;
    server_name files.example.com;

    client_max_body_size 5000M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_request_buffering off;
        proxy_http_version 1.1;
    }
}
```

Set `SERVER_URL` to the public `https://` URL so generated links and the SSO
redirect URI are correct.

---

## Database & Secrets

The SQLite database is created as **`wulfvault.db`** inside the data directory.
Back it up (it also holds the AES-256-GCM master key used to encrypt the OIDC
client secret and email API keys, so the backup is sufficient to restore
encrypted settings):

```bash
sqlite3 /var/lib/wulfvault/data/wulfvault.db ".backup backup.db"
```

For a full backup, archive the data directory (database + WAL files) and the
uploads directory together.

---

## Health Check

`GET /health` returns `200 OK` when the server is healthy.
