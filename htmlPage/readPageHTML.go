package htmlPage

const ReadPageHTML = `<!DOCTYPE html>
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

const NoteLockHTML = `<!DOCTYPE html>
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
