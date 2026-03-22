package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSHandler(hub *NotificationHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем userID из query-параметра или заголовка
		// API Gateway будет проксировать запрос и передавать userID
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			userID = r.Header.Get("X-User-ID")  // API Gateway может добавлять заголовок
		}
		
		if userID == "" {
			// Пробуем получить из token (если токен передаётся, но мы не проверяем его)
			token := r.URL.Query().Get("token")
			if token != "" {
				// Здесь можно добавить простую проверку, но для упрощения пока берём как userID
				// В реальном проекте токен должен проверяться api-gateway
				userID = token
			}
		}
		
		if userID == "" {
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}
		
		// Апгрейд до WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		
		// Регистрируем клиента
		hub.AddClient(userID, conn)
		
		// Фоновая горутина для обработки соединения
		go func() {
			defer func() {
				hub.RemoveClient(userID)
				conn.Close()
			}()
			
			// Читаем сообщения (поддерживаем соединение)
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("[WS] Ошибка чтения от user %s: %v", userID, err)
					}
					break
				}
			}
		}()
	}
}