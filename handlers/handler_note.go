package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday/v2"

	"github.com/hello--world/jot/htmlPage"
)

// HandleNote 处理笔记的 GET 和 POST 请求
func HandleNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteName := vars["note"]

	// 只检查是否为空或不安全（允许用户输入任意字符）
	if noteName == "" || !deps.IsSafeNoteName(noteName) {
		http.Redirect(w, r, "/"+deps.GenerateNoteName(), http.StatusFound)
		return
	}

	if r.Method == "GET" {
		handleNoteGet(w, r, noteName)
		return
	}

	if r.Method == "POST" {
		handleNotePost(w, r, noteName)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleNoteGet(w http.ResponseWriter, r *http.Request, noteName string) {
	// 检查 access token（如果站点有 token，不带 /read 的路径需要 token）
	if deps.AccessToken != "" {
		token := deps.GetTokenFromRequest(r)
		if token != deps.AccessToken {
			// 如果是浏览器请求（不是 curl/wget），显示登录页面
			if !strings.HasPrefix(r.UserAgent(), "curl") && !strings.HasPrefix(r.UserAgent(), "Wget") {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(htmlPage.AccessLoginHTML))
				return
			}
			http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
			return
		}
	}

	// 检查是否是 raw 请求或 curl/wget
	if r.URL.Query().Get("raw") != "" || strings.HasPrefix(r.UserAgent(), "curl") || strings.HasPrefix(r.UserAgent(), "Wget") {
		content, err := deps.LoadNote(noteName)
		if err != nil || content == "" {
			http.NotFound(w, r)
			return
		}
		// Check note lock for raw requests
		if deps.HasNoteLock(content) {
			lockToken := deps.GetNoteLockToken(content)
			providedToken := deps.GetLockTokenFromRequest(r, noteName)
			if providedToken != lockToken {
				http.Error(w, "Unauthorized: Note is locked. Provide lock_token parameter or Authorization header.", http.StatusUnauthorized)
				return
			}
			content = deps.GetNoteContent(content)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(content))
		return
	}

	// Serve HTML page
	rawContent, err := deps.LoadNote(noteName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check note lock
	if deps.HasNoteLock(rawContent) {
		lockToken := deps.GetNoteLockToken(rawContent)
		providedToken := deps.GetLockTokenFromRequest(r, noteName)
		if providedToken != lockToken {
			// Show lock login page
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl := template.Must(template.New("lock").Parse(htmlPage.NoteLockHTML))
			tmpl.Execute(w, map[string]interface{}{
				"NoteName": noteName,
			})
			return
		}
		// Token is correct, extract actual content
		rawContent = deps.GetNoteContent(rawContent)
	}

	content := rawContent

	// 获取文件信息（大小和修改时间）
	var fileSize int64
	var modTime time.Time
	var createTime time.Time
	notePath := deps.GetNotePath(noteName)
	now := time.Now()

	if info, err := os.Stat(notePath); err == nil {
		fileSize = info.Size()
		modTime = info.ModTime()
		// 尝试获取创建时间（Windows 上可用，其他平台回退到修改时间）
		if ct, err := deps.GetFileCreationTime(notePath); err == nil {
			createTime = ct
		} else {
			// 如果无法获取创建时间，回退到修改时间
			createTime = modTime
		}
	} else {
		// 文件不存在（新建笔记），使用当前时间
		modTime = now
		createTime = now
	}

	// Format size
	sizeStr := fmt.Sprintf("%d B", fileSize)
	if fileSize >= 1024 {
		sizeStr = fmt.Sprintf("%.2f KB", float64(fileSize)/1024.0)
	}

	tmpl := template.Must(template.New("note").Parse(htmlPage.NotePageHTML))
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
}

func handleNotePost(w http.ResponseWriter, r *http.Request, noteName string) {
	// Check access token for POST requests (creating/updating notes)
	if deps.AccessToken != "" {
		token := deps.GetTokenFromRequest(r)
		if token != deps.AccessToken {
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
	if contentSize > deps.GetMaxFileSize() {
		http.Error(w, fmt.Sprintf("File size exceeds maximum limit of %d bytes (%d MB)", deps.GetMaxFileSize(), deps.GetMaxFileSize()/(1024*1024)), http.StatusRequestEntityTooLarge)
		return
	}

	// Check note count limit (only for new notes)
	wasNewNote := !deps.IsNoteExists(noteName)
	if wasNewNote {
		deps.RLockMaxNoteCount()
		currentMaxNoteCount := deps.GetMaxNoteCount()
		deps.RUnlockMaxNoteCount()

		// Count existing notes
		notes, err := deps.GetAllNotes()
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
	deps.RLockMaxTotalSize()
	currentMaxTotalSize := deps.GetMaxTotalSize()
	deps.RUnlockMaxTotalSize()

	currentTotalSize, err := deps.GetTotalFileSize()
	if err != nil {
		log.Printf("Error calculating total file size: %v", err)
	} else {
		// Get current note size if it exists
		var currentNoteSize int64
		if info, err := os.Stat(deps.GetNotePath(noteName)); err == nil {
			currentNoteSize = info.Size()
		}
		// Calculate new total size
		newTotalSize := currentTotalSize - currentNoteSize + contentSize
		if newTotalSize > currentMaxTotalSize {
			http.Error(w, fmt.Sprintf("Total file size would exceed maximum limit of %d MB (current: %.2f MB, would be: %.2f MB)", currentMaxTotalSize/(1024*1024), float64(currentTotalSize)/(1024*1024), float64(newTotalSize)/(1024*1024)), http.StatusRequestEntityTooLarge)
			return
		}
	}

	if err := deps.SaveNote(noteName, content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast update to WebSocket clients
	deps.BroadcastUpdate(noteName, content)

	w.WriteHeader(http.StatusOK)
}

// HandleReadNote 处理只读笔记页面
func HandleReadNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	noteName := vars["note"]

	// 只检查是否为空或不安全
	if noteName == "" || !deps.IsSafeNoteName(noteName) {
		http.NotFound(w, r)
		return
	}

	// Load note content
	rawContent, err := deps.LoadNote(noteName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 检查是否是 raw 请求（下载原始内容）
	isRawRequest := r.URL.Query().Get("raw") != ""

	// 如果是 raw 请求，直接返回原始文件内容（不带锁标记）
	if isRawRequest {
		// Check note lock
		if deps.HasNoteLock(rawContent) {
			lockToken := deps.GetNoteLockToken(rawContent)
			providedToken := deps.GetLockTokenFromRequest(r, noteName)
			if providedToken != lockToken {
				// Token 不正确，返回错误
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				http.Error(w, "Unauthorized: Note is locked. Provide correct lock_token parameter.", http.StatusUnauthorized)
				return
			}
			// Token is correct, extract actual content (remove lock marker)
			rawContent = deps.GetNoteContent(rawContent)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(rawContent))
		return
	}

	// 非 raw 请求，检查锁
	if deps.HasNoteLock(rawContent) {
		lockToken := deps.GetNoteLockToken(rawContent)
		providedToken := deps.GetLockTokenFromRequest(r, noteName)
		if providedToken != lockToken {
			// 显示 HTML 锁登录页面
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl := template.Must(template.New("lock").Parse(htmlPage.NoteLockHTML))
			tmpl.Execute(w, map[string]interface{}{
				"NoteName": noteName,
			})
			return
		}
		// Token is correct, extract actual content (remove lock marker)
		rawContent = deps.GetNoteContent(rawContent)
	}

	content := rawContent

	// 获取文件信息（大小和修改时间）
	var fileSize int64
	var modTime time.Time
	var createTime time.Time
	notePath := deps.GetNotePath(noteName)

	if info, err := os.Stat(notePath); err == nil {
		fileSize = info.Size()
		modTime = info.ModTime()
		// 尝试获取创建时间（Windows 上可用，其他平台回退到修改时间）
		if ct, err := deps.GetFileCreationTime(notePath); err == nil {
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
	tmpl := template.Must(template.New("read").Parse(htmlPage.ReadPageHTML))
	tmpl.Execute(w, map[string]interface{}{
		"NoteName":   noteName,
		"Content":    template.HTML(htmlContent),
		"FileSize":   sizeStr,
		"ModTime":    modTime.Format("2006-01-02 15:04:05"),
		"CreateTime": createTime.Format("2006-01-02 15:04:05"),
	})
}
