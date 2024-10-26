package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func comparePasswords(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
