package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/AndreaCasaluci/go-chat-app/utils"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/AndreaCasaluci/go-chat-app/db"
	"github.com/AndreaCasaluci/go-chat-app/repository"
	"github.com/dgrijalva/jwt-go"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
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

	validationErr := utils.ValidateStruct(loginReq)
	if validationErr != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", validationErr), http.StatusBadRequest)
		return
	}

	db, err := database.GetDb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not connect to the database: %v", err), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	user, err := repository.AuthenticateUser(ctx, db, loginReq.Email, loginReq.Password)
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
	jwtSecret := utils.GetJwtSecret()

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
