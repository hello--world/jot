package setup

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hello--world/jot/handlers"
)

// ConfigLoader 用于加载配置
type ConfigLoader struct {
	// 这些函数需要从 main 包传递过来
	LoadEnvFile       func() error
	LoadConfig        func() bool
	SaveConfig        func()
	ParseFileSize     func(string) (int64, error)
	LoadExistingNotes func() error
	GetConfigLoaded   func() bool
	SetConfigLoaded   func(bool)

	// 变量设置函数
	SetAdminPath     func(string)
	SetPort          func(string)
	SetNoteNameLen   func(int)
	SetBackupDays    func(int)
	SetNoteChars     func(string)
	SetMaxFileSize   func(int64)
	SetMaxPathLength func(int)
	SetMaxTotalSize  func(int64)
	SetMaxNoteCount  func(int)
	SetAdminToken    func(string)
	SetAccessToken   func(string)

	// 变量获取函数
	GetAdminPath     func() string
	GetPort          func() string
	GetNoteNameLen   func() int
	GetBackupDays    func() int
	GetNoteChars     func() string
	GetMaxFileSize   func() int64
	GetMaxPathLength func() int
	GetMaxTotalSize  func() int64
	GetMaxNoteCount  func() int
	GetAdminToken    func() string
	GetAccessToken   func() string
}

var loader *ConfigLoader

// InitConfigLoader 初始化配置加载器
func InitConfigLoader(l *ConfigLoader) {
	loader = l
}

