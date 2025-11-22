package htmlPage

const NotePageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.NoteName}}</title>
<style>
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: #ebeef1;
    height: 100vh;
    overflow: hidden;
}
.container {
    display: flex;
    height: 100vh;
    gap: 0;
    align-items: flex-start;
}
.editor-panel, .preview-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    height: 100vh;
    align-items: stretch;
}
.panel-header {
    background: #fff;
    padding: 10px 20px;
    border-bottom: 1px solid #ddd;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    color: #666;
    min-height: 40px;
    height: 40px;
    flex-shrink: 0;
}
.file-info {
    display: flex;
    gap: 15px;
    align-items: center;
}
.file-info-item {
    display: flex;
    align-items: center;
    gap: 5px;
}
.file-info-label {
    color: #999;
}
.file-info-value {
    color: #333;
    font-weight: 500;
}
.panel-header a {
    color: #0066cc;
    text-decoration: none;
}
.panel-header a:hover {
    text-decoration: underline;
}
.header-btn {
    padding: 6px 10px;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 11px;
    line-height: 1.3;
    text-align: center;
    transition: all 0.2s;
    white-space: nowrap;
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-width: 50px;
    height: 40px;
}
.header-btn:hover {
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0,0,0,0.2);
    opacity: 0.9;
}
.header-btn:active {
    transform: translateY(0);
}
#editor {
    flex: 1;
    padding: 20px;
    overflow-y: auto;
    resize: none;
    width: 100%;
    border: none;
    outline: none;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 14px;
    line-height: 1.6;
    background: #fff;
    color: #333;
}
#preview {
    flex: 1;
    padding: 20px;
    overflow-y: auto;
    background: #fff;
    border-left: 1px solid #ddd;
}
#preview h1, #preview h2, #preview h3 {
    margin-top: 1em;
    margin-bottom: 0.5em;
}
#preview code {
    background: #f5f5f5;
    padding: 2px 6px;
    border-radius: 3px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 0.9em;
}
#preview pre {
    background: #f5f5f5;
    padding: 12px;
    border-radius: 5px;
    overflow-x: auto;
    margin: 1em 0;
}
#preview pre code {
    background: none;
    padding: 0;
}
#preview blockquote {
    border-left: 4px solid #ddd;
    padding-left: 1em;
    margin: 1em 0;
    color: #666;
}
#preview img {
    max-width: 100%;
    height: auto;
    display: block;
    margin: 1em auto;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
#preview table {
    border-collapse: collapse;
    width: 100%;
    margin: 1em 0;
}
#preview table th, #preview table td {
    border: 1px solid #ddd;
    padding: 8px;
    text-align: left;
}
#preview table th {
    background: #f5f5f5;
    font-weight: bold;
}
.status {
    position: fixed;
    bottom: 20px;
    right: 20px;
    padding: 8px 16px;
    background: #4caf50;
    color: white;
    border-radius: 4px;
    font-size: 12px;
    opacity: 0;
    transition: opacity 0.3s;
}
.status.show {
    opacity: 1;
}
.status.error {
    background: #f44336;
}
@media (max-width: 768px) {
    .container {
        flex-direction: column;
    }
    .editor-panel, .preview-panel {
        height: 50vh;
    }
    #preview {
        border-left: none;
        border-top: 1px solid #ddd;
    }
}
@media (prefers-color-scheme: dark) {
    body {
        background: #333b4d;
    }
    .panel-header {
        background: #24262b;
        color: #fff;
        border-color: #495265;
    }
    .file-info-label {
        color: #aaa;
    }
    .file-info-value {
        color: #fff;
    }
    #editor, #preview {
        background: #24262b;
        color: #fff;
    }
    #preview code {
        background: #1a1a1a;
    }
    #preview pre {
        background: #1a1a1a;
    }
    #preview blockquote {
        border-color: #495265;
    }
    #preview img {
        box-shadow: 0 2px 8px rgba(0,0,0,0.3);
    }
    #preview table th, #preview table td {
        border-color: #495265;
    }
    #preview table th {
        background: #1a1a1a;
    }
}
</style>
</head>
<body>
<div class="container">
    <div class="editor-panel">
        <div class="panel-header">
            <div class="file-info">
                <div class="file-info-item">
                    <span class="file-info-label">大小:</span>
                    <span class="file-info-value" id="file-size">{{.FileSize}}</span>
                </div>
                <div class="file-info-item">
                    <span class="file-info-label">创建:</span>
                    <span class="file-info-value" id="create-time">{{.CreateTime}}</span>
                </div>
                <div class="file-info-item">
                    <span class="file-info-label">修改:</span>
                    <span class="file-info-value" id="mod-time">{{.ModTime}}</span>
                </div>
            </div>
            <div style="display: flex; gap: 8px; align-items: center;">
                <button onclick="shareNote()" class="header-btn" style="background: #28a745;">🔗<br>分享</button>
                <button onclick="copyRawUrl()" class="header-btn" style="background: #17a2b8;">📋<br>复制下载地址</button>
                <button onclick="toggleLock()" id="lockBtn" class="header-btn" style="background: #0066cc;">🔓<br>设置锁</button>
                <a href="/" class="header-btn" style="background: #666; text-decoration: none;">📝<br>新建笔记</a>
            </div>
        </div>
        <textarea id="editor" placeholder="开始输入 Markdown 内容...">{{.Content}}</textarea>
    </div>
    <div class="preview-panel">
        <div class="panel-header">
            <span></span>
            <span id="connection-status">连接中</span>
        </div>
        <div id="preview"></div>
    </div>
