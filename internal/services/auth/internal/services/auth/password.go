package auth

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

func (a *AuthService) generateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", errors.Wrap(err, "generate salt")
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

func (a *AuthService) hashPassword(password, salt string) (string, error) {
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", errors.Wrap(err, "decode salt")
	}

	hash := argon2.IDKey([]byte(password), saltBytes, 1, 64*1024, 4, 32)
	return base64.StdEncoding.EncodeToString(hash), nil
}

func (a *AuthService) verifyPassword(password, hashedPassword string) error {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 4 {
		return entities.ErrInvalidPasswordFormat
	}

	salt := parts[2]
	calculatedHash, err := a.hashPassword(password, salt)
	if err != nil {
		return errors.Wrap(err, "hash password")
	}

	if calculatedHash != hashedPassword {
		return entities.ErrInvalidPassword
	}

	return nil
}
