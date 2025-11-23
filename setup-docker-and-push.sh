#!/bin/bash

# WulfVault v4.8.0 - Complete Docker Setup and Push Script
# This script will:
# 1. Install Docker (if not already installed)
# 2. Login to Docker Hub
# 3. Build the image
# 4. Push to Docker Hub

set -e

echo "🔍 Checking if Docker is installed..."

if ! command -v docker &> /dev/null; then
    echo "📦 Docker not found. Installing Docker..."
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sudo sh /tmp/get-docker.sh

    # Add current user to docker group to run without sudo
    sudo usermod -aG docker $USER

    echo "✅ Docker installed successfully"
    echo "⚠️  You may need to log out and back in for group changes to take effect"
    echo "   Or run: newgrp docker"
else
    echo "✅ Docker is already installed"
fi

echo ""
echo "🔐 Logging in to Docker Hub..."
echo "frimurare" | docker login --username frimurare --password-stdin

echo ""
echo "🐳 Building WulfVault v4.8.0 Docker image..."
docker build -t frimurare/wulfvault:4.8.0 -t frimurare/wulfvault:latest .

echo ""
echo "📦 Pushing to Docker Hub..."
docker push frimurare/wulfvault:4.8.0
docker push frimurare/wulfvault:latest

echo ""
echo "✅ Successfully pushed to Docker Hub!"
echo "   - frimurare/wulfvault:4.8.0"
echo "   - frimurare/wulfvault:latest"
echo ""
echo "🎉 Done! Image available at: https://hub.docker.com/r/frimurare/wulfvault"
