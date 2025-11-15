# WulfVault Teams API Guide

## Overview

Teams functionality is now fully implemented in WulfVault! This guide shows you how to use teams via the API endpoints.

## What's Implemented

### ✅ **Database Layer**
- `Teams` table - Stores team information
- `TeamMembers` table - Tracks team membership
- `TeamFiles` table - Links files to teams
- Full CRUD operations for all tables

### ✅ **Features**
- **Admin-only team creation** - Only Admins/SuperAdmins can create teams
- **Team member management** - Team members can add/remove other members
- **File sharing to teams** - Share files with your team
- **Email notifications** - Welcome emails when added to a team
- **Storage quotas** - Per-team storage limits (default 10GB)
- **Role-based permissions** - Owner, Admin, Member roles
- **Access control** - Download users excluded from teams

---

## API Endpoints

### 1. Admin: Create Team

**Endpoint:** `POST /api/admin/teams/create`
**Auth:** Admin required
**Request Body:**
```json
{
  "name": "Prudencia",
  "description": "Prudencia team workspace",
  "storageQuotaMB": 10240
}
```

**Response:**
```json
{
  "success": true,
  "team": {
    "id": 1,
    "name": "Prudencia",
    "description": "Prudencia team workspace",
    "createdBy": 1,
    "createdAt": 1736883600,
    "storageQuotaMB": 10240,
    "storageUsedMB": 0,
    "isActive": true
  }
}
```

**Test with curl:**
```bash
# First, login and get session cookie
curl -X POST http://localhost:8080/login \
  -d "username=admin&password=yourpassword" \
  -c cookies.txt

# Create team
curl -X POST http://localhost:8080/api/admin/teams/create \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "Prudencia",
    "description": "Prudencia team workspace",
    "storageQuotaMB": 10240
  }'
```

---

### 2. Get My Teams

**Endpoint:** `GET /api/teams/my`
**Auth:** User required
**Response:**
```json
{
  "success": true,
  "teams": [
    {
      "id": 1,
      "name": "Prudencia",
      "description": "Prudencia team workspace",
      "createdBy": 1,
      "createdAt": 1736883600,
      "storageQuotaMB": 10240,
      "storageUsedMB": 0,
      "isActive": true,
      "memberCount": 3,
      "userRole": 0
    }
  ]
}
```

**Test with curl:**
```bash
curl -X GET http://localhost:8080/api/teams/my \
  -b cookies.txt
```

---

### 3. Add Member to Team

**Endpoint:** `POST /api/teams/add-member`
**Auth:** User required (must be team owner/admin OR system admin)
**Request Body:**
```json
{
  "teamId": 1,
  "userId": 2,
  "role": 2
}
```

**Roles:**
- `0` = Owner (full control)
- `1` = Admin (manage members & files)
- `2` = Member (upload & view files)

**Response:**
```json
{
  "success": true,
  "member": {
    "id": 1,
    "teamId": 1,
    "userId": 2,
    "role": 2,
    "joinedAt": 1736883600,
    "addedBy": 1
  }
}
```

**Email sent automatically:**
```
Subject: Welcome to teamshare group Prudencia in the WulfVault fileshare

Body: You have been added to the teamshare group "Prudencia"...
```

**Test with curl:**
```bash
curl -X POST http://localhost:8080/api/teams/add-member \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "teamId": 1,
    "userId": 2,
    "role": 2
  }'
```

---

### 4. Get Team Members

**Endpoint:** `GET /api/teams/members?teamId=1`
**Auth:** User required (must be team member OR admin)
**Response:**
```json
{
  "success": true,
  "members": [
    {
      "id": 1,
      "teamId": 1,
      "userId": 1,
      "role": 0,
      "joinedAt": 1736883600,
      "addedBy": 1,
      "userName": "admin",
      "userEmail": "admin@example.com"
    },
    {
      "id": 2,
      "teamId": 1,
      "userId": 2,
      "role": 2,
      "joinedAt": 1736883700,
      "addedBy": 1,
      "userName": "john",
      "userEmail": "john@example.com"
    }
  ]
}
```

**Test with curl:**
```bash
curl -X GET "http://localhost:8080/api/teams/members?teamId=1" \
  -b cookies.txt
```

---

### 5. Share File to Team

**Endpoint:** `POST /api/teams/share-file`
**Auth:** User required (must be team member)
**Request Body:**
```json
{
  "fileId": "abc123",
  "teamId": 1
}
```

**Response:**
```json
{
  "success": true
}
```

