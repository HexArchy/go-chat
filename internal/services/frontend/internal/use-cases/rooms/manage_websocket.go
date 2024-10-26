package rooms

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/chat"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
)

// ManageWebSocketUseCase defines the interface for managing WebSocket connections.
type ManageWebSocketUseCase interface {
	Connect(ctx context.Context, roomID uuid.UUID, conn *websocket.Conn) error
}

// manageWebSocketUseCase is the concrete implementation of ManageWebSocketUseCase.
type manageWebSocketUseCase struct {
	chatClient *chat.Client
	logger     *zap.Logger

	mu          sync.RWMutex
	roomClients map[uuid.UUID]map[*websocket.Conn]bool
}

// NewManageWebSocketUseCase creates a new instance of ManageWebSocketUseCase.
func NewManageWebSocketUseCase(chatClient *chat.Client, logger *zap.Logger) ManageWebSocketUseCase {
	return &manageWebSocketUseCase{
		chatClient:  chatClient,
		logger:      logger,
		roomClients: make(map[uuid.UUID]map[*websocket.Conn]bool),
	}
}

// Connect handles the WebSocket connection lifecycle.
func (uc *manageWebSocketUseCase) Connect(ctx context.Context, roomID uuid.UUID, conn *websocket.Conn) error {
	// Retrieve userID from the context.
	userID, err := entities.GetUserIDFromContext(ctx)
	if err != nil {
		uc.logger.Error("ManageWebSocketUseCase: failed to get user from context", zap.Error(err))
		return errors.Wrap(err, "user not authenticated")
	}

	uc.logger.Debug("ManageWebSocketUseCase: establishing WebSocket connection",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()))

	// Add connection to roomClients.
	uc.mu.Lock()
	if uc.roomClients[roomID] == nil {
		uc.roomClients[roomID] = make(map[*websocket.Conn]bool)
	}
	uc.roomClients[roomID][conn] = true
	uc.mu.Unlock()

	// Set up a cancellable context to manage the connection.
	eventCtx, cancel := context.WithCancel(ctx)

	// Ensure connection removal upon exit.
	defer func() {
		uc.mu.Lock()
		delete(uc.roomClients[roomID], conn)
		if len(uc.roomClients[roomID]) == 0 {
			delete(uc.roomClients, roomID)
		}
		uc.mu.Unlock()
		cancel()
		// Broadcast user departure event.
		uc.broadcastUserEvent(roomID, userID.String(), "leave")
	}()

	// Broadcast user join event.
	uc.broadcastUserEvent(roomID, userID.String(), "join")

	// Connect to the chat service.
	eventChan, errChan, cancelConn, err := uc.chatClient.Connect(eventCtx, roomID, userID)
	if err != nil {
		uc.logger.Error("ManageWebSocketUseCase: failed to connect to chat service", zap.Error(err))
		return errors.Wrap(err, "failed to connect to chat service")
	}
	// Ensure connection cancellation upon exit.
	defer cancelConn()

	// Launch a goroutine to read messages from WebSocket.
	go func() {
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					uc.logger.Error("ManageWebSocketUseCase: WebSocket read error", zap.Error(err))
				}
				// Send error to errChan and terminate the goroutine.
				errChan <- errors.Wrap(err, "WebSocket read error")
				return
			}

			if messageType == websocket.TextMessage {
				var message struct {
					Content string `json:"content"`
				}
				if err := json.Unmarshal(p, &message); err != nil {
					uc.logger.Warn("ManageWebSocketUseCase: failed to unmarshal message", zap.Error(err))
					continue
				}

				// Send message to chat service.
				if err := uc.chatClient.SendMessage(ctx, roomID, userID, message.Content); err != nil {
					uc.logger.Error("ManageWebSocketUseCase: failed to send message", zap.Error(err))
				}
			}
		}
	}()

	// Main loop for listening to events and errors.
	for {
		select {
		case event := <-eventChan:
			data := map[string]interface{}{
				"type":      event.Type,
				"userId":    event.UserID.String(),
				"timestamp": event.Timestamp,
			}

			if event.Message != nil {
				data["message"] = map[string]interface{}{
					"id":        event.Message.ID.String(),
					"content":   event.Message.Content,
					"userId":    event.Message.UserID.String(),
					"createdAt": event.Message.CreatedAt,
				}
			}

			if err := conn.WriteJSON(data); err != nil {
				uc.logger.Error("ManageWebSocketUseCase: failed to write JSON to WebSocket", zap.Error(err))
				return err
			}

		case err := <-errChan:
			uc.logger.Error("ManageWebSocketUseCase: chat connection error", zap.Error(err))
			return err

		case <-ctx.Done():
			return nil
		}
	}
}

// broadcastUserEvent broadcasts user join/leave events to all connected clients in the room.
func (uc *manageWebSocketUseCase) broadcastUserEvent(roomID uuid.UUID, userID, eventType string) {
	event := map[string]interface{}{
		"type":      eventType,
		"userId":    userID,
		"timestamp": time.Now(),
	}

	uc.mu.RLock()
	defer uc.mu.RUnlock()

	for conn := range uc.roomClients[roomID] {
		if err := conn.WriteJSON(event); err != nil {
			uc.logger.Error("ManageWebSocketUseCase: failed to broadcast event", zap.Error(err))
			conn.Close()
			delete(uc.roomClients[roomID], conn)
		}
	}
}
