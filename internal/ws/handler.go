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
		userID, err := auth.GetUserFromJWT(r)
		if userID == "" {
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[ws] upgrade error: %v", err)
			return
		}

		hub.AddClient(userID, conn)
		log.Printf("[ws] user %s connected", userID)

		// ловим момент дисконнекта
		go func() {
			defer func() {
				hub.RemoveClient(userID)
				conn.Close()
				log.Printf("[ws] user %s disconnected", userID)
			}()

			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
	}
}
