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
.preview-panel {
    display: flex;
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
.preview-header-left {
    flex: 0 0 auto;
}
.preview-header-right {
    flex: 0 0 auto;
    margin-left: auto;
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
    white-space: nowrap;
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
    padding: 6px 12px;
    margin: 4px 0;
    color: white;
    border: none;
    border-radius: 12px;
    cursor: pointer;
    font-size: 12px;
    line-height: 1.4;
    text-align: center;
    transition: all 0.2s;
    white-space: nowrap;
    display: inline-flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 4px;
    min-width: 60px;
    height: auto;
    font-weight: 500;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
.header-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0,0,0,0.15);
    opacity: 0.95;
}
.header-btn:active {
    transform: translateY(0);
    box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}
.header-btn br {
    line-height: 1.2;
    margin: 2px 0;
}
#connection-status {
    padding: 4px 12px;
    margin: 4px 0;
    background: #4caf50;
    color: white;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;
    white-space: nowrap;
    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}
#connection-status.error {
    background: #f44336;
    color: white;
}
#connection-status.disconnected {
    background: #999;
    color: white;
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
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
    padding: 10px 20px;
    background: #4caf50;
    color: white;
    border-radius: 8px;
    font-size: 14px;
    opacity: 0;
    transition: opacity 0.3s;
    z-index: 1000;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}
.status.show {
    opacity: 1;
    transform: translateX(-50%);
}
.status.error {
    background: #f44336;
}
@media (max-width: 768px) {
    body {
        overflow: auto;
        height: auto;
        padding: 0;
    }
    .container {
        flex-direction: column;
        height: auto;
        min-height: 100vh;
        width: 100%;
        max-width: 100%;
    }
    /* 移动端默认只显示编辑面板 */
    .preview-panel {
        display: none !important;
        width: 0 !important;
        flex: 0 !important;
    }
    .preview-panel.show {
        display: flex !important;
        width: 100% !important;
        height: auto;
        min-height: 50vh;
        flex: 1;
    }
    .editor-panel {
        width: 100% !important;
        max-width: 100% !important;
        height: auto;
        min-height: 100vh;
        flex: 1 1 100% !important;
    }
    .editor-panel.hide {
        display: none !important;
        width: 0 !important;
        flex: 0 !important;
    }
    .panel-header {
        padding: 12px 15px;
        font-size: 12px;
        min-height: auto;
        height: auto;
        flex-wrap: wrap;
        gap: 10px;
    }
    .file-info {
        gap: 10px;
        flex-wrap: wrap;
        width: 100%;
    }
    .file-info-item {
        font-size: 12px;
    }
    .header-btn {
        padding: 10px 14px;
        font-size: 13px;
        min-width: 75px;
        height: 44px;
    }
    #editor, #preview {
        padding: 15px;
        font-size: 16px; /* 防止 iOS 自动缩放 */
        min-height: calc(100vh - 100px);
        width: 100% !important;
        max-width: 100% !important;
        box-sizing: border-box;
    }
    #preview {
        border-left: none;
        border-top: 1px solid #ddd;
    }
    .status {
        top: 15px !important;
        left: 50% !important;
        transform: translateX(-50%) !important;
        right: auto !important;
        bottom: auto !important;
        font-size: 14px;
        padding: 10px 20px;
    }
    /* 浮动按钮在移动端优化 */
    .floating-btn {
        padding: 14px 20px !important;
        font-size: 15px !important;
        min-width: 85px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.2) !important;
    }
    /* 上传窗口在移动端全屏 */
    #upload-window {
        min-width: 90% !important;
        max-width: 90% !important;
        padding: 25px !important;
        border-radius: 12px !important;
    }
    #connection-status {
        font-size: 12px;
        white-space: nowrap;
        padding: 4px 12px;
        margin: 4px 0;
        background: #4caf50;
        color: white;
        border-radius: 12px;
        font-weight: 500;
        box-shadow: 0 1px 3px rgba(0,0,0,0.1);
    }
    #connection-status.error {
        background: #f44336;
        color: white;
    }
    #connection-status.disconnected {
        background: #999;
        color: white;
    }
}

