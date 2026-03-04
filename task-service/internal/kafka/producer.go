package kafka

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/segmentio/kafka-go"
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/known/timestamppb"
    
    events "github.com/nschenikov/gosprints/task-service/proto/events"
)

type TaskEventProducer struct {
    writer *kafka.Writer
}

func NewTaskEventProducer(brokers []string, topic string) *TaskEventProducer {
    return &TaskEventProducer{
        writer: &kafka.Writer{
            Addr:     kafka.TCP(brokers...),
            Topic:    topic,
            Balancer: &kafka.LeastBytes{},
        },
    }
}

func (p *TaskEventProducer) PublishTaskEvent(ctx context.Context, eventType string, taskID int32, text, status, userID string) error {
    event := &events.TaskEvent{
        EventId:    fmt.Sprintf("%d-%d", taskID, time.Now().UnixNano()),
        EventType:  eventType,
        TaskId:     taskID,
        TaskText:   text,
        TaskStatus: status,
        UserId:     userID,
        Timestamp:  timestamppb.Now(),
    }
    
    data, err := proto.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }
    
    err = p.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(fmt.Sprintf("task-%d", taskID)),
        Value: data,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte(eventType)},
        },
    })
    
    if err != nil {
        return fmt.Errorf("failed to write message: %w", err)
    }
    
    return nil
}

func (p *TaskEventProducer) Close() error {
    return p.writer.Close()
}