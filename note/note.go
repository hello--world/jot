package note

import (
	"encoding/json"
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
	DateDir   string    `json:"date_dir"`  // 日期目录，用于分组（格式：YYYYMMDD）
	IsBackup  bool      `json:"is_backup"` // 是否在备份文件夹
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
	NoteIndex     *sync.Map  // 存储 noteName -> dateDir 的映射
	indexFile     string     // 索引文件路径
	indexLock     sync.Mutex // 索引文件读写锁
}

// NewManager 创建新的笔记管理器
func NewManager(savePath, backupPath string, maxPathLength, noteNameLen, backupDays int, noteChars string) *Manager {
	indexFile := filepath.Join(savePath, ".notes_index")
	m := &Manager{
		SavePath:      savePath,
		BackupPath:    backupPath,
		MaxPathLength: maxPathLength,
		NoteNameLen:   noteNameLen,
		NoteChars:     noteChars,
		BackupDays:    backupDays,
		ExistingNotes: &sync.Map{},
		NoteIndex:     &sync.Map{},
		indexFile:     indexFile,
	}
	// 加载索引文件
	m.LoadNoteIndex()
	return m
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

// LoadNoteIndex 从索引文件加载笔记索引
func (m *Manager) LoadNoteIndex() {
	m.indexLock.Lock()
	defer m.indexLock.Unlock()

	data, err := os.ReadFile(m.indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 索引文件不存在，扫描目录重建索引
			m.rebuildIndex()
			return
		}
		log.Printf("Error reading note index: %v", err)
		return
	}

	var index map[string]string
	if err := json.Unmarshal(data, &index); err != nil {
		log.Printf("Error parsing note index: %v, rebuilding...", err)
		m.rebuildIndex()
		return
	}

	// 加载到内存
	for noteName, dateDir := range index {
		m.NoteIndex.Store(noteName, dateDir)
		m.ExistingNotes.Store(noteName, true)
	}
	log.Printf("Loaded %d notes from index", len(index))
}

// SaveNoteIndex 保存笔记索引到文件
func (m *Manager) SaveNoteIndex() {
	m.indexLock.Lock()
	defer m.indexLock.Unlock()

	index := make(map[string]string)
	m.NoteIndex.Range(func(key, value interface{}) bool {
		noteName := key.(string)
		dateDir := value.(string)
		index[noteName] = dateDir
		return true
	})

	data, err := json.Marshal(index)
	if err != nil {
		log.Printf("Error marshaling note index: %v", err)
		return
	}

	if err := os.WriteFile(m.indexFile, data, 0644); err != nil {
		log.Printf("Error saving note index: %v", err)
	}
}

