package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
)

const ConfigFile = "config.json" // Configuration file path

// Config structure for saving/loading configuration
type Config struct {
	AdminToken    string `json:"adminToken"`
	AccessToken   string `json:"accessToken"`
	AdminPath     string `json:"adminPath"`
	NoteNameLen   int    `json:"noteNameLen"`
	BackupDays    int    `json:"backupDays"`
	NoteChars     string `json:"noteChars"`
	MaxFileSize   int64  `json:"maxFileSize"`
	MaxPathLength int    `json:"maxPathLength"`
	MaxTotalSize  int64  `json:"maxTotalSize"`
	MaxNoteCount  int    `json:"maxNoteCount"`
}

// Manager 管理配置
type Manager struct {
	configLoaded bool
	config       *Config

	// 变量引用（通过 setter/getter 访问）
	adminToken       *string
	accessToken      *string
	adminPath        *string
	noteNameLen      *int
	backupDays       *int
	noteChars        *string
	maxFileSize      *int64
	maxPathLength    *int
	maxTotalSize     *int64
	maxTotalSizeLock *sync.RWMutex
	maxNoteCount     *int
	maxNoteCountLock *sync.RWMutex
}

// NewManager 创建新的配置管理器
func NewManager(
	adminToken, accessToken, adminPath *string,
	noteNameLen, backupDays, maxPathLength, maxNoteCount *int,
	noteChars *string,
	maxFileSize, maxTotalSize *int64,
	maxTotalSizeLock, maxNoteCountLock *sync.RWMutex,
) *Manager {
	return &Manager{
		configLoaded:     false,
		config:           &Config{},
		adminToken:       adminToken,
		accessToken:      accessToken,
		adminPath:        adminPath,
		noteNameLen:      noteNameLen,
		backupDays:       backupDays,
		noteChars:        noteChars,
		maxFileSize:      maxFileSize,
		maxPathLength:    maxPathLength,
		maxTotalSize:     maxTotalSize,
		maxTotalSizeLock: maxTotalSizeLock,
		maxNoteCount:     maxNoteCount,
		maxNoteCountLock: maxNoteCountLock,
	}
}

// LoadConfig loads configuration from config.json file
// Returns true if config file exists and was loaded successfully
func (m *Manager) LoadConfig() bool {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, will use defaults from env/command line
			m.configLoaded = false
			return false
		}
		log.Printf("Warning: Failed to read config file: %v", err)
		m.configLoaded = false
		return false
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("Warning: Failed to parse config file: %v", err)
		m.configLoaded = false
		return false
	}

	// Update all values from config file
	if cfg.AdminToken != "" {
		*m.adminToken = cfg.AdminToken
	}
	if cfg.AccessToken != "" {
		*m.accessToken = cfg.AccessToken
	}
	if cfg.AdminPath != "" {
		*m.adminPath = cfg.AdminPath
		if !strings.HasPrefix(*m.adminPath, "/") {
			*m.adminPath = "/" + *m.adminPath
		}
	}
	if cfg.NoteNameLen > 0 {
		*m.noteNameLen = cfg.NoteNameLen
	}
	if cfg.BackupDays > 0 {
		*m.backupDays = cfg.BackupDays
	}
	if cfg.NoteChars != "" {
		*m.noteChars = cfg.NoteChars
	}
	if cfg.MaxFileSize > 0 {
		*m.maxFileSize = cfg.MaxFileSize
	}
	if cfg.MaxPathLength > 0 {
		*m.maxPathLength = cfg.MaxPathLength
	}
	if cfg.MaxTotalSize > 0 {
		m.maxTotalSizeLock.Lock()
		*m.maxTotalSize = cfg.MaxTotalSize
		m.maxTotalSizeLock.Unlock()
	}
	if cfg.MaxNoteCount > 0 {
		m.maxNoteCountLock.Lock()
		*m.maxNoteCount = cfg.MaxNoteCount
		m.maxNoteCountLock.Unlock()
	}

	m.configLoaded = true
	return true
}

// IsConfigLoaded 返回配置是否已加载
func (m *Manager) IsConfigLoaded() bool {
	return m.configLoaded
}

// SaveConfig saves current configuration to config.json file
func (m *Manager) SaveConfig() {
	m.maxTotalSizeLock.RLock()
	currentMaxTotalSize := *m.maxTotalSize
	m.maxTotalSizeLock.RUnlock()

	m.maxNoteCountLock.RLock()
	currentMaxNoteCount := *m.maxNoteCount
	m.maxNoteCountLock.RUnlock()

	cfg := Config{
		AdminToken:    *m.adminToken,
		AccessToken:   *m.accessToken,
		AdminPath:     *m.adminPath,
		NoteNameLen:   *m.noteNameLen,
		BackupDays:    *m.backupDays,
		NoteChars:     *m.noteChars,
		MaxFileSize:   *m.maxFileSize,
		MaxPathLength: *m.maxPathLength,
		MaxTotalSize:  currentMaxTotalSize,
		MaxNoteCount:  currentMaxNoteCount,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("Warning: Failed to marshal config: %v", err)
		return
	}

	if err := os.WriteFile(ConfigFile, data, 0644); err != nil {
		log.Printf("Warning: Failed to save config file: %v", err)
		return
	}

	m.configLoaded = true
	log.Printf("Configuration saved to %s", ConfigFile)
}