**Test with curl:**
```bash
# First, upload a file and get its ID
# Then share it to team:
curl -X POST http://localhost:8080/api/teams/share-file \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "fileId": "your-file-id",
    "teamId": 1
  }'
```

---

### 6. Get Team Files

**Endpoint:** `GET /api/teams/files?teamId=1`
**Auth:** User required (must be team member OR admin)
**Response:**
```json
{
  "success": true,
  "files": [
    {
      "file": {
        "Id": "abc123",
        "Name": "document.pdf",
        "Size": "1.5 MB",
        "UserId": 1,
        "UploadDate": 1736883600,
        "DownloadsRemaining": 10
      },
      "sharedBy": 1,
      "sharedAt": 1736883700,
      "ownerName": "admin",
      "teamFileId": 1
    }
  ]
}
```

**Test with curl:**
```bash
curl -X GET "http://localhost:8080/api/teams/files?teamId=1" \
  -b cookies.txt
```

---

### 7. Remove Member from Team

**Endpoint:** `POST /api/teams/remove-member`
**Auth:** User required (must be team owner/admin OR system admin)
**Request Body:**
```json
{
  "teamId": 1,
  "userId": 2
}
```

**Response:**
```json
{
  "success": true
}
```

**Test with curl:**
```bash
curl -X POST http://localhost:8080/api/teams/remove-member \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "teamId": 1,
    "userId": 2
  }'
```

---

### 8. Admin: Update Team

**Endpoint:** `POST /api/admin/teams/update`
**Auth:** Admin required
**Request Body:**
```json
{
  "teamId": 1,
  "name": "Prudencia Team",
  "description": "Updated description",
  "storageQuotaMB": 20480
}
```

**Response:**
```json
{
  "success": true,
  "team": {
    "id": 1,
    "name": "Prudencia Team",
    "description": "Updated description",
    "storageQuotaMB": 20480,
    ...
  }
}
```

---

### 9. Admin: Delete Team

**Endpoint:** `POST /api/admin/teams/delete`
**Auth:** Admin required
**Request Body:**
```json
{
  "teamId": 1
}
```

**Response:**
```json
{
  "success": true
}
```

**Note:** This is a soft delete (sets `IsActive = 0`). Files are not deleted.

---

## Testing Workflow

### Scenario: Create "Prudencia" Team and Share Files

```bash
#!/bin/bash

# 1. Login as admin
curl -X POST http://localhost:8080/login \
  -d "username=admin&password=admin" \
  -c cookies.txt

# 2. Create Prudencia team
curl -X POST http://localhost:8080/api/admin/teams/create \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "Prudencia",
    "description": "Prudencia team workspace",
    "storageQuotaMB": 10240
  }' | jq '.'

# 3. Get team ID from response (assume it's 1)
TEAM_ID=1

# 4. Add user with ID 2 to team
curl -X POST http://localhost:8080/api/teams/add-member \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d "{
    \"teamId\": $TEAM_ID,
    \"userId\": 2,
    \"role\": 2
  }" | jq '.'

# User 2 will receive email: "Welcome to teamshare group Prudencia..."

# 5. List team members
curl -X GET "http://localhost:8080/api/teams/members?teamId=$TEAM_ID" \
  -b cookies.txt | jq '.'

# 6. Upload a file (assume you get file ID: file123)
# (Use existing file upload endpoint)

# 7. Share file to team
curl -X POST http://localhost:8080/api/teams/share-file \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d "{
    \"fileId\": \"file123\",
    \"teamId\": $TEAM_ID
  }" | jq '.'

# 8. View team files
curl -X GET "http://localhost:8080/api/teams/files?teamId=$TEAM_ID" \
  -b cookies.txt | jq '.'

# 9. Login as user 2 and access team files
curl -X POST http://localhost:8080/login \
  -d "username=user2&password=user2password" \
  -c cookies_user2.txt

curl -X GET "http://localhost:8080/api/teams/files?teamId=$TEAM_ID" \
  -b cookies_user2.txt | jq '.'

# User 2 can now see all files shared with Prudencia team!
```

---

## Database Schema

### Teams Table
```sql
CREATE TABLE Teams (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Name TEXT NOT NULL,
    Description TEXT,
    CreatedBy INTEGER NOT NULL,
    CreatedAt INTEGER NOT NULL,
    StorageQuotaMB INTEGER NOT NULL DEFAULT 10240,
    StorageUsedMB INTEGER NOT NULL DEFAULT 0,
    IsActive INTEGER DEFAULT 1,
    FOREIGN KEY (CreatedBy) REFERENCES Users(Id)
);
```

