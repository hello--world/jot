package backup

import (
	"log"
	"time"

	"github.com/hello--world/jot/note"
)

// Manager 备份管理器
type Manager struct {
	noteManager *note.Manager
}

// NewManager 创建新的备份管理器
func NewManager(noteManager *note.Manager) *Manager {
	return &Manager{
		noteManager: noteManager,
	}
}

// StartBackupScheduler 启动备份调度器
// 启动时立即执行一次，然后每天执行一次
func (m *Manager) StartBackupScheduler() {
	log.Printf("Starting backup scheduler...")
	go func() {
		// 启动时立即执行一次
		log.Printf("Running initial backup check...")
		if err := m.noteManager.MoveOldNotesToBackup(); err != nil {
			log.Printf("Error running initial backup: %v", err)
		} else {
			log.Printf("Initial backup check completed")
		}

		// 然后每天执行一次
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			log.Printf("Running scheduled backup check...")
			if err := m.noteManager.MoveOldNotesToBackup(); err != nil {
				log.Printf("Error running scheduled backup: %v", err)
			} else {
				log.Printf("Scheduled backup check completed")
			}
		}
	}()
}
