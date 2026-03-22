package middleware

import (
    "context"
    "net/http"
    "strings"
    
    "api-gateway/pkg/auth"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Создаём новый запрос с токеном в заголовке для GetUserFromJWT
        // или просто передаём оригинальный r, так как GetUserFromJWT сама достанет токен
        userID, err := auth.GetUserFromJWT(r)
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Добавляем user_id в контекст
        ctx := context.WithValue(r.Context(), "user_id", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}