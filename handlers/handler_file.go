package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// HandleFileUpload 处理文件上传请求
func HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check access token for file uploads
	if deps.AccessToken != "" {
		token := deps.GetTokenFromRequest(r)
		if token != deps.AccessToken {
			http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
			return
		}
	}

	// Parse multipart form (max 100MB)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file size
	if handler.Size > deps.GetMaxFileSize() {
		http.Error(w, fmt.Sprintf("File size exceeds maximum limit of %d MB", deps.GetMaxFileSize()/(1024*1024)), http.StatusRequestEntityTooLarge)
		return
	}

	// Check total file size limit
	deps.RLockMaxTotalSize()
	currentMaxTotalSize := deps.GetMaxTotalSize()
	deps.RUnlockMaxTotalSize()

	currentTotalSize, err := deps.GetTotalFileSize()
	if err != nil {
		log.Printf("Error calculating total file size: %v", err)
	} else {
		newTotalSize := currentTotalSize + handler.Size
		if newTotalSize > currentMaxTotalSize {
			http.Error(w, fmt.Sprintf("Total file size would exceed maximum limit of %d MB", currentMaxTotalSize/(1024*1024)), http.StatusRequestEntityTooLarge)
			return
		}
	}

	// Generate unique filename with date directory
	now := time.Now()
	dateDir := now.Format("20060102")
	timestamp := now.UnixNano()

	originalFilename := handler.Filename
	// Sanitize filename
	originalFilename = filepath.Base(originalFilename)
	if originalFilename == "" || originalFilename == "." || originalFilename == ".." {
		originalFilename = "upload"
	}

	// Generate filename: timestamp-originalFilename
	ext := filepath.Ext(originalFilename)
	name := strings.TrimSuffix(originalFilename, ext)
	filename := fmt.Sprintf("%d-%s%s", timestamp, name, ext)

	// Create date directory: upload/YYYYMMDD/
	dateUploadPath := filepath.Join(deps.GetUploadPath(), dateDir)
	if err := os.MkdirAll(dateUploadPath, 0755); err != nil {
		http.Error(w, "Error creating upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Full path: upload/YYYYMMDD/timestamp-filename
	uploadFilePath := filepath.Join(dateUploadPath, filename)

	// Create file
	dst, err := os.Create(uploadFilePath)
	if err != nil {
		http.Error(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine if it's an image
	isImage := false
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	fileExt := strings.ToLower(filepath.Ext(filename))
	for _, imgExt := range imageExts {
		if fileExt == imgExt {
			isImage = true
			break
		}
	}

	// Return markdown format (URL includes date directory)
	fileURL := fmt.Sprintf("/uploads/%s/%s", dateDir, filename)
	var markdown string
	if isImage {
		markdown = fmt.Sprintf("![%s](%s)", filename, fileURL)
	} else {
		markdown = fmt.Sprintf("[下载 %s](%s)", filename, fileURL)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"filename": filename,
		"url":      fileURL,
		"markdown": markdown,
	})
}

// HandleFileDownload 处理文件下载请求
func HandleFileDownload(w http.ResponseWriter, r *http.Request) {
	// 检查 access token（如果站点有 token，需要验证）
	if deps.AccessToken != "" {
		token := deps.GetTokenFromRequest(r)
		if token != deps.AccessToken {
			http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
			return
		}
	}

	vars := mux.Vars(r)
	dateDir := vars["date"]
	filename := vars["filename"]

	// Security check: prevent path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	if strings.Contains(dateDir, "..") || strings.Contains(dateDir, "/") || strings.Contains(dateDir, "\\") {
		http.Error(w, "Invalid date directory", http.StatusBadRequest)
		return
	}

	// File path: upload/YYYYMMDD/filename
	filePath := filepath.Join(deps.GetUploadPath(), dateDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Check if it's an image - serve directly, otherwise force download
	isImage := false
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, imgExt := range imageExts {
		if ext == imgExt {
			isImage = true
			break
		}
	}

	if !isImage {
		// Force download for non-image files
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}
	http.ServeFile(w, r, filePath)
}
