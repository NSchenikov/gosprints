package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var mySignKey = []byte("johenews")

    func GenerateJWT(username string) (string, error) {
        token := jwt.New(jwt.SigningMethodHS256)

        claims := token.Claims.(jwt.MapClaims)

        claims["exp"] = time.Now().Add(time.Hour * 1000).Unix()
        claims["user"] = username
        claims["authorized"] = true

        tokenString, err := token.SignedString(mySignKey)

        if err != nil {
            return "", fmt.Errorf("failed to generate token: %v", err)
        }

        return tokenString, nil
    }

func GetSignKey() []byte {
    return mySignKey
}