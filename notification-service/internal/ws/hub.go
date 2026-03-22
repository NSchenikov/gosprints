package ws

import (
	"encoding/json"
	"log"
	"sync"

	// "notification-service/internal/models"  // временно 
	// "notification-service/internal/metrics" // временно 
	"github.com/gorilla/websocket"
)

type TaskStatusEvent struct {
	Type      string `json:"type"`
	TaskID    int    `json:"task_id"`
	Text      string `json:"text"`
	Status    string `json:"status"`
	UserID    string `json:"user_id"` 
	Timestamp string `json:"timestamp"`
}

type NotificationHub struct {
	mu      sync.RWMutex
	clients map[string]*websocket.Conn
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients: make(map[string]*websocket.Conn),
	}
}

func (h *NotificationHub) AddClient(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	oldConn, userHasConnection := h.clients[userID]

	if userHasConnection {
		oldConn.Close()
		// metrics.Get().DecWSConnections() // временно 
	}

	h.clients[userID] = conn
	// metrics.Get().IncWSConnections() // временно 
}

func (h *NotificationHub) RemoveClient(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
	// metrics.Get().DecWSConnections() // временно 
}

func (h *NotificationHub) SendToUser(userID string, event TaskStatusEvent) {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[ws] marshal error: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("[ws] write error for user %s: %v", userID, err)
	}
}