
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
    log.Printf("[Kafka] Creating consumer: brokers=%v, topic=%s, groupID=%s", brokers, topic, groupID)
    
    if len(brokers) == 0 || brokers[0] == "" {
        log.Println("[Kafka] WARNING: No brokers provided, consumer will not work")
        return &TaskEventConsumer{}
    }
    
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  brokers,
        Topic:    topic,
        GroupID:  groupID,
        MinBytes: 10e3,
        MaxBytes: 10e6,
    })
    
    return &TaskEventConsumer{
        reader:   reader,
        notifier: notifiers.NewManager(hub),
    }
}

func (c *TaskEventConsumer) Start(ctx context.Context) {
    if c.reader == nil {
        log.Println("[Kafka] No reader configured, consumer stopped")
        return
    }
    
    log.Println("[Kafka] Starting consumer...")
    
    for {
        msg, err := c.reader.ReadMessage(ctx)
        if err != nil {
            if err == context.Canceled {
                log.Println("[Kafka] Consumer stopped")
                return
            }
            log.Printf("[Kafka] Error reading message: %v", err)
            continue
        }
        
        log.Printf("[Kafka] Received message: topic=%s, partition=%d, offset=%d", 
            msg.Topic, msg.Partition, msg.Offset)
        
        var event events.TaskEvent
        if err := proto.Unmarshal(msg.Value, &event); err != nil {
            log.Printf("[Kafka] Failed to unmarshal event: %v", err)
            continue
        }
        
        log.Printf("[Kafka] Event: type=%s, taskID=%d, userID=%s", 
            event.EventType, event.TaskId, event.UserId)
        
        switch event.EventType {
        case "CREATED":
            c.notifier.NotifyTaskCreated(ctx, &event)
        case "UPDATED":
            c.notifier.NotifyTaskUpdated(ctx, &event)
        case "COMPLETED":
            c.notifier.NotifyTaskCompleted(ctx, &event)
        case "DELETED":
            c.notifier.NotifyTaskDeleted(ctx, &event)
        }
    }
}

func (c *TaskEventConsumer) Close() error {
    if c.reader != nil {
        log.Println("[Kafka] Closing consumer...")
        return c.reader.Close()
    }
    return nil
}
