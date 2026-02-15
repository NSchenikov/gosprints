package notifiers

import (
    "context"
    "log"
    
    events "github.com/nschenikov/gosprints/notification-service/proto/events"
    "github.com/nschenikov/gosprints/notification-service/internal/ws"
)

type Manager struct {
    wsHub   *ws.Hub          // WebSocket hub из пакета ws старого корневого internal 
    emailer *EmailNotifier   // решить понадобится или нет?
}

func (m *Manager) NotifyTaskCreated(ctx context.Context, event *events.TaskEvent) {
    // уведомление по ws
    m.wsHub.SendToUser(event.UserId, ws.Message{
        Type: "task_created",
        Data: event,
    })
    
    // Email (если понадобится)
    m.emailer.Send(ctx, event.UserId, "Task Created", event.TaskText)
    
    log.Printf("Notification sent for task %d: %s", event.TaskId, event.EventType)
}