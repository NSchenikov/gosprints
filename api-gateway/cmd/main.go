package main

import (
    "log"
    "net/http"
    
    "api-gateway/internal/grpc/client"
    "api-gateway/internal/handlers"
    "api-gateway/internal/router"
    "api-gateway/pkg/auth"
)

func main() {
    // Подключение к task-service по gRPC
    taskClient, err := client.NewTaskClient("task-service:50051")
    if err != nil {
        log.Fatal(err)
    }
    defer taskClient.Close()
    
    // прокси-хендлеры
    taskProxy := &handlers.TaskProxyHandler{
        taskClient: taskClient,
    }
    
    authHandler := &handlers.AuthHandler{
        // заглушка. Сюда нужен userRepo, но пока не понимаю где должны быть users
        // Может, auth тоже вынести в отдельный микросервис?
    }
    
    // Настраиваем роутер
    r := router.NewRouter(taskProxy, authHandler)
    
    log.Println("API Gateway starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}