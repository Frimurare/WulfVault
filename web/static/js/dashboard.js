// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// Upload handling
const uploadZone = document.getElementById('uploadZone');
const fileInput = document.getElementById('fileInput');
const uploadForm = document.getElementById('uploadForm');
const uploadOptions = document.getElementById('uploadOptions');
const fileList = document.getElementById('fileList');

// Set default expiration date to 7 days from now
if (document.getElementById('expireDate')) {
    const defaultDate = new Date();
    defaultDate.setDate(defaultDate.getDate() + 7);
    document.getElementById('expireDate').valueAsDate = defaultDate;
}

// Drag and drop handlers
if (uploadZone) {
    uploadZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadZone.classList.add('drag-over');
    });

    uploadZone.addEventListener('dragleave', () => {
        uploadZone.classList.remove('drag-over');
    });

    uploadZone.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadZone.classList.remove('drag-over');
        const files = e.dataTransfer.files;
        if (files.length > 0) {
            fileInput.files = files;
            showUploadOptions(files[0]);
        }
    });
}

// File input change handler
if (fileInput) {
    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            showUploadOptions(e.target.files[0]);
        }
    });
}

// Show upload options when file is selected
function showUploadOptions(file) {
    const uploadZone = document.getElementById('uploadZone');

    // Create visual feedback div (but keep the file input intact!)
    const existingVisual = uploadZone.querySelector('.upload-visual');
    if (existingVisual) {
        existingVisual.remove();
    }

    const visualDiv = document.createElement('div');
    visualDiv.className = 'upload-visual';
    visualDiv.innerHTML = `
        <div style="text-align: center; padding: 20px;">
            <svg style="width: 48px; height: 48px; color: #4caf50; margin-bottom: 12px;" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            <h3 style="color: #333; margin-bottom: 8px;">File Selected</h3>
            <p style="color: #666; font-weight: 600;">${file.name}</p>
            <p style="color: #999; font-size: 14px;">${formatFileSize(file.size)}</p>
        </div>
    `;

    // Hide original content but keep it in DOM
    const children = Array.from(uploadZone.children);
    children.forEach(child => {
        if (child.tagName !== 'INPUT') {
            child.style.display = 'none';
        }
    });

    // Add visual feedback at the beginning
    uploadZone.insertBefore(visualDiv, uploadZone.firstChild);

    uploadZone.style.border = '3px solid #4caf50';
    uploadOptions.style.display = 'block';

    // Load user's teams for the team selector
    loadUserTeamsForUpload();
}

// Load user's teams for upload form
function loadUserTeamsForUpload() {
    const container = document.getElementById('teamSelectContainer');
    if (!container) return;

    fetch('/api/teams/my', { credentials: 'same-origin' })
        .then(response => response.json())
        .then(data => {
            if (data && data.teams && data.teams.length > 0) {
                container.innerHTML = '';
                data.teams.forEach(team => {
                    const checkbox = document.createElement('div');
                    checkbox.style.cssText = 'padding: 8px; margin-bottom: 4px; background: white; border-radius: 4px; cursor: pointer; display: flex; align-items: center; gap: 8px;';
                    checkbox.innerHTML = `
                        <input type="checkbox" id="team_${team.id}" name="team_ids[]" value="${team.id}"
                               style="width: 16px; height: 16px; cursor: pointer;">
                        <label for="team_${team.id}" style="cursor: pointer; flex: 1; margin: 0;">
                            👥 ${escapeHtml(team.name)}
                        </label>
                    `;
                    container.appendChild(checkbox);
                });
            } else {
                container.innerHTML = '<div style="color: #999; font-style: italic;">No teams available</div>';
            }
        })
        .catch(error => {
            console.error('Failed to load teams:', error);
            container.innerHTML = '<div style="color: #f44336;">Failed to load teams</div>';
        });
}

// Handle checkbox toggles
const unlimitedTimeEl = document.getElementById('unlimitedTime');
if (unlimitedTimeEl) {
    unlimitedTimeEl.addEventListener('change', function() {
        const expireDateEl = document.getElementById('expireDate');
        if (expireDateEl) {
            expireDateEl.disabled = this.checked;
        }
    });
}

