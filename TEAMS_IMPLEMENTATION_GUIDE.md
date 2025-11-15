# Teams Frontend Implementation Guide

This document provides code snippets and guidance for implementing missing teams features.

---

## 1. Adding Team Files to Dashboard

### Current Code (handlers_user.go:370)
```go
// Only shows user's own files
files, _ := database.DB.GetFilesByUser(user.Id)
```

### Option A: Show Only Personal Files (Current)
No changes needed.

### Option B: Show Personal + Team Files Combined
```go
// Get user's personal files
userFiles, _ := database.DB.GetFilesByUser(user.Id)

// Get user's team memberships
teams, _ := database.DB.GetTeamsByUser(user.Id)

// Get files shared with each team
var teamFiles []*database.FileInfo
for _, team := range teams {
    teamFileRecords, _ := database.DB.GetTeamFiles(team.Id)
    for _, tf := range teamFileRecords {
        file, _ := database.DB.GetFileByID(tf.FileId)
        if file != nil {
            teamFiles = append(teamFiles, file)
        }
    }
}

// Combine both lists
files = append(userFiles, teamFiles...)
```

### Option C: Use New Combined Database Function
Proposed new function in `database/teams.go`:
```go
// GetFilesByUserWithTeams returns both user files and team files
func (d *Database) GetFilesByUserWithTeams(userId int) ([]*FileInfo, error) {
    // Get personal files
    userFiles, err := d.GetFilesByUser(userId)
    if err != nil {
        return nil, err
    }
    
    // Get team files
    teams, err := d.GetTeamsByUser(userId)
    if err != nil {
        return userFiles, nil // Return just personal files if team fetch fails
    }
    
    var allFiles []*FileInfo
    allFiles = append(allFiles, userFiles...)
    
    for _, team := range teams {
        teamFileRecords, err := d.GetTeamFiles(team.Id)
        if err != nil {
            continue
        }
        
        for _, tf := range teamFileRecords {
            file, err := d.GetFileByID(tf.FileId)
            if err == nil && file != nil {
                // Add indicator that this is a team file
                file.TeamId = team.Id
                file.TeamName = team.Name
                allFiles = append(allFiles, file)
            }
        }
    }
    
    return allFiles, nil
}
```

### UI Enhancement: Add Team Badge to Files
In handlers_user.go around line 1006, enhance file display:

```html
<!-- Current: Just filename -->
<h3>📄 %s %s%s</h3>

<!-- Enhanced: Show team indicator -->
<h3>📄 %s %s%s</h3>
```

Then in the Go rendering code:
```go
teamBadge := ""
if f.TeamId > 0 {
    teamBadge = fmt.Sprintf(` + "`" + `<span style="background: #e0e7ff; color: #3730a3; padding: 2px 8px; border-radius: 4px; font-size: 12px; margin-left: 8px;">👥 Team: %s</span>` + "`" + `, f.TeamName)
}

html += fmt.Sprintf(`...%s%s</h3>...`, authBadge, passwordBadge, teamBadge)
```

---

## 2. Add Tab/Filter for File Types

### HTML Structure (handlers_user.go)
Insert before file list:
```html
<div style="margin-bottom: 20px; display: flex; gap: 12px; padding: 0 24px; padding-top: 20px; border-bottom: 1px solid #e0e0e0;">
    <button class="file-filter-btn active" data-filter="all" onclick="filterFiles('all')" 
            style="padding: 8px 16px; background: transparent; border: none; border-bottom: 3px solid #2563eb; color: #2563eb; cursor: pointer; font-weight: 600;">
        All Files
    </button>
    <button class="file-filter-btn" data-filter="personal" onclick="filterFiles('personal')" 
            style="padding: 8px 16px; background: transparent; border: none; border-bottom: 3px solid transparent; color: #999; cursor: pointer; font-weight: 600;">
        My Files
    </button>
    <button class="file-filter-btn" data-filter="team" onclick="filterFiles('team')" 
            style="padding: 8px 16px; background: transparent; border: none; border-bottom: 3px solid transparent; color: #999; cursor: pointer; font-weight: 600;">
        Team Files
    </button>
</div>
```

