# Docker Build Instructions for WulfVault v6.1.1

## Prerequisites

1. Docker installed and running
2. Docker Hub account (username: wulfvault)
3. Docker buildx enabled for multi-platform builds

## Steps to Build and Push

### 1. Login to Docker Hub

```bash
sudo docker login
# Enter username: wulfvault
# Enter password: [your Docker Hub token/password]
```

### 2. Setup buildx builder (one-time setup)

```bash
sudo docker buildx create --use --name wulfvault-builder
sudo docker buildx inspect --bootstrap
```

### 3. Run the build script

```bash
./build-and-push-docker.sh
```

This will:
- Build for both linux/amd64 and linux/arm64
- Tag as wulfvault/wulfvault:6.1.1
- Tag as wulfvault/wulfvault:latest
- Push to Docker Hub automatically

## Manual Build (Alternative)

If the script doesn't work, run manually:

```bash
sudo docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag wulfvault/wulfvault:6.1.1 \
  --tag wulfvault/wulfvault:latest \
  --push \
  .
```

## Verify Upload

After successful build, check Docker Hub:
https://hub.docker.com/r/wulfvault/wulfvault/tags

You should see:
- Tag: 6.1.1
- Tag: latest
- Both with linux/amd64 and linux/arm64 platforms

## Test the Image

```bash
docker pull wulfvault/wulfvault:6.1.1
docker run -d -p 8080:8080 wulfvault/wulfvault:6.1.1
```

## Important Notes

- The Dockerfile already includes `web/static/` with all CSS and JS files
- Frontend files are copied in the build stage: `COPY --from=builder /app/web/static ./web/static`
- This includes:
  - web/static/css/style.css
  - web/static/js/dashboard.js (with cache busting v6.1.1)
  - web/static/js/inactivity-tracker.js
  - web/static/js/mobile-nav.js

## Troubleshooting

### Error: "permission denied"
Solution: Make sure you're using `sudo` with all docker commands

### Error: "buildx not found"
Solution:
```bash
sudo apt-get install docker-buildx-plugin
```

### Error: "builder not found"
Solution: Create the builder first:
```bash
sudo docker buildx create --use --name wulfvault-builder
```
