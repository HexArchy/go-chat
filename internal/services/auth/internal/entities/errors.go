package entities

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrTokenNotFound       = errors.New("token not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidTokenClaims  = errors.New("invalid token claims")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrUserAlreadyExists   = errors.New("email already in use")
	ErrPasswordValidation  = errors.New("password validation failed")
	ErrInvalidEmailFormat  = errors.New("invalid email format")
	ErrPermissionDenied    = errors.New("permission denied")
)