### JavaScript (In handlers_user.go script section)
```javascript
function filterFiles(filterType) {
    // Update active button
    document.querySelectorAll('.file-filter-btn').forEach(btn => {
        btn.classList.remove('active');
        btn.style.color = '#999';
        btn.style.borderBottom = '3px solid transparent';
    });
    
    document.querySelector(`[data-filter="${filterType}"]`).classList.add('active');
    document.querySelector(`[data-filter="${filterType}"]`).style.color = '#2563eb';
    document.querySelector(`[data-filter="${filterType}"]`).style.borderBottom = '3px solid #2563eb';
    
    // Filter files
    const fileItems = document.querySelectorAll('.file-item');
    fileItems.forEach(item => {
        const isTeamFile = item.getAttribute('data-team-id') !== null && 
                          item.getAttribute('data-team-id') !== '';
        
        if (filterType === 'all') {
            item.style.display = 'flex';
        } else if (filterType === 'personal') {
            item.style.display = isTeamFile ? 'none' : 'flex';
        } else if (filterType === 'team') {
            item.style.display = isTeamFile ? 'flex' : 'none';
        }
    });
}
```

### Add data attributes to file items (handlers_user.go line ~972)
```go
teamId := ""
if f.TeamId > 0 {
    teamId = fmt.Sprintf("%d", f.TeamId)
}

html += fmt.Sprintf(`
    <li class="file-item" data-team-id="%s">
        ...
    </li>`, teamId)
```

---

## 3. Add Team Selector to Upload Form

### Option A: Simple Radio Button During Upload

In handlers_user.go upload form (around line 770), add before button:

```html
<div class="form-group" id="uploadTeamShareSection" style="display: none;">
    <label>📁 Share with Team (optional)</label>
    <select id="uploadTeamSelect" name="team_id">
        <option value="">-- Don't share --</option>
    </select>
    <p style="color: #666; font-size: 12px; margin-top: 4px;">
        Select a team to share this file with immediately after upload
    </p>
</div>
```

### JavaScript to Load Teams on Upload Form (handlers_user.go)
```javascript
// Load user's teams for upload form
function loadUserTeamsForUpload() {
    fetch('/api/teams/my', {credentials: 'same-origin'})
        .then(response => response.json())
        .then(data => {
            const select = document.getElementById('uploadTeamSelect');
            if (!select) return;
            
            if (data.success && data.teams && data.teams.length > 0) {
                data.teams.forEach(team => {
                    const option = document.createElement('option');
                    option.value = team.id;
                    option.textContent = team.name;
                    select.appendChild(option);
                });
                
                // Show team section only if user has teams
                document.getElementById('uploadTeamShareSection').style.display = 'block';
            }
        });
}

// Call when page loads
window.addEventListener('load', function() {
    loadUserTeamsForUpload();
});
```

### Modify Form Submission (dashboard.js line 115)
```javascript
// In uploadForm.addEventListener('submit'...):

// Add team_id if selected
const teamId = document.getElementById('uploadTeamSelect')?.value;
if (teamId) {
    formData.append('team_id', teamId);
}
```

### Backend Update (handlers_files.go::handleUpload)

After file is saved to database (around line 170), check for team_id:

```go
// Check if file should be shared to team
teamIDStr := r.FormValue("team_id")
if teamIDStr != "" {
    teamID, err := strconv.Atoi(teamIDStr)
    if err == nil {
        // Verify user is team member
        isMember, err := database.DB.IsTeamMember(teamID, user.Id)
        if err == nil && isMember {
            // Share file to team
            if err := database.DB.ShareFileToTeam(fileID, teamID, user.Id); err != nil {
                log.Printf("Warning: Failed to share uploaded file to team: %v", err)
            } else {
                log.Printf("File %s shared to team %d on upload", fileID, teamID)
            }
        }
    }
}
```

---

## 4. Add Team Navigation to Header

### In handlers_user.go renderUserDashboard() header section (around line 730):

