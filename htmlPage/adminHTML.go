package htmlPage

const AdminLoginHTML = `<!DOCTYPE html>
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
    <form class="login-form" id="loginForm">
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
    e.preventDefault();
    const token = tokenInput.value.trim();
    if (!token) {
        errorMessage.textContent = '请输入访问令牌';
        errorMessage.classList.add('show');
        tokenInput.focus();
        return false;
    }
    
    // 隐藏错误信息
    errorMessage.classList.remove('show');
    
    // 通过 fetch 发送请求，token 放在 Authorization header 中
    fetch(window.location.pathname, {
        method: 'GET',
        headers: {
            'Authorization': 'Bearer ' + token
        },
        credentials: 'include', // 包含 cookies
        redirect: 'follow' // 跟随重定向
    })
    .then(response => {
        // fetch 会自动跟随重定向，如果最终返回 200，说明登录成功
        if (response.ok) {
            // 登录成功，刷新页面（现在有 session cookie 了）
            window.location.href = window.location.pathname;
        } else {
            // 登录失败
            errorMessage.textContent = '令牌无效，请重试';
            errorMessage.classList.add('show');
            tokenInput.focus();
        }
    })
    .catch(err => {
        console.error('Login error:', err);
        errorMessage.textContent = '登录失败，请重试';
        errorMessage.classList.add('show');
    });
    
    return false;
});
</script>
</body>
</html>`

