package router

import (
    "net/http"

    "api-gateway/internal/handlers"
    "api-gateway/internal/middleware"
)

func NewRouter(taskProxy *handlers.TaskProxyHandler, authHandler *handlers.AuthHandler,) *http.ServeMux {
    r := http.NewServeMux()

    //публичные
    r.HandleFunc("POST /login",    authHandler.Login)
    r.HandleFunc("POST /register", authHandler.Register)

    //защищенные
    r.Handle("GET /tasks",       middleware.AuthMiddleware(taskProxy.GetTasks))
    r.Handle("POST /tasks",      middleware.AuthMiddleware(taskProxy.CreateTask))
    r.Handle("GET /tasks/{id}",  middleware.AuthMiddleware(taskProxy.GetTaskByID))
    r.Handle("PUT /tasks/{id}",  middleware.AuthMiddleware(taskProxy.UpdateTask))
    r.Handle("DELETE /tasks/{id}", middleware.AuthMiddleware(taskProxy.DeleteTask))
    r.HandleFunc("GET /tasks/search", middleware.AuthMiddleware(taskProxy.SearchTasks))
    r.HandleFunc("POST /tasks/{id}/close", middleware.AuthMiddleware(taskProxy.CloseTask))
    r.HandleFunc("GET /users/{user_id}/tasks", middleware.AuthMiddleware(taskProxy.GetUserTasks))

    //метрики и кэш пока оставлю закомментированными
    // r.HandleFunc("GET /admin/cache/stats", cacheHandler.GetCacheStats)
    // r.HandleFunc("POST /admin/cache/clear", cacheHandler.ClearCache)
    // r.HandleFunc("POST /admin/cache/warmup", cacheHandler.WarmUpCache)

    // r.HandleFunc("GET /metrics", metricsHandler.GetMetrics)
    // r.HandleFunc("GET /metrics/prometheus", metricsHandler.GetPrometheusMetrics)

    return r
}
