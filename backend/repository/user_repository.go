package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/AndreaCasaluci/go-chat-app/models"
	"github.com/AndreaCasaluci/go-chat-app/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type UserExistsResult struct {
	Exists bool
	Field  string
	Err    error
}

type CreateUserParams struct {
	Username       string
	Email          string
	HashedPassword string
}

type UpdateUserParams struct {
	UserUUID *uuid.UUID
	Username *string
	Email    *string
	Password *string
}

func IsUserExists(ctx context.Context, db *sql.DB, username, email *string) UserExistsResult {
	resultChan := make(chan UserExistsResult, 2)

	checkExists := func(fieldName string, value *string) {
		if value == nil {
			return
		}

		go func() {
			var exists bool
			err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM users WHERE %s ILIKE $1)", fieldName), *value).Scan(&exists)

			select {
			case resultChan <- UserExistsResult{exists, fieldName, err}:
			case <-ctx.Done():
				return
			}
		}()
	}

	checkExists("username", username)
	checkExists("email", email)

	for i := 0; i < 2; i++ {
		select {
		case res := <-resultChan:
			if res.Err != nil {
				return UserExistsResult{false, "", res.Err}
			}
			if res.Exists {
				return res
			}
		case <-ctx.Done():
			return UserExistsResult{false, "", ctx.Err()}
		}
	}

	return UserExistsResult{false, "", nil}
}

func CreateUser(ctx context.Context, db *sql.DB, params CreateUserParams) (*models.User, error) {
	resultChan := make(chan struct {
		user *models.User
		err  error
	})

	go func() {
		var user models.User
		err := db.QueryRowContext(ctx, `
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

	select {
	case result := <-resultChan:
		if result.err != nil {
			log.Printf("Error inserting user: %v", result.err)
			return nil, fmt.Errorf("could not create user: %w", result.err)
		}
		return result.user, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("operation canceled: %w", ctx.Err())
	}
}

func AuthenticateUser(ctx context.Context, db *sql.DB, email, password string) (*models.User, error) {
	userChan := make(chan *models.User, 1)
	errChan := make(chan error, 1)

	go func() {
		var user models.User
		err := db.QueryRowContext(ctx, "SELECT id, uuid, username, email, password FROM users WHERE LOWER(email) = LOWER($1)", email).
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
	case <-ctx.Done():
		return nil, fmt.Errorf("operation canceled: %w", ctx.Err())
	}
}

func UpdateUser(ctx context.Context, db *sql.DB, params UpdateUserParams) (*models.User, error) {
	userChan := make(chan *models.User)
	errChan := make(chan error)

	defer close(userChan)
	defer close(errChan)

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
			hashedPassword, err := utils.HashPassword(*params.Password)
			if err != nil {
				errChan <- fmt.Errorf("could not hash password: %v", err)
				return
			}
			query += fmt.Sprintf(", password=$%d", argCount)
			args = append(args, *hashedPassword)
			argCount++
		}

		query += fmt.Sprintf(" WHERE uuid=$%d RETURNING id, username, email, uuid, created_at, updated_at", argCount)
		args = append(args, params.UserUUID)
		err := db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Username, &user.Email, &user.UUID, &user.CreatedAt, &user.UpdatedAt)
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
	case <-ctx.Done():
		return nil, fmt.Errorf("operation canceled: %w", ctx.Err())
	}
}
