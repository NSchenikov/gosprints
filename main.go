//sources used
//1) https://purpleschool.ru/knowledge-base/article/creating-rest-api
// 2) https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e
// 3) https://www.youtube.com/watch?v=k6kiNivraJ8
// 4) https://tutorialedge.net/golang/authenticating-golang-rest-api-with-jwts/

package main

import (
    "encoding/json"
    "fmt"
    "strings"
    "log"
    "net/http"
    "database/sql"

    "github.com/dgrijalva/jwt-go"

    "gosprints/internal/models"
    "gosprints/pkg/auth"
    "gosprints/pkg/database"
    "gosprints/internal/repositories"
)

func getTasksHandler(taskRepo repositories.TaskRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tasks, err := taskRepo.GetAll()

        if err != nil {
            fmt.Printf("Error getting tasks: %v\n", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
    }
}

func getTaskByID(taskRepo repositories.TaskRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        idStr := path[len("/tasks/"):]

        if idStr == "" {
            http.Error(w, "Task ID is required", http.StatusBadRequest)
            return
        }

        var id int
        if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
            http.Error(w, "Invalid task ID", http.StatusBadRequest)
            return
        }

        task, err := taskRepo.GetByID(id)
        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": err.Error(),
            })
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(task)
    }
}

func createTaskHandler(taskRepo repositories.TaskRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

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
        
        task := &models.Task{Text: newTask.Text}
		err := taskRepo.Create(task)
		if err != nil {
			fmt.Printf("Error creating task: %v\n", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(task)
		fmt.Printf("Task created: ID=%d\n", task.ID)
        
    }
}

func updateTaskHandler(taskRepo repositories.TaskRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        idStr := r.URL.Path[len("/tasks/"):]
        if idStr == "" {
            http.Error(w, "Task ID is required", http.StatusBadRequest)
            return
        }

        var id int
        if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
            http.Error(w, "Invalid task ID", http.StatusBadRequest)
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

        err := taskRepo.Update(id, updatedTask.Text)
        if err != nil {
            fmt.Printf("Error updating task: %v\n", err)
            if strings.Contains(err.Error(), "not found") {
                http.Error(w, "Task not found", http.StatusNotFound)
            } else {
                http.Error(w, "Database error", http.StatusInternalServerError)
            }
            return
        }

        fmt.Printf("Task updated successfully: ID=%d\n", id)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "success",
            "message": "Task updated successfully",
        })        
    }
}

func deleteTaskHandler(taskRepo repositories.TaskRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        idStr := r.URL.Path[len("/tasks/"):]
        if idStr == "" {
            http.Error(w, "Task ID is required", http.StatusBadRequest)
            return
        }

        var id int
        if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
            http.Error(w, "Invalid task ID", http.StatusBadRequest)
            return
        }

        task, err := taskRepo.GetByID(id)
        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": err.Error(),
            })
            return
        }

        err = taskRepo.Delete(id)
        if err != nil {
            fmt.Printf("Error deleting task: %v\n", err)
            if strings.Contains(err.Error(), "not found") {
                http.Error(w, "Task not found", http.StatusNotFound)
            } else {
                http.Error(w, "Database error", http.StatusInternalServerError)
            }
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

func registerUser(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        if r.Method != "POST" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var newUser struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        
        if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        if newUser.Username == "" || newUser.Password == "" {
            http.Error(w, "Username and password are required", http.StatusBadRequest)
            return
        }

        var id int
        err := db.QueryRow(
            "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", 
            newUser.Username, 
            newUser.Password,
        ).Scan(&id)
        
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Database error: " + err.Error(),
            })
            return
        }

        user := models.User{
            ID:       id,
            Username: newUser.Username,
        }
        
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(user)
    }
}

func login(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        var u models.User
        if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }
        
        token, err := checkLogin(db, u)
        if err != nil {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Invalid credentials",
            })
            return
        }
        
        json.NewEncoder(w).Encode(map[string]string{
            "token": token,
            "status": "success",
        })
    }
}

    func checkLogin(db *sql.DB, u models.User) (string, error) {
        var dbUser models.User
        var storedPassword string
        
        // поиск пользователя по db
        err := db.QueryRow(
            "SELECT id, username, password FROM users WHERE username = $1", 
            u.Username,
        ).Scan(&dbUser.ID, &dbUser.Username, &storedPassword)
        
        if err != nil {
            if err == sql.ErrNoRows {
                return "", fmt.Errorf("user not found")
            }
            return "", fmt.Errorf("database error: %v", err)
        }
        
        if !CheckPassword(u.Password, storedPassword) {
            return "", fmt.Errorf("invalid password")
        }
        
        validToken, err := auth.GenerateJWT(u.Username)
        if err != nil {
            return "", fmt.Errorf("error generating token: %v", err)
        }
        
        fmt.Printf("User %s logged in successfully\n", u.Username)
        return validToken, nil
    }

    func CheckPassword(inputPassword, storedPassword string) bool {
        return inputPassword == storedPassword
    }

func checkAuth(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            fmt.Fprintf(w, `{"error": "Authorization header required"}`)
            return
        }
        
        if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
            fmt.Fprintf(w, `{"error": "Invalid authorization format. Use: Bearer <token>"}`)
            return
        }
        
        tokenString := authHeader[7:]
        
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return auth.GetSignKey(), nil
        })

        if err != nil {
            fmt.Fprintf(w, `{"error": "Invalid token: %s"}`, err.Error())
            return
        }

        if token.Valid {
            endpoint(w, r)
        } else {
            fmt.Fprintf(w, `{"error": "Invalid token"}`)
        }
    })
}

func main() {

    db := database.InitDB()
    defer db.Close()

    taskRepo := repositories.NewTaskRepository(db)

	r := http.NewServeMux()

    r.Handle("GET /tasks", checkAuth(getTasksHandler(taskRepo)))
	r.Handle("POST /tasks", checkAuth(createTaskHandler(taskRepo)))
	r.Handle("GET /tasks/{id}", checkAuth(getTaskByID(taskRepo)))
	r.Handle("PUT /tasks/{id}", checkAuth(updateTaskHandler(taskRepo)))
	r.Handle("DELETE /tasks/{id}", checkAuth(deleteTaskHandler(taskRepo)))

    r.HandleFunc("POST /login", login(db))
    r.HandleFunc("POST /register", registerUser(db))

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