const unlimitedDownloadsEl = document.getElementById('unlimitedDownloads');
if (unlimitedDownloadsEl) {
    unlimitedDownloadsEl.addEventListener('change', function() {
        const downloadsLimitEl = document.getElementById('downloadsLimit');
        if (downloadsLimitEl) {
            downloadsLimitEl.disabled = this.checked;
        }
    });
}

// Form submit handler
if (uploadForm) {
    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const formData = new FormData(uploadForm);

        // Handle link type - the backend doesn't use this, but we can log it
        const linkType = formData.get('link_type');
        console.log('Selected link type:', linkType);

        // Convert checkboxes to proper values
        formData.set('unlimited_time', document.getElementById('unlimitedTime').checked ? 'true' : 'false');
        formData.set('unlimited_downloads', document.getElementById('unlimitedDownloads').checked ? 'true' : 'false');
        formData.set('require_auth', document.getElementById('requireAuth').checked ? 'true' : 'false');

        // Handle password field - only include if checkbox is checked
        const enablePasswordCheckbox = document.getElementById('enablePassword');
        const filePasswordInput = document.getElementById('filePassword');
        if (enablePasswordCheckbox && !enablePasswordCheckbox.checked) {
            // Remove password from form if checkbox is not checked
            formData.delete('file_password');
            console.log('Password protection: DISABLED');
        } else if (filePasswordInput && filePasswordInput.value) {
            // Ensure password is included
            formData.set('file_password', filePasswordInput.value);
            console.log('Password protection: ENABLED, password:', filePasswordInput.value);
        }

        // Debug: Log all form data
        console.log('=== UPLOAD FORM DATA ===');
        for (let [key, value] of formData.entries()) {
            if (key === 'file') {
                console.log(key + ':', value.name, '(' + formatFileSize(value.size) + ')');
            } else {
                console.log(key + ':', value);
            }
        }
        console.log('========================');

        // If unlimited time, remove expire_date
        if (document.getElementById('unlimitedTime').checked) {
            formData.delete('expire_date');
        }

        // If unlimited downloads, set limit to 0
        if (document.getElementById('unlimitedDownloads').checked) {
            formData.set('downloads_limit', '0');
        }

        // Show progress
        const uploadButton = document.getElementById('uploadButton');
        uploadButton.textContent = '⏳ Uploading...';
        uploadButton.disabled = true;

        // Create large upload progress overlay
        const file = formData.get('file');
        showUploadProgressOverlay(file.name, file.size);

        // Mark transfer as active to prevent inactivity timeout
        if (window.inactivityTracker) {
            window.inactivityTracker.markTransferActive();
        }

        // Get current user ID from page context (set by server)
        const userIdElement = document.querySelector('[data-user-id]');
        const userId = userIdElement ? userIdElement.getAttribute('data-user-id') : '0';

        // Prepare metadata for tus
        const metadata = {
            user_id: userId,
            filename: file.name,
            filetype: file.type,
            expire_date: formData.get('expire_date') || '',
            downloads_limit: formData.get('downloads_limit') || '10',
            require_auth: formData.get('require_auth') || 'false',
            unlimited_time: formData.get('unlimited_time') || 'false',
            unlimited_downloads: formData.get('unlimited_downloads') || 'false',
            file_password: formData.get('file_password') || '',
            file_comment: formData.get('file_comment') || '',
            client_ip: '', // Server will fill this
            user_agent: navigator.userAgent
        };

        // Start chunked upload
        uploadFileInChunks(file, metadata, uploadButton);
    });
}

// Reset upload form
function resetUploadForm() {
    uploadForm.reset();
    uploadOptions.style.display = 'none';

    const uploadZone = document.getElementById('uploadZone');

    // Remove visual feedback if it exists
    const existingVisual = uploadZone.querySelector('.upload-visual');
    if (existingVisual) {
        existingVisual.remove();
    }

    // Show all original children again
    const children = Array.from(uploadZone.children);
    children.forEach(child => {
        child.style.display = '';
    });

    uploadZone.style.border = '3px dashed #ddd';

    // Reset the file input value
    const existingFileInput = document.getElementById('fileInput');
    if (existingFileInput) {
        existingFileInput.value = '';
    }

    // Reset date to 7 days from now
    const expireDateInput = document.getElementById('expireDate');
    if (expireDateInput) {
        const defaultDate = new Date();
        defaultDate.setDate(defaultDate.getDate() + 7);
        expireDateInput.valueAsDate = defaultDate;
    }
}

