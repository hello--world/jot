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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/russross/blackfriday/v2"
)

const (
	savePath    = "_tmp"
	backupPath  = "bak"
	port        = ":8080"
	noteNameLen = 3
	backupDays  = 7 // Move notes to backup after 7 days of inactivity
)

var (
	adminPath = "/admin" // Can be set via ADMIN_PATH env var or -admin-path flag
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
        </div>
        <textarea id="editor" placeholder="开始输入 Markdown 内容...">{{.Content}}</textarea>
    </div>
    <div class="preview-panel">
        <div class="panel-header">
            <span id="connection-status">●</span>
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
    fetch('/api/markdown', {
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

    fetch(window.location.pathname, {
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
        connectionStatus.textContent = '●';
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
        connectionStatus.textContent = '●';
        connectionStatus.style.color = '#f44336';
    };
    
    ws.onclose = () => {
        connectionStatus.textContent = '●';
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

editor.focus();
updatePreview();
connectWebSocket();

// Auto-save every 2 seconds
setInterval(saveNote, 2000);
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
}

// loadEnvFile loads environment variables from .env file
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

func generateNoteName() string {
	// Use lowercase letters and numbers only (no uppercase to avoid input method switching)
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	name := make([]byte, noteNameLen)
	for i := range name {
		name[i] = chars[rand.Intn(len(chars))]
	}
	return string(name)
}

func isValidNoteName(name string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	return matched && len(name) <= 64
}

func getNotePath(name string) string {
	return filepath.Join(savePath, name)
}

func saveNote(name, content string) error {
	path := getNotePath(name)
	if content == "" {
		os.Remove(path)
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
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

func getAllNotes() ([]Note, error) {
	files, err := os.ReadDir(savePath)
	if err != nil {
		return nil, err
	}

	notes := make([]Note, 0)
	for _, file := range files {
		if !file.IsDir() && isValidNoteName(file.Name()) {
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

// moveOldNotesToBackup moves notes that haven't been modified for backupDays days to backup folder
// Backup folder structure: bak/YYYYMMDD/note_name
func moveOldNotesToBackup() error {
	files, err := os.ReadDir(savePath)
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -backupDays)
	movedCount := 0

	for _, file := range files {
		if file.IsDir() || !isValidNoteName(file.Name()) {
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

// getAllBackupNotes returns all notes from backup folders
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
			if file.IsDir() || !isValidNoteName(file.Name()) {
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

	if noteName == "" || !isValidNoteName(noteName) {
		http.Redirect(w, r, "/"+generateNoteName(), http.StatusFound)
		return
	}

	if r.Method == "GET" {
		// Check if raw or curl/wget
		if r.URL.Query().Get("raw") != "" || strings.HasPrefix(r.UserAgent(), "curl") || strings.HasPrefix(r.UserAgent(), "Wget") {
			content, err := loadNote(noteName)
			if err != nil || content == "" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(content))
			return
		}

		// Serve HTML page
		content, _ := loadNote(noteName)

		// Get file info for size and modification time
		var fileSize int64
		var modTime time.Time
		var createTime time.Time
		notePath := getNotePath(noteName)
		if info, err := os.Stat(notePath); err == nil {
			fileSize = info.Size()
			modTime = info.ModTime()
			// For creation time, use ModTime as fallback (creation time is platform-specific)
			createTime = modTime
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
		return
	}

	if r.Method == "POST" {
		body, _ := io.ReadAll(r.Body)
		content := string(body)

		// Handle form-encoded data
		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			r.ParseForm()
			if text := r.FormValue("text"); text != "" {
				content = text
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

	// Prepare template functions
	funcMap := template.FuncMap{
		"formatSize": func(size int64) string {
			if size < 1024 {
				return fmt.Sprintf("%d B", size)
			}
			return fmt.Sprintf("%.2f KB", float64(size)/1024.0)
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
		"Notes":           notes,
		"BackupNotes":     backupNotes,
		"TotalSize":       totalSize,
		"BackupTotalSize": backupTotalSize,
		"TotalCount":      len(notes),
		"BackupCount":     len(backupNotes),
	})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteName := vars["note"]

	if !isValidNoteName(noteName) {
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

func main() {
	// Load token and admin path from command line, environment variable, or .env file
	tokenFlag := flag.String("token", "", "Admin access token (required)")
	adminPathFlag := flag.String("admin-path", "", "Admin panel path (default: /admin)")
	flag.Parse()

	// Try to load from .env file first
	if err := loadEnvFile(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// Get admin path from: command line > environment variable > default
	if *adminPathFlag != "" {
		adminPath = *adminPathFlag
	} else if envPath := os.Getenv("ADMIN_PATH"); envPath != "" {
		adminPath = envPath
	}
	// Ensure admin path starts with /
	if !strings.HasPrefix(adminPath, "/") {
		adminPath = "/" + adminPath
	}

	// Get token from: command line > environment variable > .env file
	if *tokenFlag != "" {
		adminToken = *tokenFlag
	} else if envToken := os.Getenv("ADMIN_TOKEN"); envToken != "" {
		adminToken = envToken
	}

	// Token is required
	if adminToken == "" {
		log.Fatal("Error: Admin token is required. Set it via -token flag, ADMIN_TOKEN environment variable, or ADMIN_TOKEN in .env file")
	}

	r := mux.NewRouter()

	// Admin routes (must be before /{note} route)
	r.HandleFunc(adminPath, handleAdmin).Methods("GET")

	// WebSocket route
	r.HandleFunc("/ws/{note}", handleWebSocket)

	// Markdown render route
	r.HandleFunc("/api/markdown", handleMarkdownRender).Methods("POST")

	// Note routes (must be after specific routes)
	r.HandleFunc("/{note}", handleNote).Methods("GET", "POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/"+generateNoteName(), http.StatusFound)
	}).Methods("GET")

	fmt.Printf("Server starting on http://localhost%s\n", port)
	fmt.Printf("Admin panel: http://localhost%s%s\n", port, adminPath)
	log.Fatal(http.ListenAndServe(port, r))
}