</div>
<div class="status" id="status"></div>
<script>
const editor = document.getElementById('editor');
const preview = document.getElementById('preview');
const status = document.getElementById('status');
const connectionStatus = document.getElementById('connection-status');
let lastContent = editor.value;
let ws = null;
let saveTimeout = null;

function updatePreview() {
    const content = editor.value;
    const { url } = addTokenToRequest('/api/markdown');
    fetch(url, {
        method: 'POST',
        headers: {'Content-Type': 'text/plain'},
        body: content
    })
    .then(res => res.text())
    .then(html => {
        preview.innerHTML = html;
    })
    .catch(err => {
        console.error('Preview error:', err);
    });
}

function saveNote() {
    const content = editor.value;
    if (content === lastContent) return;

    const { url } = addTokenToRequest(window.location.pathname);
    fetch(url, {
        method: 'POST',
        headers: {'Content-Type': 'text/plain'},
        body: content
    })
    .then(res => {
        if (res.ok) {
            lastContent = content;
            showStatus('已保存', false);
            // Update file size and modification time
            updateFileInfo();
        } else {
            showStatus('保存失败', true);
        }
    })
    .catch(err => {
        console.error('Save error:', err);
        showStatus('保存失败', true);
    });
}

function updateFileInfo() {
    const content = editor.value;
    const size = new Blob([content]).size;
    const now = new Date();
    
    // Format time as YYYY-MM-DD HH:mm:ss
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    const hours = String(now.getHours()).padStart(2, '0');
    const minutes = String(now.getMinutes()).padStart(2, '0');
    const seconds = String(now.getSeconds()).padStart(2, '0');
    const timeStr = year + '-' + month + '-' + day + ' ' + hours + ':' + minutes + ':' + seconds;
    
    // Update file size
    let sizeStr = size + ' B';
    if (size >= 1024) {
        sizeStr = (size / 1024).toFixed(2) + ' KB';
    }
    document.getElementById('file-size').textContent = sizeStr;
    
    // Update modification time
    document.getElementById('mod-time').textContent = timeStr;
}

function showStatus(message, isError) {
    status.textContent = message;
    status.className = 'status show' + (isError ? ' error' : '');
    setTimeout(() => {
        status.className = 'status';
    }, 2000);
}

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    let wsUrl = protocol + '//' + window.location.host + '/ws' + window.location.pathname;
    
    // Add token to WebSocket URL if available
    if (savedToken) {
        wsUrl += '?token=' + encodeURIComponent(savedToken);
    }
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        connectionStatus.textContent = '已连接';
        connectionStatus.style.color = '#4caf50';
    };
    
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'update' && data.content !== editor.value) {
            editor.value = data.content;
            lastContent = data.content;
            updatePreview();
        }
    };
    
    ws.onerror = () => {
        connectionStatus.textContent = '连接错误';
        connectionStatus.style.color = '#f44336';
    };
    
    ws.onclose = () => {
        connectionStatus.textContent = '已断开';
        connectionStatus.style.color = '#999';
        setTimeout(connectWebSocket, 3000);
    };
}

editor.addEventListener('input', () => {
    updatePreview();
    clearTimeout(saveTimeout);
    saveTimeout = setTimeout(saveNote, 500);
});

editor.addEventListener('paste', () => {
    setTimeout(() => {
        updatePreview();
        saveNote();
    }, 100);
});

