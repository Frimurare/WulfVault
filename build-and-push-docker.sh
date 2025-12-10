#!/bin/bash
# WulfVault Docker Build and Push Script
# Version: 6.1.1 BloodMoon

set -e

echo "🐺 Building WulfVault v6.1.1 Docker Image..."
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_USERNAME="wulfvault"
IMAGE_NAME="wulfvault"
VERSION="6.1.1"
LATEST_TAG="latest"

echo -e "${BLUE}Step 1: Building multi-platform image${NC}"
echo "Platforms: linux/amd64, linux/arm64"
echo ""

# Build and push multi-platform image
sudo docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag ${DOCKER_USERNAME}/${IMAGE_NAME}:${VERSION} \
  --tag ${DOCKER_USERNAME}/${IMAGE_NAME}:${LATEST_TAG} \
  --push \
  .

echo ""
echo -e "${GREEN}✅ Successfully built and pushed:${NC}"
echo "  - ${DOCKER_USERNAME}/${IMAGE_NAME}:${VERSION}"
echo "  - ${DOCKER_USERNAME}/${IMAGE_NAME}:${LATEST_TAG}"
echo ""
echo -e "${BLUE}Docker Hub:${NC} https://hub.docker.com/r/${DOCKER_USERNAME}/${IMAGE_NAME}"
echo ""
echo "🎉 Done! Users can now pull:"
echo "  docker pull ${DOCKER_USERNAME}/${IMAGE_NAME}:${VERSION}"
echo "  docker pull ${DOCKER_USERNAME}/${IMAGE_NAME}:latest"
