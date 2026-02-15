
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
    "context"
    "os"
    "os/signal"
    "syscall"

    "gosprints/pkg/database"
    "gosprints/internal/repositories"
    "gosprints/task-service/internal/handlers" //подправил импорт хэндлеров, но пока не перенес весь старый проект в текущий сервис
    "gosprints/internal/worker"
    qpkg "gosprints/internal/queue"
    "gosprints/internal/router"
    "gosprints/internal/scheduler"
    // "gosprints/internal/services"
    "gosprints/internal/ws"
    "gosprints/internal/cache"
    "gosprints/internal/middleware"
    "gosprints/internal/grpc/task/client"
	"gosprints/internal/grpc/task/server"
    "gosprints/task-service/internal/kafka" //kafka из текущего сервиса
)

func main() {

    db := database.InitDB()
    defer db.Close()

    cacheConfig := cache.CacheConfig{
        DefaultTTL:      10 * time.Minute,
        CleanupInterval: 5 * time.Minute,
        MaxItems:        10000,
    }
    
    appCache := cache.NewMemoryCache(cacheConfig)
    defer appCache.Stop()

    baseTaskRepo := repositories.NewTaskRepository(db)
    userRepo := repositories.NewUserRepository(db)

    //обернули в кэширующий репозиторий
    // taskRepo := repositories.NewTaskCacheRepository(baseTaskRepo, appCache)
    apiTaskRepo := repositories.NewTaskCacheRepository(baseTaskRepo, appCache)
    workerTaskRepo := baseTaskRepo

    queue := qpkg.NewTaskQueue(100)

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    hub := ws.NewNotificationHub()
    notifier := ws.NewWSNotifier(hub)

    for i := 1; i <= 3; i++ {
		w := worker.NewWorker(i, workerTaskRepo, queue, notifier)
		w.Start(ctx)
	}

    dispatcher := scheduler.NewDispatcher(workerTaskRepo, queue, 30*time.Second)
    dispatcher.Start(ctx)

    //!Запуск gRPC Task Service ===
    log.Println("[main] Запуск gRPC Task Service...")
    go func() {
        if err := server.StartServer(baseTaskRepo, "50051"); err != nil {
            log.Fatalf("[main] Ошибка запуска gRPC сервера: %v", err)
        }
    }()
    time.Sleep(2 * time.Second)

    log.Println("[main] Подключение к gRPC Task Service...")
    taskClient, err := client.NewTaskClient("localhost:50051")
    if err != nil {
        log.Fatalf("[main] Ошибка создания gRPC клиента: %v", err)
    }
    defer taskClient.Close()

    //!Используем gRPC клиент вместо прямого сервиса ===
    // taskService := services.NewTaskService(apiTaskRepo)

    // taskHandler := handlers.NewTaskHandler(taskService)

    // Kafka producer
    kafkaProducer := kafka.NewTaskEventProducer(
        []string{"localhost:9092"},
        "task-events",
    )
    defer kafkaProducer.Close()

    // Создание taskHandler с kafka продюсером вместо hub
    taskHandler := handlers.NewTaskHandler(taskClient, kafkaProducer)

	authHandler := handlers.NewAuthHandler(userRepo)

    //хэндлер управления кэшем
    cacheHandler := handlers.NewCacheHandler(apiTaskRepo)

    metricsHandler := handlers.NewMetricsHandler(hub, apiTaskRepo, appCache)

    r := router.NewRouter(taskHandler, authHandler, hub, cacheHandler, metricsHandler)

    handler := middleware.Metrics(r)

    srv := &http.Server{
        Addr:    ":8080",
        Handler: handler,
    }

    go func() {
        time.Sleep(2 * time.Second)
        log.Println("[main] Warming up cache...")
        if err := apiTaskRepo.WarmUpCache(context.Background()); err != nil {
            log.Printf("[main] Failed to warm up cache: %v", err)
        } else {
            log.Println("[main] Cache warmed up successfully")
        }
    }()

    go func() {
        fmt.Println("Сервер запущен на http://localhost:8080")
        log.Println("[main] gRPC сервер запущен на порту 50051")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe error: %v", err)
        }
    }()

    <-ctx.Done()
    log.Println("[main] shutdown signal received")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("[main] HTTP server Shutdown error: %v", err)
    } else {
        log.Println("[main] HTTP server stopped gracefully")
    }

    log.Println("[main] waiting a bit for workers to finish...")
    time.Sleep(1 * time.Second)
    log.Println("[main] shutdown complete")
}

//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJleHAiOjE3NzIxMDc5MDUsInVzZXIiOiJ1c2VyIn0.7qwYIZhaQxRUw7lu6f7pBqrgz5PjbcrmHuzTUqwfZVg