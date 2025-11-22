package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Session 存储 session 信息
type Session struct {
	Token     string
	ExpiresAt time.Time
}

var (
	// adminSessions 存储管理员 session（token -> 过期时间）
	adminSessions = sync.Map{}
	// sessionCleanupInterval session 清理间隔
	sessionCleanupInterval = 30 * time.Minute
	// sessionExpiry session 过期时间（2小时）
	sessionExpiry = 2 * time.Hour
)

func init() {
	// 启动定期清理过期 session 的 goroutine
	go cleanupExpiredSessions()
}

// generateSessionToken 生成随机的 session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// createAdminSession 创建管理员 session
func createAdminSession() (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}
	
	session := &Session{
		Token:     token,
		ExpiresAt: time.Now().Add(sessionExpiry),
	}
	
	adminSessions.Store(token, session)
	return token, nil
}

// validateAdminSession 验证管理员 session token
func validateAdminSession(token string) bool {
	if token == "" {
		return false
	}
	
	value, ok := adminSessions.Load(token)
	if !ok {
		return false
	}
	
	session, ok := value.(*Session)
	if !ok {
		return false
	}
	
	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		adminSessions.Delete(token)
		return false
	}
	
	return true
}

// deleteAdminSession 删除管理员 session
func deleteAdminSession(token string) {
	adminSessions.Delete(token)
}

// cleanupExpiredSessions 定期清理过期的 session
func cleanupExpiredSessions() {
	ticker := time.NewTicker(sessionCleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		adminSessions.Range(func(key, value interface{}) bool {
			session, ok := value.(*Session)
			if !ok {
				adminSessions.Delete(key)
				return true
			}
			
			if now.After(session.ExpiresAt) {
				adminSessions.Delete(key)
			}
			return true
		})
	}
}

