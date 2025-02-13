package repository

import (
	"database/sql"
	"fmt"
	"github.com/AndreaCasaluci/go-chat-app/models"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type UserExistsResult struct {
	Exists bool
	Field  string
	Err    error
}

func IsUserExists(db *sql.DB, username, email *string) UserExistsResult {
	resultChan := make(chan UserExistsResult, 2)

	if username != nil {
		go func() {
			var exists bool
			err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE username ILIKE $1)", *username).Scan(&exists)
			resultChan <- UserExistsResult{exists, "username", err}
		}()
	}

	if email != nil {
		go func() {
			var exists bool
			err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email ILIKE $1)", *email).Scan(&exists)
			resultChan <- UserExistsResult{exists, "email", err}
		}()
	}

	for i := 0; i < 2; i++ {
		res := <-resultChan
		if res.Err != nil {
			return UserExistsResult{false, "", res.Err}
		}
		if res.Exists {
			return UserExistsResult{true, res.Field, nil}
		}
	}

	return UserExistsResult{false, "", nil}
}

type CreateUserParams struct {
	Username       string
	Email          string
	HashedPassword string
}

func CreateUser(db *sql.DB, params CreateUserParams) (*models.User, error) {
	resultChan := make(chan struct {
		user *models.User
		err  error
	})

	go func() {
		var user models.User
		err := db.QueryRow(`
			INSERT INTO users (uuid, username, email, password, created_at, updated_at)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5)
			RETURNING id, uuid, username, email, created_at, updated_at`,
			params.Username, params.Email, params.HashedPassword, time.Now(), time.Now(),
		).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

		resultChan <- struct {
			user *models.User
			err  error
		}{user: &user, err: err}
	}()

	result := <-resultChan
	if result.err != nil {
		log.Printf("Error inserting user: %v", result.err)
		return nil, fmt.Errorf("could not create user: %w", result.err)
	}

	return result.user, nil
}

func AuthenticateUser(db *sql.DB, email, password string) (*models.User, error) {
	userChan := make(chan *models.User)
	errChan := make(chan error)

	go func() {
		var user models.User
		err := db.QueryRow("SELECT id, uuid, username, email, password FROM users WHERE LOWER(email) = LOWER($1)", email).
			Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.Password)

		if err != nil {
			if err == sql.ErrNoRows {
				errChan <- fmt.Errorf("invalid email or password")
			} else {
				errChan <- fmt.Errorf("error querying user: %v", err)
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			errChan <- fmt.Errorf("invalid email or password")
			return
		}

		user.Password = ""

		userChan <- &user
	}()

	select {
	case user := <-userChan:
		return user, nil
	case err := <-errChan:
		return nil, err
	}
}

type UpdateUserParams struct {
	UserUUID *string
	Username *string
	Email    *string
	Password *string
}

func UpdateUser(db *sql.DB, params UpdateUserParams) (*models.User, error) {
	userChan := make(chan *models.User)
	errChan := make(chan error)

	go func() {
		var user models.User

		query := `UPDATE users SET updated_at=CURRENT_TIMESTAMP`
		args := []interface{}{}
		argCount := 1

		if params.Username != nil {
			query += fmt.Sprintf(", username=$%d", argCount)
			args = append(args, *params.Username)
			argCount++
		}

		if params.Email != nil {
			query += fmt.Sprintf(", email=$%d", argCount)
			args = append(args, *params.Email)
			argCount++
		}

		if params.Password != nil {
			query += fmt.Sprintf(", password=$%d", argCount)
			args = append(args, *params.Password)
			argCount++
		}

		query += fmt.Sprintf(" WHERE uuid=$%d RETURNING id, username, email, uuid, created_at, updated_at", argCount)
		args = append(args, params.UserUUID)

		err := db.QueryRow(query, args...).Scan(&user.ID, &user.Username, &user.Email, &user.UUID, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			errChan <- fmt.Errorf("could not update user: %v", err)
			return
		}

		user.Password = ""

		userChan <- &user
	}()

	select {
	case user := <-userChan:
		return user, nil
	case err := <-errChan:
		return nil, err
	}
}