// rebuildIndex 重建索引（扫描所有日期目录）
// 注意：调用此函数时，调用者必须已经持有 indexLock
func (m *Manager) rebuildIndex() {
	m.NoteIndex = &sync.Map{}
	m.ExistingNotes = &sync.Map{}

	// 确保 SavePath 目录存在
	if err := os.MkdirAll(m.SavePath, 0755); err != nil {
		log.Printf("Error creating save path: %v", err)
		return
	}

	files, err := os.ReadDir(m.SavePath)
	if err != nil {
		log.Printf("Error reading save path: %v", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		dirName := file.Name()
		// 检查是否是日期格式目录（YYYYMMDD，8位数字）
		if len(dirName) == 8 {
			isDateDir := true
			for _, r := range dirName {
				if r < '0' || r > '9' {
					isDateDir = false
					break
				}
			}
			if !isDateDir {
				continue
			}
		} else {
			continue
		}

		// 读取日期目录中的笔记
		datePath := filepath.Join(m.SavePath, dirName)
		noteFiles, err := os.ReadDir(datePath)
		if err != nil {
			continue
		}

		for _, noteFile := range noteFiles {
			if !noteFile.IsDir() && m.IsSafeNoteName(noteFile.Name()) {
				noteName := noteFile.Name()
				m.NoteIndex.Store(noteName, dirName)
				m.ExistingNotes.Store(noteName, true)
			}
		}
	}

	// 保存重建的索引（不获取锁，因为调用者已经持有）
	index := make(map[string]string)
	m.NoteIndex.Range(func(key, value interface{}) bool {
		noteName := key.(string)
		dateDir := value.(string)
		index[noteName] = dateDir
		return true
	})

	data, err := json.Marshal(index)
	if err != nil {
		log.Printf("Error marshaling note index: %v", err)
		return
	}

	if err := os.WriteFile(m.indexFile, data, 0644); err != nil {
		log.Printf("Error saving note index: %v", err)
		return
	}

	log.Printf("Rebuilt note index with %d notes", len(index))
}

// getIndexMap 获取索引的副本（用于统计）
func (m *Manager) getIndexMap() map[string]string {
	index := make(map[string]string)
	m.NoteIndex.Range(func(key, value interface{}) bool {
		index[key.(string)] = value.(string)
		return true
	})
	return index
}

// GetNotePath 获取笔记文件路径（保存时使用当前日期目录）
func (m *Manager) GetNotePath(name string) string {
	dateDir := time.Now().Format("20060102")
	return filepath.Join(m.SavePath, dateDir, name)
}

// FindNotePath 从索引中查找笔记文件路径（包括备份文件夹）
func (m *Manager) FindNotePath(name string) (string, error) {
	// 首先从活跃笔记索引中查找
	value, exists := m.NoteIndex.Load(name)
	if exists {
		dateDir := value.(string)
		notePath := filepath.Join(m.SavePath, dateDir, name)
		if _, err := os.Stat(notePath); err == nil {
			return notePath, nil
		}
		// 文件不存在，从索引中移除
		m.NoteIndex.Delete(name)
		m.ExistingNotes.Delete(name)
		m.SaveNoteIndex()
	}

	// 如果活跃笔记中找不到，从备份文件夹中查找
	if _, err := os.Stat(m.BackupPath); err == nil {
		dateDirs, err := os.ReadDir(m.BackupPath)
		if err == nil {
			// 按日期倒序查找（最新的优先）
			for _, dateDir := range dateDirs {
				if !dateDir.IsDir() {
					continue
				}
				dirName := dateDir.Name()
				// 检查是否是日期格式目录
				if len(dirName) == 8 {
					isDateDir := true
					for _, r := range dirName {
						if r < '0' || r > '9' {
							isDateDir = false
							break
						}
					}
					if isDateDir {
						notePath := filepath.Join(m.BackupPath, dirName, name)
						if _, err := os.Stat(notePath); err == nil {
							return notePath, nil
						}
					}
				}
			}
		}
	}

	return "", os.ErrNotExist
}

// LoadExistingNotes 将所有已存在的笔记名称加载到内存中（已由 LoadNoteIndex 处理）
func (m *Manager) LoadExistingNotes() error {
	// 索引加载已在 NewManager 中完成
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

// SaveNote 保存笔记（保存到当前日期目录）
func (m *Manager) SaveNote(name, content string) error {
	// 检查这是否是新笔记
	wasNewNote := !m.IsNoteExists(name)

	// 如果内容为空，删除笔记
	if content == "" {
		notePath, err := m.FindNotePath(name)
		if err == nil {
			os.Remove(notePath)
		}
		m.NoteIndex.Delete(name)
		m.RemoveNoteFromCache(name)
		m.SaveNoteIndex()
		return nil
	}

	// 获取当前日期目录
	currentDateDir := time.Now().Format("20060102")
	path := filepath.Join(m.SavePath, currentDateDir, name)

	// 如果笔记已存在但在其他日期目录，先删除旧文件
	if !wasNewNote {
		oldPath, err := m.FindNotePath(name)
		if err == nil && oldPath != path {
			os.Remove(oldPath)
		}
	}

	// 创建日期目录（如果不存在）
	dateDir := filepath.Join(m.SavePath, currentDateDir)
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return err
	}

	// 保存到当前日期目录
	err := os.WriteFile(path, []byte(content), 0644)
	if err == nil {
		// 更新索引
		m.NoteIndex.Store(name, currentDateDir)
		if wasNewNote {
			m.AddNoteToCache(name)
		}
		// 保存索引文件
		m.SaveNoteIndex()
	}
	return err
}

// LoadNote 加载笔记（从所有日期目录中查找）
func (m *Manager) LoadNote(name string) (string, error) {
	path, err := m.FindNotePath(name)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// GetAllNotes 获取所有笔记（从索引中读取）
func (m *Manager) GetAllNotes() ([]Note, error) {
	notes := make([]Note, 0)

	m.NoteIndex.Range(func(key, value interface{}) bool {
		noteName := key.(string)
		dateDir := value.(string)
		notePath := filepath.Join(m.SavePath, dateDir, noteName)

		// 读取文件信息
		info, err := os.Stat(notePath)
		if err != nil {
			// 文件不存在，从索引中移除
			m.NoteIndex.Delete(noteName)
			m.ExistingNotes.Delete(noteName)
			return true
		}

		// 读取文件内容
		content, err := os.ReadFile(notePath)
		if err != nil {
			return true
		}

		notes = append(notes, Note{
			Name:      noteName,
			Content:   string(content),
			UpdatedAt: info.ModTime(),
			Size:      info.Size(),
			DateDir:   dateDir,
			IsBackup:  false,
		})
		return true
	})

	// 如果索引中有无效条目，保存更新后的索引
	m.SaveNoteIndex()

	return notes, nil
}

// MoveOldNotesToBackup 将超过 backupDays 天未修改的日期目录移动到备份文件夹
// 备份文件夹结构: bak/YYYYMMDD/（整个日期目录）
func (m *Manager) MoveOldNotesToBackup() error {
	files, err := os.ReadDir(m.SavePath)
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -m.BackupDays)
	movedCount := 0

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		// 检查是否是日期格式目录（YYYYMMDD，8位数字）
		dirName := file.Name()
		if len(dirName) != 8 {
			continue
		}
		isDateDir := true
		for _, r := range dirName {
			if r < '0' || r > '9' {
				isDateDir = false
				break
			}
		}
		if !isDateDir {
			continue
		}

		// 检查日期目录的最后修改时间
		// 使用目录中最新文件的修改时间作为目录的修改时间
		datePath := filepath.Join(m.SavePath, dirName)
		noteFiles, err := os.ReadDir(datePath)
		if err != nil {
			continue
		}

		// 找到目录中最新的文件修改时间
		var latestModTime time.Time
		hasNotes := false
		for _, noteFile := range noteFiles {
			if noteFile.IsDir() {
				continue
			}
			info, err := noteFile.Info()
			if err != nil {
				continue
			}
			if !hasNotes || info.ModTime().After(latestModTime) {
				latestModTime = info.ModTime()
				hasNotes = true
			}
		}

		// 如果目录中没有笔记，或者最新修改时间早于截止时间，移动整个目录
		if !hasNotes || latestModTime.Before(cutoffTime) {
			sourcePath := filepath.Join(m.SavePath, dirName)
			backupPath := filepath.Join(m.BackupPath, dirName)

			// 如果备份目录已存在，合并文件
			if _, err := os.Stat(backupPath); err == nil {
				// 备份目录已存在，移动目录中的文件
				for _, noteFile := range noteFiles {
					if noteFile.IsDir() {
						continue
					}
					noteName := noteFile.Name()
					if !m.IsSafeNoteName(noteName) {
						continue
					}
					sourceFilePath := filepath.Join(sourcePath, noteName)
					backupFilePath := filepath.Join(backupPath, noteName)
					if err := os.Rename(sourceFilePath, backupFilePath); err != nil {
						log.Printf("Failed to move note %s/%s to backup: %v", dirName, noteName, err)
						continue
					}
					m.RemoveNoteFromCache(noteName)
					m.NoteIndex.Delete(noteName)
					movedCount++
				}
				// 删除空的源目录
				os.Remove(sourcePath)
				// 保存更新后的索引
				m.SaveNoteIndex()
			} else {
				// 备份目录不存在，直接移动整个目录
				if err := os.Rename(sourcePath, backupPath); err != nil {
					log.Printf("Failed to move date directory %s to backup: %v", dirName, err)
					continue
				}

				// 从缓存和索引中移除该目录下的所有笔记
				for _, noteFile := range noteFiles {
					if !noteFile.IsDir() && m.IsSafeNoteName(noteFile.Name()) {
						m.RemoveNoteFromCache(noteFile.Name())
						m.NoteIndex.Delete(noteFile.Name())
					}
				}
				// 保存更新后的索引
				m.SaveNoteIndex()

				movedCount++
				log.Printf("Moved date directory %s to backup (latest modified: %s)", dirName, latestModTime.Format("2006-01-02 15:04:05"))
			}
		}
	}

	if movedCount > 0 {
		log.Printf("Moved %d date directory/directories to backup folder", movedCount)
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
				DateDir:   dateDir.Name(),
				IsBackup:  true,
			})
		}
	}

	return backupNotes, nil
}
