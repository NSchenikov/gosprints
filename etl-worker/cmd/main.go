package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "etl-worker/internal/kafka"
    "etl-worker/internal/processor"
    "etl-worker/internal/storage"
)

func main() {
    // Подключаемс к ClickHouse
    ch, err := storage.NewClickHouseStorage("localhost:9000", "default")
    if err != nil {
        log.Fatal(err)
    }
    defer ch.Close()
    
    // Создание процессора
    eventProcessor := processor.NewEventProcessor(ch)
    
    // Создание Kafka consumer
    consumer := kafka.NewETLConsumer(
        []string{os.Getenv("KAFKA_BROKERS")},
        os.Getenv("KAFKA_TOPIC"),
        os.Getenv("ETL_GROUP_ID"),
        eventProcessor,
    )
    
    // Запускаем consumer
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go consumer.Start(ctx)

    log.Println("ETL Worker started")
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down ETL worker...")
    consumer.Close()
}