package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/AndreaCasaluci/go-chat-app/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AndreaCasaluci/go-chat-app/db"
	"github.com/AndreaCasaluci/go-chat-app/repository"
)

type UserResponse struct {
	UUID      string    `json:"uuid"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

	db, err := database.GetDb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not connect to the database: %v", err), http.StatusInternalServerError)
		return
	}

	userExistsResult := repository.IsUserExists(db, &userReq.Username, &userReq.Email)
	if userExistsResult.Err != nil {
		http.Error(w, fmt.Sprintf("Error checking user existence: %v", userExistsResult.Err), http.StatusInternalServerError)
		return
	}
	if userExistsResult.Exists {
		http.Error(w, fmt.Sprintf("%s is already taken", userExistsResult.Field), http.StatusConflict)
		return
	}

	hashedPassword, err := utils.HashPassword(userReq.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error hashing password: %v", err), http.StatusInternalServerError)
		return
	}

	user, err := repository.CreateUser(db, repository.CreateUserParams{userReq.Username, userReq.Email, *hashedPassword})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		UUID:      user.UUID.String(),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userUUID := vars["uuid"]

	claimUUID, ok := r.Context().Value("user_uuid").(uuid.UUID)
	if !ok || claimUUID.String() == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if userUUID != claimUUID.String() {
		http.Error(w, "You can only update your own account", http.StatusForbidden)
		return
	}

	var userReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	db, err := database.GetDb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not connect to the database: %v", err), http.StatusInternalServerError)
		return
	}

	isExistResult := repository.IsUserExists(db, userReq.Username, userReq.Email)
	if isExistResult.Err != nil {
		http.Error(w, fmt.Sprintf("Error checking user existence: %v", err), http.StatusInternalServerError)
		return
	}

	if isExistResult.Exists {
		http.Error(w, fmt.Sprintf("%s is already taken", isExistResult.Field), http.StatusConflict)
		return
	}

	updatedUser, err := repository.UpdateUser(db, repository.UpdateUserParams{&claimUUID, userReq.Username, userReq.Email, userReq.Password})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating user: %v", err), http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		UUID:      updatedUser.UUID.String(),
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
