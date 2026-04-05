package kafka

import (
    "context"
    "fmt"
    "time"
    "log"

    "github.com/segmentio/kafka-go"
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/known/timestamppb"
    
    events "task-service/proto/events"
)

type TaskEventProducer struct {
    writer *kafka.Writer
}

func NewTaskEventProducer(brokers []string, topic string) *TaskEventProducer {
    log.Printf("[Kafka] Initializing producer: brokers=%v, topic=%s", brokers, topic)
    
    if len(brokers) == 0 || brokers[0] == "" {
        log.Println("[Kafka] No brokers provided, producer disabled")
        return nil
    }
    
    return &TaskEventProducer{
        writer: &kafka.Writer{
            Addr:     kafka.TCP(brokers...),
            Topic:    topic,
            Balancer: &kafka.LeastBytes{},
        },
    }
}

func (p *TaskEventProducer) PublishTaskEvent(ctx context.Context, eventType string, taskID int32, text, status, userID string) error {

    if p.writer == nil {
        log.Println("[Kafka] Producer not configured, skipping event")
        return nil
    }
    
    log.Printf("[Kafka] Publishing event: type=%s, taskID=%d, userID=%s", eventType, taskID, userID)

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
    
    //таймаут для Kafka
    kafkaCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    err = p.writer.WriteMessages(kafkaCtx, kafka.Message{
        Key:   []byte(fmt.Sprintf("task-%d", taskID)),
        Value: data,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte(eventType)},
        },
    })
    
    if err != nil {
        log.Printf("[Kafka] Write error: %v", err)
        return fmt.Errorf("failed to write message: %w", err)
    }
    
    log.Printf("[Kafka] Event published successfully")
    return nil
}

func (p *TaskEventProducer) Close() error {
    return p.writer.Close()
}