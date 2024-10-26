package entities

type EventType string

const (
	EventTypeUserJoined EventType = "user_joined"
	EventTypeUserLeft   EventType = "user_left"
	EventTypeNewMessage EventType = "new_message"
	EventTypeUnknown    EventType = "unknown"
)
