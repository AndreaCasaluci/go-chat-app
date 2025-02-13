package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (*string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	stringHashedPassword := string(hashedPassword)
	return &stringHashedPassword, nil
}
