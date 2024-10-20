package entities

import (
	"regexp"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type User struct {
	ID          uuid.UUID
	Email       string
	Password    string
	Username    string
	Phone       string
	Age         int
	Bio         string
	Permissions []Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Permission string

const (
	PermissionRead   Permission = "read"
	PermissionWrite  Permission = "write"
	PermissionDelete Permission = "delete"
	PermissionAdmin  Permission = "admin"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func (u *User) ValidateEmail() error {
	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmailFormat
	}
	return nil
}

func (u *User) ValidatePassword() error {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	if len(u.Password) >= 8 {
		hasMinLen = true
	}

	for _, char := range u.Password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasMinLen {
		return errors.Wrap(ErrPasswordValidation, "password must be at least 8 characters long")
	}
	if !hasUpper {
		return errors.Wrap(ErrPasswordValidation, "password must have at least one uppercase letter")
	}
	if !hasLower {
		return errors.Wrap(ErrPasswordValidation, "password must have at least one lowercase letter")
	}
	if !hasNumber {
		return errors.Wrap(ErrPasswordValidation, "password must have at least one digit")
	}
	if !hasSpecial {
		return errors.Wrap(ErrPasswordValidation, "password must have at least one special character")
	}

	return nil
}
