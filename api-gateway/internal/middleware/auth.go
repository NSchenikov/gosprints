package handlers

import (
	"fmt"
	"net/http"

	"gosprints/pkg/auth"
	"github.com/dgrijalva/jwt-go"
)

func (h *AuthHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

        if !token.Valid {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnauthorized)
            fmt.Fprint(w, `{"error": "Invalid token"}`)
            return
        }

        next(w, r)
	}
}
//