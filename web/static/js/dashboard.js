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
document.getElementById('unlimitedTime').addEventListener('change', function() {
    document.getElementById('expireDate').disabled = this.checked;
});

document.getElementById('unlimitedDownloads').addEventListener('change', function() {
    document.getElementById('downloadsLimit').disabled = this.checked;
});

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

// Edit modal functions (if needed)
function showEditModal(fileId, fileName, downloadsRemaining, expireAt, unlimitedDownloads, unlimitedTime) {
    // TODO: Implement edit modal if needed
    alert('Edit functionality coming soon!');
}
