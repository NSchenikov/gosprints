package kafka

import (
    "context"
    "log"

    "github.com/segmentio/kafka-go"
    "google.golang.org/protobuf/proto"
    
    events "notification-service/proto/events"
    "notification-service/internal/notifiers"
    "notification-service/internal/ws"
)

type TaskEventConsumer struct {
    reader   *kafka.Reader
    notifier *notifiers.Manager
}

func NewTaskEventConsumer(brokers []string, topic string, groupID string, hub *ws.NotificationHub) *TaskEventConsumer {
    return &TaskEventConsumer{
        reader: kafka.NewReader(kafka.ReaderConfig{
            Brokers:  brokers,
            Topic:    topic,
            GroupID:  groupID,
            MinBytes: 10e3,
            MaxBytes: 10e6,
        }),
        notifier: notifiers.NewManager(hub),  // передаём hub
    }
}

func (c *TaskEventConsumer) Start(ctx context.Context) {
    for {
        msg, err := c.reader.ReadMessage(ctx)
        if err != nil {
            log.Printf("Error reading message: %v", err)
            continue
        }
        
        var event events.TaskEvent
        if err := proto.Unmarshal(msg.Value, &event); err != nil {
            log.Printf("Error unmarshaling event: %v", err)
            continue
        }
        
        // Определяется тип события и отправляется уведомление
        switch event.EventType {
        case "CREATED":
            c.notifier.NotifyTaskCreated(ctx, &event)
        case "COMPLETED":
            c.notifier.NotifyTaskCompleted(ctx, &event)
        case "UPDATED":
            c.notifier.NotifyTaskUpdated(ctx, &event)
        case "DELETED":
            c.notifier.NotifyTaskDeleted(ctx, &event)
        }
    }
}

func (c *TaskEventConsumer) Close() error {
    return c.reader.Close()
}