### TeamMembers Table
```sql
CREATE TABLE TeamMembers (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    TeamId INTEGER NOT NULL,
    UserId INTEGER NOT NULL,
    Role INTEGER DEFAULT 2,  -- 0=Owner, 1=Admin, 2=Member
    JoinedAt INTEGER NOT NULL,
    AddedBy INTEGER,
    FOREIGN KEY (TeamId) REFERENCES Teams(Id) ON DELETE CASCADE,
    FOREIGN KEY (UserId) REFERENCES Users(Id) ON DELETE CASCADE,
    UNIQUE(TeamId, UserId)
);
```

### TeamFiles Table
```sql
CREATE TABLE TeamFiles (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    FileId TEXT NOT NULL,
    TeamId INTEGER NOT NULL,
    SharedBy INTEGER NOT NULL,
    SharedAt INTEGER NOT NULL,
    FOREIGN KEY (FileId) REFERENCES Files(Id) ON DELETE CASCADE,
    FOREIGN KEY (TeamId) REFERENCES Teams(Id) ON DELETE CASCADE,
    UNIQUE(FileId, TeamId)
);
```

---

## Permission Model

### Who Can Do What?

| Action | Admin | Team Owner | Team Admin | Team Member |
|--------|-------|------------|------------|-------------|
| Create team | ✅ | ❌ | ❌ | ❌ |
| Delete team | ✅ | ❌ | ❌ | ❌ |
| Update team settings | ✅ | ❌ | ❌ | ❌ |
| Add members | ✅ | ✅ | ✅ | ❌ |
| Remove members | ✅ | ✅ | ✅ | ❌ |
| Share files to team | ✅ | ✅ | ✅ | ✅ |
| View team files | ✅ | ✅ | ✅ | ✅ |
| Download team files | ✅ | ✅ | ✅ | ✅ |

### Important Notes:
- **Download users** (DownloadAccounts) cannot be added to teams
- Only regular **Users**, **Admins**, and **SuperAdmins** can be team members
- Team members can only add/remove OTHER members (not themselves initially)
- File ownership remains with original uploader
- Sharing a file to multiple teams is allowed

---

## Email Notification

When a user is added to a team, they receive this email:

```
From: WulfVault <noreply@yourcompany.com>
To: user@example.com
Subject: Welcome to teamshare group Prudencia in the WulfVault fileshare

[Beautiful HTML email with:]
- Team name: "Prudencia"
- Company name from config
- Login link
- User's email address
- List of what they can do as a team member
```

---

## What's Next: Frontend Integration

To complete the full user experience, you can add:

1. **Admin Teams Management Page** (`/admin/teams`)
   - List all teams
   - Create/edit/delete teams
   - View team members and files

2. **User Teams Page** (`/teams`)
   - View my teams
   - Browse team files
   - Upload to team (use existing upload + share workflow)

3. **File Upload Enhancement**
   - Add "Share to team" dropdown in upload form
   - Show which teams a file is shared with

4. **Dashboard Enhancement**
   - Show team files alongside personal files
   - Add team filter/tab

---

## Already Implemented Features

✅ Complete database schema with migrations
✅ Full CRUD operations for teams
✅ Team member management (add/remove)
✅ File sharing to teams
✅ Role-based permissions (Owner/Admin/Member)
✅ Email notifications for team invitations
✅ Storage quota management per team
✅ Access control (download users excluded)
✅ All API endpoints registered and working
✅ Proper error handling and logging
✅ Admin-only team creation
✅ Team member can add/remove members
✅ Email uses branding from config

---

## Quick Start for Testing

1. **Start WulfVault:**
   ```bash
   ./wulfvault
   ```

2. **Login as admin** (via web UI at http://localhost:8080)

3. **Use curl or Postman** to test API endpoints (see examples above)

4. **Check database:**
   ```bash
   sqlite3 data/wulfvault.db
   SELECT * FROM Teams;
   SELECT * FROM TeamMembers;
   SELECT * FROM TeamFiles;
   ```

5. **Check logs** for email sent confirmations

---

## Summary

You now have a **fully functional teams system** with:
- WeTransfer Teams-like collaboration
- Admin-controlled team creation
- Member-managed membership
- File sharing across teams
- Email notifications
- Storage quotas
- Role-based access control

All backend functionality is complete and ready to use via API!
