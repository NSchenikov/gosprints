package ws

import "gosprints/internal/models"

type WSNotifier struct {
	hub *NotificationHub
}

func NewWSNotifier(hub *NotificationHub) *WSNotifier {
	return &WSNotifier{hub: hub}
}

func (n *WSNotifier) NotifyTaskStatusChanged(userID string, evt models.TaskStatusEvent) {
	n.hub.SendToUser(userID, evt)
}
