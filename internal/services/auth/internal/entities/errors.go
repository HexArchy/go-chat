package entities

import "errors"

var ErrInvalidCredantials = errors.New("invalid credentials")

var (
	ErrInvalidTokenType      = errors.New("invalid token type")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrInvalidRole           = errors.New("invalid role")
	ErrInvalidPasswordFormat = errors.New("invalid password format")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrInvalidToken          = errors.New("invalid token")
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
)