// Format file size
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// Copy to clipboard function with fallback for HTTP connections
function copyToClipboard(text, button) {
    // Try modern clipboard API first (requires HTTPS or localhost)
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
            const originalText = button.textContent;
            button.textContent = '✓ Copied!';
            button.style.background = '#28a745';
            setTimeout(() => {
                button.textContent = originalText;
                button.style.background = '';
            }, 2000);
        }).catch(() => {
            // If clipboard API fails, use fallback
            fallbackCopyToClipboard(text, button);
        });
    } else {
        // Use fallback for older browsers or HTTP connections
        fallbackCopyToClipboard(text, button);
    }
}

// Fallback copy function using execCommand (works on HTTP)
function fallbackCopyToClipboard(text, button) {
    const textArea = document.createElement("textarea");
    textArea.value = text;
    textArea.style.position = "fixed";
    textArea.style.left = "-999999px";
    textArea.style.top = "0";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
        const successful = document.execCommand('copy');
        if (successful) {
            const originalText = button.textContent;
            button.textContent = '✓ Copied!';
            button.style.background = '#28a745';
            setTimeout(() => {
                button.textContent = originalText;
                button.style.background = '';
            }, 2000);
        } else {
            showError('Failed to copy link');
        }
    } catch (err) {
        showError('Failed to copy: ' + err);
    }

    document.body.removeChild(textArea);
}

// Delete file function
function deleteFile(fileId, fileName) {
    if (!confirm(`Delete "${fileName}"?`)) return;

    fetch('/file/delete', {
        method: 'POST',
        headers: {'Content-Type': 'application/x-www-form-urlencoded'},
        body: 'file_id=' + fileId
    })
    .then(res => res.json())
    .then(data => {
        showSuccess('File moved to trash');
        setTimeout(() => window.location.reload(), 1000);
    })
    .catch(err => {
        showError('Failed to delete file');
    });
}

// Show success message
function showSuccess(message) {
    const toast = document.createElement('div');
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: #4caf50;
        color: white;
        padding: 16px 24px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        z-index: 10000;
        font-weight: 500;
    `;
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}

// Show error message
function showError(message) {
    const toast = document.createElement('div');
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: #f44336;
        color: white;
        padding: 16px 24px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        z-index: 10000;
        font-weight: 500;
    `;
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 4000);
}

// Note: showEditModal, togglePasswordField, togglePasswordVisibility, and showDownloadHistory
// are defined in the inline script in handlers_user.go and handlers_admin.go
// Do not define them here to avoid conflicts

// File Request functions
function showCreateRequestModal() {
    const modal = document.getElementById('fileRequestModal');
    if (modal) {
        modal.style.display = 'flex';
        // Reset form
        document.getElementById('fileRequestForm').reset();
        document.getElementById('requestMaxSize').value = 1; // Default 1 GB
    }
}

function closeFileRequestModal() {
    const modal = document.getElementById('fileRequestModal');
    if (modal) {
        modal.style.display = 'none';
    }
}

