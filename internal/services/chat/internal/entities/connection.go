package entities

import "github.com/google/uuid"

type ConnectionType string

const (
	WebSocket ConnectionType = "websocket"
	GRPC      ConnectionType = "grpc"
)

type Connection interface {
	Send(message []byte) error
	Close() error
	Type() ConnectionType
	ID() uuid.UUID
	IsClosed() bool
}
