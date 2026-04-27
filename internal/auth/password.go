package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
