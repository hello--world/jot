package vars

import (
	"os"
	"sync"
)

const (
	SavePath   = "_tmp"
	BackupPath = "bak"
	UploadPath = "uploads" // Directory for uploaded files
)

// Vars 存储全局变量
type Vars struct {
	AdminPath        string
	Port             string
	NoteNameLen      int
	BackupDays       int
	NoteChars        string
	ExistingNotes    *sync.Map
	MaxFileSize      int64
	MaxPathLength    int
	MaxTotalSize     int64
	MaxTotalSizeLock *sync.RWMutex
	MaxNoteCount     int
	MaxNoteCountLock *sync.RWMutex
	AdminToken       string
	AccessToken      string
}

// NewVars 创建新的变量管理器
func NewVars() *Vars {
	v := &Vars{
		AdminPath:        "/admin",
		Port:             ":8080",
		NoteNameLen:      3,
		BackupDays:       7,
		NoteChars:        "0123456789abcdefghijklmnopqrstuvwxyz",
		ExistingNotes:    &sync.Map{},
		MaxFileSize:      10 * 1024 * 1024,
		MaxPathLength:    20,
		MaxTotalSize:     500 * 1024 * 1024,
		MaxTotalSizeLock: &sync.RWMutex{},
		MaxNoteCount:     500,
		MaxNoteCountLock: &sync.RWMutex{},
		AdminToken:       "",
		AccessToken:      "",
	}

	// 创建必要的目录
	os.MkdirAll(SavePath, 0755)
	os.MkdirAll(BackupPath, 0755)
	os.MkdirAll(UploadPath, 0755)

	return v
}
