package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hello--world/jot/config"
	"github.com/hello--world/jot/handlers"
	"github.com/hello--world/jot/note"
	"github.com/hello--world/jot/router"
	"github.com/hello--world/jot/setup"
	"github.com/hello--world/jot/utils"
	"github.com/hello--world/jot/vars"
	"github.com/hello--world/jot/websocket"
)

var (
	// 全局变量管理器
	v *vars.Vars
	// 笔记管理器
	noteManager *note.Manager
	// 配置管理器
	configManager *config.Manager
	// WebSocket 管理器
	wsManager *websocket.Manager
)

// convertNoteToHandlerNote 将 note.Note 转换为 handlers.Note
func convertNoteToHandlerNote(n note.Note) handlers.Note {
	return handlers.Note{
		Name:      n.Name,
		Content:   n.Content,
		UpdatedAt: n.UpdatedAt,
		Size:      n.Size,
	}
}

// initSetup 初始化 setup 包
func initSetup() {
	loader := &setup.ConfigLoader{
		LoadEnvFile:       utils.LoadEnvFile,
		LoadConfig:        func() bool { return configManager.LoadConfig() },
		SaveConfig:        func() { configManager.SaveConfig() },
		ParseFileSize:     utils.ParseFileSize,
		LoadExistingNotes: func() error { return noteManager.LoadExistingNotes() },
		GetConfigLoaded:   func() bool { return configManager.IsConfigLoaded() },
		SetConfigLoaded:   func(v bool) { /* 由 configManager 管理 */ },

		SetAdminPath:     func(val string) { v.AdminPath = val },
		SetPort:          func(val string) { v.Port = val },
		SetNoteNameLen:   func(val int) { v.NoteNameLen = val },
		SetBackupDays:    func(val int) { v.BackupDays = val },
		SetNoteChars:     func(val string) { v.NoteChars = val },
		SetMaxFileSize:   func(val int64) { v.MaxFileSize = val },
		SetMaxPathLength: func(val int) { v.MaxPathLength = val },
		SetMaxTotalSize:  func(val int64) { v.MaxTotalSizeLock.Lock(); v.MaxTotalSize = val; v.MaxTotalSizeLock.Unlock() },
		SetMaxNoteCount:  func(val int) { v.MaxNoteCountLock.Lock(); v.MaxNoteCount = val; v.MaxNoteCountLock.Unlock() },
		SetAdminToken:    func(val string) { v.AdminToken = val },
		SetAccessToken:   func(val string) { v.AccessToken = val },

		GetAdminPath:     func() string { return v.AdminPath },
		GetPort:          func() string { return v.Port },
		GetNoteNameLen:   func() int { return v.NoteNameLen },
		GetBackupDays:    func() int { return v.BackupDays },
		GetNoteChars:     func() string { return v.NoteChars },
		GetMaxFileSize:   func() int64 { return v.MaxFileSize },
		GetMaxPathLength: func() int { return v.MaxPathLength },
		GetMaxTotalSize:  func() int64 { v.MaxTotalSizeLock.RLock(); defer v.MaxTotalSizeLock.RUnlock(); return v.MaxTotalSize },
		GetMaxNoteCount:  func() int { v.MaxNoteCountLock.RLock(); defer v.MaxNoteCountLock.RUnlock(); return v.MaxNoteCount },
		GetAdminToken:    func() string { return v.AdminToken },
		GetAccessToken:   func() string { return v.AccessToken },
	}
	setup.InitConfigLoader(loader)
}

