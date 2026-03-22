package notifiers

import (
    "context"
    "log"
    
    "notification-service/internal/ws"
    events "notification-service/proto/events"
)

type Manager struct {
    wsHub *ws.NotificationHub  // используем NotificationHub, а не Hub
    // emailer *EmailNotifier   // уведомлений пока не будет
}

func NewManager(wsHub *ws.NotificationHub) *Manager {
    return &Manager{
        wsHub: wsHub,
        // emailer: emailer,
    }
}

func (m *Manager) NotifyTaskCreated(ctx context.Context, event *events.TaskEvent) {
    // WebSocket уведомление
    wsEvent := ws.TaskStatusEvent{
        Type:      "task_created",
        TaskID:    int(event.TaskId),
        Text:      event.TaskText,
        Status:    event.TaskStatus,
        UserID:    event.UserId,
        Timestamp: event.Timestamp.AsTime().String(),
    }
    
    m.wsHub.SendToUser(event.UserId, wsEvent)
    log.Printf("Notification sent for task %d: %s", event.TaskId, event.EventType)
}

func (m *Manager) NotifyTaskUpdated(ctx context.Context, event *events.TaskEvent) {
    wsEvent := ws.TaskStatusEvent{
        Type:      "task_updated",
        TaskID:    int(event.TaskId),
        Text:      event.TaskText,
        Status:    event.TaskStatus,
        UserID:    event.UserId,
        Timestamp: event.Timestamp.AsTime().String(),
    }
    
    m.wsHub.SendToUser(event.UserId, wsEvent)
    log.Printf("Notification sent for task %d: %s", event.TaskId, event.EventType)
}

func (m *Manager) NotifyTaskCompleted(ctx context.Context, event *events.TaskEvent) {
    wsEvent := ws.TaskStatusEvent{
        Type:      "task_completed",
        TaskID:    int(event.TaskId),
        Text:      event.TaskText,
        Status:    event.TaskStatus,
        UserID:    event.UserId,
        Timestamp: event.Timestamp.AsTime().String(),
    }
    
    m.wsHub.SendToUser(event.UserId, wsEvent)
    log.Printf("Notification sent for task %d: %s", event.TaskId, event.EventType)
}

func (m *Manager) NotifyTaskDeleted(ctx context.Context, event *events.TaskEvent) {
    wsEvent := ws.TaskStatusEvent{
        Type:      "task_deleted",
        TaskID:    int(event.TaskId),
        Text:      event.TaskText,
        Status:    event.TaskStatus,
        UserID:    event.UserId,
        Timestamp: event.Timestamp.AsTime().String(),
    }
    
    m.wsHub.SendToUser(event.UserId, wsEvent)
    log.Printf("Notification sent for task %d: %s", event.TaskId, event.EventType)
}