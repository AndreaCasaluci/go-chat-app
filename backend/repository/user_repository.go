package repository

import (
	"database/sql"
	"github.com/AndreaCasaluci/go-chat-app/models"
	"golang.org/x/crypto/bcrypt"
)

func IsUserExists(db *sql.DB, username, email string) (bool, string, error) {
	var exists bool

	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE username ILIKE $1)", username).Scan(&exists)
	if err != nil {
		return false, "", err
	}
	if exists {
		return true, "username", nil
	}

	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email ILIKE $1)", email).Scan(&exists)
	if err != nil {
		return false, "", err
	}
	if exists {
		return true, "email", nil
	}

	return false, "", nil
}

func CreateUser(db *sql.DB, username, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = db.QueryRow(`
		INSERT INTO users (uuid, username, email, password)
		VALUES (uuid_generate_v4(), $1, $2, $3)
		RETURNING id, uuid, username, email, created_at, updated_at`,
		username, email, string(hashedPassword),
	).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
