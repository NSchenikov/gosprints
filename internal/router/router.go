package router

import (
    "net/http"

    "gosprints/internal/handlers"
    "gosprints/internal/ws"
)

func NewRouter(taskHandler *handlers.TaskHandler, authHandler *handlers.AuthHandler, hub *ws.NotificationHub, cacheHandler *handlers.CacheHandler, metricsHandler *handlers.MetricsHandler) *http.ServeMux {
    r := http.NewServeMux()

    r.Handle("GET /tasks",       authHandler.AuthMiddleware(taskHandler.GetTasks))
    r.Handle("POST /tasks",      authHandler.AuthMiddleware(taskHandler.CreateTask))
    r.Handle("GET /tasks/{id}",  authHandler.AuthMiddleware(taskHandler.GetTaskByID))
    r.Handle("PUT /tasks/{id}",  authHandler.AuthMiddleware(taskHandler.UpdateTask))
    r.Handle("DELETE /tasks/{id}", authHandler.AuthMiddleware(taskHandler.DeleteTask))


    wsHandler := ws.NewWSHandler(hub)
    r.HandleFunc("GET /ws", wsHandler)

    r.HandleFunc("GET /admin/cache/stats", cacheHandler.GetCacheStats)
    r.HandleFunc("POST /admin/cache/clear", cacheHandler.ClearCache)
    r.HandleFunc("POST /admin/cache/warmup", cacheHandler.WarmUpCache)

    r.HandleFunc("GET /metrics", metricsHandler.GetMetrics)
    r.HandleFunc("GET /metrics/prometheus", metricsHandler.GetPrometheusMetrics)

    r.HandleFunc("POST /login",    authHandler.Login)
    r.HandleFunc("POST /register", authHandler.Register)

    return r
}
