package entities

import "errors"

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidToken     = errors.New("invalid token")
	ErrConnectionClosed = errors.New("connection closed")
)
