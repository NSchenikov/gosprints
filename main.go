//sources used
//1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e

package main

import (
    "encoding/json"
    "fmt"
    "log"
	"strings"
	"strconv"
    "net/http"
    "database/sql"

    _ "github.com/lib/pq"
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

func writingTaskHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
    idStr := strings.Split(path, "/")[0] // способ достать id из эндпоинта
	id, _ := strconv.Atoi(idStr) //переводим string в int игнорируя второй параметр Atoi

    for i, task := range tasks {
		idx, _ := strconv.Atoi(task.ID)
        if idx == id {
            json.NewDecoder(r.Body).Decode(&tasks[i])
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(tasks[i])
            return
        }
    }

    http.Error(w, "Задача с указанным id не найдена", http.StatusNotFound)
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

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
    idStr := strings.Split(path, "/")[0] 
	id, _ := strconv.Atoi(idStr) 

	var updatedTask Task
	json.NewDecoder(r.Body).Decode(&updatedTask)

    for i, task := range tasks {
		idx, _ := strconv.Atoi(task.ID)
        if idx == id {
            tasks[i].Text = updatedTask.Text
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(tasks[i])
            return
        }
    }

    http.Error(w, "Задача с указанным id не найдена", http.StatusNotFound)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
    idStr := strings.Split(path, "/")[0]
	id, _ := strconv.Atoi(idStr)

    for i, task := range tasks {
		idx, _ := strconv.Atoi(task.ID)
        if idx == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
            w.WriteHeader(http.StatusNoContent)
            return
        }
    }

    http.Error(w, "Задача с указанным id не найдена", http.StatusNotFound)
}

func main() {

    connStr := "postgresql://postgres:4840707101@localhost:8000/gosprints?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка открытия соединения:", err)
		return
	}
	defer db.Close()

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		fmt.Println("Ошибка пинга базы данных:", err)
		return
	}

	fmt.Println("Успешное подключение к PostgreSQL!")

	r := http.NewServeMux()

    r.HandleFunc("GET /tasks", getTasksHandler)
	r.HandleFunc("POST /tasks", createTaskHandler)
	r.HandleFunc("GET /tasks/{id}", writingTaskHandler)
	r.HandleFunc("PUT /tasks/{id}", updateTaskHandler)
	r.HandleFunc("DELETE /tasks/{id}", deleteTaskHandler)

    fmt.Println("Сервер запущен на http://localhost:8080")
    fmt.Println("Для проверки откройте браузер или используйте curl http://localhost:8080/tasks")
	fmt.Println(`Для добавления задачи curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"text": "Задача"}'`)
	fmt.Println("Для чтения задачи curl http://localhost:8080/tasks/{id}")
	fmt.Println(`Для обновления задачи curl -X PUT http://localhost:8080/tasks/{id} -H "Content-Type: application/json" -d '{"text": "Обновленный текст задачи"}'`)
	fmt.Println("Для удаления задачи curl -X DELETE http://localhost:8080/tasks/{id}")
    
    log.Fatal(http.ListenAndServe(":8080", r))
}