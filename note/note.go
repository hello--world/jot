package note

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Note represents a note with its metadata
type Note struct {
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
	Size      int64     `json:"size"`
}

// NoteListResponse represents a response containing a list of notes
type NoteListResponse struct {
	Notes []Note `json:"notes"`
}

// Manager 管理笔记操作
type Manager struct {
	SavePath      string
	BackupPath    string
	MaxPathLength int
	NoteNameLen   int
	NoteChars     string
	BackupDays    int
	ExistingNotes *sync.Map
}

// NewManager 创建新的笔记管理器
func NewManager(savePath, backupPath string, maxPathLength, noteNameLen, backupDays int, noteChars string) *Manager {
	return &Manager{
		SavePath:      savePath,
		BackupPath:    backupPath,
		MaxPathLength: maxPathLength,
		NoteNameLen:   noteNameLen,
		NoteChars:     noteChars,
		BackupDays:    backupDays,
		ExistingNotes: &sync.Map{},
	}
}

// Note lock functions
const lockPrefix = "<!-- LOCK:"
const lockSuffix = " -->\n"

// HasNoteLock checks if a note has a lock
func HasNoteLock(content string) bool {
	return strings.HasPrefix(content, lockPrefix)
}

// GetNoteLockToken extracts the lock token from note content
// Returns empty string if no lock
func GetNoteLockToken(content string) string {
	if !HasNoteLock(content) {
		return ""
	}
	// Extract token from <!-- LOCK:token -->
	endIdx := strings.Index(content, lockSuffix)
	if endIdx == -1 {
		return ""
	}
	return content[len(lockPrefix):endIdx]
}

// GetNoteContent extracts the actual content from a locked note
func GetNoteContent(content string) string {
	if !HasNoteLock(content) {
		return content
	}
	// Remove <!-- LOCK:token -->\n prefix
	endIdx := strings.Index(content, lockSuffix)
	if endIdx == -1 {
		return content
	}
	return content[endIdx+len(lockSuffix):]
}

// SetNoteLock adds a lock to note content
func SetNoteLock(content, token string) string {
	if token == "" {
		// Remove lock if token is empty
		return GetNoteContent(content)
	}
	// If already locked, replace the token
	if HasNoteLock(content) {
		actualContent := GetNoteContent(content)
		return lockPrefix + token + lockSuffix + actualContent
	}
	// Add new lock
	return lockPrefix + token + lockSuffix + content
}

// IsSafeNoteName checks if a note name is safe (prevents path traversal attacks)
// This is only for basic security, not for restricting user input
func (m *Manager) IsSafeNoteName(name string) bool {
	if name == "" || len(name) > m.MaxPathLength {
		return false
	}
	// Prevent path traversal and other dangerous patterns
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	// Prevent control characters
	for _, r := range name {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

// IsValidGeneratedName 检查名称是否对自动生成的笔记有效
// 这限制生成的名称只能使用 noteChars 字符集中的字符
func (m *Manager) IsValidGeneratedName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !strings.ContainsRune(m.NoteChars, r) {
			return false
		}
	}
	return true
}

// GetNotePath 获取笔记文件路径
func (m *Manager) GetNotePath(name string) string {
	return filepath.Join(m.SavePath, name)
}

// LoadExistingNotes 将所有已存在的笔记名称加载到内存中
func (m *Manager) LoadExistingNotes() error {
	files, err := os.ReadDir(m.SavePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist yet, no notes to load
		}
		return err
	}

	for _, file := range files {
		if !file.IsDir() && m.IsSafeNoteName(file.Name()) {
			m.ExistingNotes.Store(file.Name(), true)
		}
	}

	return nil
}

// IsNoteExists 检查笔记名称是否已存在（使用内存缓存）
func (m *Manager) IsNoteExists(name string) bool {
	_, exists := m.ExistingNotes.Load(name)
	return exists
}

// AddNoteToCache 将笔记名称添加到内存缓存中
func (m *Manager) AddNoteToCache(name string) {
	m.ExistingNotes.Store(name, true)
}

// RemoveNoteFromCache 从内存缓存中移除笔记名称
func (m *Manager) RemoveNoteFromCache(name string) {
	m.ExistingNotes.Delete(name)
}