function submitFileRequest(event) {
    event.preventDefault();

    const title = document.getElementById('requestTitle').value;
    const message = document.getElementById('requestMessage').value;
    const maxSizeGB = document.getElementById('requestMaxSize').value;
    const recipientEmail = document.getElementById('requestRecipientEmail').value;

    // Convert GB to MB for backend (backend expects MB)
    const maxSizeMB = Math.round(parseFloat(maxSizeGB) * 1024);

    console.log('Creating file request:', {title, message, maxSizeGB, maxSizeMB, recipientEmail});

    const data = new FormData();
    data.append('title', title);
    data.append('message', message);
    data.append('max_file_size_mb', maxSizeMB);
    if (recipientEmail) {
        data.append('recipient_email', recipientEmail);
    }

    fetch('/file-request/create', {
        method: 'POST',
        body: data,
        credentials: 'same-origin'
    })
    .then(response => response.json())
    .then(result => {
        console.log('File request result:', result);
        if (result.success) {
            closeFileRequestModal();
            showSuccess('Upload request created! The link is shown below.');
            loadFileRequests();
        } else {
            alert('Error: ' + (result.error || 'Unknown error'));
        }
    })
    .catch(error => {
        console.error('Error creating request:', error);
        alert('Error creating request: ' + error);
    });
}

// Close modal when clicking outside
window.addEventListener('click', function(event) {
    const modal = document.getElementById('fileRequestModal');
    if (event.target === modal) {
        closeFileRequestModal();
    }
});

function loadFileRequests() {
    fetch('/file-request/list', {
        credentials: 'same-origin'
    })
        .then(response => response.json())
        .then(data => {
            const container = document.getElementById('requestsList');
            if (!container) return;

            if (!data.requests || data.requests.length === 0) {
                container.innerHTML = '<p style="color: #999; font-style: italic;">No upload requests yet</p>';
                return;
            }

            const now = Math.floor(Date.now() / 1000);
            let html = '<div style="margin-top: 20px;">';

            data.requests.forEach(req => {
                const expiresAt = req.expires_at;
                const timeDiff = expiresAt - now;

                let expiryStatus = '';
                let borderColor = '#e0e0e0';
                let bgColor = 'white';

                if (req.is_expired) {
                    // Calculate days until auto-removal
                    const expiredFor = now - expiresAt;
                    const fiveDays = 5 * 24 * 60 * 60;
                    const daysUntilRemoval = Math.max(0, Math.ceil((fiveDays - expiredFor) / (24 * 60 * 60)));

                    expiryStatus = '<span style="color: #f44336; font-weight: 600;">⏰ EXPIRED</span> - ' +
                                   '<span style="color: #ff9800;">Auto-removal in ' + daysUntilRemoval + ' day' + (daysUntilRemoval !== 1 ? 's' : '') + '</span>';
                    borderColor = '#f44336';
                    bgColor = '#fff5f5';
                } else {
                    // Calculate hours until expiry
                    const hoursUntilExpiry = Math.max(0, Math.floor(timeDiff / 3600));
                    const minutesRemaining = Math.max(0, Math.floor((timeDiff % 3600) / 60));

                    if (hoursUntilExpiry > 0) {
                        expiryStatus = '<span style="color: #4caf50; font-weight: 500;">✓ Expires in ' + hoursUntilExpiry + ' hour' + (hoursUntilExpiry !== 1 ? 's' : '') + '</span>';
                    } else if (minutesRemaining > 0) {
                        expiryStatus = '<span style="color: #ff9800; font-weight: 500;">⚠️ Expires in ' + minutesRemaining + ' minute' + (minutesRemaining !== 1 ? 's' : '') + '</span>';
                        borderColor = '#ff9800';
                        bgColor = '#fff8e1';
                    } else {
                        expiryStatus = '<span style="color: #f44336; font-weight: 600;">⏰ Expiring soon...</span>';
                        borderColor = '#f44336';
                        bgColor = '#fff5f5';
                    }
                }

                const active = req.is_active ? '✅' : '❌';
                html += '<div style="border: 2px solid ' + borderColor + '; background: ' + bgColor + '; padding: 16px; margin-bottom: 12px; border-radius: 8px; transition: all 0.3s;">';
                html += '<div style="display: flex; justify-content: space-between; align-items: start; margin-bottom: 8px;">';
                html += '<h4 style="margin: 0; flex: 1;">' + active + ' ' + escapeHtml(req.title) + '</h4>';
                html += '<div style="text-align: right; font-size: 13px;">' + expiryStatus + '</div>';
                html += '</div>';

                if (req.message) {
                    html += '<p style="color: #666; font-size: 14px; margin-bottom: 12px;">' + escapeHtml(req.message) + '</p>';
                }

                html += '<div style="display: flex; gap: 12px; align-items: center; flex-wrap: wrap;">';
                html += '<input type="text" value="' + req.upload_url + '" readonly style="flex: 1; padding: 8px; border: 1px solid #ddd; border-radius: 4px; font-family: monospace; font-size: 12px;">';
                html += '<button onclick="copyToClipboard(\''+req.upload_url+'\', this)" style="padding: 8px 16px; background: #2196f3; color: white; border: none; border-radius: 4px; cursor: pointer;">📋 Copy</button>';
                html += '<button class="delete-request-btn" data-request-id="'+req.id+'" data-request-title="'+escapeHtml(req.title)+'" style="padding: 8px 16px; background: #f44336; color: white; border: none; border-radius: 4px; cursor: pointer;">🗑️ Delete</button>';
                html += '</div></div>';
            });
            html += '</div>';
            container.innerHTML = html;

            // Add event listeners for delete buttons
            document.querySelectorAll('.delete-request-btn').forEach(btn => {
                btn.addEventListener('click', function() {
                    const id = this.getAttribute('data-request-id');
                    const title = this.getAttribute('data-request-title');
                    deleteFileRequest(parseInt(id), title);
                });
            });

            // Refresh every minute to update countdowns
            setTimeout(loadFileRequests, 60000);
        });
}

