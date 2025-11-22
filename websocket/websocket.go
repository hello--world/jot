package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients     = make(map[string]map[*websocket.Conn]bool)
	clientsLock = sync.RWMutex{}
)

// Manager 管理 WebSocket 连接
type Manager struct {
	IsSafeNoteName      func(string) bool
	GetAccessToken      func() string
	GetTokenFromRequest func(*http.Request) string
}

// NewManager 创建新的 WebSocket 管理器
func NewManager(isSafeNoteName func(string) bool, getAccessToken func() string, getTokenFromRequest func(*http.Request) string) *Manager {
	return &Manager{
		IsSafeNoteName:      isSafeNoteName,
		GetAccessToken:      getAccessToken,
		GetTokenFromRequest: getTokenFromRequest,
	}
}

// HandleWebSocket 处理 WebSocket 连接
func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 检查 access token（如果站点有 token，需要验证）
	accessToken := m.GetAccessToken()
	if accessToken != "" {
		token := m.GetTokenFromRequest(r)
		if token != accessToken {
			http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
			return
		}
	}

	vars := mux.Vars(r)
	noteName := vars["note"]

	// Only check if unsafe (allow any characters for user input)
	if !m.IsSafeNoteName(noteName) {
		http.Error(w, "Invalid note name", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	clientsLock.Lock()
	if clients[noteName] == nil {
		clients[noteName] = make(map[*websocket.Conn]bool)
	}
	clients[noteName][conn] = true
	clientsLock.Unlock()

	// Unregister on disconnect
	defer func() {
		clientsLock.Lock()
		delete(clients[noteName], conn)
		if len(clients[noteName]) == 0 {
			delete(clients, noteName)
		}
		clientsLock.Unlock()
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// BroadcastUpdate 广播更新
func BroadcastUpdate(noteName, content string) {
	clientsLock.RLock()
	defer clientsLock.RUnlock()

	if clients[noteName] == nil {
		return
	}

	message := map[string]interface{}{
		"type":    "update",
		"content": content,
	}

	for conn := range clients[noteName] {
		if err := conn.WriteJSON(message); err != nil {
			delete(clients[noteName], conn)
			conn.Close()
		}
	}
}
