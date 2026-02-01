package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"gosprints/pkg/auth"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSHandler(hub *NotificationHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// log.Printf("[WS DEBUG] === Новый WebSocket запрос ===")
		// log.Printf("[WS DEBUG] Метод: %s, Путь: %s", r.Method, r.URL.Path)
		// log.Printf("[WS DEBUG] Полный URL: %s", r.URL.String())
		// log.Printf("[WS DEBUG] Query параметры: %v", r.URL.Query())
		
		// Пробуем извлечь токен разными способами
		tokenFromQuery := r.URL.Query().Get("token")
		// log.Printf("[WS DEBUG] Токен из query (token): %s", tokenFromQuery)
		
		authHeader := r.Header.Get("Authorization")
		log.Printf("[WS DEBUG] Заголовок Authorization: %s", authHeader)
		
		// пытаемся получить userID
		userID, err := auth.GetUserFromJWT(r)
		if err != nil {
			// log.Printf("[WS DEBUG] Ошибка GetUserFromJWT: %v", err)
			
			// Пробуем альтернативный способ: напрямую из query
			if tokenFromQuery != "" {
				// log.Printf("[WS DEBUG] Пробуем напрямую распарсить токен из query...")
				// Создаем временный запрос с заголовком
				r2 := r.Clone(r.Context())
				r2.Header.Set("Authorization", "Bearer "+tokenFromQuery)
				userID, err = auth.GetUserFromJWT(r2)
				// if err != nil {
				// 	log.Printf("[WS DEBUG] Ошибка при парсинге токена из query: %v", err)
				// } else {
				// 	log.Printf("[WS DEBUG] Успешно извлекли userID из query токена: %s", userID)
				// }
			}
		} else {
			// log.Printf("[WS DEBUG] Успешно извлекли userID: %s", userID)
		}
		
		if userID == "" {
			// log.Printf("[WS DEBUG] userID пустой, возвращаем 400")
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}
		
		// Проверка WebSocket заголовков
		// log.Printf("[WS DEBUG] Заголовок Upgrade: %s", r.Header.Get("Upgrade"))
		// log.Printf("[WS DEBUG] Заголовок Connection: %s", r.Header.Get("Connection"))
		
		// Апгрейд до WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// log.Printf("[WS DEBUG] Ошибка апгрейда до WebSocket: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		
		// log.Printf("[WS DEBUG] WebSocket апгрейд успешен для user: %s", userID)
		
		// Регистрируем клиента
		hub.AddClient(userID, conn)
		// log.Printf("[WS] user %s connected", userID)
		
		// Фоновая горутина для обработки соединения
		go func() {
			defer func() {
				hub.RemoveClient(userID)
				conn.Close()
				// log.Printf("[WS] user %s disconnected", userID)
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