// Floating upload window
const uploadWindow = document.createElement('div');
uploadWindow.id = 'upload-window';
uploadWindow.style.cssText = 'display: none; position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%); background: white; border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.3); z-index: 1000; padding: 30px; min-width: 400px; max-width: 500px;';
uploadWindow.innerHTML = '<div style="text-align: center; margin-bottom: 20px;"><h3 style="margin: 0 0 10px 0; font-size: 18px; color: #333;">上传文件</h3><p style="margin: 0; font-size: 14px; color: #666;">拖拽文件到此处或点击按钮选择</p></div><div id="upload-drop-zone" style="border: 2px dashed #0066cc; border-radius: 8px; padding: 40px; text-align: center; background: #f5f9ff; margin-bottom: 15px; transition: all 0.3s;"><div style="font-size: 48px; margin-bottom: 10px;">📁</div><div style="color: #0066cc; font-size: 16px; font-weight: 500;">拖拽文件到此处</div><div style="color: #999; font-size: 12px; margin-top: 5px;">或点击下方按钮选择文件</div></div><input type="file" id="file-input" style="display: none;" multiple><button id="upload-select-btn" style="width: 100%; padding: 12px; background: #0066cc; color: white; border: none; border-radius: 4px; font-size: 14px; font-weight: 500; cursor: pointer; margin-bottom: 10px;">选择文件</button><button id="upload-close-btn" style="width: 100%; padding: 10px; background: #f5f5f5; color: #666; border: none; border-radius: 4px; font-size: 14px; cursor: pointer;">取消</button>';
document.body.appendChild(uploadWindow);

const uploadDropZone = document.getElementById('upload-drop-zone');
const fileInput = document.getElementById('file-input');
const uploadSelectBtn = document.getElementById('upload-select-btn');
const uploadCloseBtn = document.getElementById('upload-close-btn');

// Upload button (floating button)
const uploadFloatingBtn = document.createElement('button');
uploadFloatingBtn.innerHTML = '📤 上传';
uploadFloatingBtn.style.cssText = 'position: fixed; bottom: 20px; left: 20px; padding: 12px 20px; background: #0066cc; color: white; border: none; border-radius: 25px; font-size: 14px; font-weight: 500; cursor: pointer; box-shadow: 0 2px 10px rgba(0,102,204,0.3); z-index: 100; transition: all 0.3s;';
uploadFloatingBtn.onmouseover = function() { uploadFloatingBtn.style.transform = 'scale(1.05)'; uploadFloatingBtn.style.boxShadow = '0 4px 15px rgba(0,102,204,0.4)'; };
uploadFloatingBtn.onmouseout = function() { uploadFloatingBtn.style.transform = 'scale(1)'; uploadFloatingBtn.style.boxShadow = '0 2px 10px rgba(0,102,204,0.3)'; };
uploadFloatingBtn.onclick = function() { uploadWindow.style.display = 'block'; };
document.body.appendChild(uploadFloatingBtn);

uploadSelectBtn.addEventListener('click', () => {
    fileInput.click();
});

uploadCloseBtn.addEventListener('click', () => {
    uploadWindow.style.display = 'none';
    uploadDropZone.style.borderColor = '#0066cc';
    uploadDropZone.style.background = '#f5f9ff';
});

fileInput.addEventListener('change', (e) => {
    handleFiles(e.target.files);
});

// Drag and drop on upload window
uploadDropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    e.stopPropagation();
    uploadDropZone.style.borderColor = '#0052a3';
    uploadDropZone.style.background = '#e6f2ff';
});

uploadDropZone.addEventListener('dragleave', (e) => {
    e.preventDefault();
    e.stopPropagation();
    uploadDropZone.style.borderColor = '#0066cc';
    uploadDropZone.style.background = '#f5f9ff';
});

uploadDropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    e.stopPropagation();
    uploadDropZone.style.borderColor = '#0066cc';
    uploadDropZone.style.background = '#f5f9ff';
    const files = e.dataTransfer.files;
    if (files.length > 0) {
        handleFiles(files);
    }
});

// Close window when clicking outside
uploadWindow.addEventListener('click', (e) => {
    if (e.target === uploadWindow) {
        uploadWindow.style.display = 'none';
    }
});

function handleFiles(files) {
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const formData = new FormData();
        formData.append('file', file);

        showStatus('上传中...', false);

        const { url: uploadUrl } = addTokenToRequest('/api/upload');
        fetch(uploadUrl, {
            method: 'POST',
            body: formData
        })
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                // Insert markdown at cursor position
                const cursorPos = editor.selectionStart;
                const textBefore = editor.value.substring(0, cursorPos);
                const textAfter = editor.value.substring(cursorPos);
                editor.value = textBefore + data.markdown + '\n' + textAfter;
                editor.selectionStart = editor.selectionEnd = cursorPos + data.markdown.length + 1;
                lastContent = editor.value;
                updatePreview();
                saveNote();
                showStatus('上传成功', false);
                uploadWindow.style.display = 'none';
            } else {
                showStatus('上传失败: ' + (data.error || '未知错误'), true);
            }
        })
        .catch(err => {
            console.error('Upload error:', err);
            showStatus('上传失败', true);
        });
    }
    fileInput.value = '';
}

// Access token management
// Get token from cookie (set by backend) or localStorage (fallback)
function getAccessToken() {
    // Try to get from cookie first
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'access_token' && value) {
            return decodeURIComponent(value);
        }
    }
    // Fallback to localStorage
    return localStorage.getItem('jot_access_token') || '';
}

