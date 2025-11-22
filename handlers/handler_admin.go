package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hello--world/jot/htmlPage"
)

// getAdminSessionTokenFromRequest 从请求中获取 admin session token（从 cookie）
func getAdminSessionTokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie("admin_session")
	if err == nil && cookie.Value != "" {
		return strings.TrimSpace(cookie.Value)
	}
	return ""
}

// HandleAdmin 处理管理后台请求
func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	// 首先检查是否有 session token（cookie）
	sessionToken := getAdminSessionTokenFromRequest(r)
	if sessionToken != "" {
		// 验证 session token
		if validateAdminSession(sessionToken) {
			// Session 有效，显示管理页面
			serveAdminPage(w, r, sessionToken)
			return
		}
		// Session 无效或过期，清除 cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})
	}

	// 如果没有 session token，检查是否是登录请求（使用原始 admin token）
	// Admin token 只从 Authorization header 中读取，不从 URL 或 cookie 中读取
	adminToken := ""
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		adminToken = strings.TrimPrefix(authHeader, "Bearer ")
	} else if authHeader != "" {
		adminToken = authHeader
	}
	adminToken = strings.TrimSpace(adminToken)

	// 如果没有提供 token，显示登录页面
	if adminToken == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlPage.AdminLoginHTML))
		return
	}

	// 验证原始 admin token
	if adminToken != deps.AdminToken || deps.AdminToken == "" {
		http.Redirect(w, r, deps.AdminPath+"?error=invalid", http.StatusFound)
		return
	}

	// Admin token 有效，创建 session token
	sessionToken, err := createAdminSession()
	if err != nil {
		log.Printf("Error creating admin session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 设置 session token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   int(sessionExpiry.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // 如果使用 HTTPS，可以设置为 true
	})

	// 重定向到不带 token 的 URL
	http.Redirect(w, r, deps.AdminPath, http.StatusFound)
}

// serveAdminPage 显示管理页面
func serveAdminPage(w http.ResponseWriter, r *http.Request, sessionToken string) {

	notes, err := deps.GetAllNotes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	backupNotes, err := deps.GetAllBackupNotes()
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
	currentTotalSize, _ := deps.GetTotalFileSize()

	// Get current max total size
	deps.RLockMaxTotalSize()
	currentMaxTotalSize := deps.GetMaxTotalSize()
	deps.RUnlockMaxTotalSize()

	// Get current max note count
	deps.RLockMaxNoteCount()
	currentMaxNoteCount := deps.GetMaxNoteCount()
	deps.RUnlockMaxNoteCount()

	// Get current values for display
	currentMaxFileSizeMB := int(deps.GetMaxFileSize() / (1024 * 1024))

	// 按日期分组笔记
	type NotesByDate struct {
		Date  string // 日期（格式：YYYY-MM-DD）
		Notes []Note
	}

	groupNotesByDate := func(notes []Note) []NotesByDate {
		dateMap := make(map[string][]Note)
		for _, note := range notes {
			dateDir := note.DateDir
			if dateDir == "" {
				// 如果没有日期目录，使用更新时间的日期
				dateDir = note.UpdatedAt.Format("20060102")
			}
			// 格式化日期为 YYYY-MM-DD
			if len(dateDir) == 8 {
				dateStr := dateDir[:4] + "-" + dateDir[4:6] + "-" + dateDir[6:8]
				dateMap[dateStr] = append(dateMap[dateStr], note)
			}
		}

		// 转换为切片并按日期倒序排序
		grouped := make([]NotesByDate, 0, len(dateMap))
		for date, notes := range dateMap {
			grouped = append(grouped, NotesByDate{
				Date:  date,
				Notes: notes,
			})
		}

		// 按日期倒序排序
		for i := 0; i < len(grouped)-1; i++ {
			for j := i + 1; j < len(grouped); j++ {
				if grouped[i].Date < grouped[j].Date {
					grouped[i], grouped[j] = grouped[j], grouped[i]
				}
			}
		}

		return grouped
	}

	groupedNotes := groupNotesByDate(notes)
	groupedBackupNotes := groupNotesByDate(backupNotes)

	// 获取所有日期用于过滤
	allDates := make(map[string]bool)
	for _, group := range groupedNotes {
		allDates[group.Date] = true
	}
	for _, group := range groupedBackupNotes {
		allDates[group.Date] = true
	}
	dateList := make([]string, 0, len(allDates))
	for date := range allDates {
		dateList = append(dateList, date)
	}
	// 按日期倒序排序
	for i := 0; i < len(dateList)-1; i++ {
		for j := i + 1; j < len(dateList); j++ {
			if dateList[i] < dateList[j] {
				dateList[i], dateList[j] = dateList[j], dateList[i]
			}
		}
	}

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

	tmpl := template.Must(template.New("admin").Funcs(funcMap).Parse(htmlPage.AdminPageHTML))
	tmpl.Execute(w, map[string]interface{}{
		"Notes":              notes,
		"BackupNotes":        backupNotes,
		"GroupedNotes":       groupedNotes,
		"GroupedBackupNotes": groupedBackupNotes,
		"DateList":           dateList,
		"TotalSize":          totalSize,
		"BackupTotalSize":    backupTotalSize,
		"TotalCount":         len(notes),
		"BackupCount":        len(backupNotes),
		"CurrentTotalSize":   currentTotalSize,
		"MaxTotalSize":       currentMaxTotalSize,
		"MaxNoteCount":       currentMaxNoteCount,
		"AdminPath":          deps.AdminPath,
		"NoteNameLen":        deps.GetNoteNameLen(),
		"BackupDays":         deps.GetBackupDays(),
		"NoteChars":          deps.GetNoteChars(),
		"MaxFileSize":        deps.GetMaxFileSize(),
		"MaxFileSizeMB":      currentMaxFileSizeMB,
		"MaxPathLength":      deps.GetMaxPathLength(),
		"AccessToken":        deps.AccessToken,
	})
}

