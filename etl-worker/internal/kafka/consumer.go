package kafka

import (
    "context"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
    "google.golang.org/protobuf/proto"
    
    events "etl-worker/proto/events"
    "etl-worker/internal/processor"
)

type ETLConsumer struct {
    reader    *kafka.Reader
    processor *processor.EventProcessor
}

func NewETLConsumer(brokers []string, topic string, groupID string, processor *processor.EventProcessor) *ETLConsumer {
    log.Printf("[ETL] Creating consumer: brokers=%v, topic=%s, groupID=%s", brokers, topic, groupID)
    
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:     brokers,
        Topic:       topic,
        GroupID:     groupID,
        MinBytes:    10e3,
        MaxBytes:    10e6,
        MaxWait:     1 * time.Second,
        StartOffset: kafka.LastOffset,
    })
    
    return &ETLConsumer{
        reader:    reader,
        processor: processor,
    }
}

func (c *ETLConsumer) Start(ctx context.Context) {
    log.Println("[ETL] Starting consumer...")
    
    for {
        msg, err := c.reader.ReadMessage(ctx)
        if err != nil {
            if err == context.Canceled {
                log.Println("[ETL] Consumer stopped")
                return
            }
            log.Printf("[ETL] Error reading message: %v", err)
            continue
        }
        
        log.Printf("[ETL] Received message: topic=%s, partition=%d, offset=%d", 
            msg.Topic, msg.Partition, msg.Offset)
        
        var event events.TaskEvent
        if err := proto.Unmarshal(msg.Value, &event); err != nil {
            log.Printf("[ETL] Failed to unmarshal event: %v", err)
            continue
        }
        
        // Обрабатываем события
        if err := c.processor.ProcessEvent(ctx, &event); err != nil {
            log.Printf("[ETL] Failed to process event: %v", err)
            // TODO: добавить retry механику
            continue
        }
        
        log.Printf("[ETL] Processed event: type=%s, taskID=%d, userID=%s", 
            event.EventType, event.TaskId, event.UserId)
    }
}

func (c *ETLConsumer) Close() error {
    return c.reader.Close()
}