const AdminPageHTML = `<!DOCTYPE html>
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
    padding: 10px;
}
.container {
    max-width: 1200px;
    margin: 0 auto;
    background: #fff;
    border-radius: 6px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    overflow: hidden;
}
.header {
    background: #0066cc;
    color: white;
    padding: 12px 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.header h1 {
    font-size: 20px;
    font-weight: 500;
}
.header a {
    color: white;
    text-decoration: none;
    padding: 6px 12px;
    background: rgba(255,255,255,0.2);
    border-radius: 4px;
    transition: background 0.2s;
    font-size: 13px;
}
.header a:hover {
    background: rgba(255,255,255,0.3);
}
.stats {
    padding: 12px 16px;
    background: #f5f5f5;
    border-bottom: 1px solid #ddd;
    display: flex;
    gap: 20px;
    flex-wrap: wrap;
}
.stat-item {
    display: flex;
    flex-direction: column;
}
.stat-label {
    font-size: 11px;
    color: #666;
    margin-bottom: 2px;
}
.stat-value {
    font-size: 18px;
    font-weight: 600;
    color: #333;
}
.notes-list {
    padding: 12px 16px;
}
.notes-table {
    width: 100%;
    border-collapse: collapse;
}
.notes-table th {
    background: #f5f5f5;
    padding: 8px 10px;
    text-align: left;
    font-weight: 600;
    color: #333;
    border-bottom: 2px solid #ddd;
    font-size: 13px;
}
.notes-table td {
    padding: 8px 10px;
    border-bottom: 1px solid #eee;
    font-size: 13px;
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
    gap: 8px;
    padding: 10px 16px;
    background: #f5f5f5;
    border-bottom: 1px solid #ddd;
}
.tab-button {
    padding: 6px 14px;
    background: #fff;
    border: 1px solid #ddd;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
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
.sub-tabs {
    display: flex;
    gap: 8px;
    padding: 10px 16px;
    background: #f9f9f9;
    border-bottom: 1px solid #ddd;
}
.sub-tab-button {
    padding: 6px 14px;
    background: #fff;
    border: 1px solid #ddd;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    color: #666;
    transition: all 0.2s;
}
.sub-tab-button:hover {
    background: #f0f0f0;
}
.sub-tab-button.active {
    background: #0066cc;
    color: white;
    border-color: #0066cc;
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
        <button class="tab-button active" onclick="showTab('active')">📝 活跃笔记 ({{.TotalCount}})</button>
        <button class="tab-button" onclick="showTab('backup')">📦 备份笔记 ({{.BackupCount}})</button>
        <button class="tab-button" onclick="showTab('settings')">⚙️ 系统设置</button>
    </div>
    <div id="active-tab" class="tab-content">
    <div class="notes-list">
        <div id="active-notes">
            {{if .GroupedNotes}}
            {{range .GroupedNotes}}
            <div class="date-group" data-date="{{.Date}}">
                <div style="margin: 10px 16px; padding: 8px 12px; background: #f0f0f0; border-left: 4px solid #0066cc; display: flex; align-items: center; justify-content: space-between; gap: 10px;">
                    <h3 style="margin: 0; font-size: 14px; color: #333; font-weight: 600;">📅 {{.Date}} ({{len .Notes}} 条笔记)</h3>
                    <select class="date-filter-select" onchange="filterByDate()" style="padding: 4px 8px; border: 1px solid #ddd; border-radius: 3px; font-size: 12px; background: white; cursor: pointer;">
                        <option value="">全部日期</option>
                        {{range $.DateList}}
                        <option value="{{.}}" {{if eq . $.Date}}selected{{end}}>{{.}}</option>
                        {{end}}
                    </select>
                </div>
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
            </div>
            {{end}}
            {{else}}
            <div class="empty">
                <div class="empty-icon">📄</div>
                <p>还没有笔记，<a href="/" style="color: #0066cc;">创建第一个笔记</a></p>
            </div>
            {{end}}
        </div>
    </div>
    </div>
    <div id="backup-tab" class="tab-content" style="display: none;">
    <div class="notes-list">
        <div id="backup-notes">
            {{if .GroupedBackupNotes}}
            {{range .GroupedBackupNotes}}
            <div class="date-group" data-date="{{.Date}}">
                <div style="margin: 10px 16px; padding: 8px 12px; background: #f0f0f0; border-left: 4px solid #ff9800; display: flex; align-items: center; justify-content: space-between; gap: 10px;">
                    <h3 style="margin: 0; font-size: 14px; color: #333; font-weight: 600;">📅 {{.Date}} ({{len .Notes}} 条笔记)</h3>
                    <select class="date-filter-select" onchange="filterByDate()" style="padding: 4px 8px; border: 1px solid #ddd; border-radius: 3px; font-size: 12px; background: white; cursor: pointer;">
                        <option value="">全部日期</option>
                        {{range $.DateList}}
                        <option value="{{.}}" {{if eq . $.Date}}selected{{end}}>{{.}}</option>
                        {{end}}
                    </select>
                </div>
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
                            <td><a href="/read/{{.Name}}" class="note-name">{{.Name}}</a></td>
                            <td class="note-content" title="{{.Content}}">{{if .Content}}{{preview .Content 50}}{{else}}<em>空笔记</em>{{end}}</td>
                            <td class="note-size">{{formatSize .Size}}</td>
                            <td class="note-date">{{formatDate .UpdatedAt}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            {{end}}
            {{else}}
            <div class="empty">
                <div class="empty-icon">📦</div>
                <p>还没有备份笔记</p>
            </div>
            {{end}}
        </div>
    </div>
    </div>
    <div id="settings-tab" class="tab-content" style="display: none;">
    <div class="stats" style="margin-bottom: 0;">
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
            <div style="margin-top: 4px; display: flex; gap: 6px; align-items: center;">
                <input type="text" id="max-total-size-input" placeholder="如: 500MB" style="padding: 3px 6px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px; width: 100px;">
                <button onclick="updateMaxTotalSize()" style="padding: 3px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
            </div>
        </div>
        <div class="stat-item">
            <span class="stat-label">最大笔记数量限制</span>
            <span class="stat-value" id="max-note-count">{{.MaxNoteCount}}</span>
            <div style="margin-top: 4px; display: flex; gap: 6px; align-items: center;">
                <input type="number" id="max-note-count-input" placeholder="如: 500" min="1" style="padding: 3px 6px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px; width: 100px;">
                <button onclick="updateConfig('maxNoteCount')" style="padding: 3px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
            </div>
        </div>
    </div>
    <div style="padding: 12px 16px; background: #f9f9f9; border-top: 1px solid #ddd;">
        <h3 style="margin-bottom: 10px; font-size: 14px; color: #333; font-weight: 600;">配置管理</h3>
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 10px;">
            <div style="background: white; padding: 10px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 4px; font-size: 11px; color: #666;">访问令牌（用于访问笔记）</label>
                <div style="display: flex; gap: 6px;">
                    <input type="text" id="access-token-input" value="{{.AccessToken}}" placeholder="留空表示无需授权" style="flex: 1; padding: 5px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px;">
                    <button onclick="updateConfig('accessToken')" style="padding: 5px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
                </div>
                <div style="margin-top: 3px; font-size: 10px; color: #999;">留空表示无需授权即可访问笔记</div>
            </div>
            <div style="background: white; padding: 10px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 4px; font-size: 11px; color: #666;">管理后台路径</label>
                <div style="display: flex; gap: 6px;">
                    <input type="text" id="admin-path-input" value="{{.AdminPath}}" style="flex: 1; padding: 5px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px;">
                    <button onclick="updateConfig('adminPath')" style="padding: 5px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
                </div>
            </div>
            <div style="background: white; padding: 10px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 4px; font-size: 11px; color: #666;">笔记名称最小长度</label>
                <div style="display: flex; gap: 6px;">
                    <input type="number" id="note-name-len-input" value="{{.NoteNameLen}}" min="1" style="flex: 1; padding: 5px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px;">
                    <button onclick="updateConfig('noteNameLen')" style="padding: 5px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
                </div>
            </div>
            <div style="background: white; padding: 10px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 4px; font-size: 11px; color: #666;">备份天数</label>
                <div style="display: flex; gap: 6px;">
                    <input type="number" id="backup-days-input" value="{{.BackupDays}}" min="1" style="flex: 1; padding: 5px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px;">
                    <button onclick="updateConfig('backupDays')" style="padding: 5px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
                </div>
            </div>
            <div style="background: white; padding: 10px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 4px; font-size: 11px; color: #666;">随机字符串字符集</label>
                <div style="display: flex; gap: 6px;">
                    <input type="text" id="note-chars-input" value="{{.NoteChars}}" style="flex: 1; padding: 5px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px;">
                    <button onclick="updateConfig('noteChars')" style="padding: 5px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
                </div>
            </div>
            <div style="background: white; padding: 10px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 4px; font-size: 11px; color: #666;">最大文件大小</label>
                <div style="display: flex; gap: 6px;">
                    <input type="text" id="max-file-size-input" placeholder="如: 10MB" style="flex: 1; padding: 5px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px;">
                    <button onclick="updateConfig('maxFileSize')" style="padding: 5px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
                </div>
                <div style="margin-top: 3px; font-size: 10px; color: #999;">当前: {{.MaxFileSizeMB}} MB</div>
            </div>
            <div style="background: white; padding: 10px; border-radius: 4px; border: 1px solid #ddd;">
                <label style="display: block; margin-bottom: 4px; font-size: 11px; color: #666;">最大路径长度</label>
                <div style="display: flex; gap: 6px;">
                    <input type="number" id="max-path-length-input" value="{{.MaxPathLength}}" min="1" style="flex: 1; padding: 5px; border: 1px solid #ddd; border-radius: 3px; font-size: 11px;">
                    <button onclick="updateConfig('maxPathLength')" style="padding: 5px 10px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 11px;">更新</button>
                </div>
            </div>
        </div>
    </div>
    </div>
</div>
<script>
// Session token is stored in HttpOnly cookie, not accessible from JavaScript
// All requests will automatically include the cookie


function showTab(tabName) {
    // Hide all tab contents
    document.getElementById('active-tab').style.display = 'none';
    document.getElementById('backup-tab').style.display = 'none';
    document.getElementById('settings-tab').style.display = 'none';
    
    // Remove active class from all buttons
    document.querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected tab
    if (tabName === 'active') {
        document.getElementById('active-tab').style.display = 'block';
        document.querySelector('.tab-button:first-child').classList.add('active');
    } else if (tabName === 'backup') {
        document.getElementById('backup-tab').style.display = 'block';
        document.querySelector('.tab-button:nth-child(2)').classList.add('active');
    } else if (tabName === 'settings') {
        document.getElementById('settings-tab').style.display = 'block';
        document.querySelector('.tab-button:last-child').classList.add('active');
    }
    
    // 重新应用日期过滤
    filterByDate();
}

function filterByDate() {
    // 获取所有日期过滤器
    const filters = document.querySelectorAll('.date-filter-select');
    if (filters.length === 0) return;
    
    // 使用第一个过滤器的值（所有过滤器应该同步）
    const selectedDate = filters[0].value;
    
    // 同步所有过滤器的值
    filters.forEach(filter => {
        if (filter.value !== selectedDate) {
            filter.value = selectedDate;
        }
    });
    
    // 过滤活跃笔记
    const activeGroups = document.querySelectorAll('#active-notes .date-group');
    activeGroups.forEach(group => {
        if (!selectedDate || group.getAttribute('data-date') === selectedDate) {
            group.style.display = 'block';
        } else {
            group.style.display = 'none';
        }
    });
    
    // 过滤备份笔记
    const backupGroups = document.querySelectorAll('#backup-notes .date-group');
    backupGroups.forEach(group => {
        if (!selectedDate || group.getAttribute('data-date') === selectedDate) {
            group.style.display = 'block';
        } else {
            group.style.display = 'none';
        }
    });
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
    // Session token is in HttpOnly cookie, automatically sent with request
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

    fetch('/api/max-total-size', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: 'include', // Include cookies
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
