// sources used
// 1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e
// 3) https://www.youtube.com/watch?v=k6kiNivraJ8
// 4) https://tutorialedge.net/golang/authenticating-golang-rest-api-with-jwts/
// 5) https://www.youtube.com/watch?v=f9IrbW13C_c&t=29s
// 6) https://www.youtube.com/watch?v=wHQBMDInWEg
// 7) https://proglib.io/p/parallelizm-v-golang-i-workerpool-chast-1-2020-12-24
// 8) https://proglib.io/p/parallelizm-v-golang-i-workerpool-chast-2-2020-12-26
// 9) https://habr.com/ru/articles/948866/

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
    "gosprints/internal/handlers"
    "gosprints/internal/worker"
    qpkg "gosprints/internal/queue"
    "gosprints/internal/router"
    "gosprints/internal/scheduler"
    "gosprints/internal/services"
    "gosprints/internal/ws"
    "gosprints/internal/cache"
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
    taskRepo := repositories.NewTaskCacheRepository(baseTaskRepo, appCache)

    queue := qpkg.NewTaskQueue(100)

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    hub := ws.NewNotificationHub()
    notifier := ws.NewWSNotifier(hub)

    for i := 1; i <= 3; i++ {
		w := worker.NewWorker(i, taskRepo, queue, notifier)
		w.Start(ctx)
	}

    dispatcher := scheduler.NewDispatcher(taskRepo, queue, 5*time.Second)
    dispatcher.Start(ctx)

    taskService := services.NewTaskService(taskRepo)

    taskHandler := handlers.NewTaskHandler(taskService)
	authHandler := handlers.NewAuthHandler(userRepo)

    //хэндлер управления кэшем
    cacheHandler := handlers.NewCacheHandler(taskRepo)

    r := router.NewRouter(taskHandler, authHandler, hub, cacheHandler)

    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    go func() {
        time.Sleep(2 * time.Second)
        log.Println("[main] Warming up cache...")
        if err := taskRepo.WarmUpCache(context.Background()); err != nil {
            log.Printf("[main] Failed to warm up cache: %v", err)
        } else {
            log.Println("[main] Cache warmed up successfully")
        }
    }()

    go func() {
        fmt.Println("Сервер запущен на http://localhost:8080")
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