// initHandlerInitializer 初始化 handler 初始化器
func initHandlerInitializer() {
	init := &setup.HandlerInitializer{
		ConvertNoteToHandlerNote: func(n interface{}) handlers.Note {
			note := n.(note.Note)
			return convertNoteToHandlerNote(note)
		},
		GetAllNotes: func() ([]interface{}, error) {
			notes, err := noteManager.GetAllNotes()
			if err != nil {
				return nil, err
			}
			result := make([]interface{}, len(notes))
			for i := range notes {
				result[i] = notes[i]
			}
			return result, nil
		},
		GetAllBackupNotes: func() ([]interface{}, error) {
			notes, err := noteManager.GetAllBackupNotes()
			if err != nil {
				return nil, err
			}
			result := make([]interface{}, len(notes))
			for i := range notes {
				result[i] = notes[i]
			}
			return result, nil
		},
		LoadNote:            func(name string) (string, error) { return noteManager.LoadNote(name) },
		SaveNote:            func(name, content string) error { return noteManager.SaveNote(name, content) },
		GenerateNoteName:    func() string { return noteManager.GenerateNoteName() },
		IsSafeNoteName:      func(name string) bool { return noteManager.IsSafeNoteName(name) },
		GetNotePath:         func(name string) string { return noteManager.GetNotePath(name) },
		IsNoteExists:        func(name string) bool { return noteManager.IsNoteExists(name) },
		GetFileCreationTime: func(path string) (time.Time, error) { return utils.GetFileCreationTime(path) },
		HasNoteLock:         func(content string) bool { return note.HasNoteLock(content) },
		GetNoteLockToken:    func(content string) string { return note.GetNoteLockToken(content) },
		GetNoteContent:      func(content string) string { return note.GetNoteContent(content) },
		GetTotalFileSize:    func() (int64, error) { return utils.GetTotalFileSize(vars.SavePath, vars.UploadPath) },
		ParseFileSize:       utils.ParseFileSize,
		BroadcastUpdate:     websocket.BroadcastUpdate,
		SaveConfig:          func() { configManager.SaveConfig() },
		GetMaxFileSize:      func() int64 { return v.MaxFileSize },
		SetMaxFileSize:      func(val int64) { v.MaxFileSize = val },
		GetMaxPathLength:    func() int { return v.MaxPathLength },
		SetMaxPathLength:    func(val int) { v.MaxPathLength = val },
		GetMaxTotalSize:     func() int64 { v.MaxTotalSizeLock.RLock(); defer v.MaxTotalSizeLock.RUnlock(); return v.MaxTotalSize },
		SetMaxTotalSize:     func(val int64) { v.MaxTotalSizeLock.Lock(); v.MaxTotalSize = val; v.MaxTotalSizeLock.Unlock() },
		GetMaxNoteCount:     func() int { v.MaxNoteCountLock.RLock(); defer v.MaxNoteCountLock.RUnlock(); return v.MaxNoteCount },
		SetMaxNoteCount:     func(val int) { v.MaxNoteCountLock.Lock(); v.MaxNoteCount = val; v.MaxNoteCountLock.Unlock() },
		GetNoteNameLen:      func() int { return v.NoteNameLen },
		SetNoteNameLen:      func(val int) { v.NoteNameLen = val },
		GetBackupDays:       func() int { return v.BackupDays },
		SetBackupDays:       func(val int) { v.BackupDays = val },
		GetNoteChars:        func() string { return v.NoteChars },
		SetNoteChars:        func(val string) { v.NoteChars = val },
		GetSavePath:         func() string { return vars.SavePath },
		GetUploadPath:       func() string { return vars.UploadPath },
		SetAdminPath:        func(val string) { v.AdminPath = val },
		SetAccessToken:      func(val string) { v.AccessToken = val },
		SetAdminToken:       func(val string) { v.AdminToken = val },
		GetAdminToken:       func() string { return v.AdminToken },
		GetAccessToken:      func() string { return v.AccessToken },
		GetAdminPath:        func() string { return v.AdminPath },
		RLockMaxTotalSize:   func() { v.MaxTotalSizeLock.RLock() },
		RUnlockMaxTotalSize: func() { v.MaxTotalSizeLock.RUnlock() },
		LockMaxTotalSize:    func() { v.MaxTotalSizeLock.Lock() },
		UnlockMaxTotalSize:  func() { v.MaxTotalSizeLock.Unlock() },
		RLockMaxNoteCount:   func() { v.MaxNoteCountLock.RLock() },
		RUnlockMaxNoteCount: func() { v.MaxNoteCountLock.RUnlock() },
		LockMaxNoteCount:    func() { v.MaxNoteCountLock.Lock() },
		UnlockMaxNoteCount:  func() { v.MaxNoteCountLock.Unlock() },
	}
	setup.InitHandlerInitializer(init)
}

func main() {
	// 初始化全局变量
	v = vars.NewVars()

	// 初始化笔记管理器
	noteManager = note.NewManager(
		vars.SavePath,
		vars.BackupPath,
		v.MaxPathLength,
		v.NoteNameLen,
		v.BackupDays,
		v.NoteChars,
	)
	// 加载现有笔记到缓存
	noteManager.LoadExistingNotes()

	// 初始化配置管理器
	configManager = config.NewManager(
		&v.AdminToken,
		&v.AccessToken,
		&v.AdminPath,
		&v.NoteNameLen,
		&v.BackupDays,
		&v.MaxPathLength,
		&v.MaxNoteCount,
		&v.NoteChars,
		&v.MaxFileSize,
		&v.MaxTotalSize,
		v.MaxTotalSizeLock,
		v.MaxNoteCountLock,
	)
	// 先尝试从配置文件加载
	configManager.LoadConfig()

	// 初始化 WebSocket 管理器
	wsManager = websocket.NewManager(
		noteManager.IsSafeNoteName,
		func() string { return v.AccessToken },
		handlers.GetTokenFromRequest,
	)

	// 初始化 setup 包
	initSetup()

	// 加载配置（从命令行、环境变量等）
	setup.LoadConfiguration()

	// 初始化 handler 初始化器
	initHandlerInitializer()

	// 初始化 handlers
	setup.InitHandlers()

	// 设置路由
	routerConfig := &router.RouterConfig{
		AdminPath:        v.AdminPath,
		UploadPath:       vars.UploadPath,
		HandleWebSocket:  wsManager.HandleWebSocket,
		GenerateNoteName: func() string { return noteManager.GenerateNoteName() },
		GetAccessToken:   func() string { return v.AccessToken },
	}
	router.InitRouter(routerConfig)
	r := router.SetupRoutes()

	fmt.Printf("Server starting on http://localhost%s\n", v.Port)
	fmt.Printf("Admin panel: http://localhost%s%s\n", v.Port, v.AdminPath)
	log.Fatal(http.ListenAndServe(v.Port, r))
}
