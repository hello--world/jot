package router

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/hello--world/jot/handlers"
	"github.com/hello--world/jot/htmlPage"
)

// requireAccessToken 中间件：如果站点有 access token，需要验证
func requireAccessToken(next http.Handler, getAccessToken func() string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := getAccessToken()
		if accessToken != "" {
			token := r.URL.Query().Get("token")
			if token == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}
			token = strings.TrimSpace(token)
			if token != accessToken {
				http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RouterConfig 路由配置
type RouterConfig struct {
	AdminPath        string
	UploadPath       string
	HandleWebSocket  func(http.ResponseWriter, *http.Request)
	GenerateNoteName func() string
	GetAccessToken   func() string
}

var config *RouterConfig

// InitRouter 初始化路由配置
func InitRouter(c *RouterConfig) {
	config = c
}

// SetupRoutes 设置所有路由
func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Admin routes (must be before /{note} route)
	r.HandleFunc(config.AdminPath, handlers.HandleAdmin).Methods("GET")

	// Read-only route (must be before /{note} route)
	r.HandleFunc("/read/{note}", handlers.HandleReadNote).Methods("GET")

	// WebSocket route
	r.HandleFunc("/ws/{note}", config.HandleWebSocket)

	// Markdown render route
	r.HandleFunc("/api/markdown", handlers.HandleMarkdownRender).Methods("POST")

	// File upload route
	r.HandleFunc("/api/upload", handlers.HandleFileUpload).Methods("POST")

	// File download route with date directory: /uploads/{date}/{filename}
	r.HandleFunc("/uploads/{date}/{filename}", handlers.HandleFileDownload).Methods("GET")

	// Update max total size route (admin only)
	r.HandleFunc("/api/max-total-size", handlers.HandleUpdateMaxTotalSize).Methods("POST")

	// Static file server for uploads (需要 access token 验证)
	// Support both old format (without date) and new format (with date)
	uploadsHandler := http.StripPrefix("/uploads/", http.FileServer(http.Dir(config.UploadPath)))
	r.PathPrefix("/uploads/").Handler(requireAccessToken(uploadsHandler, config.GetAccessToken))

	// Note routes (must be after specific routes)
	r.HandleFunc("/{note}", handlers.HandleNote).Methods("GET", "POST")
	r.HandleFunc("/", handleRoot).Methods("GET")

	return r
}

// handleRoot 处理根路径请求
func handleRoot(w http.ResponseWriter, r *http.Request) {
	// Check access token if set (only for browser requests, not curl/wget)
	accessToken := config.GetAccessToken()
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
			w.Write([]byte(htmlPage.AccessLoginHTML))
			return
		}
		// If token is valid, set cookie and redirect without token in URL
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    token,
			Path:     "/",
			MaxAge:   86400 * 30, // 30 days
			HttpOnly: false,      // Allow JavaScript to read it
			SameSite: http.SameSiteStrictMode,
		})
		noteName := config.GenerateNoteName()
		http.Redirect(w, r, "/"+noteName, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/"+config.GenerateNoteName(), http.StatusFound)
}
