package controllers

import (
	"sync"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WebSocketMessage struct {
	Type    string         `json:"type"`
	Content string         `json:"content,omitempty"`
	Limit   int            `json:"limit,omitempty"`
	Offset  int            `json:"offset,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

type WebSocketConnection struct {
	conn      *websocket.Conn
	logger    *zap.Logger
	userID    uuid.UUID
	roomID    uuid.UUID
	mu        sync.Mutex
	closed    bool
	closeChan chan struct{}
}

func NewWebSocketConnection(conn *websocket.Conn, logger *zap.Logger, userID, roomID uuid.UUID) *WebSocketConnection {
	return &WebSocketConnection{
		conn:      conn,
		logger:    logger,
		userID:    userID,
		roomID:    roomID,
		closeChan: make(chan struct{}),
	}
}

func (c *WebSocketConnection) Send(message []byte) error {
	if c.IsClosed() {
		return entities.ErrConnectionClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteMessage(websocket.TextMessage, message)
}

func (c *WebSocketConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	close(c.closeChan)
	return c.conn.Close()
}

func (c *WebSocketConnection) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *WebSocketConnection) Type() entities.ConnectionType {
	return entities.WebSocket
}

func (c *WebSocketConnection) ID() uuid.UUID {
	return c.userID
}
