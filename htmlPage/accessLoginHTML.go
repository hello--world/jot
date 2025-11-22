package htmlPage

const AccessLoginHTML = `<!DOCTYPE html>
<html>
	、<head>
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

// Check for saved token in localStorage
const ACCESS_TOKEN_KEY = 'jot_access_token';
const savedToken = localStorage.getItem(ACCESS_TOKEN_KEY);

// Check if there's a token in URL (from form submission or direct access)
const urlParams = new URLSearchParams(window.location.search);
const urlToken = urlParams.get('token');

// If there's a token in URL, save it and redirect
if (urlToken) {
    localStorage.setItem(ACCESS_TOKEN_KEY, urlToken);
    // Redirect to new note with token
    window.location.href = '/?token=' + encodeURIComponent(urlToken);
    // Stop execution
    throw new Error('Redirecting...');
}

// If there's a saved token and no error, try to use it automatically
if (savedToken && urlParams.get('error') !== 'invalid') {
    // Auto-login with saved token
    window.location.href = '/?token=' + encodeURIComponent(savedToken);
    // Stop execution
    throw new Error('Auto-login...');
}

// Set form action to current path
const currentPath = window.location.pathname;
const currentSearch = window.location.search;
form.action = currentPath + (currentSearch ? currentSearch + '&' : '?') + 'token=';

// Check if there's an error in URL
if (urlParams.get('error') === 'invalid') {
    errorMessage.textContent = '令牌无效，请重试';
    errorMessage.classList.add('show');
    // Clear invalid token from localStorage
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    tokenInput.focus();
} else if (savedToken) {
    // Pre-fill the input with saved token (user can still change it)
    tokenInput.value = savedToken;
    tokenInput.type = 'text'; // Show saved token
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
    // Save token to localStorage before redirecting
    localStorage.setItem(ACCESS_TOKEN_KEY, token);
    // Update form action with token
    form.action = currentPath + (currentSearch ? currentSearch + '&' : '?') + 'token=' + encodeURIComponent(token);
});
</script>
</body>
</html>`
