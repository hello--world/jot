package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/russross/blackfriday/v2"
)

const (
	savePath   = "_tmp"
	backupPath = "bak"
	uploadPath = "uploads" // Directory for uploaded files
)

var (
	adminPath              = "/admin" // Can be set via ADMIN_PATH env var or -admin-path flag
	port                   = ":8080"
	noteNameLen            = 3
	backupDays             = 7                                      // Move notes to backup after 7 days of inactivity
	noteChars              = "0123456789abcdefghijklmnopqrstuvwxyz" // Characters used for generating note names
	existingNotes          = sync.Map{}                             // In-memory cache of existing note names
	maxFileSize      int64 = 10 * 1024 * 1024                       // Maximum file size in bytes (default: 10MB)
	maxPathLength          = 20                                     // Maximum path/note name length (default: 20)
	maxTotalSize     int64 = 500 * 1024 * 1024                      // Maximum total file size in bytes (default: 500MB)
	maxTotalSizeLock       = sync.RWMutex{}                         // Lock for maxTotalSize
	maxNoteCount           = 500                                    // Maximum number of notes (default: 500)
	maxNoteCountLock       = sync.RWMutex{}                         // Lock for maxNoteCount
	configFile             = "config.json"                          // Configuration file path
)

const notePageHTML = `<!DOCTYPE html>
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
}
.editor-panel, .preview-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    height: 100vh;
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
            <div style="display: flex; gap: 10px; align-items: center;">
                <button onclick="toggleLock()" id="lockBtn" style="padding: 6px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">🔓 设置锁</button>
                <a href="/" style="padding: 6px 12px; background: #666; color: white; text-decoration: none; border-radius: 4px; font-size: 12px;">新建笔记</a>
            </div>
        </div>
        <textarea id="editor" placeholder="开始输入 Markdown 内容...">{{.Content}}</textarea>
    </div>
    <div class="preview-panel">
        <div class="panel-header">
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
    const url = addTokenToUrl('/api/markdown');
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

    const url = addTokenToUrl(window.location.pathname);
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
    const wsUrl = protocol + '//' + window.location.host + '/ws' + window.location.pathname;
    
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

        const uploadUrl = addTokenToUrl('/api/upload');
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
const ACCESS_TOKEN_KEY = 'jot_access_token';
let savedToken = localStorage.getItem(ACCESS_TOKEN_KEY);

// Get token from URL or use saved token
const urlParams = new URLSearchParams(window.location.search);
const urlToken = urlParams.get('token');
if (urlToken) {
    // Save token from URL to localStorage
    localStorage.setItem(ACCESS_TOKEN_KEY, urlToken);
    savedToken = urlToken;
    // Remove token from URL to keep it clean
    const newUrl = window.location.pathname;
    window.history.replaceState({}, '', newUrl);
}

// Add token to all requests
function addTokenToUrl(url) {
    if (savedToken) {
        const separator = url.includes('?') ? '&' : '?';
        return url + separator + 'token=' + encodeURIComponent(savedToken);
    }
    return url;
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
                document.getElementById('lockBtn').style.background = '';
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

const readPageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.NoteName}} - 只读模式</title>
<style>
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: #ebeef1;
    min-height: 100vh;
    padding: 20px;
}
.container {
    max-width: 900px;
    margin: 0 auto;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    overflow: hidden;
}
.header {
    background: #f5f5f5;
    padding: 15px 20px;
    border-bottom: 1px solid #ddd;
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.header-info {
    display: flex;
    gap: 20px;
    align-items: center;
    font-size: 13px;
    color: #666;
}
.header-info-item {
    display: flex;
    align-items: center;
    gap: 5px;
}
.header-info-label {
    color: #999;
}
.header-info-value {
    color: #333;
    font-weight: 500;
}
.header-actions {
    display: flex;
    gap: 10px;
}
.btn {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    font-size: 13px;
    cursor: pointer;
    text-decoration: none;
    display: inline-block;
    transition: all 0.2s;
}
.btn-primary {
    background: #0066cc;
    color: white;
}
.btn-primary:hover {
    background: #0052a3;
}
.btn-secondary {
    background: #f5f5f5;
    color: #666;
    border: 1px solid #ddd;
}
.btn-secondary:hover {
    background: #e9e9e9;
}
.content {
    padding: 30px;
}
#preview {
    line-height: 1.8;
    color: #333;
}
#preview h1, #preview h2, #preview h3, #preview h4, #preview h5, #preview h6 {
    margin-top: 1.5em;
    margin-bottom: 0.8em;
    font-weight: 600;
}
#preview h1 {
    font-size: 2em;
    border-bottom: 2px solid #eee;
    padding-bottom: 0.3em;
}
#preview h2 {
    font-size: 1.5em;
    border-bottom: 1px solid #eee;
    padding-bottom: 0.3em;
}
#preview p {
    margin-bottom: 1em;
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
    padding: 15px;
    border-radius: 5px;
    overflow-x: auto;
    margin: 1.5em 0;
    border-left: 4px solid #0066cc;
}
#preview pre code {
    background: none;
    padding: 0;
}
#preview blockquote {
    border-left: 4px solid #ddd;
    padding-left: 1em;
    margin: 1.5em 0;
    color: #666;
    font-style: italic;
}
#preview table {
    border-collapse: collapse;
    width: 100%;
    margin: 1.5em 0;
}
#preview table th, #preview table td {
    border: 1px solid #ddd;
    padding: 10px;
    text-align: left;
}
#preview table th {
    background: #f5f5f5;
    font-weight: 600;
}
#preview ul, #preview ol {
    margin: 1em 0;
    padding-left: 2em;
}
#preview li {
    margin: 0.5em 0;
}
#preview img {
    max-width: 100%;
    height: auto;
    border-radius: 4px;
    margin: 1em 0;
}
#preview a {
    color: #0066cc;
    text-decoration: none;
}
#preview a:hover {
    text-decoration: underline;
}
.empty {
    text-align: center;
    padding: 60px 20px;
    color: #999;
}
.empty-icon {
    font-size: 48px;
    margin-bottom: 16px;
}
@media (prefers-color-scheme: dark) {
    body {
        background: #333b4d;
    }
    .container {
        background: #24262b;
    }
    .header {
        background: #1a1a1a;
        border-color: #495265;
    }
    .header-info-label {
        color: #aaa;
    }
    .header-info-value {
        color: #fff;
    }
    .content {
        background: #24262b;
    }
    #preview {
        color: #fff;
    }
    #preview h1, #preview h2 {
        border-color: #495265;
    }
    #preview code {
        background: #1a1a1a;
    }
    #preview pre {
        background: #1a1a1a;
        border-color: #0066cc;
    }
    #preview blockquote {
        border-color: #495265;
    }
    #preview table th, #preview table td {
        border-color: #495265;
    }
    #preview table th {
        background: #1a1a1a;
    }
    .btn-secondary {
        background: #1a1a1a;
        color: #aaa;
        border-color: #495265;
    }
    .btn-secondary:hover {
        background: #2a2a2a;
    }
}
</style>
</head>
<body>
<div class="container">
    <div class="header">
        <div class="header-info">
            <div class="header-info-item">
                <span class="header-info-label">大小:</span>
                <span class="header-info-value">{{.FileSize}}</span>
            </div>
            <div class="header-info-item">
                <span class="header-info-label">创建:</span>
                <span class="header-info-value">{{.CreateTime}}</span>
            </div>
            <div class="header-info-item">
                <span class="header-info-label">修改:</span>
                <span class="header-info-value">{{.ModTime}}</span>
            </div>
        </div>
        <div class="header-actions">
            <a href="/{{.NoteName}}" class="btn btn-primary">编辑</a>
            <a href="/" class="btn btn-secondary">新建笔记</a>
        </div>
    </div>
    <div class="content">
        {{if .Content}}
        <div id="preview">{{.Content}}</div>
        {{else}}
        <div class="empty">
            <div class="empty-icon">📄</div>
            <p>笔记为空</p>
            <a href="/{{.NoteName}}" class="btn btn-primary" style="margin-top: 20px;">开始编辑</a>
        </div>
        {{end}}
    </div>
</div>
</body>
</html>`

const noteLockHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>笔记已锁定</title>
<style>
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    padding: 20px;
}
.container {
    background: white;
    border-radius: 12px;
    box-shadow: 0 10px 40px rgba(0,0,0,0.2);
    padding: 40px;
    max-width: 400px;
    width: 100%;
}
h1 {
    color: #333;
    margin-bottom: 10px;
    font-size: 24px;
}
.subtitle {
    color: #666;
    margin-bottom: 30px;
    font-size: 14px;
}
.form-group {
    margin-bottom: 20px;
}
label {
    display: block;
    color: #333;
    margin-bottom: 8px;
    font-size: 14px;
    font-weight: 500;
}
input[type="text"] {
    width: 100%;
    padding: 12px;
    border: 2px solid #e0e0e0;
    border-radius: 6px;
    font-size: 14px;
    transition: border-color 0.3s;
}
input[type="text"]:focus {
    outline: none;
    border-color: #667eea;
}
button {
    width: 100%;
    padding: 12px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
}
button:hover {
    transform: translateY(-2px);
    box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
}
button:active {
    transform: translateY(0);
}
.error {
    color: #e74c3c;
    font-size: 12px;
    margin-top: 8px;
    display: none;
}
.error.show {
    display: block;
}
</style>
</head>
<body>
<div class="container">
    <h1>🔒 笔记已锁定</h1>
    <p class="subtitle">此笔记需要解锁令牌才能查看</p>
    <form id="lockForm" onsubmit="handleLockSubmit(event)">
        <div class="form-group">
            <label for="lockToken">解锁令牌</label>
            <input type="text" id="lockToken" name="lockToken" placeholder="请输入解锁令牌" required autofocus>
            <div class="error" id="errorMsg"></div>
        </div>
        <button type="submit">解锁</button>
    </form>
</div>
<script>
function handleLockSubmit(event) {
    event.preventDefault();
    const token = document.getElementById('lockToken').value.trim();
    const noteName = '{{.NoteName}}';
    const errorMsg = document.getElementById('errorMsg');
    
    if (!token) {
        errorMsg.textContent = '请输入解锁令牌';
        errorMsg.classList.add('show');
        return;
    }
    
    // Set cookie and redirect
    document.cookie = 'note_lock_' + noteName + '=' + encodeURIComponent(token) + '; path=/; max-age=86400'; // 24 hours
    window.location.href = window.location.pathname + '?lock_token=' + encodeURIComponent(token);
}
</script>
</body>
</html>`

const accessLoginHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>访问授权</title>
<style>
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: #ebeef1;
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    padding: 20px;
}
.login-container {
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    padding: 40px;
    width: 100%;
    max-width: 400px;
}
.login-header {
    text-align: center;
    margin-bottom: 30px;
}
.login-header h1 {
    font-size: 24px;
    color: #333;
    margin-bottom: 8px;
}
.login-header p {
    color: #666;
    font-size: 14px;
}
.login-form {
    display: flex;
    flex-direction: column;
    gap: 20px;
}
.form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.form-group label {
    font-size: 14px;
    color: #333;
    font-weight: 500;
}
.form-group input {
    padding: 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    transition: border-color 0.2s;
}
.form-group input:focus {
    outline: none;
    border-color: #0066cc;
}
.error-message {
    color: #f44336;
    font-size: 14px;
    margin-top: -10px;
    display: none;
}
.error-message.show {
    display: block;
}
.login-button {
    padding: 12px;
    background: #0066cc;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
}
.login-button:hover {
    background: #0052a3;
}
.login-button:active {
    background: #003d7a;
}
@media (prefers-color-scheme: dark) {
    body {
        background: #333b4d;
    }
    .login-container {
        background: #24262b;
    }
    .login-header h1 {
        color: #fff;
    }
    .login-header p {
        color: #aaa;
    }
    .form-group label {
        color: #fff;
    }
    .form-group input {
        background: #1a1a1a;
        border-color: #495265;
        color: #fff;
    }
    .form-group input:focus {
        border-color: #0066cc;
    }
}
</style>
</head>
<body>
<div class="login-container">
    <div class="login-header">
        <h1>🔐 访问授权</h1>
        <p>请输入访问令牌以使用笔记功能</p>
    </div>
    <form class="login-form" id="loginForm" method="GET" action="">
        <div class="form-group">
            <label for="token">访问令牌</label>
            <input type="password" id="token" name="token" placeholder="输入访问令牌" required autofocus>
            <div class="error-message" id="errorMessage"></div>
        </div>
        <button type="submit" class="login-button">登录</button>
    </form>
</div>
<script>
const form = document.getElementById('loginForm');
const errorMessage = document.getElementById('errorMessage');
const tokenInput = document.getElementById('token');

// Set form action to current path
const currentPath = window.location.pathname;
const currentSearch = window.location.search;
form.action = currentPath + (currentSearch ? currentSearch + '&' : '?') + 'token=';

// Check if there's an error in URL
const urlParams = new URLSearchParams(window.location.search);
if (urlParams.get('error') === 'invalid') {
    errorMessage.textContent = '令牌无效，请重试';
    errorMessage.classList.add('show');
    tokenInput.focus();
}

form.addEventListener('submit', function(e) {
    const token = tokenInput.value.trim();
    if (!token) {
        e.preventDefault();
        errorMessage.textContent = '请输入访问令牌';
        errorMessage.classList.add('show');
        tokenInput.focus();
        return false;
    }
    // Update form action with token
    form.action = currentPath + (currentSearch ? currentSearch + '&' : '?') + 'token=' + encodeURIComponent(token);
});
</script>
</body>
</html>`

const adminLoginHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>管理后台登录</title>
<style>
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: #ebeef1;
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    padding: 20px;
}
.login-container {
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    padding: 40px;
    width: 100%;
    max-width: 400px;
}
.login-header {
    text-align: center;
    margin-bottom: 30px;
}
.login-header h1 {
    font-size: 24px;
    color: #333;
    margin-bottom: 8px;
}
.login-header p {
    color: #666;
    font-size: 14px;
}
.login-form {
    display: flex;
    flex-direction: column;
    gap: 20px;
}
.form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.form-group label {
    font-size: 14px;
    color: #333;
    font-weight: 500;
}
.form-group input {
    padding: 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    transition: border-color 0.2s;
}
.form-group input:focus {
    outline: none;
    border-color: #0066cc;
}
.error-message {
    color: #f44336;
    font-size: 14px;
    margin-top: -10px;
    display: none;
}
.error-message.show {
    display: block;
}
.login-button {
    padding: 12px;
    background: #0066cc;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
}
.login-button:hover {
    background: #0052a3;
}
.login-button:active {
    background: #003d7a;
}
@media (prefers-color-scheme: dark) {
    body {
        background: #333b4d;
    }
    .login-container {
        background: #24262b;
    }
    .login-header h1 {
        color: #fff;
    }
    .login-header p {
        color: #aaa;
    }
    .form-group label {
        color: #fff;
    }
    .form-group input {
        background: #1a1a1a;
        border-color: #495265;
        color: #fff;
    }
    .form-group input:focus {
        border-color: #0066cc;
    }
}
</style>
</head>
<body>
<div class="login-container">
    <div class="login-header">
        <h1>🔐 管理后台登录</h1>
        <p>请输入访问令牌</p>
    </div>
    <form class="login-form" id="loginForm" method="GET" action="/admin">
        <div class="form-group">
            <label for="token">访问令牌</label>
            <input type="password" id="token" name="token" placeholder="输入访问令牌" required autofocus>
            <div class="error-message" id="errorMessage"></div>
        </div>
        <button type="submit" class="login-button">登录</button>
    </form>
</div>
<script>
const form = document.getElementById('loginForm');
const errorMessage = document.getElementById('errorMessage');
const tokenInput = document.getElementById('token');

// Check if there's an error in URL
const urlParams = new URLSearchParams(window.location.search);
if (urlParams.get('error') === 'invalid') {
    errorMessage.textContent = '令牌无效，请重试';
    errorMessage.classList.add('show');
    tokenInput.focus();
}

