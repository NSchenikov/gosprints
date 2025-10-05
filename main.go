//sources used
//1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e
// 3) https://www.youtube.com/watch?v=k6kiNivraJ8
// 4) https://tutorialedge.net/golang/authenticating-golang-rest-api-with-jwts/

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "database/sql"
    "time"

    _ "github.com/lib/pq"
    "github.com/dgrijalva/jwt-go"
)


type Task struct {
    ID    int `json:"id"`
    Text  string `json:"text"`
}

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

var user = User{
    Username: "1",
    Password: "1",
}

var mySignKey = []byte("johenews")

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

func createTaskHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        
        if r.Method != "POST" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        if r.Header.Get("Content-Type") != "application/json" {
            http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
            return
        }

        var newTask struct {
            Text string `json:"text"`
        }
        
        if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        if newTask.Text == "" {
            http.Error(w, "Text field is required", http.StatusBadRequest)
            return
        }
        
        var id int
        err := db.QueryRow("INSERT INTO \"Tasks\" (text) VALUES ($1) RETURNING id", newTask.Text).Scan(&id)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        
        task := Task{
            ID:   id,
            Text: newTask.Text,
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        if err := json.NewEncoder(w).Encode(task); err != nil {
            return
        }
        
    }
}

func updateTaskHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "PUT" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        if r.Header.Get("Content-Type") != "application/json" {
            http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
            return
        }

        id := r.URL.Path[len("/tasks/"):]
        if id == "" {
            http.Error(w, "Task ID is required", http.StatusBadRequest)
            return
        }

        var updatedTask struct {
            Text string `json:"text"`
        }

        if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
            fmt.Printf("JSON decode error: %v\n", err)
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        if updatedTask.Text == "" {
            http.Error(w, "Text field is required", http.StatusBadRequest)
            return
        }
        
        result, err := db.Exec("UPDATE \"Tasks\" SET text = $1 WHERE id = $2", updatedTask.Text, id)
        if err != nil {
                fmt.Printf("Database update error: %v\n", err)
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
        }

        rowsAffected, err := result.RowsAffected()
        if err != nil {
            fmt.Printf("Error checking rows affected: %v\n", err)
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        if rowsAffected == 0 {
            fmt.Printf("Task with ID %s not found\n", id)
            http.Error(w, "Task not found", http.StatusNotFound)
            return
        }

        var task Task
        err = db.QueryRow("SELECT id, text FROM \"Tasks\" WHERE id = $1", id).Scan(&task.ID, &task.Text)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(task); err != nil {
            return
        }
    }
}

func deleteTaskHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "DELETE" {
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
            fmt.Printf("Database select error: %v\n", err)
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        
        result, err := db.Exec("DELETE FROM \"Tasks\" WHERE id = $1", id)
        if err != nil {
                fmt.Printf("Database delete error: %v\n", err)
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
        }

        rowsAffected, err := result.RowsAffected()
        if err != nil {
            fmt.Printf("Error checking rows affected: %v\n", err)
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        if rowsAffected == 0 {
            fmt.Printf("Task with ID %s not found\n", id)
            http.Error(w, "Task not found", http.StatusNotFound)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "Task deleted successfully",
            "deleted_task": task,
        }); err != nil {
            fmt.Printf("JSON encoding error: %v\n", err)
            return
        }
    }
}

    func login(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        var u User
        json.NewDecoder(r.Body).Decode(&u)
        // fmt.Println("user: ", u)
        checkLogin(u)
    }

    func checkLogin(u User) string {
        if user.Username != u.Username || user.Password != u.Password {
            fmt.Println("NOT CORRECT")
            err := "error"
            return err
        }

        validToken, err := GenerateJWT()
        fmt.Println(validToken)

        if err != nil {
            fmt.Println(err)
        }

        return validToken
    }

    func GenerateJWT() (string, error) {
        token := jwt.New(jwt.SigningMethodHS256)

        claims := token.Claims.(jwt.MapClaims)

        claims["exp"] = time.Now().Add(time.Hour * 1000).Unix()
        claims["user"] = "Johenews"
        claims["authorized"] = true

        tokenString, err := token.SignedString(mySignKey)

        if err != nil {
            log.Fatal(err)
        }

        return tokenString, nil
    }

    func checkAuth(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        if r.Header["Token"] != nil {

            token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("There was an error")
                }
                return mySignKey, nil
            })

            if err != nil {
                fmt.Fprintf(w, err.Error())
            }

            if token.Valid {
                endpoint(w, r)
            }
        } else {

            fmt.Fprintf(w, "Not Authorized")
        }
    })
}

func main() {

    initDB()
    defer db.Close()

	r := http.NewServeMux()

    r.Handle("GET /tasks", checkAuth(getTasksHandler(db)))
	r.Handle("POST /tasks", checkAuth(createTaskHandler(db)))
	r.Handle("GET /tasks/{id}", checkAuth(writingTaskHandler(db)))
	r.Handle("PUT /tasks/{id}", checkAuth(updateTaskHandler(db)))
	r.Handle("DELETE /tasks/{id}", checkAuth(deleteTaskHandler(db)))

    r.HandleFunc("POST /login", login)

    fmt.Println("Сервер запущен на http://localhost:8080")
    fmt.Println("Для проверки откройте браузер или используйте curl http://localhost:8080/tasks")

    fmt.Println(`Чтобы отправить пароль curl -X POST -H "Content-Type: application/json" -d '{"username":"ваш_логин", "password":"ваш_пароль"}' http://localhost:8080/login`)

	fmt.Println(`Для добавления задачи curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"text": "Задача"}'`)
	fmt.Println("Для чтения задачи curl http://localhost:8080/tasks/{id}")
	fmt.Println(`Для обновления задачи curl -X PUT http://localhost:8080/tasks/{id} -H "Content-Type: application/json" -d '{"text": "Обновленный текст задачи"}'`)
	fmt.Println("Для удаления задачи curl -X DELETE http://localhost:8080/tasks/{id}")

    log.Fatal(http.ListenAndServe(":8080", r))
}