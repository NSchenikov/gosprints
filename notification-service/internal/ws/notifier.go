package ws

import (
	"log"
)

type WSNotifier struct {
	hub *NotificationHub
}

func NewWSNotifier(hub *NotificationHub) *WSNotifier {
	return &WSNotifier{hub: hub}
}

func (n *WSNotifier) Notify(event TaskStatusEvent) {
	log.Printf("[WSNotifier] Sending event: %+v", event)
	n.hub.SendToUser(event.UserID, event)
}