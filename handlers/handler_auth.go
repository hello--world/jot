package handlers

import (
	"net/http"
	"strings"
)

// GetTokenFromRequest 从请求中提取 token（优先级：cookie > query 参数 > Authorization header）
func GetTokenFromRequest(r *http.Request) string {
	// First try cookie
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return strings.TrimSpace(cookie.Value)
	}
	// Then try query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else if authHeader != "" {
			// 如果没有 "Bearer " 前缀，直接使用整个 header 值
			token = authHeader
		}
	}
	// 去除首尾空格
	token = strings.TrimSpace(token)
	return token
}

// GetLockTokenFromRequest 从请求中提取锁 token（从 query 参数、cookie 或 Authorization header）
func GetLockTokenFromRequest(r *http.Request, noteName string) string {
	token := r.URL.Query().Get("lock_token")
	if token == "" {
		// Try to get from cookie
		cookie, err := r.Cookie("note_lock_" + noteName)
		if err == nil {
			token = cookie.Value
		}
	}
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	// 去除首尾空格
	return strings.TrimSpace(token)
}
