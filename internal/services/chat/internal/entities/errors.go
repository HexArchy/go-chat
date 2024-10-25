package entities

import "errors"

var (
	ErrMessageNotFound    = errors.New("message not found")
	ErrRoomNotFound       = errors.New("room not found")
	ErrInvalidMessageData = errors.New("invalid message data")
	ErrUserNotFound       = errors.New("user not found")
)
