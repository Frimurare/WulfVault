#!/bin/bash
# Sharecare Update Script
# Run this script to update and rebuild sharecare

set -e

echo "🔄 Updating Sharecare..."
echo ""

# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin

# Change to sharecare directory
cd /home/ulf/sharecare

# Stop the service
echo "⏸️  Stopping sharecare service..."
sudo systemctl stop sharecare.service

# Pull latest changes
echo "📥 Pulling latest changes from git..."
git pull

# Download Go dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Build the application
echo "🔨 Building sharecare..."
go build -o sharecare cmd/server/main.go

# Start the service
echo "▶️  Starting sharecare service..."
sudo systemctl start sharecare.service

# Show status
echo ""
echo "✅ Update complete! Service status:"
sudo systemctl status sharecare.service --no-pager -l

echo ""
echo "📝 To view logs, run: sudo journalctl -u sharecare -f"