form.addEventListener('submit', function(e) {
    const token = tokenInput.value.trim();
    if (!token) {
        e.preventDefault();
        errorMessage.textContent = '请输入访问令牌';
        errorMessage.classList.add('show');
        tokenInput.focus();
        return false;
    }
});
</script>
</body>
</html>`

const adminPageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>管理后台 - 所有笔记</title>
<style>
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    background: #ebeef1;
    padding: 20px;
}
.container {
    max-width: 1200px;
    margin: 0 auto;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    overflow: hidden;
}
.header {
    background: #0066cc;
    color: white;
    padding: 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.header h1 {
    font-size: 24px;
    font-weight: 500;
}
.header a {
    color: white;
    text-decoration: none;
    padding: 8px 16px;
    background: rgba(255,255,255,0.2);
    border-radius: 4px;
    transition: background 0.2s;
}
.header a:hover {
    background: rgba(255,255,255,0.3);
}
.stats {
    padding: 20px;
    background: #f5f5f5;
    border-bottom: 1px solid #ddd;
    display: flex;
    gap: 30px;
}
.stat-item {
    display: flex;
    flex-direction: column;
}
.stat-label {
    font-size: 12px;
    color: #666;
    margin-bottom: 4px;
}
.stat-value {
    font-size: 24px;
    font-weight: 600;
    color: #333;
}
.notes-list {
    padding: 20px;
}
.notes-table {
    width: 100%;
    border-collapse: collapse;
}
.notes-table th {
    background: #f5f5f5;
    padding: 12px;
    text-align: left;
    font-weight: 600;
    color: #333;
    border-bottom: 2px solid #ddd;
}
.notes-table td {
    padding: 12px;
    border-bottom: 1px solid #eee;
}
.notes-table tr:hover {
    background: #f9f9f9;
}
.note-name {
    font-weight: 500;
    color: #0066cc;
    text-decoration: none;
}
.note-name:hover {
    text-decoration: underline;
}
.note-content {
    max-width: 400px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #666;
    font-size: 14px;
}
.note-date {
    color: #999;
    font-size: 13px;
}
.note-size {
    color: #999;
    font-size: 13px;
}
.empty {
    text-align: center;
    padding: 60px 20px;
    color: #999;
}
.empty-icon {
    font-size: 48px;
    margin-bottom: 16px;
}
.tabs {
    display: flex;
    gap: 10px;
    padding: 20px;
    background: #f5f5f5;
    border-bottom: 1px solid #ddd;
}
.tab-button {
    padding: 10px 20px;
    background: #fff;
    border: 1px solid #ddd;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    color: #666;
    transition: all 0.2s;
}
.tab-button:hover {
    background: #f0f0f0;
}
.tab-button.active {
    background: #0066cc;
    color: white;
    border-color: #0066cc;
}
.tab-content {
    display: block;
}
@media (prefers-color-scheme: dark) {
    body {
        background: #333b4d;
    }
    .container {
        background: #24262b;
    }
    .header {
        background: #0066cc;
    }
    .stats {
        background: #1a1a1a;
        border-color: #495265;
    }
    .stat-label {
        color: #aaa;
    }
    .stat-value {
        color: #fff;
    }
    .notes-table th {
        background: #1a1a1a;
        color: #fff;
        border-color: #495265;
    }
    .notes-table td {
        border-color: #495265;
    }
    .notes-table tr:hover {
        background: #1a1a1a;
    }
    .note-content {
        color: #aaa;
    }
    .note-date, .note-size {
        color: #666;
    }
    .empty {
        color: #666;
    }
}
</style>
</head>
<body>
<div class="container">
    <div class="header">
        <h1>📝 笔记管理后台</h1>
        <a href="/">新建笔记</a>
    </div>
    <div class="tabs">
        <button class="tab-button active" onclick="showTab('active')">活跃笔记 ({{.TotalCount}})</button>
        <button class="tab-button" onclick="showTab('backup')">备份笔记 ({{.BackupCount}})</button>
    </div>
    <div class="stats">
        <div class="stat-item">
            <span class="stat-label" id="stat-label">总笔记数</span>
            <span class="stat-value" id="total-notes">{{.TotalCount}}</span>
        </div>
        <div class="stat-item">
            <span class="stat-label">总大小</span>
            <span class="stat-value" id="total-size">{{formatSize .TotalSize}}</span>
        </div>
        <div class="stat-item">
            <span class="stat-label">当前总文件大小（含上传）</span>
            <span class="stat-value" id="current-total-size">{{formatSize .CurrentTotalSize}}</span>
        </div>
        <div class="stat-item">
            <span class="stat-label">最大总文件大小限制</span>
            <span class="stat-value" id="max-total-size">{{formatSize .MaxTotalSize}}</span>
            <div style="margin-top: 8px; display: flex; gap: 8px; align-items: center;">
                <input type="text" id="max-total-size-input" placeholder="如: 500MB" style="padding: 4px 8px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px; width: 120px;">
                <button onclick="updateMaxTotalSize()" style="padding: 4px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
            </div>
        </div>
        <div class="stat-item">
            <span class="stat-label">最大笔记数量限制</span>
            <span class="stat-value" id="max-note-count">{{.MaxNoteCount}}</span>
            <div style="margin-top: 8px; display: flex; gap: 8px; align-items: center;">
                <input type="number" id="max-note-count-input" placeholder="如: 500" min="1" style="padding: 4px 8px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px; width: 120px;">
                <button onclick="updateConfig('maxNoteCount')" style="padding: 4px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
            </div>
        </div>
    </div>
    <div style="padding: 20px; background: #f9f9f9; border-top: 1px solid #ddd;">
        <h3 style="margin-bottom: 15px; font-size: 16px; color: #333;">配置管理</h3>
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 15px;">
            <div style="background: white; padding: 15px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 5px; font-size: 12px; color: #666;">访问令牌（用于访问笔记）</label>
                <div style="display: flex; gap: 8px;">
                    <input type="text" id="access-token-input" value="{{.AccessToken}}" placeholder="留空表示无需授权" style="flex: 1; padding: 6px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px;">
                    <button onclick="updateConfig('accessToken')" style="padding: 6px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
                </div>
                <div style="margin-top: 5px; font-size: 11px; color: #999;">留空表示无需授权即可访问笔记</div>
            </div>
            <div style="background: white; padding: 15px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 5px; font-size: 12px; color: #666;">管理后台路径</label>
                <div style="display: flex; gap: 8px;">
                    <input type="text" id="admin-path-input" value="{{.AdminPath}}" style="flex: 1; padding: 6px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px;">
                    <button onclick="updateConfig('adminPath')" style="padding: 6px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
                </div>
            </div>
            <div style="background: white; padding: 15px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 5px; font-size: 12px; color: #666;">笔记名称最小长度</label>
                <div style="display: flex; gap: 8px;">
                    <input type="number" id="note-name-len-input" value="{{.NoteNameLen}}" min="1" style="flex: 1; padding: 6px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px;">
                    <button onclick="updateConfig('noteNameLen')" style="padding: 6px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
                </div>
            </div>
            <div style="background: white; padding: 15px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 5px; font-size: 12px; color: #666;">备份天数</label>
                <div style="display: flex; gap: 8px;">
                    <input type="number" id="backup-days-input" value="{{.BackupDays}}" min="1" style="flex: 1; padding: 6px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px;">
                    <button onclick="updateConfig('backupDays')" style="padding: 6px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
                </div>
            </div>
            <div style="background: white; padding: 15px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 5px; font-size: 12px; color: #666;">随机字符串字符集</label>
                <div style="display: flex; gap: 8px;">
                    <input type="text" id="note-chars-input" value="{{.NoteChars}}" style="flex: 1; padding: 6px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px;">
                    <button onclick="updateConfig('noteChars')" style="padding: 6px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
                </div>
            </div>
            <div style="background: white; padding: 15px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 5px; font-size: 12px; color: #666;">最大文件大小</label>
                <div style="display: flex; gap: 8px;">
                    <input type="text" id="max-file-size-input" placeholder="如: 10MB" style="flex: 1; padding: 6px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px;">
                    <button onclick="updateConfig('maxFileSize')" style="padding: 6px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
                </div>
                <div style="margin-top: 5px; font-size: 11px; color: #999;">当前: {{.MaxFileSizeMB}} MB</div>
            </div>
            <div style="background: white; padding: 15px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 5px; font-size: 12px; color: #666;">最大路径长度</label>
                <div style="display: flex; gap: 8px;">
                    <input type="number" id="max-path-length-input" value="{{.MaxPathLength}}" min="1" style="flex: 1; padding: 6px; border: 1px solid #ddd; border-radius: 4px; font-size: 12px;">
                    <button onclick="updateConfig('maxPathLength')" style="padding: 6px 12px; background: #0066cc; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">更新</button>
                </div>
            </div>
        </div>
    </div>
    <div class="notes-list">
        <div id="active-notes" class="tab-content">
            {{if .Notes}}
            <table class="notes-table">
                <thead>
                    <tr>
                        <th>笔记名称</th>
                        <th>内容预览</th>
                        <th>大小</th>
                        <th>更新时间</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Notes}}
                    <tr>
                        <td><a href="/{{.Name}}" class="note-name">{{.Name}}</a></td>
                        <td class="note-content" title="{{.Content}}">{{if .Content}}{{preview .Content 50}}{{else}}<em>空笔记</em>{{end}}</td>
                        <td class="note-size">{{formatSize .Size}}</td>
                        <td class="note-date">{{formatDate .UpdatedAt}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <div class="empty">
                <div class="empty-icon">📄</div>
                <p>还没有笔记，<a href="/" style="color: #0066cc;">创建第一个笔记</a></p>
            </div>
            {{end}}
        </div>
        <div id="backup-notes" class="tab-content" style="display: none;">
            {{if .BackupNotes}}
            <table class="notes-table">
                <thead>
                    <tr>
                        <th>笔记名称</th>
                        <th>内容预览</th>
                        <th>大小</th>
                        <th>更新时间</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .BackupNotes}}
                    <tr>
                        <td><span class="note-name">{{.Name}}</span></td>
                        <td class="note-content" title="{{.Content}}">{{if .Content}}{{preview .Content 50}}{{else}}<em>空笔记</em>{{end}}</td>
                        <td class="note-size">{{formatSize .Size}}</td>
                        <td class="note-date">{{formatDate .UpdatedAt}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <div class="empty">
                <div class="empty-icon">📦</div>
                <p>还没有备份笔记</p>
            </div>
            {{end}}
        </div>
    </div>
</div>
<script>
function showTab(tabName) {
    // Hide all tab contents
    document.getElementById('active-notes').style.display = 'none';
    document.getElementById('backup-notes').style.display = 'none';
    
    // Remove active class from all buttons
    document.querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected tab
    if (tabName === 'active') {
        document.getElementById('active-notes').style.display = 'block';
        document.querySelector('.tab-button:first-child').classList.add('active');
        document.getElementById('total-notes').textContent = '{{.TotalCount}}';
        document.getElementById('total-size').textContent = '{{formatSize .TotalSize}}';
        document.getElementById('stat-label').textContent = '总笔记数';
    } else {
        document.getElementById('backup-notes').style.display = 'block';
        document.querySelector('.tab-button:last-child').classList.add('active');
        document.getElementById('total-notes').textContent = '{{.BackupCount}}';
        document.getElementById('total-size').textContent = '{{formatSize .BackupTotalSize}}';
        document.getElementById('stat-label').textContent = '备份笔记数';
    }
}

// Auto refresh every 30 seconds
setInterval(() => {
    location.reload();
}, 30000);

function updateMaxTotalSize() {
    updateConfig('maxTotalSize');
}

function updateMaxNoteCount() {
    updateConfig('maxNoteCount');
}

function updateConfig(field) {
    const token = '{{.Token}}';
    let payload = {};
    let value;

    switch(field) {
        case 'accessToken':
            value = document.getElementById('access-token-input').value.trim();
            payload.accessToken = value;
            break;
        case 'adminPath':
            value = document.getElementById('admin-path-input').value.trim();
            if (!value) {
                alert('请输入管理后台路径');
                return;
            }
            payload.adminPath = value;
            break;
        case 'noteNameLen':
            value = parseInt(document.getElementById('note-name-len-input').value);
            if (isNaN(value) || value <= 0) {
                alert('请输入有效的数字');
                return;
            }
            payload.noteNameLen = value;
            break;
        case 'backupDays':
            value = parseInt(document.getElementById('backup-days-input').value);
            if (isNaN(value) || value <= 0) {
                alert('请输入有效的数字');
                return;
            }
            payload.backupDays = value;
            break;
        case 'noteChars':
            value = document.getElementById('note-chars-input').value.trim();
            if (!value) {
                alert('请输入字符集');
                return;
            }
            payload.noteChars = value;
            break;
        case 'maxFileSize':
            value = document.getElementById('max-file-size-input').value.trim();
            if (!value) {
                alert('请输入文件大小限制（如: 10MB）');
                return;
            }
            payload.maxFileSize = value;
            break;
        case 'maxPathLength':
            value = parseInt(document.getElementById('max-path-length-input').value);
            if (isNaN(value) || value <= 0) {
                alert('请输入有效的数字');
                return;
            }
            payload.maxPathLength = value;
            break;
        case 'maxTotalSize':
            value = document.getElementById('max-total-size-input').value.trim();
            if (!value) {
                alert('请输入大小限制（如: 500MB）');
                return;
            }
            payload.maxTotalSize = value;
            break;
        case 'maxNoteCount':
            value = parseInt(document.getElementById('max-note-count-input').value);
            if (isNaN(value) || value <= 0) {
                alert('请输入有效的数字');
                return;
            }
            payload.maxNoteCount = value;
            break;
        default:
            alert('未知的配置项');
            return;
    }

    fetch('/api/max-total-size?token=' + encodeURIComponent(token), {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
    })
    .then(res => res.json())
    .then(data => {
        if (data.success) {
            alert('配置已更新并保存到配置文件');
            location.reload();
        } else {
            alert('更新失败: ' + (data.error || '未知错误'));
        }
    })
    .catch(err => {
        console.error('Update error:', err);
        alert('更新失败');
    });
}
</script>
</body>
</html>`

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients     = make(map[string]map[*websocket.Conn]bool)
	clientsLock = sync.RWMutex{}
	adminToken  string
	accessToken string // Token for accessing notes (different from admin token)
)

