#!/bin/bash

# WulfVault v4.8.0 Docker Hub Push Script
# Run this on a machine with Docker installed

set -e

echo "🐳 Building WulfVault v4.8.0 Docker image..."

# Build the Docker image with both version tag and latest tag
docker build -t frimurare/wulfvault:4.8.0 -t frimurare/wulfvault:latest .

echo "✅ Docker image built successfully"
echo ""
echo "📦 Pushing to Docker Hub..."

# Push both tags
docker push frimurare/wulfvault:4.8.0
docker push frimurare/wulfvault:latest

echo ""
echo "✅ Successfully pushed to Docker Hub!"
echo "   - frimurare/wulfvault:4.8.0"
echo "   - frimurare/wulfvault:latest"
echo ""
echo "🎉 Done! Image available at: https://hub.docker.com/r/frimurare/wulfvault"
