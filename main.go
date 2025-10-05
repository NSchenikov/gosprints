//sources used
//1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e

package main

import (
    "encoding/json"
    "fmt"
    "log"
	// "strings"
	// "strconv"
    "net/http"
    "database/sql"

    _ "github.com/lib/pq"
)


type Task struct {
    ID    string `json:"id"`
    Text  string `json:"text"`
}

var db *sql.DB

func initDB() {
    connStr := "postgresql://postgres:4840707101@localhost:8000/gosprints?sslmode=disable"
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Error opening database:", err)
    }

    if err = db.Ping(); err != nil {
        log.Fatal("Error connecting to database:", err)
    }
    fmt.Println("Connected to PostgreSQL!")
}

func getTasksHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        
        if r.Method != "GET" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        rows, err := db.Query("SELECT id, text FROM \"Tasks\"") //уточняем что таблица называется Tasks
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var tasks []Task
        for rows.Next() {
            var task Task
            if err := rows.Scan(&task.ID, &task.Text); err != nil {
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
            }
            tasks = append(tasks, task)
        }
        
        if err = rows.Err(); err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(tasks); err != nil {
            return
        }
    }
}

func writingTaskHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        
        if r.Method != "GET" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        id := r.URL.Path[len("/tasks/"):]
        if id == "" {
            http.Error(w, "Task ID is required", http.StatusBadRequest)
            return
        }
        
        var task Task
        err := db.QueryRow("SELECT id, text FROM \"Tasks\" WHERE id = $1", id).Scan(&task.ID, &task.Text)
        if err != nil {
            if err == sql.ErrNoRows {
                fmt.Printf("Task with ID %s not found\n", id)
                http.Error(w, "Task not found", http.StatusNotFound)
                return
            }
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(task); err != nil {
            return
        }
    }
}

// func createTaskHandler(w http.ResponseWriter, r *http.Request) {
//     var task Task
//     json.NewDecoder(r.Body).Decode(&task)
//     task.ID = fmt.Sprintf("%d", len(tasks)+1)
//     tasks = append(tasks, task)

//     w.Header().Set("Content-Type", "application/json")
//     w.WriteHeader(http.StatusCreated)
//     json.NewEncoder(w).Encode(task)
// }

// func updateTaskHandler(w http.ResponseWriter, r *http.Request) {

// 	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
//     idStr := strings.Split(path, "/")[0] 
// 	id, _ := strconv.Atoi(idStr) 

// 	var updatedTask Task
// 	json.NewDecoder(r.Body).Decode(&updatedTask)

//     for i, task := range tasks {
// 		idx, _ := strconv.Atoi(task.ID)
//         if idx == id {
//             tasks[i].Text = updatedTask.Text
//             w.Header().Set("Content-Type", "application/json")
//             json.NewEncoder(w).Encode(tasks[i])
//             return
//         }
//     }

//     http.Error(w, "Задача с указанным id не найдена", http.StatusNotFound)
// }

// func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {

// 	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
//     idStr := strings.Split(path, "/")[0]
// 	id, _ := strconv.Atoi(idStr)

//     for i, task := range tasks {
// 		idx, _ := strconv.Atoi(task.ID)
//         if idx == id {
// 			tasks = append(tasks[:i], tasks[i+1:]...)
//             w.WriteHeader(http.StatusNoContent)
//             return
//         }
//     }

//     http.Error(w, "Задача с указанным id не найдена", http.StatusNotFound)
// }

func main() {

    initDB()
    defer db.Close()

	r := http.NewServeMux()

    r.HandleFunc("GET /tasks", getTasksHandler(db))
	// r.HandleFunc("POST /tasks", createTaskHandler)
	r.HandleFunc("GET /tasks/{id}", writingTaskHandler(db))
	// r.HandleFunc("PUT /tasks/{id}", updateTaskHandler)
	// r.HandleFunc("DELETE /tasks/{id}", deleteTaskHandler)

    fmt.Println("Сервер запущен на http://localhost:8080")
    fmt.Println("Для проверки откройте браузер или используйте curl http://localhost:8080/tasks")
	// fmt.Println(`Для добавления задачи curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"text": "Задача"}'`)
	fmt.Println("Для чтения задачи curl http://localhost:8080/tasks/{id}")
	// fmt.Println(`Для обновления задачи curl -X PUT http://localhost:8080/tasks/{id} -H "Content-Type: application/json" -d '{"text": "Обновленный текст задачи"}'`)
	// fmt.Println("Для удаления задачи curl -X DELETE http://localhost:8080/tasks/{id}")
    
    log.Fatal(http.ListenAndServe(":8080", r))
}