type Note struct {
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
	Size      int64     `json:"size"`
}

type NoteListResponse struct {
	Notes []Note `json:"notes"`
}

func init() {
	os.MkdirAll(savePath, 0755)
	os.MkdirAll(backupPath, 0755)
	os.MkdirAll(uploadPath, 0755)
	// Try to load configuration from file (if exists)
	// If config file doesn't exist, will load from env/command line in main()
	loadConfig()
}

// Config structure for saving/loading configuration
type Config struct {
	AdminToken    string `json:"adminToken"`
	AccessToken   string `json:"accessToken"`
	AdminPath     string `json:"adminPath"`
	NoteNameLen   int    `json:"noteNameLen"`
	BackupDays    int    `json:"backupDays"`
	NoteChars     string `json:"noteChars"`
	MaxFileSize   int64  `json:"maxFileSize"`
	MaxPathLength int    `json:"maxPathLength"`
	MaxTotalSize  int64  `json:"maxTotalSize"`
	MaxNoteCount  int    `json:"maxNoteCount"`
}

// configLoaded indicates if config file exists and was loaded
var configLoaded = false

// loadConfig loads configuration from config.json file
// Returns true if config file exists and was loaded successfully
func loadConfig() bool {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, will use defaults from env/command line
			configLoaded = false
			return false
		}
		log.Printf("Warning: Failed to read config file: %v", err)
		configLoaded = false
		return false
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Warning: Failed to parse config file: %v", err)
		configLoaded = false
		return false
	}

	// Update all values from config file
	if config.AdminToken != "" {
		adminToken = config.AdminToken
	}
	if config.AccessToken != "" {
		accessToken = config.AccessToken
	}
	if config.AdminPath != "" {
		adminPath = config.AdminPath
		if !strings.HasPrefix(adminPath, "/") {
			adminPath = "/" + adminPath
		}
	}
	if config.NoteNameLen > 0 {
		noteNameLen = config.NoteNameLen
	}
	if config.BackupDays > 0 {
		backupDays = config.BackupDays
	}
	if config.NoteChars != "" {
		noteChars = config.NoteChars
	}
	if config.MaxFileSize > 0 {
		maxFileSize = config.MaxFileSize
	}
	if config.MaxPathLength > 0 {
		maxPathLength = config.MaxPathLength
	}
	if config.MaxTotalSize > 0 {
		maxTotalSizeLock.Lock()
		maxTotalSize = config.MaxTotalSize
		maxTotalSizeLock.Unlock()
	}
	if config.MaxNoteCount > 0 {
		maxNoteCountLock.Lock()
		maxNoteCount = config.MaxNoteCount
		maxNoteCountLock.Unlock()
	}

	configLoaded = true
	return true
}

// saveConfig saves current configuration to config.json file
func saveConfig() {
	maxTotalSizeLock.RLock()
	currentMaxTotalSize := maxTotalSize
	maxTotalSizeLock.RUnlock()

	maxNoteCountLock.RLock()
	currentMaxNoteCount := maxNoteCount
	maxNoteCountLock.RUnlock()

	config := Config{
		AdminToken:    adminToken,
		AccessToken:   accessToken,
		AdminPath:     adminPath,
		NoteNameLen:   noteNameLen,
		BackupDays:    backupDays,
		NoteChars:     noteChars,
		MaxFileSize:   maxFileSize,
		MaxPathLength: maxPathLength,
		MaxTotalSize:  currentMaxTotalSize,
		MaxNoteCount:  currentMaxNoteCount,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("Warning: Failed to marshal config: %v", err)
		return
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		log.Printf("Warning: Failed to save config file: %v", err)
		return
	}

	configLoaded = true
	log.Printf("Configuration saved to %s", configFile)
}

