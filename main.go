//sources used
//1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)


type Task struct {
    ID    string `json:"id"`
    Text  string `json:"text"`
}

var tasks = []Task{}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
    var task Task
    json.NewDecoder(r.Body).Decode(&task)
    task.ID = fmt.Sprintf("%d", len(tasks)+1)
    tasks = append(tasks, task)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}

func main() {

	r := http.NewServeMux()

    r.HandleFunc("GET /tasks", getTasksHandler)
	r.HandleFunc("POST /tasks", createTaskHandler)

    fmt.Println("Сервер запущен на http://localhost:8080")
    fmt.Println("Для проверки откройте браузер или используйте curl http://localhost:8080/tasks")
	fmt.Println(`Для добавления задачи curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"text": "Задача"}'`)
    
    log.Fatal(http.ListenAndServe(":8080", r))
}