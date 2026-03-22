package main

import (
    "log"
    "net/http"
    
    "api-gateway/internal/grpc/client"
    "api-gateway/internal/handlers"
    "api-gateway/internal/router"
)

func main() {
    log.Println("Starting API Gateway...")
    
    log.Println("Connecting to task-service...")
    // Подключение к task-service по gRPC
    taskClient, err := client.NewTaskClient("localhost:50051")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer taskClient.Close()
    log.Println("Connected to task-service")
    
    log.Println("Creating handlers...")
    // прокси-хендлеры
    taskProxy := handlers.NewTaskProxyHandler(taskClient)
    authHandler := handlers.NewAuthHandler()
    log.Println("Handlers created")
    
    log.Println("Setting up router...")
    // Настраиваем роутер
    r := router.NewRouter(taskProxy, authHandler)
    log.Println("Router created")
    
    log.Println("API Gateway starting on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}