```html
<!-- Current navigation -->
<a href="/dashboard">Dashboard</a>
<a href="/teams">Teams</a>
<a href="/settings">Settings</a>

<!-- Enhanced with dropdown or direct link -->
<!-- Option A: Direct link (simple) -->
<a href="/teams">👥 Teams</a>

<!-- Option B: With badge showing team count -->
<a href="/teams">👥 Teams <span id="teamCountBadge"></span></a>
```

### Add team count JavaScript (handlers_user.go):

```javascript
// Load team count for header badge
function loadTeamCountBadge() {
    fetch('/api/teams/my', {credentials: 'same-origin'})
        .then(response => response.json())
        .then(data => {
            const badge = document.getElementById('teamCountBadge');
            if (badge && data.teams && data.teams.length > 0) {
                badge.textContent = `(${data.teams.length})`;
                badge.style.fontSize = '12px';
                badge.style.color = '#999';
            }
        });
}

window.addEventListener('load', loadTeamCountBadge);
```

---

## 5. File Models Enhancement

### In database/files.go or models/FileList.go, add fields:

```go
type FileInfo struct {
    // ... existing fields ...
    
    // Team fields (new)
    TeamId    int    `json:"teamId,omitempty"`
    TeamName  string `json:"teamName,omitempty"`
    SharedAt  int64  `json:"sharedAt,omitempty"`
    SharedBy  string `json:"sharedBy,omitempty"`
}
```

---

## 6. Team Settings Page (User-Facing)

### New Handler (handlers_teams.go):

```go
// handleUserTeamSettings displays team settings for members
func (s *Server) handleUserTeamSettings(w http.ResponseWriter, r *http.Request) {
    user, _ := userFromContext(r.Context())
    
    teamIdStr := r.URL.Query().Get("id")
    teamId, err := strconv.Atoi(teamIdStr)
    if err != nil {
        http.Error(w, "Invalid team ID", http.StatusBadRequest)
        return
    }
    
    // Verify user is team member
    isMember, _ := database.DB.IsTeamMember(teamId, user.Id)
    if !isMember && !user.IsAdmin() {
        http.Error(w, "Access denied", http.StatusForbidden)
        return
    }
    
    team, _ := database.DB.GetTeamByID(teamId)
    members, _ := database.DB.GetTeamMembers(teamId)
    
    // Render team settings page
    // Should show:
    // - Team info (name, description, storage)
    // - Team members list
    // - Leave team button (if not owner)
    // - Edit permissions (if owner/admin)
}
```

### Add route in server.go:

```go
mux.HandleFunc("/teams/settings", s.requireAuth(s.handleUserTeamSettings))
```

---

## 7. File Upload to Team Storage (Future)

### Database schema addition:

```sql
ALTER TABLE files ADD COLUMN team_id INT DEFAULT NULL;
ALTER TABLE files ADD CONSTRAINT fk_files_team FOREIGN KEY (team_id) REFERENCES teams(id);
```

### Then files could be uploaded directly to team with team quota enforcement:

```go
// In handleUpload():
teamId := r.FormValue("team_id")
if teamId != "" {
    // Upload to team storage instead of user storage
    // Deduct from team quota
    // etc.
}
```

---

## Database Functions Already Available

### Use these existing functions:

```go
// From database/teams.go:
GetTeamsByUser(userId int)           // Get teams for user
GetTeamMembers(teamId int)           // Get team members
GetTeamFiles(teamId int)             // Get files in team
ShareFileToTeam(fileId, teamId, by)  // Share file to team
GetFileTeams(fileId string)          // Get teams for file
GetFilesByUserWithTeams(userId int)  // Get personal + team files
```

---

## Testing Checklist

- [ ] Team selector appears on upload form
- [ ] Files can be shared to teams during upload
- [ ] Team files appear in dashboard
- [ ] Filter tabs work correctly
- [ ] Team badges display properly
- [ ] Team count updates in header
- [ ] File edit modal still works for post-upload sharing
- [ ] Download history shows team files
- [ ] Unshare functionality works
- [ ] Team member invitations send emails

