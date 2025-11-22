package handlers

import (
	"net/http"
	"time"
)

// Note 表示笔记
type Note struct {
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
	Size      int64     `json:"size"`
}

// Dependencies 包含 handlers 需要的所有依赖
type Dependencies struct {
	// 配置变量
	AdminToken  string
	AccessToken string
	AdminPath   string

	// 笔记操作函数
	GetAllNotes         func() ([]Note, error)
	GetAllBackupNotes   func() ([]Note, error)
	LoadNote            func(string) (string, error)
	SaveNote            func(string, string) error
	GenerateNoteName    func() string
	IsSafeNoteName      func(string) bool
	GetNotePath         func(string) string
	IsNoteExists        func(string) bool
	GetFileCreationTime func(string) (time.Time, error)

	// 锁相关函数
	HasNoteLock             func(string) bool
	GetNoteLockToken        func(string) string
	GetNoteContent          func(string) string
	GetLockTokenFromRequest func(*http.Request, string) string
	GetTokenFromRequest     func(*http.Request) string

	// 文件相关
	GetTotalFileSize func() (int64, error)
	ParseFileSize    func(string) (int64, error)

	// WebSocket
	BroadcastUpdate func(string, string)

	// 配置保存
	SaveConfig func()

	// 变量（通过 getter/setter 访问）
	GetMaxFileSize   func() int64
	SetMaxFileSize   func(int64)
	GetMaxPathLength func() int
	SetMaxPathLength func(int)
	GetMaxTotalSize  func() int64
	SetMaxTotalSize  func(int64)
	GetMaxNoteCount  func() int
	SetMaxNoteCount  func(int)
	GetNoteNameLen   func() int
	SetNoteNameLen   func(int)
	GetBackupDays    func() int
	SetBackupDays    func(int)
	GetNoteChars     func() string
	SetNoteChars     func(string)
	GetSavePath      func() string
	GetUploadPath    func() string
	SetAdminPath     func(string)
	SetAccessToken   func(string)
	SetAdminToken    func(string)

	// 锁操作
	RLockMaxTotalSize   func()
	RUnlockMaxTotalSize func()
	LockMaxTotalSize    func()
	UnlockMaxTotalSize  func()
	RLockMaxNoteCount   func()
	RUnlockMaxNoteCount func()
	LockMaxNoteCount    func()
	UnlockMaxNoteCount  func()
}

// 全局依赖变量
var deps *Dependencies

// Init 初始化 handlers 包的依赖
func Init(d *Dependencies) {
	deps = d
}