@media (max-width: 480px) {
    body {
        padding: 0;
        margin: 0;
    }
    .container {
        width: 100% !important;
        max-width: 100% !important;
    }
    .editor-panel {
        width: 100% !important;
        max-width: 100% !important;
    }
    .panel-header {
        padding: 10px 12px;
        font-size: 11px;
        flex-direction: column;
        align-items: flex-start;
        gap: 8px;
        width: 100%;
    }
    .file-info {
        gap: 8px;
        width: 100%;
        order: 2;
    }
    .file-info-item {
        font-size: 11px;
    }
    .file-info-label {
        display: none; /* 在小屏幕上隐藏标签，只显示值 */
    }
    .header-btn {
        padding: 10px 14px;
        font-size: 13px;
        min-width: 70px;
        height: 42px;
    }
    .header-btn br {
        display: none; /* 小屏幕上按钮文字单行显示 */
    }
    #editor, #preview {
        padding: 15px;
        font-size: 16px;
    }
    .status {
        top: 10px !important;
        left: 50% !important;
        transform: translateX(-50%) !important;
        right: auto !important;
        bottom: auto !important;
        font-size: 13px;
        padding: 8px 16px;
    }
    /* 浮动按钮在小屏幕优化 */
    .floating-btn {
        padding: 14px 18px !important;
        font-size: 15px !important;
        min-width: 75px !important;
    }
    /* 浮动按钮容器在小屏幕调整 */
    .floating-actions {
        bottom: 15px !important;
        left: 15px !important;
        gap: 8px !important;
        flex-direction: column;
    }
    #connection-status {
        font-size: 11px;
        padding: 3px 10px;
        margin: 3px 0;
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
            </div>
        </div>
        <textarea id="editor" placeholder="开始输入 Markdown 内容...">{{.Content}}</textarea>
    </div>
    <div class="preview-panel" id="preview-panel">
        <div class="panel-header">
            <div class="preview-header-left" style="display: flex; gap: 8px; align-items: center;">
            </div>
            <div class="preview-header-right" style="display: flex; gap: 8px; align-items: center;">
                <button onclick="shareNote()" class="header-btn" style="background: #28a745;">🔗 复制地址</button>
                <button onclick="copyRawUrl()" class="header-btn" style="background: #17a2b8;">📋 下载地址</button>
                <span id="connection-status">连接中</span>
            </div>
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
        connectionStatus.className = '';
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
        connectionStatus.className = 'error';
    };
    
    ws.onclose = () => {
        connectionStatus.textContent = '已断开';
        connectionStatus.className = 'disconnected';
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

// Floating action buttons container
const floatingActions = document.createElement('div');
floatingActions.className = 'floating-actions';
floatingActions.style.cssText = 'position: fixed; bottom: 20px; left: 20px; display: flex; gap: 10px; align-items: center; z-index: 100; flex-wrap: wrap;';

// Upload button
const uploadFloatingBtn = document.createElement('button');
uploadFloatingBtn.innerHTML = '📤 上传';
uploadFloatingBtn.className = 'floating-btn';
uploadFloatingBtn.style.cssText = 'padding: 12px 20px; background: #0066cc; color: white; border: none; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; box-shadow: 0 2px 8px rgba(0,102,204,0.3); transition: all 0.2s;';
uploadFloatingBtn.onmouseover = function() { uploadFloatingBtn.style.transform = 'translateY(-2px)'; uploadFloatingBtn.style.boxShadow = '0 4px 12px rgba(0,102,204,0.4)'; };
uploadFloatingBtn.onmouseout = function() { uploadFloatingBtn.style.transform = 'translateY(0)'; uploadFloatingBtn.style.boxShadow = '0 2px 8px rgba(0,102,204,0.3)'; };
uploadFloatingBtn.onclick = function() { uploadWindow.style.display = 'block'; };

// Lock button
const lockBtn = document.createElement('button');
lockBtn.id = 'lockBtn';
lockBtn.innerHTML = '🔓 加锁';
lockBtn.className = 'floating-btn';
lockBtn.style.cssText = 'padding: 12px 20px; background: #0066cc; color: white; border: none; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; box-shadow: 0 2px 8px rgba(0,102,204,0.3); transition: all 0.2s;';
lockBtn.onmouseover = function() { lockBtn.style.transform = 'translateY(-2px)'; lockBtn.style.boxShadow = '0 4px 12px rgba(0,102,204,0.4)'; };
lockBtn.onmouseout = function() { lockBtn.style.transform = 'translateY(0)'; lockBtn.style.boxShadow = '0 2px 8px rgba(0,102,204,0.3)'; };
lockBtn.onclick = function() { toggleLock(); };

// Preview/Edit toggle button (only shown on mobile)
const previewToggleBtn = document.createElement('button');
previewToggleBtn.id = 'preview-toggle-btn';
previewToggleBtn.innerHTML = '👁️ 预览';
previewToggleBtn.className = 'floating-btn';
previewToggleBtn.style.cssText = 'padding: 12px 20px; background: #0066cc; color: white; border: none; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; box-shadow: 0 2px 8px rgba(0,102,204,0.3); transition: all 0.2s; display: none;';
previewToggleBtn.onmouseover = function() { previewToggleBtn.style.transform = 'translateY(-2px)'; previewToggleBtn.style.boxShadow = '0 4px 12px rgba(0,102,204,0.4)'; };
previewToggleBtn.onmouseout = function() { previewToggleBtn.style.transform = 'translateY(0)'; previewToggleBtn.style.boxShadow = '0 2px 8px rgba(0,102,204,0.3)'; };
previewToggleBtn.onclick = function() { togglePreview(); };

// New note button
const newNoteBtn = document.createElement('a');
newNoteBtn.href = '/';
newNoteBtn.innerHTML = '📝 新建';
newNoteBtn.className = 'floating-btn';
newNoteBtn.style.cssText = 'padding: 12px 20px; background: #0066cc; color: white; border: none; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; box-shadow: 0 2px 8px rgba(0,102,204,0.3); transition: all 0.2s; text-decoration: none; display: inline-block;';
newNoteBtn.onmouseover = function() { newNoteBtn.style.transform = 'translateY(-2px)'; newNoteBtn.style.boxShadow = '0 4px 12px rgba(0,102,204,0.4)'; };
newNoteBtn.onmouseout = function() { newNoteBtn.style.transform = 'translateY(0)'; newNoteBtn.style.boxShadow = '0 2px 8px rgba(0,102,204,0.3)'; };

floatingActions.appendChild(uploadFloatingBtn);
floatingActions.appendChild(lockBtn);
floatingActions.appendChild(previewToggleBtn);
floatingActions.appendChild(newNoteBtn);
document.body.appendChild(floatingActions);

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

// Toggle preview panel (mobile only)
function togglePreview() {
    const editorPanel = document.querySelector('.editor-panel');
    const previewPanel = document.getElementById('preview-panel');
    const previewToggleBtn = document.getElementById('preview-toggle-btn');
    
    if (previewPanel.classList.contains('show')) {
        // Hide preview, show editor (currently in preview mode, switch to edit mode)
        previewPanel.classList.remove('show');
        editorPanel.classList.remove('hide');
        if (previewToggleBtn) previewToggleBtn.innerHTML = '👁️ 预览';
        editor.focus();
    } else {
        // Show preview, hide editor (currently in edit mode, switch to preview mode)
        previewPanel.classList.add('show');
        editorPanel.classList.add('hide');
        if (previewToggleBtn) previewToggleBtn.innerHTML = '✏️ 编辑';
        updatePreview();
    }
}

// Show toggle button on mobile
if (window.innerWidth <= 768) {
    const previewToggleBtn = document.getElementById('preview-toggle-btn');
    if (previewToggleBtn) previewToggleBtn.style.display = 'inline-block';
}

// Update on resize
window.addEventListener('resize', () => {
    const previewToggleBtn = document.getElementById('preview-toggle-btn');
    if (window.innerWidth <= 768) {
        if (previewToggleBtn) previewToggleBtn.style.display = 'inline-block';
    } else {
        if (previewToggleBtn) previewToggleBtn.style.display = 'none';
        // Show both panels on desktop
        const previewPanel = document.getElementById('preview-panel');
        const editorPanel = document.querySelector('.editor-panel');
        if (previewPanel) previewPanel.classList.add('show');
        if (editorPanel) editorPanel.classList.remove('hide');
    }
});

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
        document.getElementById('lockBtn').textContent = '🔒 解锁';
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
    
    // Copy share URL
    copyToClipboard(shareUrl);
    showStatus('地址已复制到剪贴板', false);
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
    showStatus('下载地址已复制到剪贴板', false);
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
    const lockBtn = document.getElementById('lockBtn');
    if (isLocked) {
        // Remove lock
        if (confirm('确定要移除笔记锁吗？')) {
            const currentContent = editor.value;
            const endIdx = currentContent.indexOf(' -->\n');
            if (endIdx !== -1) {
                editor.value = currentContent.substring(endIdx + 6); // Remove ' -->\n'
                isLocked = false;
                noteLockToken = '';
                if (lockBtn) {
                    lockBtn.textContent = '🔓 加锁';
                    lockBtn.style.background = '#0066cc';
                }
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
            if (lockBtn) {
                lockBtn.textContent = '🔒 解锁';
                lockBtn.style.background = '#e74c3c';
            }
            saveNote();
        }
    }
}
</script>
</body>
</html>`
