package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type Response struct {
    Message string `json:"message"`
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(Response{Message: "Добро пожаловать в API"})
}

func main() {

	r := http.NewServeMux()

    r.HandleFunc("GET /", homeHandler)

    fmt.Println("Сервер запущен на http://localhost:8080")
    fmt.Println("Для проверки откройте браузер или используйте curl:")
    fmt.Println("curl http://localhost:8080/")
    
    log.Fatal(http.ListenAndServe(":8080", r))
}