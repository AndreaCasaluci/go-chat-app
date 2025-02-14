package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/AndreaCasaluci/go-chat-app/db"
	"github.com/AndreaCasaluci/go-chat-app/repository"
	"github.com/AndreaCasaluci/go-chat-app/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type UserResponse struct {
	UUID      string    `json:"uuid"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20,usernamechars"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}

type UpdateUserRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var userReq RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	err := utils.ValidateStruct(userReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	db, err := database.GetDb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not connect to the database: %v", err), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	userExistsResult := repository.IsUserExists(ctx, db, &userReq.Username, &userReq.Email)
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

	user, err := repository.CreateUser(ctx, db, repository.CreateUserParams{userReq.Username, userReq.Email, *hashedPassword})
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

	ctx := r.Context()

	isExistResult := repository.IsUserExists(ctx, db, userReq.Username, userReq.Email)
	if isExistResult.Err != nil {
		http.Error(w, fmt.Sprintf("Error checking user existence: %v", err), http.StatusInternalServerError)
		return
	}

	if isExistResult.Exists {
		http.Error(w, fmt.Sprintf("%s is already taken", isExistResult.Field), http.StatusConflict)
		return
	}

	updatedUser, err := repository.UpdateUser(ctx, db, repository.UpdateUserParams{&claimUUID, userReq.Username, userReq.Email, userReq.Password})
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
