// sources used
// 1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e
// 3) https://www.youtube.com/watch?v=k6kiNivraJ8
// 4) https://tutorialedge.net/golang/authenticating-golang-rest-api-with-jwts/
// 5) https://www.youtube.com/watch?v=f9IrbW13C_c&t=29s

package main

import (
    "fmt"
    "log"
    "net/http"

    "gosprints/pkg/database"
    "gosprints/internal/repositories"
    "gosprints/internal/handlers"
)

func main() {

    db := database.InitDB()
    defer db.Close()

    taskRepo := repositories.NewTaskRepository(db)
    userRepo := repositories.NewUserRepository(db)

    taskHandler := handlers.NewTaskHandler(taskRepo)
	authHandler := handlers.NewAuthHandler(userRepo)

	r := http.NewServeMux()

    r.Handle("GET /tasks", authHandler.AuthMiddleware(taskHandler.GetTasks))
	r.Handle("POST /tasks", authHandler.AuthMiddleware(taskHandler.CreateTask))
	r.Handle("GET /tasks/{id}", authHandler.AuthMiddleware(taskHandler.GetTaskByID))
	r.Handle("PUT /tasks/{id}", authHandler.AuthMiddleware(taskHandler.UpdateTask))
	r.Handle("DELETE /tasks/{id}", authHandler.AuthMiddleware(taskHandler.DeleteTask)) //вынести в роутер

    r.HandleFunc("POST /login", authHandler.Login)
    r.HandleFunc("POST /register", authHandler.Register)

    fmt.Println("Сервер запущен на http://localhost:8080")

    log.Fatal(http.ListenAndServe(":8080", r))
}