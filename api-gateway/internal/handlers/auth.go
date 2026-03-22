package handlers

import (
    "encoding/json"
    "net/http"
    "sync"
    
    "api-gateway/pkg/auth"
)

type AuthHandler struct {
    users map[string]string // username -> password
    mu    sync.RWMutex
}

func NewAuthHandler() *AuthHandler {
    return &AuthHandler{
        users: make(map[string]string),
    }
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if _, exists := h.users[req.Username]; exists {
        http.Error(w, "User already exists", http.StatusConflict)
        return
    }
    
    h.users[req.Username] = req.Password
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    h.mu.RLock()
    password, exists := h.users[req.Username]
    h.mu.RUnlock()
    
    if !exists || password != req.Password {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    token, err := auth.GenerateJWT(req.Username)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "token":  token,
        "status": "success",
    })
}