// LoadConfiguration 加载配置（从命令行、环境变量、配置文件）
func LoadConfiguration() {
	// Load configuration from command line, environment variable, or .env file
	tokenFlag := flag.String("token", "", "Admin access token (required)")
	portFlag := flag.String("port", "", "Server port (default: :8080)")
	flag.Parse()

	// Get port from: command line > environment variable > default (port is always configurable)
	if *portFlag != "" {
		port := *portFlag
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
		loader.SetPort(port)
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		port := envPort
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
		loader.SetPort(port)
	}

	// Check if config file was loaded
	// If config file exists, use it and ignore env/command line (except port and token)
	// If config file doesn't exist, use env/command line and save to config file
	if !loader.GetConfigLoaded() {
		// Config file doesn't exist, load from env/command line
		// Try to load from .env file first
		if err := loader.LoadEnvFile(); err != nil {
			log.Printf("Warning: Failed to load .env file: %v", err)
		}

		// Get admin path from: command line > environment variable > default
		adminPathFlag := flag.String("admin-path", "", "Admin panel path (default: /admin)")
		if *adminPathFlag != "" {
			loader.SetAdminPath(*adminPathFlag)
		} else if envPath := os.Getenv("ADMIN_PATH"); envPath != "" {
			loader.SetAdminPath(envPath)
		}
		// Ensure admin path starts with /
		adminPath := loader.GetAdminPath()
		if !strings.HasPrefix(adminPath, "/") {
			loader.SetAdminPath("/" + adminPath)
		}

		// Get note name length from: command line > environment variable > default
		noteNameLenFlag := flag.Int("note-name-len", 0, "Minimum length for note names (default: 3)")
		if *noteNameLenFlag > 0 {
			loader.SetNoteNameLen(*noteNameLenFlag)
		} else if envLen := os.Getenv("NOTE_NAME_LEN"); envLen != "" {
			if len, err := strconv.Atoi(envLen); err == nil && len > 0 {
				loader.SetNoteNameLen(len)
			}
		}

		// Get backup days from: command line > environment variable > default
		backupDaysFlag := flag.Int("backup-days", 0, "Days before moving notes to backup (default: 7)")
		if *backupDaysFlag > 0 {
			loader.SetBackupDays(*backupDaysFlag)
		} else if envDays := os.Getenv("BACKUP_DAYS"); envDays != "" {
			if days, err := strconv.Atoi(envDays); err == nil && days > 0 {
				loader.SetBackupDays(days)
			}
		}

		// Get note characters from: command line > environment variable > default
		noteCharsFlag := flag.String("note-chars", "", "Characters used for generating note names (default: 0123456789abcdefghijklmnopqrstuvwxyz)")
		if *noteCharsFlag != "" {
			loader.SetNoteChars(*noteCharsFlag)
		} else if envChars := os.Getenv("NOTE_CHARS"); envChars != "" {
			loader.SetNoteChars(envChars)
		}

		// Validate noteChars is not empty
		if len(loader.GetNoteChars()) == 0 {
			log.Fatal("Error: NOTE_CHARS cannot be empty")
		}

		// Get max file size from: command line > environment variable > default
		maxFileSizeFlag := flag.String("max-file-size", "", "Maximum file size (e.g., 10M, 100MB, default: 10MB)")
		if *maxFileSizeFlag != "" {
			if size, err := loader.ParseFileSize(*maxFileSizeFlag); err == nil && size > 0 {
				loader.SetMaxFileSize(size)
			} else {
				log.Fatalf("Error: Invalid max-file-size format: %s. Use format like 10M, 100MB, 1G", *maxFileSizeFlag)
			}
		} else if envSize := os.Getenv("MAX_FILE_SIZE"); envSize != "" {
			if size, err := loader.ParseFileSize(envSize); err == nil && size > 0 {
				loader.SetMaxFileSize(size)
			} else {
				log.Fatalf("Error: Invalid MAX_FILE_SIZE format: %s. Use format like 10M, 100MB, 1G", envSize)
			}
		}

		// Get max path length from: command line > environment variable > default
		maxPathLengthFlag := flag.Int("max-path-length", 0, "Maximum path/note name length (default: 20)")
		if *maxPathLengthFlag > 0 {
			loader.SetMaxPathLength(*maxPathLengthFlag)
		} else if envPathLen := os.Getenv("MAX_PATH_LENGTH"); envPathLen != "" {
			if pathLen, err := strconv.Atoi(envPathLen); err == nil && pathLen > 0 {
				loader.SetMaxPathLength(pathLen)
			}
		}

		// Save config to file after loading from env/command line
		loader.SaveConfig()
		log.Printf("Configuration loaded from environment/command line and saved to config.json")
	} else {
		log.Printf("Configuration loaded from config.json (environment variables and command line arguments ignored)")
	}

	// Get token from: command line > environment variable > config file > .env file
	// Command line and env always take precedence (for security)
	if *tokenFlag != "" {
		loader.SetAdminToken(*tokenFlag)
	} else if envToken := os.Getenv("ADMIN_TOKEN"); envToken != "" {
		loader.SetAdminToken(envToken)
	} else if loader.GetConfigLoaded() && loader.GetAdminToken() == "" {
		// If config was loaded but token is still empty, it means config file didn't have token
		// Try to load from .env file
		if err := loader.LoadEnvFile(); err == nil {
			if envToken := os.Getenv("ADMIN_TOKEN"); envToken != "" {
				loader.SetAdminToken(envToken)
			}
		}
	}

	// Token is required
	if loader.GetAdminToken() == "" {
		log.Fatal("Error: Admin token is required. Set it via -token flag, ADMIN_TOKEN environment variable, ADMIN_TOKEN in .env file, or adminToken in config.json")
	}

	// Get access token from: environment variable > config file > .env file
	// Access token is optional - if not set, no authentication required for notes
	if !loader.GetConfigLoaded() {
		// Try to load from .env file first
		if err := loader.LoadEnvFile(); err == nil {
			if envAccessToken := os.Getenv("ACCESS_TOKEN"); envAccessToken != "" {
				loader.SetAccessToken(envAccessToken)
			}
		}
	}
	// If config was loaded, accessToken should already be set from config file
	// But we can still override with environment variable
	if envAccessToken := os.Getenv("ACCESS_TOKEN"); envAccessToken != "" {
		loader.SetAccessToken(envAccessToken)
	}

	// Load existing notes into memory cache
	if err := loader.LoadExistingNotes(); err != nil {
		log.Printf("Warning: Failed to load existing notes: %v", err)
	} else {
		log.Printf("Loaded existing notes into memory cache")
	}
}