// HandleUpdateMaxTotalSize 处理配置更新请求
func HandleUpdateMaxTotalSize(w http.ResponseWriter, r *http.Request) {
	// Check session token authentication
	sessionToken := getAdminSessionTokenFromRequest(r)

	if !validateAdminSession(sessionToken) {
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
		deps.SetAccessToken(*req.AccessToken)
		updated = true
	}

	// Update admin path if provided
	if req.AdminPath != nil && *req.AdminPath != "" {
		newPath := *req.AdminPath
		if !strings.HasPrefix(newPath, "/") {
			newPath = "/" + newPath
		}
		deps.SetAdminPath(newPath)
		updated = true
	}

	// Update note name length if provided
	if req.NoteNameLen != nil && *req.NoteNameLen > 0 {
		deps.SetNoteNameLen(*req.NoteNameLen)
		updated = true
	}

	// Update backup days if provided
	if req.BackupDays != nil && *req.BackupDays > 0 {
		deps.SetBackupDays(*req.BackupDays)
		updated = true
	}

	// Update note chars if provided
	if req.NoteChars != nil && *req.NoteChars != "" {
		deps.SetNoteChars(*req.NoteChars)
		updated = true
	}

	// Update max file size if provided
	if req.MaxFileSize != nil && *req.MaxFileSize != "" {
		size, err := deps.ParseFileSize(*req.MaxFileSize)
		if err != nil || size <= 0 {
			http.Error(w, fmt.Sprintf("Invalid maxFileSize format: %s", *req.MaxFileSize), http.StatusBadRequest)
			return
		}
		deps.SetMaxFileSize(size)
		updated = true
	}

	// Update max path length if provided
	if req.MaxPathLength != nil && *req.MaxPathLength > 0 {
		deps.SetMaxPathLength(*req.MaxPathLength)
		updated = true
	}

	// Update max total size if provided
	if req.MaxTotalSize != nil && *req.MaxTotalSize != "" {
		size, err := deps.ParseFileSize(*req.MaxTotalSize)
		if err != nil || size <= 0 {
			http.Error(w, fmt.Sprintf("Invalid maxTotalSize format: %s", *req.MaxTotalSize), http.StatusBadRequest)
			return
		}
		deps.SetMaxTotalSize(size)
		updated = true
	}

	// Update max note count if provided
	if req.MaxNoteCount != nil && *req.MaxNoteCount > 0 {
		deps.SetMaxNoteCount(*req.MaxNoteCount)
		updated = true
	}

	// Save config to file
	if updated {
		deps.SaveConfig()
	}

	deps.RLockMaxTotalSize()
	currentMaxTotalSize := deps.GetMaxTotalSize()
	deps.RUnlockMaxTotalSize()

	deps.RLockMaxNoteCount()
	currentMaxNoteCount := deps.GetMaxNoteCount()
	deps.RUnlockMaxNoteCount()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"adminPath":      deps.AdminPath,
		"noteNameLen":    deps.GetNoteNameLen(),
		"backupDays":     deps.GetBackupDays(),
		"noteChars":      deps.GetNoteChars(),
		"maxFileSize":    deps.GetMaxFileSize(),
		"maxPathLength":  deps.GetMaxPathLength(),
		"maxTotalSize":   currentMaxTotalSize,
		"maxTotalSizeMB": currentMaxTotalSize / (1024 * 1024),
		"maxNoteCount":   currentMaxNoteCount,
	})
}
