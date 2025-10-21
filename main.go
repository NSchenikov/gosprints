//sources used
//1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e
// 3) https://www.youtube.com/watch?v=k6kiNivraJ8
// 4) https://tutorialedge.net/golang/authenticating-golang-rest-api-with-jwts/

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
	r.Handle("DELETE /tasks/{id}", authHandler.AuthMiddleware(taskHandler.DeleteTask))

    r.HandleFunc("POST /login", authHandler.Login)
    r.HandleFunc("POST /register", authHandler.Register)

    fmt.Println("Сервер запущен на http://localhost:8080")

    fmt.Println(`Регистрация curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"testuser", "password":"password123"}'`)
    fmt.Println("\n Получить токен:")
    fmt.Println(`   curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"username":"admin","password":"123456"}'`)

    fmt.Println("\n Использовать токен для доступа:")
    fmt.Println("   Заменить YOUR_TOKEN на полученный токен")

    fmt.Println("\n 1) Все задачи:")
    fmt.Println(`      curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks`)

    fmt.Println("   2) Добавить задачу:")
    fmt.Println(`      curl -X POST http://localhost:8080/tasks -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{"text":"Задача"}'`)

    fmt.Println("   3) Прочитать задачу по id:")
    fmt.Println(`      curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks/1`)

    fmt.Println("   4) Обновить задачу:")
    fmt.Println(`      curl -X PUT http://localhost:8080/tasks/1 -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{"text":"Новый текст"}'`)

    fmt.Println("   5) Удалить задачу:")
    fmt.Println(`      curl -X DELETE -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks/1`)


    log.Fatal(http.ListenAndServe(":8080", r))
}