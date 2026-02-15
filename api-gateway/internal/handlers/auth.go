package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gosprints/internal/models"
	"gosprints/internal/repositories"
	"gosprints/pkg/auth"
)

type AuthHandler struct {
	userRepo repositories.UserRepository
}

func NewAuthHandler(userRepo repositories.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
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
        
        err := h.userRepo.Create(user)
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

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
        var credentials struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }

        if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        user, err := h.userRepo.GetByUsername(credentials.Username)
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