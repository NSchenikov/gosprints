package router

import (
    "net/http"

    "gosprints/internal/handlers"
)

func NewRouter(taskHandler *handlers.TaskHandler, authHandler *handlers.AuthHandler) *http.ServeMux {
    r := http.NewServeMux()

    r.Handle("GET /tasks",       authHandler.AuthMiddleware(taskHandler.GetTasks))
    r.Handle("POST /tasks",      authHandler.AuthMiddleware(taskHandler.CreateTask))
    r.Handle("GET /tasks/{id}",  authHandler.AuthMiddleware(taskHandler.GetTaskByID))
    r.Handle("PUT /tasks/{id}",  authHandler.AuthMiddleware(taskHandler.UpdateTask))
    r.Handle("DELETE /tasks/{id}", authHandler.AuthMiddleware(taskHandler.DeleteTask))

    r.HandleFunc("POST /login",    authHandler.Login)
    r.HandleFunc("POST /register", authHandler.Register)

    return r
}
