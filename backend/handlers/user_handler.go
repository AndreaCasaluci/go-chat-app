package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/AndreaCasaluci/go-chat-app/db"
	"github.com/AndreaCasaluci/go-chat-app/repository"
)

// RegisterUserRequest defines the expected fields in the registration request
type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func ValidateUsername(username string) error {
	if utf8.RuneCountInString(username) < 3 || utf8.RuneCountInString(username) > 20 {
		return fmt.Errorf("username must be between 3 and 20 characters")
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !re.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	return nil
}

func ValidateEmail(email *string) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(*email) {
		return fmt.Errorf("invalid email format")
	}

	*email = strings.ToLower(*email)

	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var userReq RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if err := ValidateUsername(userReq.Username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ValidateEmail(&userReq.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ValidatePassword(userReq.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := database.Connect()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not connect to the database: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	userExists, field, err := repository.IsUserExists(db, userReq.Username, userReq.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking user existence: %v", err), http.StatusInternalServerError)
		return
	}
	if userExists {
		http.Error(w, fmt.Sprintf("%s is already taken", field), http.StatusConflict)
		return
	}

	user, err := repository.CreateUser(db, userReq.Username, userReq.Email, userReq.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