// InitHandlers 初始化 handlers 包的依赖
type HandlerInitializer struct {
	// Note 转换函数
	ConvertNoteToHandlerNote func(interface{}) handlers.Note
	GetAllNotes              func() ([]interface{}, error)
	GetAllBackupNotes        func() ([]interface{}, error)

	// 笔记操作函数
	LoadNote            func(string) (string, error)
	SaveNote            func(string, string) error
	GenerateNoteName    func() string
	IsSafeNoteName      func(string) bool
	GetNotePath         func(string) string
	IsNoteExists        func(string) bool
	GetFileCreationTime func(string) (time.Time, error)

	// 锁相关函数
	HasNoteLock      func(string) bool
	GetNoteLockToken func(string) string
	GetNoteContent   func(string) string

	// 文件相关
	GetTotalFileSize func() (int64, error)
	ParseFileSize    func(string) (int64, error)

	// WebSocket
	BroadcastUpdate func(string, string)

	// 配置保存
	SaveConfig func()

	// 变量访问函数
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
	GetAdminToken    func() string
	GetAccessToken   func() string
	GetAdminPath     func() string

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

var initializer *HandlerInitializer

// InitHandlerInitializer 初始化 handler 初始化器
func InitHandlerInitializer(init *HandlerInitializer) {
	initializer = init
}

// InitHandlers 初始化 handlers 包
func InitHandlers() {
	// 转换 Note 的函数
	convertNotes := func(notes []interface{}) []handlers.Note {
		result := make([]handlers.Note, len(notes))
		for i, n := range notes {
			result[i] = initializer.ConvertNoteToHandlerNote(n)
		}
		return result
	}

	getAllNotesForHandlers := func() ([]handlers.Note, error) {
		notes, err := initializer.GetAllNotes()
		if err != nil {
			return nil, err
		}
		return convertNotes(notes), nil
	}

	getAllBackupNotesForHandlers := func() ([]handlers.Note, error) {
		notes, err := initializer.GetAllBackupNotes()
		if err != nil {
			return nil, err
		}
		return convertNotes(notes), nil
	}

	d := &handlers.Dependencies{
		AdminToken:  initializer.GetAdminToken(),
		AccessToken: initializer.GetAccessToken(),
		AdminPath:   initializer.GetAdminPath(),

		GetAllNotes:         getAllNotesForHandlers,
		GetAllBackupNotes:   getAllBackupNotesForHandlers,
		LoadNote:            initializer.LoadNote,
		SaveNote:            initializer.SaveNote,
		GenerateNoteName:    initializer.GenerateNoteName,
		IsSafeNoteName:      initializer.IsSafeNoteName,
		GetNotePath:         initializer.GetNotePath,
		IsNoteExists:        initializer.IsNoteExists,
		GetFileCreationTime: initializer.GetFileCreationTime,

		HasNoteLock:             initializer.HasNoteLock,
		GetNoteLockToken:        initializer.GetNoteLockToken,
		GetNoteContent:          initializer.GetNoteContent,
		GetLockTokenFromRequest: handlers.GetLockTokenFromRequest,
		GetTokenFromRequest:     handlers.GetTokenFromRequest,

		GetTotalFileSize: initializer.GetTotalFileSize,
		ParseFileSize:    initializer.ParseFileSize,

		BroadcastUpdate: initializer.BroadcastUpdate,

		SaveConfig: initializer.SaveConfig,

		GetMaxFileSize:   initializer.GetMaxFileSize,
		SetMaxFileSize:   initializer.SetMaxFileSize,
		GetMaxPathLength: initializer.GetMaxPathLength,
		SetMaxPathLength: initializer.SetMaxPathLength,
		GetMaxTotalSize:  initializer.GetMaxTotalSize,
		SetMaxTotalSize:  initializer.SetMaxTotalSize,
		GetMaxNoteCount:  initializer.GetMaxNoteCount,
		SetMaxNoteCount:  initializer.SetMaxNoteCount,
		GetNoteNameLen:   initializer.GetNoteNameLen,
		SetNoteNameLen:   initializer.SetNoteNameLen,
		GetBackupDays:    initializer.GetBackupDays,
		SetBackupDays:    initializer.SetBackupDays,
		GetNoteChars:     initializer.GetNoteChars,
		SetNoteChars:     initializer.SetNoteChars,
		GetSavePath:      initializer.GetSavePath,
		GetUploadPath:    initializer.GetUploadPath,
		SetAdminPath:     initializer.SetAdminPath,
		SetAccessToken:   initializer.SetAccessToken,
		SetAdminToken:    initializer.SetAdminToken,

		RLockMaxTotalSize:   initializer.RLockMaxTotalSize,
		RUnlockMaxTotalSize: initializer.RUnlockMaxTotalSize,
		LockMaxTotalSize:    initializer.LockMaxTotalSize,
		UnlockMaxTotalSize:  initializer.UnlockMaxTotalSize,
		RLockMaxNoteCount:   initializer.RLockMaxNoteCount,
		RUnlockMaxNoteCount: initializer.RUnlockMaxNoteCount,
		LockMaxNoteCount:    initializer.LockMaxNoteCount,
		UnlockMaxNoteCount:  initializer.UnlockMaxNoteCount,
	}
	handlers.Init(d)
}
