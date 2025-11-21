package ws

import (
	"encoding/json"
	"log"
	"sync"

	"gosprints/internal/models"
	"github.com/gorilla/websocket"
)

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
	h.clients[userID] = conn
}

func (h *NotificationHub) RemoveClient(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

func (h *NotificationHub) SendToUser(userID string, evt models.TaskStatusEvent) {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("[ws] marshal error: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("[ws] write error for user %s: %v", userID, err)
	}
}
