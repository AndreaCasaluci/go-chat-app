package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"time"

	"github.com/AndreaCasaluci/go-chat-app/db"
	"github.com/AndreaCasaluci/go-chat-app/repository"
	"github.com/dgrijalva/jwt-go"
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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	db, err := database.Connect()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not connect to the database: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	user, err := repository.AuthenticateUser(db, loginReq.Email, loginReq.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user.ID, user.UUID, user.Username, user.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating JWT: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

func generateJWT(userID int64, userUuid uuid.UUID, username string, userEmail string) (string, error) {
	retrieveJwtSecret()

	claims := jwt.MapClaims{
		"user_id":   userID,
		"user_uuid": userUuid,
		"email":     userEmail,
		"username":  username,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