function deleteFileRequest(id, title) {
    if (!confirm('Delete request: ' + title + '?')) return;

    const data = new FormData();
    data.append('request_id', id);

    fetch('/file-request/delete', {
        method: 'POST',
        body: data,
        credentials: 'same-origin'
    })
    .then(response => response.json())
    .then(result => {
        if (result.success) {
            loadFileRequests();
        } else {
            alert('Error: ' + (result.error || 'Unknown error'));
        }
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// ============================================================================
// CHUNKED UPLOAD IMPLEMENTATION
// ============================================================================

async function uploadFileInChunks(file, metadata, uploadButton) {
    const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB chunks
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
    let retryCount = 0;
    const MAX_RETRIES = 10;

    try {
        // Step 1: Initialize upload
        const initResponse = await fetch('/api/upload/init', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'same-origin',
            body: JSON.stringify({
                filename: file.name,
                total_size: file.size,
                metadata: metadata
            })
        });

        if (!initResponse.ok) {
            throw new Error('Failed to initialize upload');
        }

        const { upload_id } = await initResponse.json();
        console.log(`Upload initialized: ${upload_id}, ${totalChunks} chunks`);

        // Step 2: Upload chunks
        for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
            const start = chunkIndex * CHUNK_SIZE;
            const end = Math.min(start + CHUNK_SIZE, file.size);
            const chunk = file.slice(start, end);

            let chunkUploaded = false;
            let attempts = 0;

            while (!chunkUploaded && attempts < MAX_RETRIES) {
                try {
                    const chunkResponse = await fetch(`/api/upload/chunk?upload_id=${upload_id}&chunk_index=${chunkIndex}`, {
                        method: 'POST',
                        body: chunk,
                        credentials: 'same-origin'
                    });

                    if (!chunkResponse.ok) {
                        throw new Error(`Chunk ${chunkIndex} upload failed`);
                    }

                    const result = await chunkResponse.json();
                    chunkUploaded = true;

                    // Update progress
                    const percentComplete = Math.round((result.bytes_received / result.total_size) * 100);
                    uploadButton.textContent = `⏳ Uploading... ${percentComplete}%`;
                    updateUploadProgress(percentComplete, result.bytes_received, result.total_size);

                    console.log(`Chunk ${chunkIndex + 1}/${totalChunks} uploaded (${percentComplete}%)`);

                } catch (error) {
                    attempts++;
                    retryCount++;
                    console.error(`Chunk ${chunkIndex} failed (attempt ${attempts}/${MAX_RETRIES}):`, error);

                    if (attempts < MAX_RETRIES) {
                        showRetryIndicator(retryCount);
                        // Wait before retry with exponential backoff
                        await new Promise(resolve => setTimeout(resolve, Math.min(1000 * Math.pow(2, attempts - 1), 10000)));
                    } else {
                        throw new Error(`Chunk ${chunkIndex} failed after ${MAX_RETRIES} attempts`);
                    }
                }
            }
        }

        // Step 3: Complete upload
        const completeResponse = await fetch(`/api/upload/complete?upload_id=${upload_id}`, {
            method: 'POST',
            credentials: 'same-origin'
        });

        if (!completeResponse.ok) {
            throw new Error('Failed to complete upload');
        }

        const result = await completeResponse.json();
        console.log('Upload completed successfully:', result);

        // Mark transfer as inactive
        if (window.inactivityTracker) {
            window.inactivityTracker.markTransferInactive();
        }

        // Show success
        showUploadSuccess();

        // Reload page after showing success animation
        setTimeout(() => window.location.reload(), 3000);

    } catch (error) {
        // Mark transfer as inactive
        if (window.inactivityTracker) {
            window.inactivityTracker.markTransferInactive();
        }

        console.error('Upload failed:', error);
        showUploadError(error, retryCount);

        uploadButton.textContent = '📤 Upload File';
        uploadButton.disabled = false;
    }
}

// ============================================================================
// UPLOAD PROGRESS OVERLAY - Large Visual Feedback
// ============================================================================

function showUploadProgressOverlay(filename, filesize) {
    // Create overlay element
    const overlay = document.createElement('div');
    overlay.id = 'uploadProgressOverlay';
    overlay.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.9);
        z-index: 10000;
        display: flex;
        align-items: center;
        justify-content: center;
        animation: fadeIn 0.3s ease;
    `;

    // Create progress container
    const container = document.createElement('div');
    container.style.cssText = `
        text-align: center;
        padding: 60px;
        max-width: 700px;
        width: 90%;
    `;

    // Upload status text
    const statusText = document.createElement('div');
    statusText.id = 'uploadStatusText';
    statusText.style.cssText = `
        font-size: 72px;
        font-weight: bold;
        color: #ef4444;
        margin-bottom: 30px;
        text-shadow: 0 0 20px rgba(239, 68, 68, 0.5);
        animation: pulse 2s ease-in-out infinite;
    `;
    statusText.textContent = 'UPLOADING - 0%';

    // File info
    const fileInfo = document.createElement('div');
    fileInfo.style.cssText = `
        font-size: 24px;
        color: #e5e7eb;
        margin-bottom: 40px;
        font-weight: 500;
    `;
    fileInfo.textContent = filename;

    // File size
    const sizeInfo = document.createElement('div');
    sizeInfo.id = 'uploadSizeInfo';
    sizeInfo.style.cssText = `
        font-size: 18px;
        color: #9ca3af;
        margin-bottom: 50px;
    `;
    sizeInfo.textContent = `0 B / ${formatFileSize(filesize)}`;

    // Progress bar container
    const progressBarContainer = document.createElement('div');
    progressBarContainer.style.cssText = `
        width: 100%;
        height: 40px;
        background: rgba(255, 255, 255, 0.1);
        border-radius: 20px;
        overflow: hidden;
        margin-bottom: 20px;
        box-shadow: 0 0 30px rgba(0, 0, 0, 0.5);
    `;

    // Progress bar fill
    const progressBarFill = document.createElement('div');
    progressBarFill.id = 'uploadProgressBarFill';
    progressBarFill.style.cssText = `
        height: 100%;
        width: 0%;
        background: linear-gradient(90deg, #ef4444, #dc2626);
        transition: width 0.3s ease, background 0.5s ease;
        border-radius: 20px;
        box-shadow: 0 0 20px rgba(239, 68, 68, 0.8);
    `;

    // Speed and ETA info
    const speedInfo = document.createElement('div');
    speedInfo.id = 'uploadSpeedInfo';
    speedInfo.style.cssText = `
        font-size: 16px;
        color: #9ca3af;
        margin-top: 20px;
    `;
    speedInfo.textContent = 'Calculating speed...';

    // Retry indicator
    const retryInfo = document.createElement('div');
    retryInfo.id = 'uploadRetryInfo';
    retryInfo.style.cssText = `
        font-size: 14px;
        color: #fbbf24;
        margin-top: 15px;
        font-weight: 600;
        display: none;
    `;

    progressBarContainer.appendChild(progressBarFill);
    container.appendChild(statusText);
    container.appendChild(fileInfo);
    container.appendChild(sizeInfo);
    container.appendChild(progressBarContainer);
    container.appendChild(speedInfo);
    container.appendChild(retryInfo);
    overlay.appendChild(container);

    // Add CSS animations
    const style = document.createElement('style');
    style.textContent = `
        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }
        @keyframes pulse {
            0%, 100% { transform: scale(1); }
            50% { transform: scale(1.05); }
        }
        @keyframes successPulse {
            0%, 100% { transform: scale(1); }
            50% { transform: scale(1.1); }
        }
    `;
    document.head.appendChild(style);

    document.body.appendChild(overlay);

    // Store start time for speed calculation
    window.uploadStartTime = Date.now();
    window.uploadStartLoaded = 0;
}

function updateUploadProgress(percent, loaded, total) {
    const statusText = document.getElementById('uploadStatusText');
    const progressBarFill = document.getElementById('uploadProgressBarFill');
    const sizeInfo = document.getElementById('uploadSizeInfo');
    const speedInfo = document.getElementById('uploadSpeedInfo');

    if (!statusText) return;

    // Update status text
    statusText.textContent = `UPLOADING - ${percent}%`;

    // Update progress bar
    progressBarFill.style.width = `${percent}%`;

    // Update size info
    sizeInfo.textContent = `${formatFileSize(loaded)} / ${formatFileSize(total)}`;

    // Calculate speed and ETA
    const now = Date.now();
    const timeElapsed = (now - window.uploadStartTime) / 1000; // seconds
    const bytesUploaded = loaded - window.uploadStartLoaded;

    if (timeElapsed > 0) {
        const speed = bytesUploaded / timeElapsed; // bytes per second
        const remainingBytes = total - loaded;
        const eta = remainingBytes / speed; // seconds

        speedInfo.textContent = `Speed: ${formatFileSize(speed)}/s | ETA: ${formatTime(eta)}`;
    }

    // Update last measurement
    window.uploadStartTime = now;
    window.uploadStartLoaded = loaded;
}

function showUploadSuccess() {
    const statusText = document.getElementById('uploadStatusText');
    const progressBarFill = document.getElementById('uploadProgressBarFill');
    const speedInfo = document.getElementById('uploadSpeedInfo');

    if (!statusText) return;

    // Change to green and show 100%
    statusText.textContent = 'UPLOAD COMPLETE - 100%';
    statusText.style.color = '#10b981';
    statusText.style.textShadow = '0 0 20px rgba(16, 185, 129, 0.8)';
    statusText.style.animation = 'successPulse 0.8s ease-in-out 3';

    // Update progress bar to green
    progressBarFill.style.width = '100%';
    progressBarFill.style.background = 'linear-gradient(90deg, #10b981, #059669)';
    progressBarFill.style.boxShadow = '0 0 20px rgba(16, 185, 129, 0.8)';

    // Update speed info
    if (speedInfo) {
        speedInfo.textContent = '✓ File uploaded successfully!';
        speedInfo.style.color = '#10b981';
        speedInfo.style.fontSize = '20px';
        speedInfo.style.fontWeight = 'bold';
    }

    // Add large green "PRESS HERE TO GO BACK" button
    const overlay = document.getElementById('uploadProgressOverlay');
    if (overlay) {
        const backButton = document.createElement('button');
        backButton.textContent = 'PRESS HERE TO GO BACK';
        backButton.style.cssText = `
            margin-top: 40px;
            padding: 25px 60px;
            font-size: 24px;
            font-weight: bold;
            color: white;
            background: linear-gradient(135deg, #10b981 0%, #059669 100%);
            border: none;
            border-radius: 15px;
            cursor: pointer;
            box-shadow: 0 8px 24px rgba(16, 185, 129, 0.4);
            transition: all 0.3s ease;
            text-transform: uppercase;
            letter-spacing: 1px;
        `;
        backButton.onmouseover = () => {
            backButton.style.transform = 'translateY(-3px)';
            backButton.style.boxShadow = '0 12px 32px rgba(16, 185, 129, 0.5)';
        };
        backButton.onmouseout = () => {
            backButton.style.transform = 'translateY(0)';
            backButton.style.boxShadow = '0 8px 24px rgba(16, 185, 129, 0.4)';
        };
        backButton.onclick = () => window.location.reload();
        overlay.querySelector('div').appendChild(backButton);
    }
}

function showUploadError(error, retryCount) {
    const statusText = document.getElementById('uploadStatusText');
    const progressBarFill = document.getElementById('uploadProgressBarFill');
    const speedInfo = document.getElementById('uploadSpeedInfo');
    const retryInfo = document.getElementById('uploadRetryInfo');

    if (!statusText) return;

    // Change to red error state
    statusText.textContent = 'UPLOAD FAILED';
    statusText.style.color = '#ef4444';
    statusText.style.textShadow = '0 0 20px rgba(239, 68, 68, 0.8)';
    statusText.style.animation = 'pulse 2s ease-in-out infinite';

    // Update progress bar to red
    progressBarFill.style.background = 'linear-gradient(90deg, #ef4444, #dc2626)';
    progressBarFill.style.boxShadow = '0 0 20px rgba(239, 68, 68, 0.8)';

    // Show error details
    if (speedInfo) {
        let errorMsg = error.message || 'Unknown error';
        if (retryCount > 0) {
            errorMsg += `\n\nFailed after ${retryCount} retry attempts.`;
        }
        speedInfo.innerHTML = `<div style="color: #fca5a5; font-size: 16px; line-height: 1.6; white-space: pre-wrap; max-width: 600px; margin: 0 auto;">${errorMsg}</div>`;
    }

    // Hide retry info
    if (retryInfo) {
        retryInfo.style.display = 'none';
    }

    // Add close button
    const overlay = document.getElementById('uploadProgressOverlay');
    if (overlay) {
        const closeBtn = document.createElement('button');
        closeBtn.textContent = 'Close';
        closeBtn.style.cssText = `
            margin-top: 30px;
            padding: 15px 40px;
            font-size: 18px;
            font-weight: bold;
            color: white;
            background: #ef4444;
            border: none;
            border-radius: 10px;
            cursor: pointer;
            transition: background 0.3s ease;
        `;
        closeBtn.onmouseover = () => closeBtn.style.background = '#dc2626';
        closeBtn.onmouseout = () => closeBtn.style.background = '#ef4444';
        closeBtn.onclick = () => {
            overlay.style.animation = 'fadeOut 0.3s ease';
            setTimeout(() => overlay.remove(), 300);
        };
        overlay.querySelector('div').appendChild(closeBtn);
    }
}

function hideUploadProgressOverlay() {
    const overlay = document.getElementById('uploadProgressOverlay');
    if (overlay) {
        overlay.style.animation = 'fadeOut 0.3s ease';
        setTimeout(() => overlay.remove(), 300);
    }
}

function showRetryIndicator(retryCount) {
    const retryInfo = document.getElementById('uploadRetryInfo');
    if (retryInfo) {
        retryInfo.style.display = 'block';
        retryInfo.textContent = `⚠️ Connection interrupted - Retry attempt ${retryCount} of 10...`;

        // Flash animation
        retryInfo.style.animation = 'pulse 1s ease-in-out 3';

        console.log(`Retry ${retryCount}/10: Network interruption detected, retrying upload...`);
    }
}

function formatTime(seconds) {
    if (seconds < 60) {
        return Math.round(seconds) + 's';
    } else if (seconds < 3600) {
        const mins = Math.floor(seconds / 60);
        const secs = Math.round(seconds % 60);
        return `${mins}m ${secs}s`;
    } else {
        const hours = Math.floor(seconds / 3600);
        const mins = Math.floor((seconds % 3600) / 60);
        return `${hours}h ${mins}m`;
    }
}

// Load file requests when page loads
window.addEventListener('load', function() {
    loadFileRequests();
});
