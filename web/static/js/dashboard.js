// Sharecare Dashboard JavaScript

// Upload handling
const uploadZone = document.getElementById('uploadZone');
const fileInput = document.getElementById('fileInput');
const fileList = document.getElementById('fileList');

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
            uploadFile(files[0]);
        }
    });
}

if (fileInput) {
    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            uploadFile(e.target.files[0]);
        }
    });
}

// Upload file function
function uploadFile(file) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('expiration_days', '7'); // Default 7 days
    formData.append('downloads_limit', '0'); // Unlimited
    formData.append('require_auth', 'false'); // Direct link by default

    // Show upload progress
    const progressHTML = `
        <div class="upload-progress-container">
            <p>Uploading: ${file.name}</p>
            <div class="upload-progress">
                <div class="upload-progress-bar" id="progressBar"></div>
            </div>
        </div>
    `;
    uploadZone.innerHTML = progressHTML;

    // Create XMLHttpRequest for progress tracking
    const xhr = new XMLHttpRequest();

    xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
            const percentComplete = (e.loaded / e.total) * 100;
            document.getElementById('progressBar').style.width = percentComplete + '%';
        }
    });

    xhr.addEventListener('load', () => {
        if (xhr.status === 200) {
            const response = JSON.parse(xhr.responseText);
            showSuccess('File uploaded successfully!');
            // Reload files
            setTimeout(() => window.location.reload(), 1000);
        } else {
            showError('Upload failed: ' + xhr.statusText);
            resetUploadZone();
        }
    });

    xhr.addEventListener('error', () => {
        showError('Upload failed');
        resetUploadZone();
    });

    xhr.open('POST', '/upload');
    xhr.send(formData);
}

function resetUploadZone() {
    setTimeout(() => {
        uploadZone.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
            <h2>Drop files here or click to upload</h2>
            <p>Maximum file size: 5 GB</p>
        `;
    }, 2000);
}

// Copy to clipboard function
function copyToClipboard(text, button) {
    navigator.clipboard.writeText(text).then(() => {
        const originalText = button.textContent;
        button.textContent = '✓ Copied!';
        button.style.background = '#28a745';
        setTimeout(() => {
            button.textContent = originalText;
            button.style.background = '';
        }, 2000);
    }).catch(err => {
        showError('Failed to copy: ' + err);
    });
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
        showSuccess('File deleted');
        setTimeout(() => window.location.reload(), 1000);
    })
    .catch(err => {
        showError('Failed to delete file');
    });
}

// Show success message
function showSuccess(message) {
    const toast = document.createElement('div');
    toast.className = 'toast toast-success';
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}

// Show error message
function showError(message) {
    const toast = document.createElement('div');
    toast.className = 'toast toast-error';
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}
