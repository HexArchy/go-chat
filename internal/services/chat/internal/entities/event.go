package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventUserConnected    EventType = "user_connected"
	EventUserDisconnected EventType = "user_disconnected"
	EventNewMessage       EventType = "new_message"
	EventMessageHistory   EventType = "message_history"
	EventError            EventType = "error"
)

type Event struct {
	Type      EventType       `json:"type"`
	RoomID    uuid.UUID       `json:"room_id"`
	UserID    uuid.UUID       `json:"user_id"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

type Message struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
