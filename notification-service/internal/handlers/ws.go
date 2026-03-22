package handlers

import (
    "net/http"
    "log"
    
    "github.com/gorilla/websocket"
    "notification-service/internal/ws"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type WSHandler struct {
    hub *ws.NotificationHub
}

func NewWSHandler(hub *ws.NotificationHub) *WSHandler {
    return &WSHandler{hub: hub}
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
    // Получаем user_id из query параметра или заголовка
    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        userID = r.Header.Get("X-User-ID")
    }
    
    if userID == "" {
        http.Error(w, "user_id required", http.StatusBadRequest)
        return
    }
    
    // Upgrading HTTP to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    
    // регистрация клиента в хабе ws
    h.hub.AddClient(userID, conn)
    
    // Убираем клиента при закрытии соединения ws
    defer h.hub.RemoveClient(userID)
    
    // Держим соединение открытым
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            break
        }
    }
}