// loadEnvFile 从 .env 文件加载环境变量
func loadEnvFile() error {
	file, err := os.Open(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return nil // .env file is optional
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

// loadExistingNotes 将所有已存在的笔记名称加载到内存中
func loadExistingNotes() error {
	files, err := os.ReadDir(savePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist yet, no notes to load
		}
		return err
	}

	for _, file := range files {
		if !file.IsDir() && isSafeNoteName(file.Name()) {
			existingNotes.Store(file.Name(), true)
		}
	}

	return nil
}

// isNoteExists 检查笔记名称是否已存在（使用内存缓存）
func isNoteExists(name string) bool {
	_, exists := existingNotes.Load(name)
	return exists
}

// addNoteToCache 将笔记名称添加到内存缓存中
func addNoteToCache(name string) {
	existingNotes.Store(name, true)
}

// removeNoteFromCache 从内存缓存中移除笔记名称
func removeNoteFromCache(name string) {
	existingNotes.Delete(name)
}

func generateNoteName() string {
	// 从最小长度开始，如果名称已存在则增加长度
	length := noteNameLen
	maxAttempts := 100 // 由于使用内存缓存，增加尝试次数

	for attempt := 0; attempt < maxAttempts; attempt++ {
		name := make([]byte, length)
		for i := range name {
			name[i] = noteChars[rand.Intn(len(noteChars))]
		}
		noteName := string(name)

		// 使用内存缓存检查笔记是否已存在
		if !isNoteExists(noteName) {
			// 立即添加到缓存以防止竞态条件
			addNoteToCache(noteName)
			return noteName
		}

		// 如果名称存在且尚未达到最大长度，增加长度
		// 如果已经是 4 位或更多，则使用相同长度重试
		if length < 4 {
			length = 4
		} else if attempt%10 == 0 {
			// 每 10 次尝试，增加长度以避免太多冲突
			length++
		}
	}

	// 如果所有尝试都失败，继续尝试增加长度
	// 如果 noteChars 有足够的字符，这应该很少发生
	for length < 20 {
		length++
		for attempt := 0; attempt < 50; attempt++ {
			name := make([]byte, length)
			for i := range name {
				name[i] = noteChars[rand.Intn(len(noteChars))]
			}
			noteName := string(name)
			if !isNoteExists(noteName) {
				addNoteToCache(noteName)
				return noteName
			}
		}
	}

	// 最后手段：这在实践中不应该发生
	// 但如果发生了，返回一个错误指示名称
	log.Printf("警告: 经过多次尝试后仍无法生成唯一的笔记名称")
	return ""
}

// isSafeNoteName checks if a note name is safe (prevents path traversal attacks)
// This is only for basic security, not for restricting user input
func isSafeNoteName(name string) bool {
	if name == "" || len(name) > maxPathLength {
		return false
	}
	// Prevent path traversal and other dangerous patterns
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	// Prevent control characters
	for _, r := range name {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

// isValidGeneratedName 检查名称是否对自动生成的笔记有效
// 这限制生成的名称只能使用 noteChars 字符集中的字符
func isValidGeneratedName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !strings.ContainsRune(noteChars, r) {
			return false
		}
	}
	return true
}

func getNotePath(name string) string {
	return filepath.Join(savePath, name)
}

// parseFileSize parses a file size string like "10M", "100MB", "1G" into bytes
func parseFileSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Remove "B" or "BYTE" suffix if present (case insensitive)
	sizeStr = strings.ToUpper(sizeStr)
	sizeStr = strings.TrimSuffix(sizeStr, "BYTES")
	sizeStr = strings.TrimSuffix(sizeStr, "BYTE")
	sizeStr = strings.TrimSuffix(sizeStr, "B")
	sizeStr = strings.TrimSpace(sizeStr)

	if sizeStr == "" {
		return 0, fmt.Errorf("invalid size format")
	}

	// Find the last non-digit character to determine the unit
	var numStr string
	var unit string
	for i := len(sizeStr) - 1; i >= 0; i-- {
		if sizeStr[i] >= '0' && sizeStr[i] <= '9' || sizeStr[i] == '.' {
			numStr = sizeStr[:i+1]
			unit = sizeStr[i+1:]
			break
		}
	}
	if numStr == "" {
		numStr = sizeStr
		unit = ""
	}

	// Parse the number
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %v", err)
	}

	// Convert based on unit
	unit = strings.ToUpper(strings.TrimSpace(unit))
	switch unit {
	case "", "B":
		return int64(num), nil
	case "K", "KB":
		return int64(num * 1024), nil
	case "M", "MB":
		return int64(num * 1024 * 1024), nil
	case "G", "GB":
		return int64(num * 1024 * 1024 * 1024), nil
	case "T", "TB":
		return int64(num * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// getFileCreationTime 获取文件的创建时间
// Windows: 通过 syscall 获取真实的创建时间
// Linux/Unix: 使用修改时间作为近似值（大多数文件系统不存储创建时间）
// 其他平台: 使用修改时间作为近似值
func getFileCreationTime(path string) (time.Time, error) {
	// 获取文件信息
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}

	// Windows: 通过 Win32FileAttributeData 获取创建时间
	// 使用 build tags 来避免在非 Windows 平台上编译错误
	if runtime.GOOS == "windows" {
		creationTime := getFileCreationTimeWindows(info)
		if !creationTime.IsZero() {
			return creationTime, nil
		}
	}

	// Linux/Unix: 标准库的 Stat_t 不包含 birthtime 字段
	// 虽然 ext4 等文件系统支持创建时间，但需要通过 statx 系统调用获取
	// 这需要 CGO 或额外的系统调用，为了简化，这里使用修改时间作为近似值
	// 注意：在 Linux 上，大多数文件系统不存储创建时间，只有修改时间和访问时间
	// 如果文件系统支持，可以通过 statx 系统调用获取，但需要额外的实现
	// 这里统一使用修改时间作为回退

	// 其他平台或获取失败时，使用修改时间作为近似值
	return info.ModTime(), nil
}

// getFileCreationTimeWindows 在 Windows 平台上获取文件创建时间
// 这个函数的实现位于 windows.go（Windows）和 windows_stub.go（非 Windows）文件中

func saveNote(name, content string) error {
	path := getNotePath(name)
	if content == "" {
		os.Remove(path)
		removeNoteFromCache(name)
		return nil
	}

	// 检查这是否是新笔记
	wasNewNote := !isNoteExists(name)

	err := os.WriteFile(path, []byte(content), 0644)
	if err == nil && wasNewNote {
		// 如果是新笔记，添加到缓存
		addNoteToCache(name)
	}
	return err
}

func loadNote(name string) (string, error) {
	path := getNotePath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// Note lock functions
const lockPrefix = "<!-- LOCK:"
const lockSuffix = " -->\n"

// hasNoteLock checks if a note has a lock
func hasNoteLock(content string) bool {
	return strings.HasPrefix(content, lockPrefix)
}

// getNoteLockToken extracts the lock token from note content
// Returns empty string if no lock
func getNoteLockToken(content string) string {
	if !hasNoteLock(content) {
		return ""
	}
	// Extract token from <!-- LOCK:token -->
	endIdx := strings.Index(content, lockSuffix)
	if endIdx == -1 {
		return ""
	}
	return content[len(lockPrefix):endIdx]
}

// getNoteContent extracts the actual content from a locked note
func getNoteContent(content string) string {
	if !hasNoteLock(content) {
		return content
	}
	// Remove <!-- LOCK:token -->\n prefix
	endIdx := strings.Index(content, lockSuffix)
	if endIdx == -1 {
		return content
	}
	return content[endIdx+len(lockSuffix):]
}

// setNoteLock adds a lock to note content
func setNoteLock(content, token string) string {
	if token == "" {
		// Remove lock if token is empty
		return getNoteContent(content)
	}
	// If already locked, replace the token
	if hasNoteLock(content) {
		actualContent := getNoteContent(content)
		return lockPrefix + token + lockSuffix + actualContent
	}
	// Add new lock
	return lockPrefix + token + lockSuffix + content
}

func getAllNotes() ([]Note, error) {
	files, err := os.ReadDir(savePath)
	if err != nil {
		return nil, err
	}

	notes := make([]Note, 0)
	for _, file := range files {
		if !file.IsDir() && isSafeNoteName(file.Name()) {
			content, _ := loadNote(file.Name())
			info, _ := file.Info()
			notes = append(notes, Note{
				Name:      file.Name(),
				Content:   content,
				UpdatedAt: info.ModTime(),
				Size:      info.Size(),
			})
		}
	}
	return notes, nil
}

// getTotalFileSize calculates the total size of all files in savePath and uploadPath (excluding backupPath)
func getTotalFileSize() (int64, error) {
	var totalSize int64

	// Calculate size of notes in savePath
	if err := filepath.Walk(savePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	}); err != nil && !os.IsNotExist(err) {
		return 0, err
	}

	// Calculate size of uploaded files in uploadPath
	if err := filepath.Walk(uploadPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	}); err != nil && !os.IsNotExist(err) {
		return 0, err
	}

	return totalSize, nil
}

// moveOldNotesToBackup 将超过 backupDays 天未修改的笔记移动到备份文件夹
// 备份文件夹结构: bak/YYYYMMDD/note_name
func moveOldNotesToBackup() error {
	files, err := os.ReadDir(savePath)
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -backupDays)
	movedCount := 0

	for _, file := range files {
		if file.IsDir() || !isSafeNoteName(file.Name()) {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Check if note hasn't been modified for backupDays days
		if info.ModTime().Before(cutoffTime) {
			sourcePath := getNotePath(file.Name())
			// Use creation/modification date as directory name (YYYYMMDD format)
			dateDir := info.ModTime().Format("20060102")
			dateBackupPath := filepath.Join(backupPath, dateDir)

			// Create date directory if it doesn't exist
			if err := os.MkdirAll(dateBackupPath, 0755); err != nil {
				log.Printf("Failed to create backup directory %s: %v", dateBackupPath, err)
				continue
			}

			backupFilePath := filepath.Join(dateBackupPath, file.Name())

			// Move file to backup folder
			if err := os.Rename(sourcePath, backupFilePath); err != nil {
				log.Printf("Failed to move note %s to backup: %v", file.Name(), err)
				continue
			}

			movedCount++
			log.Printf("Moved old note %s to backup/%s (last modified: %s)", file.Name(), dateDir, info.ModTime().Format("2006-01-02 15:04:05"))
		}
	}

	if movedCount > 0 {
		log.Printf("Moved %d old note(s) to backup folder", movedCount)
	}

	return nil
}