// GenerateNoteName 生成新的笔记名称
func (m *Manager) GenerateNoteName() string {
	// 从最小长度开始，如果名称已存在则增加长度
	length := m.NoteNameLen
	maxAttempts := 100 // 由于使用内存缓存，增加尝试次数

	for attempt := 0; attempt < maxAttempts; attempt++ {
		name := make([]byte, length)
		for i := range name {
			name[i] = m.NoteChars[rand.Intn(len(m.NoteChars))]
		}
		noteName := string(name)

		// 使用内存缓存检查笔记是否已存在
		if !m.IsNoteExists(noteName) {
			// 立即添加到缓存以防止竞态条件
			m.AddNoteToCache(noteName)
			return noteName
		}

		// 如果名称存在且尚未达到最大长度，增加长度
		// 如果已经是 4 位或更多，则使用相同长度重试
		if length < 4 {
			length = 4
		} else if attempt%10 == 0 {
			// 每 10 次尝试，增加长度以避免太多冲突
			length++
		}
	}

	// 如果所有尝试都失败，继续尝试增加长度
	// 如果 noteChars 有足够的字符，这应该很少发生
	for length < 20 {
		length++
		for attempt := 0; attempt < 50; attempt++ {
			name := make([]byte, length)
			for i := range name {
				name[i] = m.NoteChars[rand.Intn(len(m.NoteChars))]
			}
			noteName := string(name)
			if !m.IsNoteExists(noteName) {
				m.AddNoteToCache(noteName)
				return noteName
			}
		}
	}

	// 最后手段：这在实践中不应该发生
	// 但如果发生了，返回一个错误指示名称
	log.Printf("警告: 经过多次尝试后仍无法生成唯一的笔记名称")
	return ""
}

// SaveNote 保存笔记
func (m *Manager) SaveNote(name, content string) error {
	path := m.GetNotePath(name)
	if content == "" {
		os.Remove(path)
		m.RemoveNoteFromCache(name)
		return nil
	}

	// 检查这是否是新笔记
	wasNewNote := !m.IsNoteExists(name)

	err := os.WriteFile(path, []byte(content), 0644)
	if err == nil && wasNewNote {
		// 如果是新笔记，添加到缓存
		m.AddNoteToCache(name)
	}
	return err
}

// LoadNote 加载笔记
func (m *Manager) LoadNote(name string) (string, error) {
	path := m.GetNotePath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// GetAllNotes 获取所有笔记
func (m *Manager) GetAllNotes() ([]Note, error) {
	files, err := os.ReadDir(m.SavePath)
	if err != nil {
		return nil, err
	}

	notes := make([]Note, 0)
	for _, file := range files {
		if !file.IsDir() && m.IsSafeNoteName(file.Name()) {
			content, _ := m.LoadNote(file.Name())
			info, _ := file.Info()
			notes = append(notes, Note{
				Name:      file.Name(),
				Content:   content,
				UpdatedAt: info.ModTime(),
				Size:      info.Size(),
			})
		}
	}
	return notes, nil
}

// MoveOldNotesToBackup 将超过 backupDays 天未修改的笔记移动到备份文件夹
// 备份文件夹结构: bak/YYYYMMDD/note_name
func (m *Manager) MoveOldNotesToBackup() error {
	files, err := os.ReadDir(m.SavePath)
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -m.BackupDays)
	movedCount := 0

	for _, file := range files {
		if file.IsDir() || !m.IsSafeNoteName(file.Name()) {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Check if note hasn't been modified for backupDays days
		if info.ModTime().Before(cutoffTime) {
			sourcePath := m.GetNotePath(file.Name())
			// Use creation/modification date as directory name (YYYYMMDD format)
			dateDir := info.ModTime().Format("20060102")
			dateBackupPath := filepath.Join(m.BackupPath, dateDir)

			// Create date directory if it doesn't exist
			if err := os.MkdirAll(dateBackupPath, 0755); err != nil {
				log.Printf("Failed to create backup directory %s: %v", dateBackupPath, err)
				continue
			}

			backupFilePath := filepath.Join(dateBackupPath, file.Name())

			// Move file to backup folder
			if err := os.Rename(sourcePath, backupFilePath); err != nil {
				log.Printf("Failed to move note %s to backup: %v", file.Name(), err)
				continue
			}

			movedCount++
			log.Printf("Moved old note %s to backup/%s (last modified: %s)", file.Name(), dateDir, info.ModTime().Format("2006-01-02 15:04:05"))
		}
	}

	if movedCount > 0 {
		log.Printf("Moved %d old note(s) to backup folder", movedCount)
	}

	return nil
}

// GetAllBackupNotes 返回备份文件夹中的所有笔记
func (m *Manager) GetAllBackupNotes() ([]Note, error) {
	backupNotes := make([]Note, 0)

	// Check if backup directory exists
	if _, err := os.Stat(m.BackupPath); os.IsNotExist(err) {
		return backupNotes, nil
	}

	// Read all date directories in backup folder
	dateDirs, err := os.ReadDir(m.BackupPath)
	if err != nil {
		return nil, err
	}

	for _, dateDir := range dateDirs {
		if !dateDir.IsDir() {
			continue
		}

		datePath := filepath.Join(m.BackupPath, dateDir.Name())
		files, err := os.ReadDir(datePath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !m.IsSafeNoteName(file.Name()) {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			// Read note content
			notePath := filepath.Join(datePath, file.Name())
			content, err := os.ReadFile(notePath)
			if err != nil {
				continue
			}

			backupNotes = append(backupNotes, Note{
				Name:      file.Name(),
				Content:   string(content),
				UpdatedAt: info.ModTime(),
				Size:      info.Size(),
			})
		}
	}

	return backupNotes, nil
}
