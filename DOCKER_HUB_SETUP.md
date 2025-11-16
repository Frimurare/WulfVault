# Docker Hub Setup Guide - WulfVault

This is a simple guide to get WulfVault automatically published to Docker Hub so others can easily use your software!

## 🎯 What Happens When This Is Done?

Every time you push code to the `main` branch, GitHub will automatically:
1. ✅ Build a Docker image
2. ✅ Push it to Docker Hub
3. ✅ Tag with the correct version number (e.g., `4.5.6 Gold`)
4. ✅ Also tag as `latest` for easy use

**Other users can then run:**
```bash
docker pull yourusername/wulfvault:latest
docker run -p 8080:8080 yourusername/wulfvault:latest
```

---

## 📋 What You Need to Do (3 Easy Steps)

### Step 1: Create Docker Hub Account (5 minutes)

1. Go to https://hub.docker.com/signup
2. Create a free account
3. Verify your email
4. Remember your username (e.g., "frimurare" or "ulfholm")

**Free plan gives you:**
- Unlimited public images
- Perfect for open source projects like WulfVault!

---

### Step 2: Create Docker Hub Access Token (3 minutes)

1. Log in to Docker Hub (https://hub.docker.com)
2. Click your username (top right) → **Account Settings**
3. Select **Security** in the left menu
4. Click **New Access Token**
5. Give it a name: `GitHub Actions WulfVault`
6. Permissions: Select **Read, Write, Delete**
7. Click **Generate**
8. **IMPORTANT:** Copy the token now! It's only shown once!

**Save this token** - you'll need it in the next step.

---

### Step 3: Add Secrets to GitHub (2 minutes)

1. Go to your GitHub repo: https://github.com/Frimurare/WulfVault
2. Click **Settings** (top menu)
3. In the left menu: select **Secrets and variables** → **Actions**
4. Click **New repository secret**

**Add two secrets:**

**Secret #1:**
- Name: `DOCKER_HUB_USERNAME`
- Value: `your-dockerhub-username` (e.g., "frimurare")
- Click **Add secret**

**Secret #2:**
- Name: `DOCKER_HUB_TOKEN`
- Value: *paste the access token from Step 2*
- Click **Add secret**

---

## 🎉 Done! What Happens Now?

When I push the GitHub Actions workflow file, it will automatically:

1. **On next push to main:**
   - GitHub builds Docker image
   - Pushes to Docker Hub as `yourusername/wulfvault:latest`
   - Also pushes as `yourusername/wulfvault:4.5.6-gold`

2. **You can watch the progress:**
   - Go to GitHub repo → **Actions** tab
   - See live build progress (~3-5 minutes first time)

3. **When complete:**
   - Your image is available at https://hub.docker.com/r/yourusername/wulfvault
   - Others can use it immediately!

---

## 📦 How Others Use Your Software

Once everything is set up, anyone can run WulfVault with:

### Quick Start (One Command)
```bash
docker run -d -p 8080:8080 \
  -v wulfvault-data:/data \
  -v wulfvault-uploads:/uploads \
  yourusername/wulfvault:latest
```

**First Login:**
1. Open browser: `http://localhost:8080`
2. Login with default admin account:
   - **Email:** `admin@wulfvault.local`
   - **Password:** `WulfVaultAdmin2024!`
3. **⚠️ IMPORTANT:** Go to Settings and change password immediately!

### Or with Docker Compose
```yaml
version: '3.8'
services:
  wulfvault:
    image: yourusername/wulfvault:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./uploads:/uploads
```

**First Login (same as above):**
- Email: `admin@wulfvault.local`
- Password: `WulfVaultAdmin2024!`
- Change password right after first login!

---

## 🔄 Automatic Updates

**Every time you update the code:**
1. You push to `main` (as we always do)
2. GitHub Actions automatically builds new image
3. It appears on Docker Hub with new version number
4. Users can update with: `docker pull yourusername/wulfvault:latest`

---

## ❓ Troubleshooting

### "Workflow failed" in GitHub Actions

**Most common causes:**

1. **Wrong Docker Hub username in secrets**
   - Go to Settings → Secrets → DOCKER_HUB_USERNAME
   - Verify it matches exactly

2. **Wrong or expired token**
   - Create new token on Docker Hub
   - Update DOCKER_HUB_TOKEN secret

3. **First time takes longer**
   - Multi-platform builds (amd64 + arm64) take ~5-10 min
   - Next times are faster (cache)

### Check workflow status
1. Go to: https://github.com/Frimurare/WulfVault/actions
2. Click on latest workflow run
3. View logs for each step

---

## 📊 What the Workflow Does

The `.github/workflows/docker-publish.yml` file I created:

✅ **Triggers (when it runs):**
- On every push to `main`
- When a release is created
- Manually (workflow_dispatch)

✅ **What it builds:**
- Multi-platform: amd64 (x86 computers) + arm64 (Raspberry Pi, Apple Silicon)
- Optimized multi-stage build (from Dockerfile)

✅ **Tagging:**
- `latest` - always latest from main
- `4.5.6-gold` - version number from code
- `main-abc123` - Git commit SHA

✅ **Extras:**
- Updates Docker Hub description from README.md
- Build cache for faster builds
- Professional metadata

---

## 🎓 Lessons for the Future

### Create Releases for Important Versions

When you want to mark a special version (e.g., v5.0.0):

1. **In GitHub:**
   ```bash
   git tag v5.0.0
   git push origin v5.0.0
   ```

2. **Or via GitHub UI:**
   - Go to repo → Releases → Create new release
   - Tag: `v5.0.0`
   - Title: "WulfVault v5.0.0 - Description"
   - Publish

3. **Result:**
   - Automatic Docker build
   - Tagged as both `5.0.0` and `latest`
   - Visible in Releases section

---

## 📢 Marketing

Once Docker Hub is set up, you can add to README.md:

```markdown
## 🐳 Docker Quick Start

docker run -d -p 8080:8080 yourusername/wulfvault:latest
```

**Docker Hub Badge:**
```markdown
[![Docker Pulls](https://img.shields.io/docker/pulls/yourusername/wulfvault)](https://hub.docker.com/r/yourusername/wulfvault)
```

This shows how many people are using your software! 📈

---

## ✅ Checklist

- [ ] Created Docker Hub account
- [ ] Created Access Token
- [ ] Added DOCKER_HUB_USERNAME to GitHub Secrets
- [ ] Added DOCKER_HUB_TOKEN to GitHub Secrets
- [ ] Pushed workflow file (Claude does this)
- [ ] Waited for first build (3-10 min)
- [ ] Verified at https://hub.docker.com/r/yourusername/wulfvault

---

## 🎉 You're Done!

Once all steps are complete, WulfVault is available to the whole world via Docker Hub!

**Questions?** Check GitHub Actions logs or ask me next session!

**Tip:** Add a badge to README.md to show that WulfVault is available on Docker Hub!

---

**Created by:** Claude Code (Anthropic)
**Date:** 2025-11-16
**For:** WulfVault Docker Hub Publishing
**Difficulty:** 🟢 Beginner-Friendly (15 minutes total)
