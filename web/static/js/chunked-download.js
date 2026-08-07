// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.
//
// Chunked download for the splash page. The download button stays an ordinary
// link to /d/<id>: this script only takes over the click when the browser can
// do better, and hands the click back to the plain link whenever it cannot.

(function () {
    'use strict';

    var DEFAULT_CHUNK_SIZE = 25 * 1024 * 1024;
    var MAX_RETRIES = 50;
    var MAX_BACKOFF_MS = 10000;
    var IDB_NAME = 'wulfvault-downloads';
    var IDB_VERSION = 1;
    var PROGRESS_PREFIX = 'wulfvault.download.';

    // ------------------------------------------------------------------
    // Translations
    //
    // The splash page writes window.WV_DOWNLOAD_I18N before this file loads.
    // Every string below carries its English original as a fallback, so the
    // download still works word for word when the object is missing - on a
    // cached page, for instance, or if the server sends nothing at all.
    // ------------------------------------------------------------------

    function t(name, fallback, values) {
        var catalog = window.WV_DOWNLOAD_I18N;
        var text = (catalog && typeof catalog[name] === 'string' && catalog[name]) || fallback;
        if (values) {
            Object.keys(values).forEach(function (placeholder) {
                text = text.split('{{' + placeholder + '}}').join(values[placeholder]);
            });
        }
        return text;
    }

    // ------------------------------------------------------------------
    // Streaming SHA-256
    //
    // Web Crypto has no incremental digest API - crypto.subtle.digest() wants
    // the whole payload in one buffer, which is not an option for a 15 GB file.
    // Each chunk is therefore folded into this streaming implementation as it
    // arrives, and Web Crypto is used to prove the implementation right before
    // any of its output is trusted (see selfTest below).
    // ------------------------------------------------------------------

    var K = new Uint32Array([
        0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
        0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
        0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
        0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
        0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
        0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
        0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
        0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
    ]);

    function rotr(x, n) {
        return (x >>> n) | (x << (32 - n));
    }

    function Sha256() {
        this.h = new Uint32Array([
            0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
            0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19
        ]);
        this.w = new Uint32Array(64);
        this.buffer = new Uint8Array(64);
        this.buffered = 0;
        this.byteCount = 0;
    }

    Sha256.prototype._block = function (bytes, offset) {
        var w = this.w;
        var i;
        for (i = 0; i < 16; i++) {
            var o = offset + i * 4;
            w[i] = (bytes[o] << 24) | (bytes[o + 1] << 16) | (bytes[o + 2] << 8) | bytes[o + 3];
        }
        for (i = 16; i < 64; i++) {
            var x = w[i - 15];
            var y = w[i - 2];
            var s0 = rotr(x, 7) ^ rotr(x, 18) ^ (x >>> 3);
            var s1 = rotr(y, 17) ^ rotr(y, 19) ^ (y >>> 10);
            w[i] = (w[i - 16] + s0 + w[i - 7] + s1) | 0;
        }

        var a = this.h[0], b = this.h[1], c = this.h[2], d = this.h[3];
        var e = this.h[4], f = this.h[5], g = this.h[6], hh = this.h[7];

        for (i = 0; i < 64; i++) {
            var S1 = rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25);
            var ch = (e & f) ^ (~e & g);
            var t1 = (hh + S1 + ch + K[i] + w[i]) | 0;
            var S0 = rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22);
            var maj = (a & b) ^ (a & c) ^ (b & c);
            var t2 = (S0 + maj) | 0;

            hh = g; g = f; f = e;
            e = (d + t1) | 0;
            d = c; c = b; b = a;
            a = (t1 + t2) | 0;
        }

        this.h[0] = (this.h[0] + a) | 0;
        this.h[1] = (this.h[1] + b) | 0;
        this.h[2] = (this.h[2] + c) | 0;
        this.h[3] = (this.h[3] + d) | 0;
        this.h[4] = (this.h[4] + e) | 0;
        this.h[5] = (this.h[5] + f) | 0;
        this.h[6] = (this.h[6] + g) | 0;
        this.h[7] = (this.h[7] + hh) | 0;
    };

    Sha256.prototype.update = function (bytes) {
        var i = 0;
        this.byteCount += bytes.length;

        if (this.buffered > 0) {
            var take = Math.min(64 - this.buffered, bytes.length);
            this.buffer.set(bytes.subarray(0, take), this.buffered);
            this.buffered += take;
            i = take;
            if (this.buffered === 64) {
                this._block(this.buffer, 0);
                this.buffered = 0;
            }
        }

        for (; i + 64 <= bytes.length; i += 64) {
            this._block(bytes, i);
        }

        if (i < bytes.length) {
            this.buffer.set(bytes.subarray(i), 0);
            this.buffered = bytes.length - i;
        }
    };

    Sha256.prototype.hex = function () {
        var clone = new Sha256();
        clone.h = new Uint32Array(this.h);
        clone.buffer = new Uint8Array(this.buffer);
        clone.buffered = this.buffered;
        clone.byteCount = this.byteCount;

        var bitLength = clone.byteCount * 8;
        var padLength = clone.buffered < 56 ? 56 - clone.buffered : 120 - clone.buffered;
        var tail = new Uint8Array(padLength + 8);
        tail[0] = 0x80;

        var high = Math.floor(bitLength / 4294967296);
        var low = bitLength % 4294967296;
        var base = padLength;
        tail[base] = (high >>> 24) & 0xff;
        tail[base + 1] = (high >>> 16) & 0xff;
        tail[base + 2] = (high >>> 8) & 0xff;
        tail[base + 3] = high & 0xff;
        tail[base + 4] = (low >>> 24) & 0xff;
        tail[base + 5] = (low >>> 16) & 0xff;
        tail[base + 6] = (low >>> 8) & 0xff;
        tail[base + 7] = low & 0xff;

        clone.byteCount = 0;
        clone.update(tail);

        var out = '';
        for (var i = 0; i < 8; i++) {
            out += ('00000000' + (clone.h[i] >>> 0).toString(16)).slice(-8);
        }
        return out;
    };

    var selfTestResult = null;

    // selfTest checks the streaming implementation against Web Crypto on a
    // fixed sample. A mismatch means verification is reported as unavailable
    // instead of producing a false alarm on a perfectly good download.
    function selfTest() {
        if (selfTestResult) {
            return selfTestResult;
        }

        selfTestResult = (async function () {
            var sample = new Uint8Array(1024);
            for (var i = 0; i < sample.length; i++) {
                sample[i] = (i * 31 + 7) & 0xff;
            }

            var streaming = new Sha256();
            streaming.update(sample.subarray(0, 100));
            streaming.update(sample.subarray(100));
            var streamed = streaming.hex();

            if (!window.crypto || !window.crypto.subtle || !window.crypto.subtle.digest) {
                return false;
            }

            var buffer = await window.crypto.subtle.digest('SHA-256', sample);
            var bytes = new Uint8Array(buffer);
            var reference = '';
            for (var j = 0; j < bytes.length; j++) {
                reference += ('00' + bytes[j].toString(16)).slice(-2);
            }

            return reference === streamed;
        })().catch(function () {
            return false;
        });

        return selfTestResult;
    }

    // ------------------------------------------------------------------
    // Persistence: IndexedDB for the payload, localStorage for the progress
    // ------------------------------------------------------------------

    function openDatabase() {
        return new Promise(function (resolve, reject) {
            if (!window.indexedDB) {
                reject(new Error('IndexedDB unavailable'));
                return;
            }
            var request = window.indexedDB.open(IDB_NAME, IDB_VERSION);
            request.onupgradeneeded = function () {
                var db = request.result;
                if (!db.objectStoreNames.contains('chunks')) {
                    db.createObjectStore('chunks');
                }
                if (!db.objectStoreNames.contains('handles')) {
                    db.createObjectStore('handles');
                }
            };
            request.onsuccess = function () { resolve(request.result); };
            request.onerror = function () { reject(request.error); };
        });
    }

    function idbRequest(store, action) {
        return new Promise(function (resolve, reject) {
            var request = action(store);
            request.onsuccess = function () { resolve(request.result); };
            request.onerror = function () { reject(request.error); };
        });
    }

    function withStore(name, mode, action) {
        return openDatabase().then(function (db) {
            var tx = db.transaction(name, mode);
            var store = tx.objectStore(name);
            return idbRequest(store, action).finally(function () { db.close(); });
        });
    }

    function putChunk(fileId, index, blob) {
        return withStore('chunks', 'readwrite', function (store) {
            return store.put(blob, [fileId, index]);
        });
    }

    function getChunk(fileId, index) {
        return withStore('chunks', 'readonly', function (store) {
            return store.get([fileId, index]);
        });
    }

    function clearChunks(fileId) {
        return withStore('chunks', 'readwrite', function (store) {
            return store.delete(IDBKeyRange.bound([fileId, -Infinity], [fileId, Infinity]));
        }).catch(function () { /* nothing stored yet */ });
    }

    function putHandle(fileId, handle) {
        return withStore('handles', 'readwrite', function (store) {
            return store.put(handle, fileId);
        }).catch(function () { /* handles are a nicety, not a requirement */ });
    }

    function getHandle(fileId) {
        return withStore('handles', 'readonly', function (store) {
            return store.get(fileId);
        }).catch(function () { return null; });
    }

    function clearHandle(fileId) {
        return withStore('handles', 'readwrite', function (store) {
            return store.delete(fileId);
        }).catch(function () { /* already gone */ });
    }

    function progressKey(fileId) {
        return PROGRESS_PREFIX + fileId;
    }

    function readProgress(fileId) {
        try {
            var raw = window.localStorage.getItem(progressKey(fileId));
            return raw ? JSON.parse(raw) : null;
        } catch (err) {
            return null;
        }
    }

    function writeProgress(record) {
        try {
            window.localStorage.setItem(progressKey(record.fileId), JSON.stringify(record));
        } catch (err) {
            // A full or disabled storage only costs us the ability to resume.
        }
    }

    function clearProgress(fileId) {
        try {
            window.localStorage.removeItem(progressKey(fileId));
        } catch (err) {
            // Ignore - nothing depends on the record being gone.
        }
    }

    // ------------------------------------------------------------------
    // Sinks
    //
    // A 15 GB download must never be held in JavaScript memory. The File
    // System Access API writes straight to the file the user picked, which is
    // both the cheapest and the only option that survives a closed tab intact.
    // Browsers without it fall back to IndexedDB, where the chunks are stored
    // as Blobs (the browser keeps those on disk, not on the JS heap) and are
    // assembled into one Blob only at the very end.
    // ------------------------------------------------------------------

    function FileSystemSink(handle, writable, written, sourceFile) {
        this.mode = 'filesystem';
        this.handle = handle;
        this.writable = writable;
        this.written = written || 0;
        // Captured before the writable was opened: while a writable stream is
        // open the browser works on a swap copy, so this is the only reliable
        // view of the bytes an earlier session already wrote.
        this.sourceFile = sourceFile || null;
    }

    FileSystemSink.prototype.write = function (chunk) {
        var self = this;
        return this.writable.write(chunk).then(function () {
            self.written += chunk.byteLength;
        });
    };

    FileSystemSink.prototype.readBack = async function (onBytes) {
        var file = this.sourceFile || await this.handle.getFile();
        var offset = 0;
        while (offset < this.written) {
            var end = Math.min(offset + DEFAULT_CHUNK_SIZE, this.written);
            var slice = await file.slice(offset, end).arrayBuffer();
            onBytes(new Uint8Array(slice));
            offset = end;
        }
    };

    FileSystemSink.prototype.finish = async function () {
        await this.writable.close();
        return null;
    };

    FileSystemSink.prototype.abort = async function () {
        try {
            await this.writable.close();
        } catch (err) {
            // The stream may already be closed.
        }
    };

    function IndexedDbSink(fileId, contentType, chunkSize, storedChunks, written) {
        this.mode = 'indexeddb';
        this.fileId = fileId;
        this.contentType = contentType;
        this.chunkSize = chunkSize;
        this.index = storedChunks || 0;
        this.written = written || 0;
    }

    IndexedDbSink.prototype.write = async function (chunk) {
        await putChunk(this.fileId, this.index, new Blob([chunk]));
        this.index += 1;
        this.written += chunk.byteLength;
    };

    IndexedDbSink.prototype.readBack = async function (onBytes) {
        for (var i = 0; i < this.index; i++) {
            var blob = await getChunk(this.fileId, i);
            if (!blob) {
                throw new Error('Stored chunk ' + i + ' is missing');
            }
            onBytes(new Uint8Array(await blob.arrayBuffer()));
        }
    };

    IndexedDbSink.prototype.finish = async function () {
        var parts = [];
        for (var i = 0; i < this.index; i++) {
            var blob = await getChunk(this.fileId, i);
            if (!blob) {
                throw new Error('Stored chunk ' + i + ' is missing');
            }
            parts.push(blob);
        }
        return new Blob(parts, { type: this.contentType || 'application/octet-stream' });
    };

    IndexedDbSink.prototype.abort = function () {
        return Promise.resolve();
    };

    // ------------------------------------------------------------------
    // Progress overlay - same look as the upload overlay on the dashboard
    // ------------------------------------------------------------------

    var overlay = null;

    function element(tag, css, text) {
        var el = document.createElement(tag);
        if (css) {
            el.style.cssText = css;
        }
        if (text) {
            el.textContent = text;
        }
        return el;
    }

    function showOverlay(name, size) {
        var style = document.createElement('style');
        style.textContent =
            '@keyframes wvFadeIn { from { opacity: 0; } to { opacity: 1; } }' +
            '@keyframes wvPulse { 0%, 100% { transform: scale(1); } 50% { transform: scale(1.05); } }';
        document.head.appendChild(style);

        overlay = element('div', [
            'position: fixed', 'top: 0', 'left: 0', 'width: 100%', 'height: 100%',
            'background: rgba(0, 0, 0, 0.92)', 'z-index: 10000', 'display: flex',
            'align-items: center', 'justify-content: center', 'overflow-y: auto',
            'animation: wvFadeIn 0.3s ease'
        ].join(';'));
        overlay.id = 'downloadProgressOverlay';

        var container = element('div', 'text-align: center; padding: 40px 20px; max-width: 700px; width: 92%;');

        var status = element('div', [
            'font-size: clamp(32px, 9vw, 72px)', 'font-weight: bold', 'color: #2563eb',
            'margin-bottom: 24px', 'text-shadow: 0 0 20px rgba(37, 99, 235, 0.5)',
            'animation: wvPulse 2s ease-in-out infinite', 'line-height: 1.1'
        ].join(';'), t('downloading', 'DOWNLOADING - {{percent}}%', { percent: 0 }));
        status.id = 'downloadStatusText';

        var fileName = element('div', 'font-size: clamp(16px, 4vw, 24px); color: #e5e7eb; margin-bottom: 20px; font-weight: 500; word-break: break-word;', name);

        var sizeInfo = element('div', 'font-size: clamp(13px, 3.5vw, 18px); color: #9ca3af; margin-bottom: 30px;', '0 B / ' + formatFileSize(size));
        sizeInfo.id = 'downloadSizeInfo';

        var barOuter = element('div', [
            'width: 100%', 'height: 32px', 'background: rgba(255, 255, 255, 0.1)',
            'border-radius: 16px', 'overflow: hidden', 'margin-bottom: 20px',
            'box-shadow: 0 0 30px rgba(0, 0, 0, 0.5)'
        ].join(';'));

        var barFill = element('div', [
            'height: 100%', 'width: 0%', 'background: linear-gradient(90deg, #3b82f6, #2563eb)',
            'transition: width 0.3s ease, background 0.5s ease', 'border-radius: 16px',
            'box-shadow: 0 0 20px rgba(37, 99, 235, 0.8)'
        ].join(';'));
        barFill.id = 'downloadProgressBarFill';

        var speed = element('div', 'font-size: clamp(13px, 3.2vw, 16px); color: #9ca3af; margin-top: 16px;', t('calculatingSpeed', 'Calculating speed...'));
        speed.id = 'downloadSpeedInfo';

        var notice = element('div', 'font-size: 14px; color: #fbbf24; margin-top: 14px; font-weight: 600; display: none;');
        notice.id = 'downloadNotice';

        var actions = element('div', 'margin-top: 30px;');
        actions.id = 'downloadActions';

        barOuter.appendChild(barFill);
        container.appendChild(status);
        container.appendChild(fileName);
        container.appendChild(sizeInfo);
        container.appendChild(barOuter);
        container.appendChild(speed);
        container.appendChild(notice);
        container.appendChild(actions);
        overlay.appendChild(container);
        document.body.appendChild(overlay);
    }

    function setStatus(text, color) {
        var el = document.getElementById('downloadStatusText');
        if (!el) {
            return;
        }
        el.textContent = text;
        if (color) {
            el.style.color = color;
            el.style.textShadow = '0 0 20px ' + color;
        }
    }

    function setBar(percent, color) {
        var el = document.getElementById('downloadProgressBarFill');
        if (!el) {
            return;
        }
        el.style.width = percent + '%';
        if (color) {
            el.style.background = color;
        }
    }

    function setSpeedText(text, color) {
        var el = document.getElementById('downloadSpeedInfo');
        if (!el) {
            return;
        }
        el.textContent = text;
        if (color) {
            el.style.color = color;
        }
    }

    function setNotice(text) {
        var el = document.getElementById('downloadNotice');
        if (!el) {
            return;
        }
        if (!text) {
            el.style.display = 'none';
            return;
        }
        el.style.display = 'block';
        el.textContent = text;
    }

    function addAction(label, background, onClick) {
        var actions = document.getElementById('downloadActions');
        if (!actions) {
            return;
        }
        var button = element('button', [
            'margin: 10px', 'padding: 16px 34px', 'font-size: 17px', 'font-weight: bold',
            'color: white', 'background: ' + background, 'border: none', 'border-radius: 12px',
            'cursor: pointer', 'letter-spacing: 0.5px'
        ].join(';'), label);
        button.onclick = onClick;
        actions.appendChild(button);
    }

    function closeOverlay() {
        if (overlay && overlay.parentNode) {
            overlay.parentNode.removeChild(overlay);
        }
        overlay = null;
    }

    function formatFileSize(bytes) {
        if (!bytes || bytes < 1024) {
            return Math.max(0, Math.round(bytes || 0)) + ' B';
        }
        var units = ['KB', 'MB', 'GB', 'TB'];
        var value = bytes / 1024;
        var unit = 0;
        while (value >= 1024 && unit < units.length - 1) {
            value /= 1024;
            unit++;
        }
        return value.toFixed(1) + ' ' + units[unit];
    }

    function formatTime(seconds) {
        if (!isFinite(seconds) || seconds < 0) {
            return '--';
        }
        if (seconds < 60) {
            return Math.round(seconds) + 's';
        }
        if (seconds < 3600) {
            return Math.floor(seconds / 60) + 'm ' + Math.round(seconds % 60) + 's';
        }
        return Math.floor(seconds / 3600) + 'h ' + Math.floor((seconds % 3600) / 60) + 'm';
    }

    // ------------------------------------------------------------------
    // Download
    // ------------------------------------------------------------------

    function sleep(ms) {
        return new Promise(function (resolve) { setTimeout(resolve, ms); });
    }

    async function fetchChunk(apiBase, offset, size) {
        var url = apiBase + '/chunk?offset=' + offset + '&size=' + size;
        var response = await fetch(url, { credentials: 'same-origin', cache: 'no-store' });
        if (!response.ok) {
            var error = new Error('Chunk request failed with status ' + response.status);
            error.status = response.status;
            throw error;
        }
        return new Uint8Array(await response.arrayBuffer());
    }

    async function fetchChunkWithRetry(apiBase, offset, size, onRetry) {
        var attempt = 0;
        for (;;) {
            try {
                return await fetchChunk(apiBase, offset, size);
            } catch (err) {
                // An expired link, a spent download counter or a lost session is
                // not going to fix itself - only network errors are worth a retry.
                if (err.status && err.status !== 408 && err.status !== 429 && err.status < 500) {
                    throw err;
                }
                attempt++;
                if (attempt >= MAX_RETRIES) {
                    throw new Error('Chunk at offset ' + offset + ' failed after ' + MAX_RETRIES + ' attempts');
                }
                onRetry(attempt);
                await sleep(Math.min(1000 * Math.pow(2, attempt - 1), MAX_BACKOFF_MS));
            }
        }
    }

    async function createSink(info, resumeFrom) {
        var supportsFileSystem = typeof window.showSaveFilePicker === 'function' && window.isSecureContext;

        if (supportsFileSystem && resumeFrom && resumeFrom.mode === 'filesystem') {
            var stored = await getHandle(info.file_id);
            if (stored && typeof stored.createWritable === 'function') {
                var permission = 'granted';
                if (typeof stored.requestPermission === 'function') {
                    permission = await stored.requestPermission({ mode: 'readwrite' });
                }
                if (permission === 'granted') {
                    var existing = await stored.getFile();
                    if (existing.size >= resumeFrom.bytesDone && resumeFrom.bytesDone > 0) {
                        var writable = await stored.createWritable({ keepExistingData: true });
                        await writable.truncate(resumeFrom.bytesDone);
                        await writable.seek(resumeFrom.bytesDone);
                        return new FileSystemSink(stored, writable, resumeFrom.bytesDone, existing);
                    }
                }
            }
        }

        if (supportsFileSystem) {
            var handle = await window.showSaveFilePicker({ suggestedName: info.name });
            await putHandle(info.file_id, handle);
            return new FileSystemSink(handle, await handle.createWritable(), 0);
        }

        if (resumeFrom && resumeFrom.mode === 'indexeddb' && resumeFrom.chunkCount > 0) {
            return new IndexedDbSink(info.file_id, info.content_type, resumeFrom.chunkSize, resumeFrom.chunkCount, resumeFrom.bytesDone);
        }

        await clearChunks(info.file_id);
        return new IndexedDbSink(info.file_id, info.content_type, DEFAULT_CHUNK_SIZE, 0, 0);
    }

    function deliverBlob(blob, name) {
        var url = URL.createObjectURL(blob);
        var link = document.createElement('a');
        link.href = url;
        link.download = name;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        setTimeout(function () { URL.revokeObjectURL(url); }, 60000);
    }

    async function runDownload(config, info, resumeFrom) {
        var chunkSize = info.chunk_size > 0 ? info.chunk_size : DEFAULT_CHUNK_SIZE;
        var total = info.size_bytes;

        var hashUsable = await selfTest();
        var hasher = new Sha256();
        var sink;

        showOverlay(info.name, total);

        try {
            sink = await createSink(info, resumeFrom);
        } catch (err) {
            closeOverlay();
            if (err && err.name === 'AbortError') {
                return;
            }
            throw err;
        }

        var beforeUnload = function (event) {
            event.preventDefault();
            event.returnValue = '';
        };
        window.addEventListener('beforeunload', beforeUnload);

        try {
            if (sink.written > 0) {
                setStatus(t('resuming', 'RESUMING...'));
                setSpeedText(t('checkingExisting', 'Checking {{size}} already downloaded...', { size: formatFileSize(sink.written) }));
                if (hashUsable) {
                    await sink.readBack(function (bytes) { hasher.update(bytes); });
                }
            }

            var offset = sink.written;
            var startedAt = Date.now();
            var startedAtBytes = offset;
            var averageSpeed = 0;

            while (offset < total) {
                var size = Math.min(chunkSize, total - offset);
                var chunk = await fetchChunkWithRetry(config.apiBase, offset, size, function (attempt) {
                    setNotice(t('retrying', 'Connection interrupted - retry attempt {{attempt}} of {{max}}...', { attempt: attempt, max: MAX_RETRIES }));
                });
                setNotice('');

                if (chunk.byteLength === 0) {
                    throw new Error('Server returned an empty chunk at offset ' + offset);
                }

                if (hashUsable) {
                    hasher.update(chunk);
                }
                await sink.write(chunk);

                offset += chunk.byteLength;

                var elapsed = (Date.now() - startedAt) / 1000;
                if (elapsed > 0) {
                    var instant = (offset - startedAtBytes) / elapsed;
                    averageSpeed = averageSpeed === 0 ? instant : (averageSpeed * 0.7 + instant * 0.3);
                }

                var percent = total > 0 ? Math.floor((offset / total) * 100) : 100;
                setStatus(t('downloading', 'DOWNLOADING - {{percent}}%', { percent: percent }));
                setBar(percent);
                var sizeInfo = document.getElementById('downloadSizeInfo');
                if (sizeInfo) {
                    sizeInfo.textContent = formatFileSize(offset) + ' / ' + formatFileSize(total);
                }
                if (averageSpeed > 0) {
                    setSpeedText(t('speedEta', 'Speed: {{speed}}/s | ETA: {{eta}}', {
                        speed: formatFileSize(averageSpeed),
                        eta: formatTime((total - offset) / averageSpeed)
                    }));
                }

                writeProgress({
                    fileId: info.file_id,
                    name: info.name,
                    size: total,
                    bytesDone: offset,
                    chunkSize: chunkSize,
                    chunkCount: sink.mode === 'indexeddb' ? sink.index : 0,
                    mode: sink.mode,
                    updatedAt: Date.now()
                });
            }

            setStatus(t('verifying', 'VERIFYING...'), '#3b82f6');
            setBar(100);
            setSpeedText(t('assembling', 'Assembling and checking the file...'));

            var blob = await sink.finish();
            sink = null;

            if (blob) {
                deliverBlob(blob, info.name);
                await clearChunks(info.file_id);
            }
            await clearHandle(info.file_id);
            clearProgress(info.file_id);

            await reportVerification(config, hasher, hashUsable, info);
        } catch (err) {
            if (sink) {
                await sink.abort();
            }
            setStatus(t('failed', 'DOWNLOAD FAILED'), '#ef4444');
            setBar(100, 'linear-gradient(90deg, #ef4444, #dc2626)');
            setSpeedText(err && err.message ? err.message : t('unknownError', 'Unknown error'), '#fca5a5');
            setNotice('');
            addAction(t('useDirectLink', 'Use the direct link instead'), '#2563eb', function () {
                window.location.href = config.directUrl;
            });
            addAction(t('close', 'Close'), '#ef4444', closeOverlay);
        } finally {
            window.removeEventListener('beforeunload', beforeUnload);
        }
    }

    async function reportVerification(config, hasher, hashUsable, info) {
        var localDigest = hashUsable ? hasher.hex() : '';
        var serverDigest = '';

        try {
            var response = await fetch(config.apiBase + '/verify', { credentials: 'same-origin', cache: 'no-store' });
            if (response.ok) {
                serverDigest = (await response.json()).sha256 || '';
            }
        } catch (err) {
            serverDigest = '';
        }

        setBar(100, 'linear-gradient(90deg, #10b981, #059669)');

        if (!hashUsable || !serverDigest) {
            setStatus(t('complete', 'DOWNLOAD COMPLETE'), '#10b981');
            setSpeedText(t('noChecksum', 'Downloaded {{size}}. Checksum verification was not available in this browser.', {
                size: formatFileSize(info.size_bytes)
            }), '#fbbf24');
        } else if (localDigest === serverDigest) {
            setStatus(t('complete', 'DOWNLOAD COMPLETE'), '#10b981');
            setSpeedText(t('checksumOk', 'SHA-256 verified: {{digest}}', { digest: localDigest }), '#10b981');
        } else {
            setStatus(t('checksumFailed', 'VERIFICATION FAILED'), '#ef4444');
            setBar(100, 'linear-gradient(90deg, #ef4444, #dc2626)');
            setSpeedText(
                t('checksumMismatch', 'The downloaded file does not match the server checksum and may be corrupt. Do not use it.') + '\n' +
                t('serverDigest', 'Server') + ': ' + serverDigest + '\n' +
                t('localDigest', 'Downloaded') + ': ' + localDigest,
                '#fca5a5'
            );
            var speedEl = document.getElementById('downloadSpeedInfo');
            if (speedEl) {
                speedEl.style.whiteSpace = 'pre-wrap';
                speedEl.style.wordBreak = 'break-all';
            }
            addAction(t('tryAgain', 'Try again'), '#2563eb', function () { window.location.reload(); });
        }

        addAction(t('close', 'Close'), '#4b5563', closeOverlay);
    }

    // ------------------------------------------------------------------
    // Wiring
    // ------------------------------------------------------------------

    function findResumePoint(fileId, total) {
        var record = readProgress(fileId);
        if (!record || record.size !== total || record.bytesDone <= 0 || record.bytesDone >= total) {
            return null;
        }
        // A week-old record is more likely to be stale than useful.
        if (Date.now() - record.updatedAt > 7 * 24 * 3600 * 1000) {
            clearProgress(fileId);
            return null;
        }
        return record;
    }

    function initialise() {
        var button = document.getElementById('wvDownloadButton');
        if (!button || !window.fetch || !window.Promise) {
            return;
        }

        var config = {
            fileId: button.getAttribute('data-file-id'),
            apiBase: button.getAttribute('data-api-base'),
            directUrl: button.getAttribute('data-direct-url')
        };
        if (!config.fileId || !config.apiBase || !config.directUrl) {
            return;
        }

        // Probe first: a file behind a password or a login answers with an error
        // here, and the button is then left alone as a plain link to /d/<id>,
        // which knows how to ask for those credentials.
        fetch(config.apiBase + '/info', { credentials: 'same-origin', cache: 'no-store' })
            .then(function (response) {
                if (!response.ok) {
                    return null;
                }
                return response.json();
            })
            .then(function (info) {
                if (!info || !info.file_id || typeof info.size_bytes !== 'number') {
                    return;
                }

                var resumeFrom = findResumePoint(config.fileId, info.size_bytes);
                if (resumeFrom) {
                    var percent = Math.floor((resumeFrom.bytesDone / info.size_bytes) * 100);
                    var label = button.querySelector('[data-download-label]');
                    if (label) {
                        label.textContent = t('resumeButton', 'Resume download ({{percent}}%)', { percent: percent });
                    }
                }

                button.addEventListener('click', function (event) {
                    if (event.metaKey || event.ctrlKey || event.shiftKey || event.button !== 0) {
                        return;
                    }
                    event.preventDefault();
                    runDownload(config, info, findResumePoint(config.fileId, info.size_bytes))
                        .catch(function (err) {
                            console.error('Chunked download failed, falling back to the direct link:', err);
                            closeOverlay();
                            window.location.href = config.directUrl;
                        });
                });
            })
            .catch(function () {
                // Leave the plain link in place.
            });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initialise);
    } else {
        initialise();
    }
})();
