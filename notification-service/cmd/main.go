
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    
    "notification-service/internal/handlers"
    "notification-service/internal/kafka"
    "notification-service/internal/ws"
)

func main() {
    // создание WebSocket hub
    hub := ws.NewNotificationHub()
    
    // создаем Kafka consumer
    consumer := kafka.NewTaskEventConsumer(
        []string{os.Getenv("KAFKA_BROKERS")},
        os.Getenv("KAFKA_TOPIC"),
        os.Getenv("KAFKA_GROUP_ID"),
        hub,
    )
    
    // запуск consumer в фоне
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go consumer.Start(ctx)
    
   // Создаем WebSocket handler
    wsHandler := handlers.NewWSHandler(hub)
    
    // Настраиваем маршруты
    http.HandleFunc("/ws", wsHandler.ServeWS)
    
    // запуск HTTP сервера для WebSocket
    go func() {
        log.Println("Notification service started on :8082")
        if err := http.ListenAndServe(":8082", nil); err != nil {
            log.Fatal(err)
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down...")
    consumer.Close()
}