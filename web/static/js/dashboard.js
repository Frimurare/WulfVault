// Manvarg Sharecare Dashboard JavaScript

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

        // Create XMLHttpRequest for progress tracking
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
                const percentComplete = Math.round((e.loaded / e.total) * 100);
                uploadButton.textContent = `⏳ Uploading... ${percentComplete}%`;
            }
        });

        xhr.addEventListener('load', () => {
            if (xhr.status === 200) {
                const response = JSON.parse(xhr.responseText);
                showSuccess('File uploaded successfully!');

                // Reload page after successful upload
                setTimeout(() => window.location.reload(), 1500);
            } else {
                let errorMsg = 'Upload failed';
                try {
                    const errorResponse = JSON.parse(xhr.responseText);
                    errorMsg = errorResponse.error || errorMsg;
                } catch (e) {
                    errorMsg = xhr.statusText || errorMsg;
                }
                showError(errorMsg);
                uploadButton.textContent = '📤 Upload File';
                uploadButton.disabled = false;
            }
        });

        xhr.addEventListener('error', () => {
            showError('Upload failed - network error');
            uploadButton.textContent = '📤 Upload File';
            uploadButton.disabled = false;
        });

        xhr.open('POST', '/upload');
        xhr.send(formData);
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
        document.getElementById('requestMaxSize').value = 100;
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
    const maxSizeMB = document.getElementById('requestMaxSize').value;
    const recipientEmail = document.getElementById('requestRecipientEmail').value;

    console.log('Creating file request:', {title, message, maxSizeMB, recipientEmail});

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

            let html = '<div style="margin-top: 20px;">';
            data.requests.forEach(req => {
                const expired = req.is_expired ? ' (EXPIRED)' : '';
                const active = req.is_active ? '✅' : '❌';
                html += '<div style="border: 1px solid #e0e0e0; padding: 16px; margin-bottom: 12px; border-radius: 8px;">';
                html += '<h4 style="margin-bottom: 8px;">' + active + ' ' + escapeHtml(req.title) + expired + '</h4>';
                if (req.message) {
                    html += '<p style="color: #666; font-size: 14px; margin-bottom: 8px;">' + escapeHtml(req.message) + '</p>';
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

// Load file requests when page loads
window.addEventListener('load', function() {
    loadFileRequests();
});
