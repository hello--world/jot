package handlers

import (
	"io"
	"net/http"

	"github.com/russross/blackfriday/v2"
)

// HandleMarkdownRender 处理 Markdown 渲染请求
func HandleMarkdownRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查 access token（如果站点有 token，需要验证）
	if deps.AccessToken != "" {
		token := deps.GetTokenFromRequest(r)
		if token != deps.AccessToken {
			http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
			return
		}
	}

	body, _ := io.ReadAll(r.Body)
	content := string(body)

	output := blackfriday.Run([]byte(content))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(output)
}