// getAllBackupNotes 返回备份文件夹中的所有笔记
func getAllBackupNotes() ([]Note, error) {
	backupNotes := make([]Note, 0)

	// Check if backup directory exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return backupNotes, nil
	}

	// Read all date directories in backup folder
	dateDirs, err := os.ReadDir(backupPath)
	if err != nil {
		return nil, err
	}

	for _, dateDir := range dateDirs {
		if !dateDir.IsDir() {
			continue
		}

		datePath := filepath.Join(backupPath, dateDir.Name())
		files, err := os.ReadDir(datePath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !isSafeNoteName(file.Name()) {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			// Read note content
			notePath := filepath.Join(datePath, file.Name())
			content, err := os.ReadFile(notePath)
			if err != nil {
				continue
			}

			backupNotes = append(backupNotes, Note{
				Name:      file.Name(),
				Content:   string(content),
				UpdatedAt: info.ModTime(),
				Size:      info.Size(),
			})
		}
	}

	return backupNotes, nil
}

func handleNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteName := vars["note"]

	// 只检查是否为空或不安全（允许用户输入任意字符）
	if noteName == "" || !isSafeNoteName(noteName) {
		http.Redirect(w, r, "/"+generateNoteName(), http.StatusFound)
		return
	}

	if r.Method == "GET" {
		// 检查是否是 raw 请求或 curl/wget
		if r.URL.Query().Get("raw") != "" || strings.HasPrefix(r.UserAgent(), "curl") || strings.HasPrefix(r.UserAgent(), "Wget") {
			content, err := loadNote(noteName)
			if err != nil || content == "" {
				http.NotFound(w, r)
				return
			}
			// Check note lock for raw requests
			if hasNoteLock(content) {
				lockToken := getNoteLockToken(content)
				providedToken := r.URL.Query().Get("lock_token")
				if providedToken == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						providedToken = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}
				if providedToken != lockToken {
					http.Error(w, "Unauthorized: Note is locked. Provide lock_token parameter or Authorization header.", http.StatusUnauthorized)
					return
				}
				content = getNoteContent(content)
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(content))
			return
		}

		// Serve HTML page
		rawContent, _ := loadNote(noteName)

		// Check note lock
		if hasNoteLock(rawContent) {
			lockToken := getNoteLockToken(rawContent)
			providedToken := r.URL.Query().Get("lock_token")
			if providedToken == "" {
				// Try to get from cookie
				cookie, err := r.Cookie("note_lock_" + noteName)
				if err == nil {
					providedToken = cookie.Value
				}
			}
			if providedToken != lockToken {
				// Show lock login page
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				tmpl := template.Must(template.New("lock").Parse(noteLockHTML))
				tmpl.Execute(w, map[string]interface{}{
					"NoteName": noteName,
				})
				return
			}
			// Token is correct, extract actual content
			rawContent = getNoteContent(rawContent)
		}

		content := rawContent

		// 获取文件信息（大小和修改时间）
		var fileSize int64
		var modTime time.Time
		var createTime time.Time
		notePath := getNotePath(noteName)

		if info, err := os.Stat(notePath); err == nil {
			fileSize = info.Size()
			modTime = info.ModTime()
			// 尝试获取创建时间（Windows 上可用，其他平台回退到修改时间）
			if ct, err := getFileCreationTime(notePath); err == nil {
				createTime = ct
			} else {
				// 如果无法获取创建时间，回退到修改时间
				createTime = modTime
			}
		}

		// Format size
		sizeStr := fmt.Sprintf("%d B", fileSize)
		if fileSize >= 1024 {
			sizeStr = fmt.Sprintf("%.2f KB", float64(fileSize)/1024.0)
		}

		tmpl := template.Must(template.New("note").Parse(notePageHTML))
		tmpl.Execute(w, map[string]interface{}{
			"NoteName":   noteName,
			"Content":    template.HTML(template.HTMLEscapeString(content)),
			"FileSize":   sizeStr,
			"ModTime":    modTime.Format("2006-01-02 15:04:05"),
			"CreateTime": createTime.Format("2006-01-02 15:04:05"),
		})

		// Set cookie if token was provided
		if lockToken := r.URL.Query().Get("lock_token"); lockToken != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     "note_lock_" + noteName,
				Value:    lockToken,
				Path:     "/",
				MaxAge:   86400, // 24 hours
				HttpOnly: false,
			})
		}
		return
	}

	if r.Method == "POST" {
		// Check access token for POST requests (creating/updating notes)
		if accessToken != "" {
			token := r.URL.Query().Get("token")
			if token == "" {
				// Try to get from Authorization header
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}
			if token != accessToken {
				http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
				return
			}
		}

		body, _ := io.ReadAll(r.Body)
		content := string(body)

		// Handle form-encoded data
		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			r.ParseForm()
			if text := r.FormValue("text"); text != "" {
				content = text
			}
		}

		// Check file size limit
		contentSize := int64(len([]byte(content)))
		if contentSize > maxFileSize {
			http.Error(w, fmt.Sprintf("File size exceeds maximum limit of %d bytes (%d MB)", maxFileSize, maxFileSize/(1024*1024)), http.StatusRequestEntityTooLarge)
			return
		}

		// Check note count limit (only for new notes)
		wasNewNote := !isNoteExists(noteName)
		if wasNewNote {
			maxNoteCountLock.RLock()
			currentMaxNoteCount := maxNoteCount
			maxNoteCountLock.RUnlock()

			// Count existing notes
			notes, err := getAllNotes()
			if err != nil {
				log.Printf("Error getting notes: %v", err)
			} else {
				if len(notes) >= currentMaxNoteCount {
					http.Error(w, fmt.Sprintf("Maximum number of notes (%d) has been reached. Please delete some notes or increase the limit in admin panel.", currentMaxNoteCount), http.StatusForbidden)
					return
				}
			}
		}

		// Check total file size limit
		maxTotalSizeLock.RLock()
		currentMaxTotalSize := maxTotalSize
		maxTotalSizeLock.RUnlock()

		currentTotalSize, err := getTotalFileSize()
		if err != nil {
			log.Printf("Error calculating total file size: %v", err)
		} else {
			// Get current note size if it exists
			var currentNoteSize int64
			if info, err := os.Stat(getNotePath(noteName)); err == nil {
				currentNoteSize = info.Size()
			}
			// Calculate new total size
			newTotalSize := currentTotalSize - currentNoteSize + contentSize
			if newTotalSize > currentMaxTotalSize {
				http.Error(w, fmt.Sprintf("Total file size would exceed maximum limit of %d MB (current: %.2f MB, would be: %.2f MB)", currentMaxTotalSize/(1024*1024), float64(currentTotalSize)/(1024*1024), float64(newTotalSize)/(1024*1024)), http.StatusRequestEntityTooLarge)
				return
			}
		}

		if err := saveNote(noteName, content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Broadcast update to WebSocket clients
		broadcastUpdate(noteName, content)

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleReadNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	noteName := vars["note"]

	// 只检查是否为空或不安全
	if noteName == "" || !isSafeNoteName(noteName) {
		http.NotFound(w, r)
		return
	}

	// Load note content
	rawContent, err := loadNote(noteName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check note lock
	if hasNoteLock(rawContent) {
		lockToken := getNoteLockToken(rawContent)
		providedToken := r.URL.Query().Get("lock_token")
		if providedToken == "" {
			// Try to get from cookie
			cookie, err := r.Cookie("note_lock_" + noteName)
			if err == nil {
				providedToken = cookie.Value
			}
		}
		if providedToken != lockToken {
			// Show lock login page
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl := template.Must(template.New("lock").Parse(noteLockHTML))
			tmpl.Execute(w, map[string]interface{}{
				"NoteName": noteName,
			})
			return
		}
		// Token is correct, extract actual content
		rawContent = getNoteContent(rawContent)
	}

	content := rawContent

	// 获取文件信息（大小和修改时间）
	var fileSize int64
	var modTime time.Time
	var createTime time.Time
	notePath := getNotePath(noteName)

	if info, err := os.Stat(notePath); err == nil {
		fileSize = info.Size()
		modTime = info.ModTime()
		// 尝试获取创建时间（Windows 上可用，其他平台回退到修改时间）
		if ct, err := getFileCreationTime(notePath); err == nil {
			createTime = ct
		} else {
			// 如果无法获取创建时间，回退到修改时间
			createTime = modTime
		}
	}

	// Format size
	sizeStr := fmt.Sprintf("%d B", fileSize)
	if fileSize >= 1024 {
		sizeStr = fmt.Sprintf("%.2f KB", float64(fileSize)/1024.0)
	}

	// Render markdown to HTML
	htmlContent := blackfriday.Run([]byte(content))

	// Parse and execute template
	tmpl := template.Must(template.New("read").Parse(readPageHTML))
	tmpl.Execute(w, map[string]interface{}{
		"NoteName":   noteName,
		"Content":    template.HTML(htmlContent),
		"FileSize":   sizeStr,
		"ModTime":    modTime.Format("2006-01-02 15:04:05"),
		"CreateTime": createTime.Format("2006-01-02 15:04:05"),
	})
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	// Check token authentication
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try to get from Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// If no token provided, show login page
	if token == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(adminLoginHTML))
		return
	}

	// If token is invalid, show login page with error
	if token != adminToken || adminToken == "" {
		http.Redirect(w, r, "/admin?error=invalid", http.StatusFound)
		return
	}

	notes, err := getAllNotes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	backupNotes, err := getAllBackupNotes()
	if err != nil {
		log.Printf("Warning: Failed to get backup notes: %v", err)
		backupNotes = []Note{}
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"notes":       notes,
			"backupNotes": backupNotes,
		})
		return
	}

	// Calculate total size
	var totalSize int64
	for _, note := range notes {
		totalSize += note.Size
	}

	var backupTotalSize int64
	for _, note := range backupNotes {
		backupTotalSize += note.Size
	}

	// Get current total file size (including uploads)
	currentTotalSize, _ := getTotalFileSize()

	// Get current max total size
	maxTotalSizeLock.RLock()
	currentMaxTotalSize := maxTotalSize
	maxTotalSizeLock.RUnlock()

	// Get current max note count
	maxNoteCountLock.RLock()
	currentMaxNoteCount := maxNoteCount
	maxNoteCountLock.RUnlock()

	// Get current values for display
	currentMaxFileSizeMB := int(maxFileSize / (1024 * 1024))

	// Prepare template functions
	funcMap := template.FuncMap{
		"formatSize": func(size int64) string {
			if size < 1024 {
				return fmt.Sprintf("%d B", size)
			}
			if size < 1024*1024 {
				return fmt.Sprintf("%.2f KB", float64(size)/1024.0)
			}
			if size < 1024*1024*1024 {
				return fmt.Sprintf("%.2f MB", float64(size)/(1024.0*1024.0))
			}
			return fmt.Sprintf("%.2f GB", float64(size)/(1024.0*1024.0*1024.0))
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"preview": func(content string, maxLen int) string {
			if len(content) <= maxLen {
				return content
			}
			return content[:maxLen] + "..."
		},
	}

	tmpl := template.Must(template.New("admin").Funcs(funcMap).Parse(adminPageHTML))
	tmpl.Execute(w, map[string]interface{}{
		"Notes":            notes,
		"BackupNotes":      backupNotes,
		"TotalSize":        totalSize,
		"BackupTotalSize":  backupTotalSize,
		"TotalCount":       len(notes),
		"BackupCount":      len(backupNotes),
		"CurrentTotalSize": currentTotalSize,
		"MaxTotalSize":     currentMaxTotalSize,
		"MaxNoteCount":     currentMaxNoteCount,
		"AdminPath":        adminPath,
		"NoteNameLen":      noteNameLen,
		"BackupDays":       backupDays,
		"NoteChars":        noteChars,
		"MaxFileSize":      maxFileSize,
		"MaxFileSizeMB":    currentMaxFileSizeMB,
		"MaxPathLength":    maxPathLength,
		"AccessToken":      accessToken,
		"Token":            token,
	})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteName := vars["note"]

	// Only check if unsafe (allow any characters for user input)
	if !isSafeNoteName(noteName) {
		http.Error(w, "Invalid note name", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	clientsLock.Lock()
	if clients[noteName] == nil {
		clients[noteName] = make(map[*websocket.Conn]bool)
	}
	clients[noteName][conn] = true
	clientsLock.Unlock()

	// Unregister on disconnect
	defer func() {
		clientsLock.Lock()
		delete(clients[noteName], conn)
		if len(clients[noteName]) == 0 {
			delete(clients, noteName)
		}
		clientsLock.Unlock()
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func broadcastUpdate(noteName, content string) {
	clientsLock.RLock()
	defer clientsLock.RUnlock()

	if clients[noteName] == nil {
		return
	}

	message := map[string]interface{}{
		"type":    "update",
		"content": content,
	}

	for conn := range clients[noteName] {
		if err := conn.WriteJSON(message); err != nil {
			delete(clients[noteName], conn)
			conn.Close()
		}
	}
}

func handleMarkdownRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	content := string(body)

	output := blackfriday.Run([]byte(content))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(output)
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check access token for file uploads
	if accessToken != "" {
		token := r.URL.Query().Get("token")
		if token == "" {
			// Try to get from Authorization header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if token != accessToken {
			http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
			return
		}
	}

	// Parse multipart form (max 100MB)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file size
	if handler.Size > maxFileSize {
		http.Error(w, fmt.Sprintf("File size exceeds maximum limit of %d MB", maxFileSize/(1024*1024)), http.StatusRequestEntityTooLarge)
		return
	}

	// Check total file size limit
	maxTotalSizeLock.RLock()
	currentMaxTotalSize := maxTotalSize
	maxTotalSizeLock.RUnlock()

	currentTotalSize, err := getTotalFileSize()
	if err != nil {
		log.Printf("Error calculating total file size: %v", err)
	} else {
		newTotalSize := currentTotalSize + handler.Size
		if newTotalSize > currentMaxTotalSize {
			http.Error(w, fmt.Sprintf("Total file size would exceed maximum limit of %d MB", currentMaxTotalSize/(1024*1024)), http.StatusRequestEntityTooLarge)
			return
		}
	}

	// Generate unique filename
	filename := handler.Filename
	// Sanitize filename
	filename = filepath.Base(filename)
	if filename == "" || filename == "." || filename == ".." {
		filename = "upload_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Check if file already exists, add suffix if needed
	uploadFilePath := filepath.Join(uploadPath, filename)
	counter := 1
	originalFilename := filename
	for {
		if _, err := os.Stat(uploadFilePath); os.IsNotExist(err) {
			break
		}
		ext := filepath.Ext(originalFilename)
		name := strings.TrimSuffix(originalFilename, ext)
		filename = fmt.Sprintf("%s_%d%s", name, counter, ext)
		uploadFilePath = filepath.Join(uploadPath, filename)
		counter++
	}

	// Create file
	dst, err := os.Create(uploadFilePath)
	if err != nil {
		http.Error(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine if it's an image
	isImage := false
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, imgExt := range imageExts {
		if ext == imgExt {
			isImage = true
			break
		}
	}

	// Return markdown format
	fileURL := "/uploads/" + filename
	var markdown string
	if isImage {
		markdown = fmt.Sprintf("![%s](%s)", filename, fileURL)
	} else {
		markdown = fmt.Sprintf("[下载 %s](%s)", filename, fileURL)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"filename": filename,
		"url":      fileURL,
		"markdown": markdown,
	})
}

func handleFileDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Security check: prevent path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadPath, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Check if it's an image - serve directly, otherwise force download
	isImage := false
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, imgExt := range imageExts {
		if ext == imgExt {
			isImage = true
			break
		}
	}

	if !isImage {
		// Force download for non-image files
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}
	http.ServeFile(w, r, filePath)
}

func handleUpdateMaxTotalSize(w http.ResponseWriter, r *http.Request) {
	// Check token authentication
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token != adminToken || adminToken == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AccessToken   *string `json:"accessToken,omitempty"`
		AdminPath     *string `json:"adminPath,omitempty"`
		NoteNameLen   *int    `json:"noteNameLen,omitempty"`
		BackupDays    *int    `json:"backupDays,omitempty"`
		NoteChars     *string `json:"noteChars,omitempty"`
		MaxFileSize   *string `json:"maxFileSize,omitempty"`
		MaxPathLength *int    `json:"maxPathLength,omitempty"`
		MaxTotalSize  *string `json:"maxTotalSize,omitempty"`
		MaxNoteCount  *int    `json:"maxNoteCount,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated := false

	// Update access token if provided
	if req.AccessToken != nil {
		accessToken = *req.AccessToken
		updated = true
	}

	// Update admin path if provided
	if req.AdminPath != nil && *req.AdminPath != "" {
		newPath := *req.AdminPath
		if !strings.HasPrefix(newPath, "/") {
			newPath = "/" + newPath
		}
		adminPath = newPath
		updated = true
	}

	// Update note name length if provided
	if req.NoteNameLen != nil && *req.NoteNameLen > 0 {
		noteNameLen = *req.NoteNameLen
		updated = true
	}

	// Update backup days if provided
	if req.BackupDays != nil && *req.BackupDays > 0 {
		backupDays = *req.BackupDays
		updated = true
	}

	// Update note chars if provided
	if req.NoteChars != nil && *req.NoteChars != "" {
		noteChars = *req.NoteChars
		updated = true
	}

	// Update max file size if provided
	if req.MaxFileSize != nil && *req.MaxFileSize != "" {
		size, err := parseFileSize(*req.MaxFileSize)
		if err != nil || size <= 0 {
			http.Error(w, fmt.Sprintf("Invalid maxFileSize format: %s", *req.MaxFileSize), http.StatusBadRequest)
			return
		}
		maxFileSize = size
		updated = true
	}

	// Update max path length if provided
	if req.MaxPathLength != nil && *req.MaxPathLength > 0 {
		maxPathLength = *req.MaxPathLength
		updated = true
	}

	// Update max total size if provided
	if req.MaxTotalSize != nil && *req.MaxTotalSize != "" {
		size, err := parseFileSize(*req.MaxTotalSize)
		if err != nil || size <= 0 {
			http.Error(w, fmt.Sprintf("Invalid maxTotalSize format: %s", *req.MaxTotalSize), http.StatusBadRequest)
			return
		}

		maxTotalSizeLock.Lock()
		maxTotalSize = size
		maxTotalSizeLock.Unlock()
		updated = true
	}

	// Update max note count if provided
	if req.MaxNoteCount != nil && *req.MaxNoteCount > 0 {
		maxNoteCountLock.Lock()
		maxNoteCount = *req.MaxNoteCount
		maxNoteCountLock.Unlock()
		updated = true
	}

	// Save config to file
	if updated {
		saveConfig()
	}

	maxTotalSizeLock.RLock()
	currentMaxTotalSize := maxTotalSize
	maxTotalSizeLock.RUnlock()

	maxNoteCountLock.RLock()
	currentMaxNoteCount := maxNoteCount
	maxNoteCountLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"adminPath":      adminPath,
		"noteNameLen":    noteNameLen,
		"backupDays":     backupDays,
		"noteChars":      noteChars,
		"maxFileSize":    maxFileSize,
		"maxPathLength":  maxPathLength,
		"maxTotalSize":   currentMaxTotalSize,
		"maxTotalSizeMB": currentMaxTotalSize / (1024 * 1024),
		"maxNoteCount":   currentMaxNoteCount,
	})
}

func main() {
	// Load configuration from command line, environment variable, or .env file
	tokenFlag := flag.String("token", "", "Admin access token (required)")
	portFlag := flag.String("port", "", "Server port (default: :8080)")
	flag.Parse()

	// Get port from: command line > environment variable > default (port is always configurable)
	if *portFlag != "" {
		port = *portFlag
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
	}

	// Check if config file was loaded
	// If config file exists, use it and ignore env/command line (except port and token)
	// If config file doesn't exist, use env/command line and save to config file
	if !configLoaded {
		// Config file doesn't exist, load from env/command line
		// Try to load from .env file first
		if err := loadEnvFile(); err != nil {
			log.Printf("Warning: Failed to load .env file: %v", err)
		}

		// Get admin path from: command line > environment variable > default
		adminPathFlag := flag.String("admin-path", "", "Admin panel path (default: /admin)")
		if *adminPathFlag != "" {
			adminPath = *adminPathFlag
		} else if envPath := os.Getenv("ADMIN_PATH"); envPath != "" {
			adminPath = envPath
		}
		// Ensure admin path starts with /
		if !strings.HasPrefix(adminPath, "/") {
			adminPath = "/" + adminPath
		}

		// Get note name length from: command line > environment variable > default
		noteNameLenFlag := flag.Int("note-name-len", 0, "Minimum length for note names (default: 3)")
		if *noteNameLenFlag > 0 {
			noteNameLen = *noteNameLenFlag
		} else if envLen := os.Getenv("NOTE_NAME_LEN"); envLen != "" {
			if len, err := strconv.Atoi(envLen); err == nil && len > 0 {
				noteNameLen = len
			}
		}

		// Get backup days from: command line > environment variable > default
		backupDaysFlag := flag.Int("backup-days", 0, "Days before moving notes to backup (default: 7)")
		if *backupDaysFlag > 0 {
			backupDays = *backupDaysFlag
		} else if envDays := os.Getenv("BACKUP_DAYS"); envDays != "" {
			if days, err := strconv.Atoi(envDays); err == nil && days > 0 {
				backupDays = days
			}
		}

		// Get note characters from: command line > environment variable > default
		noteCharsFlag := flag.String("note-chars", "", "Characters used for generating note names (default: 0123456789abcdefghijklmnopqrstuvwxyz)")
		if *noteCharsFlag != "" {
			noteChars = *noteCharsFlag
		} else if envChars := os.Getenv("NOTE_CHARS"); envChars != "" {
			noteChars = envChars
		}

		// Validate noteChars is not empty
		if len(noteChars) == 0 {
			log.Fatal("Error: NOTE_CHARS cannot be empty")
		}

		// Get max file size from: command line > environment variable > default
		maxFileSizeFlag := flag.String("max-file-size", "", "Maximum file size (e.g., 10M, 100MB, default: 10MB)")
		if *maxFileSizeFlag != "" {
			if size, err := parseFileSize(*maxFileSizeFlag); err == nil && size > 0 {
				maxFileSize = size
			} else {
				log.Fatalf("Error: Invalid max-file-size format: %s. Use format like 10M, 100MB, 1G", *maxFileSizeFlag)
			}
		} else if envSize := os.Getenv("MAX_FILE_SIZE"); envSize != "" {
			if size, err := parseFileSize(envSize); err == nil && size > 0 {
				maxFileSize = size
			} else {
				log.Fatalf("Error: Invalid MAX_FILE_SIZE format: %s. Use format like 10M, 100MB, 1G", envSize)
			}
		}

		// Get max path length from: command line > environment variable > default
		maxPathLengthFlag := flag.Int("max-path-length", 0, "Maximum path/note name length (default: 20)")
		if *maxPathLengthFlag > 0 {
			maxPathLength = *maxPathLengthFlag
		} else if envPathLen := os.Getenv("MAX_PATH_LENGTH"); envPathLen != "" {
			if pathLen, err := strconv.Atoi(envPathLen); err == nil && pathLen > 0 {
				maxPathLength = pathLen
			}
		}

		// Save config to file after loading from env/command line
		saveConfig()
		log.Printf("Configuration loaded from environment/command line and saved to %s", configFile)
	} else {
		log.Printf("Configuration loaded from %s (environment variables and command line arguments ignored)", configFile)
	}

	// Get token from: command line > environment variable > config file > .env file
	// Command line and env always take precedence (for security)
	if *tokenFlag != "" {
		adminToken = *tokenFlag
	} else if envToken := os.Getenv("ADMIN_TOKEN"); envToken != "" {
		adminToken = envToken
	} else if configLoaded && adminToken == "" {
		// If config was loaded but token is still empty, it means config file didn't have token
		// Try to load from .env file
		if err := loadEnvFile(); err == nil {
			if envToken := os.Getenv("ADMIN_TOKEN"); envToken != "" {
				adminToken = envToken
			}
		}
	}

	// Token is required
	if adminToken == "" {
		log.Fatal("Error: Admin token is required. Set it via -token flag, ADMIN_TOKEN environment variable, ADMIN_TOKEN in .env file, or adminToken in config.json")
	}

	// Get access token from: environment variable > config file > .env file
	// Access token is optional - if not set, no authentication required for notes
	if !configLoaded {
		// Try to load from .env file first
		if err := loadEnvFile(); err == nil {
			if envAccessToken := os.Getenv("ACCESS_TOKEN"); envAccessToken != "" {
				accessToken = envAccessToken
			}
		}
	}
	// If config was loaded, accessToken should already be set from config file
	// But we can still override with environment variable
	if envAccessToken := os.Getenv("ACCESS_TOKEN"); envAccessToken != "" {
		accessToken = envAccessToken
	}

	// Load existing notes into memory cache
	if err := loadExistingNotes(); err != nil {
		log.Printf("Warning: Failed to load existing notes: %v", err)
	} else {
		var count int
		existingNotes.Range(func(key, value interface{}) bool {
			count++
			return true
		})
		log.Printf("Loaded %d existing notes into memory cache", count)
	}

	r := mux.NewRouter()

	// Admin routes (must be before /{note} route)
	r.HandleFunc(adminPath, handleAdmin).Methods("GET")

	// Read-only route (must be before /{note} route)
	r.HandleFunc("/read/{note}", handleReadNote).Methods("GET")

	// WebSocket route
	r.HandleFunc("/ws/{note}", handleWebSocket)

	// Markdown render route
	r.HandleFunc("/api/markdown", handleMarkdownRender).Methods("POST")

	// File upload route
	r.HandleFunc("/api/upload", handleFileUpload).Methods("POST")

	// File download route
	r.HandleFunc("/uploads/{filename}", handleFileDownload).Methods("GET")

	// Update max total size route (admin only)
	r.HandleFunc("/api/max-total-size", handleUpdateMaxTotalSize).Methods("POST")

	// Static file server for uploads
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadPath))))

	// Note routes (must be after specific routes)
	r.HandleFunc("/{note}", handleNote).Methods("GET", "POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check access token if set (only for browser requests, not curl/wget)
		if accessToken != "" && !strings.HasPrefix(r.UserAgent(), "curl") && !strings.HasPrefix(r.UserAgent(), "Wget") {
			token := r.URL.Query().Get("token")
			if token == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}
			if token != accessToken {
				// Show login page
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(accessLoginHTML))
				return
			}
		}
		http.Redirect(w, r, "/"+generateNoteName(), http.StatusFound)
	}).Methods("GET")

	fmt.Printf("Server starting on http://localhost%s\n", port)
	fmt.Printf("Admin panel: http://localhost%s%s\n", port, adminPath)
	log.Fatal(http.ListenAndServe(port, r))
}
