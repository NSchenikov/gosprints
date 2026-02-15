package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func GetUserFromJWT(r *http.Request) (string, error) {

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		return extractUserFromToken(tokenStr)
	}
	
	// Пробуем из query parameter
	tokenFromQuery := r.URL.Query().Get("token")
	if tokenFromQuery != "" {
		return extractUserFromToken(tokenFromQuery)
	}
	
	return "", errors.New("missing or invalid token")
}

func extractUserFromToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return GetSignKey(), nil
	})
	
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	
	user, ok := claims["user"].(string)
	if !ok {
		return "", errors.New("user not found in token")
	}
	
	return user, nil
}
