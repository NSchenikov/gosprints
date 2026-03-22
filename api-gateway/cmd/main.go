package main

import (
    "log"
    "net/http"
    
    "api-gateway/internal/grpc/client"
    "api-gateway/internal/handlers"
    "api-gateway/internal/router"
)

func main() {
    // Подключение к task-service по gRPC
    taskClient, err := client.NewTaskClient("task-service:50051")
    if err != nil {
        log.Fatal(err)
    }
    defer taskClient.Close()
    
    // прокси-хендлеры
    taskProxy := handlers.NewTaskProxyHandler(taskClient)
    authHandler := handlers.NewAuthHandler()
    
    // Настраиваем роутер
    r := router.NewRouter(taskProxy, authHandler)
    
    log.Println("API Gateway starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}