let savedToken = getAccessToken();

// Get token from URL (only for first-time login)
const urlParams = new URLSearchParams(window.location.search);
const urlToken = urlParams.get('token');
if (urlToken) {
    // Save token to localStorage as backup
    localStorage.setItem('jot_access_token', urlToken);
    savedToken = urlToken;
    // Remove token from URL to keep it clean
    const newUrl = window.location.pathname;
    window.history.replaceState({}, '', newUrl);
}

// Add token to all requests (via query parameter)
function addTokenToRequest(url, options = {}) {
    if (savedToken) {
        const separator = url.includes('?') ? '&' : '?';
        url = url + separator + 'token=' + encodeURIComponent(savedToken);
    }
    return { url, options };
}

editor.focus();
updatePreview();
connectWebSocket();

// Auto-save every 2 seconds
setInterval(saveNote, 2000);

// Note lock management
let noteLockToken = '';
let isLocked = false;

// Check if note is locked on page load
const rawContent = editor.value;
if (rawContent.startsWith('<!-- LOCK:')) {
    const endIdx = rawContent.indexOf(' -->\n');
    if (endIdx !== -1) {
        noteLockToken = rawContent.substring('<!-- LOCK:'.length, endIdx);
        isLocked = true;
        document.getElementById('lockBtn').textContent = '🔒 移除锁';
        document.getElementById('lockBtn').style.background = '#e74c3c';
    }
}

// Get current lock token from editor content
function getCurrentLockToken() {
    const currentContent = editor.value;
    if (currentContent.startsWith('<!-- LOCK:')) {
        const endIdx = currentContent.indexOf(' -->\n');
        if (endIdx !== -1) {
            return currentContent.substring('<!-- LOCK:'.length, endIdx);
        }
    }
    return '';
}

// Share note function - copy URL only
function shareNote() {
    const noteName = window.location.pathname.substring(1);
    let shareUrl = window.location.origin + '/read/' + noteName;
    
    // Add lock token if note is locked (笔记锁的 token)
    // /read 路径不需要 access token，只需要笔记的 lock_token
    const lockToken = getCurrentLockToken();
    if (lockToken) {
        shareUrl += '?lock_token=' + encodeURIComponent(lockToken);
    }
    
    // Copy URL only
    copyToClipboard(shareUrl);
    showStatus('链接已复制到剪贴板', false);
}

// Copy raw download URL function - for downloading original content
// 所有读取操作都需要 /read 路径，使用 /read/xxx?raw=1 格式
function copyRawUrl() {
    const noteName = window.location.pathname.substring(1);
    let rawUrl = window.location.origin + '/read/' + noteName + '?raw=1';
    
    // Add lock token if note is locked (笔记锁的 token)
    // /read 路径不需要 access token，只需要笔记的 lock_token
    const lockToken = getCurrentLockToken();
    if (lockToken) {
        rawUrl += '&lock_token=' + encodeURIComponent(lockToken);
    }
    
    copyToClipboard(rawUrl);
    showStatus('原始下载地址已复制到剪贴板', false);
}

// Copy text to clipboard
function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).catch(err => {
            console.error('复制失败:', err);
            fallbackCopyToClipboard(text);
        });
    } else {
        fallbackCopyToClipboard(text);
    }
}

// Fallback copy method
function fallbackCopyToClipboard(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    try {
        document.execCommand('copy');
    } catch (err) {
        console.error('复制失败:', err);
    }
    document.body.removeChild(textArea);
}

function toggleLock() {
    if (isLocked) {
        // Remove lock
        if (confirm('确定要移除笔记锁吗？')) {
            const currentContent = editor.value;
            const endIdx = currentContent.indexOf(' -->\n');
            if (endIdx !== -1) {
                editor.value = currentContent.substring(endIdx + 6); // Remove ' -->\n'
                isLocked = false;
                noteLockToken = '';
                document.getElementById('lockBtn').textContent = '🔓 设置锁';
                document.getElementById('lockBtn').style.background = '#0066cc';
                saveNote();
            }
        }
    } else {
        // Set lock
        const token = prompt('请输入解锁令牌（留空取消）:');
        if (token === null) {
            return; // User cancelled
        }
        if (token.trim() === '') {
            alert('令牌不能为空');
            return;
        }
        const currentContent = editor.value;
        if (!currentContent.startsWith('<!-- LOCK:')) {
            editor.value = '<!-- LOCK:' + token.trim() + ' -->\n' + currentContent;
            isLocked = true;
            noteLockToken = token.trim();
            document.getElementById('lockBtn').textContent = '🔒 移除锁';
            document.getElementById('lockBtn').style.background = '#e74c3c';
            saveNote();
        }
    }
}
</script>
</body>
</html>`
