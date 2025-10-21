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

    "github.com/dgrijalva/jwt-go"

    "gosprints/internal/models"
    "gosprints/pkg/auth"
    "gosprints/pkg/database"
    "gosprints/internal/repositories"
    "gosprints/internal/handlers"
)

func registerUser(userRepo repositories.UserRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        var newUser struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        
        if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        if newUser.Username == "" || newUser.Password == "" {
            http.Error(w, "Username and password required", http.StatusBadRequest)
            return
        }

        user := &models.User{
            Username: newUser.Username,
            Password: newUser.Password,
        }
        
        err := userRepo.Create(user)
        if err != nil {
            fmt.Printf("Error creating user: %v\n", err)
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(user)
        fmt.Printf("User registered: ID=%d\n", user.ID)
    }
}

func login(userRepo repositories.UserRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        
        var credentials struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }

        if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        user, err := userRepo.GetByUsername(credentials.Username)
        if err != nil {
            fmt.Printf("User not found: %v\n", err)
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }

        if user.Password != credentials.Password {
            fmt.Printf("Password mismatch for user: %s\n", credentials.Username)
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }

        token, err := auth.GenerateJWT(user.Username)
        if err != nil {
            fmt.Printf("Token generation error: %v\n", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "token": token,
            "status": "success",
        })
    }
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
    userRepo := repositories.NewUserRepository(db)

    taskHandler := handlers.NewTaskHandler(taskRepo)
	// authHandler := handlers.NewAuthHandler(userRepo)

	r := http.NewServeMux()

    r.Handle("GET /tasks", checkAuth(taskHandler.GetTasks))
	r.Handle("POST /tasks", checkAuth(taskHandler.CreateTask))
	r.Handle("GET /tasks/{id}", checkAuth(taskHandler.GetTaskByID))
	r.Handle("PUT /tasks/{id}", checkAuth(taskHandler.UpdateTask))
	r.Handle("DELETE /tasks/{id}", checkAuth(taskHandler.DeleteTask))

    r.HandleFunc("POST /login", login(userRepo))
    r.HandleFunc("POST /register", registerUser(userRepo))

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