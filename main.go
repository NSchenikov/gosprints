// sources used
// 1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e
// 3) https://www.youtube.com/watch?v=k6kiNivraJ8
// 4) https://tutorialedge.net/golang/authenticating-golang-rest-api-with-jwts/
// 5) https://www.youtube.com/watch?v=f9IrbW13C_c&t=29s
// 6) https://www.youtube.com/watch?v=wHQBMDInWEg
// 7) https://proglib.io/p/parallelizm-v-golang-i-workerpool-chast-1-2020-12-24
// 8) https://proglib.io/p/parallelizm-v-golang-i-workerpool-chast-2-2020-12-26

package main

import (
    "fmt"
    "log"
    "net/http"

    "gosprints/pkg/database"
    "gosprints/internal/repositories"
    "gosprints/internal/handlers"
    "gosprints/internal/worker"
    qpkg "gosprints/internal/queue"
    "gosprints/internal/router"
)

func main() {

    db := database.InitDB()
    defer db.Close()

    taskRepo := repositories.NewTaskRepository(db)
    userRepo := repositories.NewUserRepository(db)

    queue := qpkg.NewTaskQueue(100)

    for i := 1; i <= 3; i++ {
		w := worker.NewWorker(i, taskRepo, queue)
		w.Start()
	}

    taskHandler := handlers.NewTaskHandler(taskRepo, queue)
	authHandler := handlers.NewAuthHandler(userRepo)

    r := router.NewRouter(taskHandler, authHandler)

    fmt.Println("Сервер запущен на http://localhost:8080")

    log.Fatal(http.ListenAndServe(":8080", r))
}