package utils

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var jwtSecret []byte = nil

func retrieveJwtSecret() {
	if jwtSecret != nil {
		return
	}
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}
	jwtSecretString := os.Getenv("JWT_SECRET_KEY")
	if len(jwtSecretString) == 0 {
		panic("JWT_SECRET_KEY is not set in the environment variables!")
	}
	jwtSecret = []byte(jwtSecretString)
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.StandardClaims
}

func ValidateToken(tokenString string) (*Claims, error) {
	retrieveJwtSecret()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid or expired token")
}
