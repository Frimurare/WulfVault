#!/usr/bin/env bash

# WulfVault - Simple Proxmox Installation Script
# Test version before submitting to community-scripts

set -e

echo "=================================="
echo "  WulfVault Installation"
echo "=================================="
echo ""

# Update system
echo "Updating system..."
apt-get update
apt-get upgrade -y

# Install dependencies
echo "Installing dependencies..."
apt-get install -y curl sudo git

# Install Docker
echo "Installing Docker..."
curl -fsSL https://get.docker.com | bash

# Install Docker Compose
echo "Installing Docker Compose..."
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
mkdir -p $DOCKER_CONFIG/cli-plugins
LATEST=$(curl -sL https://api.github.com/repos/docker/compose/releases/latest | grep '"tag_name":' | cut -d'"' -f4)
curl -sSL https://github.com/docker/compose/releases/download/$LATEST/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose
chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose

# Setup WulfVault
echo "Setting up WulfVault..."
mkdir -p /opt/wulfvault
cd /opt/wulfvault

cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  wulfvault:
    image: frimurare/wulfvault:latest
    container_name: wulfvault
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./uploads:/uploads
    environment:
      - SERVER_URL=http://localhost:8080
      - PORT=8080
      - MAX_FILE_SIZE_MB=5000
      - DEFAULT_QUOTA_MB=10000
      - SESSION_TIMEOUT_HOURS=24
      - TRASH_RETENTION_DAYS=5
    restart: unless-stopped
EOF

mkdir -p data uploads
chmod 755 data uploads

# Start WulfVault
echo "Starting WulfVault..."
docker compose up -d

# Create update script
cat > /opt/wulfvault/update.sh << 'UPDATEEOF'
#!/usr/bin/env bash
set -e
echo "Updating WulfVault..."
cd /opt/wulfvault
docker compose pull
docker compose up -d
docker image prune -f
echo "Update complete!"
UPDATEEOF
chmod +x /opt/wulfvault/update.sh

# Get IP
IP=$(hostname -I | awk '{print $1}')

echo ""
echo "=================================="
echo "  Installation Complete!"
echo "=================================="
echo ""
echo "Web Interface: http://${IP}:8080"
echo ""
echo "Default Admin Login:"
echo "  Email: admin@wulfvault.local"
echo "  Password: WulfVaultAdmin2024!"
echo ""
echo "⚠️  IMPORTANT: Change admin password!"
echo ""
echo "Installation: /opt/wulfvault"
echo "Update: /opt/wulfvault/update.sh"
echo ""
