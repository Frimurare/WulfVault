# WulfVault REST API Documentation

**Version:** 4.1.0

WulfVault provides a comprehensive REST API for managing users, files, teams, and system settings. This documentation covers all available endpoints, authentication methods, and usage examples.

## Table of Contents

- [Authentication](#authentication)
- [User Management API](#user-management-api)
- [File Management API](#file-management-api)
- [Download Accounts API](#download-accounts-api)
- [File Requests API](#file-requests-api)
- [Trash Management API](#trash-management-api)
- [Teams API](#teams-api)
- [Email API](#email-api)
- [Admin/System API](#adminsystem-api)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)

## Authentication

WulfVault uses session-based authentication via cookies. To authenticate API requests:

1. **Login** via the web interface at `/login`
2. The server sets a `session` cookie that's automatically included in subsequent requests
3. For programmatic access, include the session cookie in your requests

### Example: Login and API Call

```bash
# Login and save cookies
curl -c cookies.txt -X POST http://localhost:4949/login \
  -d "email=admin@wulfvault.local" \
  -d "password=your_password"

# Use the session cookie for API calls
curl -b cookies.txt http://localhost:4949/api/v1/users
```

### Authorization Levels

- **Public**: No authentication required
- **Authenticated**: Requires valid session cookie
- **Admin**: Requires valid session cookie + admin privileges

## User Management API

Manage user accounts, storage quotas, and permissions.

### List All Users

```http
GET /api/v1/users
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "users": [
    {
      "id": 1,
      "name": "Admin User",
      "email": "admin@wulfvault.local",
      "userLevel": 0,
      "permissions": 15,
      "storageQuotaMB": 102400,
      "storageUsedMB": 5120,
      "isActive": true,
      "createdAt": 1704067200
    }
  ],
  "count": 1
}
```

### Get User by ID

```http
GET /api/v1/users/{id}
```

**Authorization:** Admin
**Parameters:**
- `id` (path): User ID

**Response:**

```json
{
  "success": true,
  "user": {
    "id": 1,
    "name": "Admin User",
    "email": "admin@wulfvault.local",
    "userLevel": 0,
    "permissions": 15,
    "storageQuotaMB": 102400,
    "storageUsedMB": 5120,
    "isActive": true,
    "createdAt": 1704067200
  }
}
```

### Create User

```http
POST /api/v1/users
```

**Authorization:** Admin
**Request Body:**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePassword123!",
  "userLevel": 2,
  "permissions": 0,
  "storageQuotaMB": 10240,
  "isActive": true
}
```

**User Levels:**
- `0`: Super Admin
- `1`: Admin
- `2`: Regular User

**Response:**

```json
{
  "success": true,
  "user": {
    "id": 2,
    "name": "John Doe",
    "email": "john@example.com",
    "userLevel": 2,
    "permissions": 0,
    "storageQuotaMB": 10240,
    "storageUsedMB": 0,
    "isActive": true,
    "createdAt": 1704153600
  }
}
```

### Update User

```http
PUT /api/v1/users/{id}
```

**Authorization:** Admin
**Request Body:**

```json
{
  "name": "John Doe Updated",
  "email": "john.doe@example.com",
  "password": "NewPassword123!",
  "userLevel": 2,
  "permissions": 0,
  "storageQuotaMB": 20480,
  "isActive": true
}
```

**Note:** Password is optional. If not provided, existing password is kept.

**Response:**

```json
{
  "success": true,
  "user": {
    "id": 2,
    "name": "John Doe Updated",
    "email": "john.doe@example.com",
    "userLevel": 2,
    "permissions": 0,
    "storageQuotaMB": 20480,
    "storageUsedMB": 0,
    "isActive": true
  }
}
```

### Delete User

```http
DELETE /api/v1/users/{id}
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

### Get User's Files

```http
GET /api/v1/users/{id}/files
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "files": [
    {
      "id": "abc123",
      "name": "document.pdf",
      "sizeBytes": 1048576,
      "uploadDate": 1704153600,
      "downloadCount": 5,
      "downloadsRemaining": 95
    }
  ],
  "count": 1
}
```

### Get User's Storage Usage

```http
GET /api/v1/users/{id}/storage
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "userId": 2,
  "storageUsedMB": 1024,
  "storageQuotaMB": 10240,
  "percentage": 10,
  "fileCount": 15
}
```

## File Management API

Manage file uploads, downloads, and metadata.

### List User's Files

```http
GET /api/v1/files
```

**Authorization:** Authenticated
**Response:**

```json
{
  "success": true,
  "files": [
    {
      "id": "abc123xyz",
      "name": "presentation.pptx",
      "sizeBytes": 2097152,
      "uploadDate": 1704153600,
      "expireAt": 1704758400,
      "downloadCount": 10,
      "downloadsRemaining": 90,
      "unlimitedDownloads": false,
      "unlimitedTime": false,
      "requireAuth": true
    }
  ]
}
```

### Get File Details

```http
GET /api/v1/files/{id}
```

**Authorization:** Authenticated (own files) or Admin
**Response:**

```json
{
  "success": true,
  "file": {
    "id": "abc123xyz",
    "name": "presentation.pptx",
    "sizeBytes": 2097152,
    "contentType": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    "uploadDate": 1704153600,
    "expireAt": 1704758400,
    "downloadCount": 10,
    "downloadsRemaining": 90,
    "unlimitedDownloads": false,
    "unlimitedTime": false,
    "requireAuth": true,
    "userId": 2
  }
}
```

### Update File Metadata

```http
PUT /api/v1/files/{id}
```

**Authorization:** Authenticated (own files) or Admin
**Request Body:**

```json
{
  "downloadsRemaining": 50,
  "expireAt": 1705363200,
  "expireAtString": "2024-01-15",
  "unlimitedDownloads": false,
  "unlimitedTime": false,
  "password": "optional_file_password"
}
```

**Response:**

```json
{
  "success": true,
  "file": {
    "id": "abc123xyz",
    "downloadsRemaining": 50,
    "expireAt": 1705363200
  }
}
```

### Delete File

```http
DELETE /api/v1/files/{id}
```

**Authorization:** Authenticated (own files) or Admin
**Response:**

```json
{
  "success": true,
  "message": "File deleted successfully"
}
```

**Note:** Files are soft-deleted (moved to trash) and retained for 30 days before permanent deletion.

### Get File Download History

```http
GET /api/v1/files/{id}/downloads
```

**Authorization:** Authenticated (own files) or Admin
**Response:**

```json
{
  "success": true,
  "downloads": [
    {
      "id": 1,
      "fileId": "abc123xyz",
      "email": "downloader@example.com",
      "ipAddress": "192.168.1.100",
      "downloadedAt": 1704153600,
      "isAuthenticated": true
    }
  ],
  "count": 1
}
```

### Set/Update File Password

```http
POST /api/v1/files/{id}/password
```

**Authorization:** Authenticated (own files) or Admin
**Request Body:**

```json
{
  "password": "SecureFilePassword123"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Password updated successfully"
}
```

### Upload File

```http
POST /api/v1/upload
```

**Authorization:** Authenticated
**Content-Type:** multipart/form-data
**Form Fields:**
- `file`: The file to upload
- `requireAuth`: Boolean (optional)
- `downloadsRemaining`: Integer (optional, default: 100)
- `expireAt`: Unix timestamp (optional)
- `password`: String (optional)

**Response:**

```json
{
  "success": true,
  "fileId": "abc123xyz",
  "downloadUrl": "https://vault.example.com/d/abc123xyz",
  "splashUrl": "https://vault.example.com/s/abc123xyz"
}
```

### Download File

```http
GET /api/v1/download/{id}
```

**Authorization:** Public (may require file password if set)
**Response:** File binary data with appropriate Content-Type header

## Download Accounts API

Manage download-only user accounts.

### List All Download Accounts

```http
GET /api/v1/download-accounts
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "accounts": [
    {
      "id": 1,
      "name": "External Partner",
      "email": "partner@external.com",
      "isActive": true,
      "createdAt": 1704067200
    }
  ],
  "count": 1
}
```

### Create Download Account

```http
POST /api/v1/download-accounts
```

**Authorization:** Admin
**Request Body:**

```json
{
  "name": "External Contractor",
  "email": "contractor@example.com",
  "password": "TempPassword123!",
  "isActive": true
}
```

**Response:**

```json
{
  "success": true,
  "account": {
    "id": 2,
    "name": "External Contractor",
    "email": "contractor@example.com",
    "isActive": true,
    "createdAt": 1704153600
  }
}
```

### Update Download Account

```http
PUT /api/v1/download-accounts/{id}
```

**Authorization:** Admin
**Request Body:**

```json
{
  "name": "Updated Name",
  "email": "newemail@example.com",
  "password": "NewPassword123",
  "isActive": true
}
```

**Response:**

```json
{
  "success": true,
  "account": {
    "id": 2,
    "name": "Updated Name",
    "email": "newemail@example.com",
    "isActive": true
  }
}
```

### Delete Download Account

```http
DELETE /api/v1/download-accounts/{id}
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "message": "Download account deleted successfully"
}
```

### Toggle Download Account Status

```http
POST /api/v1/download-accounts/{id}/toggle
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "account": {
    "id": 2,
    "isActive": false
  }
}
```

## File Requests API

Create and manage file upload request portals.

### List File Requests

```http
GET /api/v1/file-requests
```

**Authorization:** Authenticated (own requests) or Admin (all requests)
**Response:**

```json
{
  "success": true,
  "requests": [
    {
      "id": 1,
      "userId": 2,
      "requestToken": "abc123token",
      "title": "Upload Documents",
      "message": "Please upload the requested documents",
      "createdAt": 1704067200,
      "expiresAt": 1704672000,
      "isActive": true,
      "maxFileSize": 104857600,
      "allowedFileTypes": ""
    }
  ],
  "count": 1
}
```

### Create File Request

```http
POST /api/v1/file-requests
```

**Authorization:** Authenticated
**Request Body:**

```json
{
  "title": "Contract Documents",
  "description": "Please upload signed contracts",
  "password": "OptionalPassword",
  "maxSizeMB": 100,
  "expiresAt": 1705363200,
  "maxUploads": 10,
  "notifyOnUpload": true
}
```

**Response:**

```json
{
  "success": true,
  "request": {
    "id": 2,
    "requestToken": "xyz789token",
    "title": "Contract Documents",
    "description": "Please upload signed contracts",
    "uploadUrl": "https://vault.example.com/upload-request/xyz789token"
  }
}
```

### Update File Request

```http
PUT /api/v1/file-requests/{id}
```

**Authorization:** Authenticated (own requests) or Admin
**Request Body:**

```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "password": "",
  "maxSizeMB": 200,
  "expiresAt": 1706659200,
  "maxUploads": 20,
  "notifyOnUpload": false,
  "isActive": true
}
```

**Response:**

```json
{
  "success": true,
  "request": {
    "id": 2,
    "title": "Updated Title",
    "isActive": true
  }
}
```

### Delete File Request

```http
DELETE /api/v1/file-requests/{id}
```

**Authorization:** Authenticated (own requests) or Admin
**Response:**

```json
{
  "success": true,
  "message": "File request deleted successfully"
}
```

### Get File Request by Token (Public)

```http
GET /api/v1/file-requests/token/{token}
```

**Authorization:** Public
**Response:**

```json
{
  "success": true,
  "request": {
    "id": 2,
    "title": "Upload Documents",
    "message": "Please upload the requested files",
    "maxFileSize": 104857600,
    "allowedFileTypes": "pdf,docx,xlsx"
  }
}
```

## Trash Management API

Manage deleted files in trash.

### List Trash

```http
GET /api/v1/trash
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "files": [
    {
      "id": "deleted123",
      "name": "old_file.pdf",
      "sizeBytes": 524288,
      "deletedAt": 1704067200,
      "deletedBy": 1
    }
  ],
  "count": 1
}
```

### Restore File from Trash

```http
POST /api/v1/trash/{id}/restore
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "message": "File restored successfully"
}
```

### Permanently Delete File

```http
DELETE /api/v1/trash/{id}
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "message": "File permanently deleted"
}
```

**Warning:** This action cannot be undone!

## Teams API

Manage teams, members, and file sharing. See [TEAMS_API_GUIDE.md](../TEAMS_API_GUIDE.md) for detailed documentation.

### Quick Reference

```http
GET    /api/teams/my                  # Get user's teams
GET    /api/teams/members?teamId={id} # List team members
GET    /api/teams/files?teamId={id}   # List team files
POST   /api/teams/add-member           # Add user to team
POST   /api/teams/remove-member        # Remove user from team
POST   /api/teams/share-file           # Share file to team
POST   /api/teams/unshare-file         # Unshare file from team

# Admin endpoints
POST   /api/admin/teams/create         # Create team
POST   /api/admin/teams/update         # Update team
POST   /api/admin/teams/delete         # Delete team
GET    /api/admin/users/list           # List all users
```

## Email API

Configure and send emails.

### Configure Email Settings

```http
POST /api/email/configure
```

**Authorization:** Admin
**Request Body:**

```json
{
  "provider": "smtp",
  "smtpHost": "smtp.gmail.com",
  "smtpPort": 587,
  "smtpUser": "noreply@example.com",
  "smtpPassword": "app_password",
  "fromEmail": "WulfVault <noreply@example.com>"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Email configuration saved"
}
```

### Test Email Configuration

```http
POST /api/email/test
```

**Authorization:** Admin
**Request Body:**

```json
{
  "recipientEmail": "test@example.com"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Test email sent successfully"
}
```

### Send File Link via Email

```http
POST /api/email/send-splash-link
```

**Authorization:** Authenticated
**Request Body:**

```json
{
  "fileId": "abc123xyz",
  "recipientEmail": "recipient@example.com",
  "message": "Here's the file you requested"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Email sent successfully"
}
```

## Admin/System API

System statistics, settings, and branding configuration.

### Get System Statistics

```http
GET /api/v1/admin/stats
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "stats": {
    "userCount": 25,
    "activeUserCount": 20,
    "fileCount": 150,
    "deletedFileCount": 10,
    "teamCount": 5,
    "totalStorageBytes": 536870912,
    "totalDownloads": 1250
  }
}
```

### Get Branding Configuration

```http
GET /api/v1/admin/branding
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "branding": {
    "branding_company_name": "WulfVault",
    "branding_primary_color": "#2563eb",
    "branding_secondary_color": "#1e40af",
    "logo_url": "/static/uploads/logo.png"
  }
}
```

### Update Branding Configuration

```http
POST /api/v1/admin/branding
```

**Authorization:** Admin
**Request Body:**

```json
{
  "companyName": "My Company",
  "primaryColor": "#ff6600",
  "secondaryColor": "#cc5200",
  "logoUrl": "/static/uploads/custom_logo.png"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Branding updated successfully"
}
```

### Get System Settings

```http
GET /api/v1/admin/settings
```

**Authorization:** Admin
**Response:**

```json
{
  "success": true,
  "settings": {
    "serverUrl": "https://vault.example.com",
    "port": "4949",
    "companyName": "WulfVault",
    "maxUploadSizeMB": 5120,
    "defaultQuotaMB": 10240,
    "trashRetentionDays": 30
  }
}
```

### Update System Settings

```http
POST /api/v1/admin/settings
```

**Authorization:** Admin
**Request Body:**

```json
{
  "maxUploadSizeMB": 10240,
  "defaultQuotaMB": 20480,
  "trashRetentionDays": 60
}
```

**Response:**

```json
{
  "success": true,
  "message": "Settings updated successfully"
}
```

## Error Handling

All API endpoints return errors in the following format:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes

- **200 OK**: Request successful
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid request parameters
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **405 Method Not Allowed**: HTTP method not supported
- **500 Internal Server Error**: Server error

### Example Error Response

```json
{
  "error": "User not found"
}
```

## Rate Limiting

Currently, WulfVault does not implement rate limiting. For production deployments, consider implementing rate limiting at the reverse proxy level (nginx, Apache, etc.).

## Best Practices

1. **Always use HTTPS** in production environments
2. **Implement proper error handling** in your API clients
3. **Store session cookies securely** (HTTPOnly, Secure flags)
4. **Validate file uploads** before sending to API
5. **Monitor storage quotas** to prevent disk space issues
6. **Regularly clean up trash** to reclaim storage space
7. **Use strong passwords** for all accounts
8. **Enable 2FA** for admin accounts
9. **Rotate API sessions** periodically
10. **Implement request logging** for security auditing

## Code Examples

### Python Example

```python
import requests

# Login
session = requests.Session()
login_data = {
    'email': 'admin@wulfvault.local',
    'password': 'your_password'
}
session.post('http://localhost:4949/login', data=login_data)

# List users
response = session.get('http://localhost:4949/api/v1/users')
users = response.json()
print(f"Total users: {users['count']}")

# Create user
new_user = {
    'name': 'API User',
    'email': 'apiuser@example.com',
    'password': 'SecurePassword123!',
    'userLevel': 2,
    'storageQuotaMB': 10240,
    'isActive': True
}
response = session.post('http://localhost:4949/api/v1/users', json=new_user)
print(response.json())

# Upload file
files = {'file': open('document.pdf', 'rb')}
data = {
    'requireAuth': 'true',
    'downloadsRemaining': '100'
}
response = session.post('http://localhost:4949/api/v1/upload', files=files, data=data)
print(response.json())
```

### JavaScript/Node.js Example

```javascript
const axios = require('axios');
const FormData = require('form-data');
const fs = require('fs');

const BASE_URL = 'http://localhost:4949';
const axiosInstance = axios.create({
  withCredentials: true,
  baseURL: BASE_URL
});

// Login
async function login() {
  const formData = new URLSearchParams();
  formData.append('email', 'admin@wulfvault.local');
  formData.append('password', 'your_password');

  await axiosInstance.post('/login', formData);
}

// List users
async function listUsers() {
  const response = await axiosInstance.get('/api/v1/users');
  console.log(`Total users: ${response.data.count}`);
  return response.data.users;
}

// Create user
async function createUser() {
  const newUser = {
    name: 'API User',
    email: 'apiuser@example.com',
    password: 'SecurePassword123!',
    userLevel: 2,
    storageQuotaMB: 10240,
    isActive: true
  };

  const response = await axiosInstance.post('/api/v1/users', newUser);
  return response.data;
}

// Upload file
async function uploadFile(filePath) {
  const form = new FormData();
  form.append('file', fs.createReadStream(filePath));
  form.append('requireAuth', 'true');
  form.append('downloadsRemaining', '100');

  const response = await axiosInstance.post('/api/v1/upload', form, {
    headers: form.getHeaders()
  });

  return response.data;
}

// Main execution
(async () => {
  await login();
  const users = await listUsers();
  const newUser = await createUser();
  const upload = await uploadFile('./document.pdf');
  console.log('Upload successful:', upload.fileId);
})();
```

### cURL Example

```bash
#!/bin/bash

# Login and save cookies
curl -c cookies.txt -X POST http://localhost:4949/login \
  -d "email=admin@wulfvault.local" \
  -d "password=your_password"

# List users
curl -b cookies.txt http://localhost:4949/api/v1/users | jq

# Create user
curl -b cookies.txt -X POST http://localhost:4949/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API User",
    "email": "apiuser@example.com",
    "password": "SecurePassword123!",
    "userLevel": 2,
    "storageQuotaMB": 10240,
    "isActive": true
  }' | jq

# Upload file
curl -b cookies.txt -X POST http://localhost:4949/api/v1/upload \
  -F "file=@document.pdf" \
  -F "requireAuth=true" \
  -F "downloadsRemaining=100" | jq

# Get system stats
curl -b cookies.txt http://localhost:4949/api/v1/admin/stats | jq
```

## Support

For issues, questions, or feature requests, please visit:
- GitHub: https://github.com/Frimurare/WulfVault
- Documentation: See README.md and USER_GUIDE.md

## License

WulfVault is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).
Copyright (c) 2025 Ulf Holmström (Frimurare)
