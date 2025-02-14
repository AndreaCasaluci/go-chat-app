package utils

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var jwtSecret []byte = nil

func retrieveJwtSecret() {
	if jwtSecret != nil {
		return
	}

	config, err := GetConfig()
	if err != nil {
		panic(err)
	}

	jwtSecretString := config.JwtSecretKey
	if len(jwtSecretString) == 0 {
		panic("JWT_SECRET_KEY is not set in the environment variables!")
	}
	jwtSecret = []byte(jwtSecretString)
}

type Claims struct {
	UserID int64 `json:"user_id"`
	UserUUID uuid.UUID `json:"user_uuid"`
	Username string `json:"username"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func ValidateToken(tokenString string) (*Claims, error) {
	jwtSecretKey := GetJwtSecret()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid or expired token")
}

func GetJwtSecret() []byte {
	if jwtSecret != nil {
		return jwtSecret
	}
	retrieveJwtSecret()
	return jwtSecret
}