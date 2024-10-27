package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/HexArch/go-chat/internal/services/chat/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/connect"
	"github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/disconnect"
	getmessages "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/get-messages"
	sendmessage "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/send-message"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	logger        *zap.Logger
	connectUC     *connect.UseCase
	disconnectUC  *disconnect.UseCase
	messageUC     *sendmessage.UseCase
	getMessagesUC *getmessages.UseCase
	authClient    *auth.Client
	upgrader      websocket.Upgrader
}

func NewWebSocketHandler(
	logger *zap.Logger,
	connectUC *connect.UseCase,
	disconnectUC *disconnect.UseCase,
	messageUC *sendmessage.UseCase,
	getMessagesUC *getmessages.UseCase,
	authClient *auth.Client,
) *WebSocketHandler {
	return &WebSocketHandler{
		logger:        logger,
		connectUC:     connectUC,
		disconnectUC:  disconnectUC,
		messageUC:     messageUC,
		getMessagesUC: getMessagesUC,
		authClient:    authClient,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("starting serving")

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token query parameter required", http.StatusUnauthorized)
		return
	}

	userInfo, err := h.authClient.ValidateToken(r.Context(), token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomIDStr := vars["roomID"]
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Processing connection",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userInfo.UserID.String()),
	)

	wsConn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	conn := NewWebSocketConnection(wsConn, h.logger, userInfo.UserID, roomID)

	connectEvent := &entities.Event{
		Type:      entities.EventUserConnected,
		RoomID:    roomID,
		UserID:    userInfo.UserID,
		Timestamp: time.Now(),
	}

	if err := h.connectUC.Execute(r.Context(), connect.ConnectInput{
		RoomID:     roomID,
		UserID:     userInfo.UserID,
		Connection: conn,
		Event:      connectEvent,
	}); err != nil {
		h.logger.Error("Failed to connect to room", zap.Error(err))
		conn.Close()
		return
	}

	go h.handleMessages(conn, roomID, userInfo.UserID)

	defer func() {
		disconnectEvent := &entities.Event{
			Type:      entities.EventUserDisconnected,
			RoomID:    roomID,
			UserID:    userInfo.UserID,
			Timestamp: time.Now(),
		}

		if err := h.disconnectUC.Execute(context.Background(), disconnect.DisconnectInput{
			RoomID: roomID,
			UserID: userInfo.UserID,
			Event:  disconnectEvent,
		}); err != nil {
			h.logger.Error("Failed to disconnect", zap.Error(err))
		}

		conn.Close()
	}()

	<-conn.closeChan
}

func (h *WebSocketHandler) handleMessages(conn *WebSocketConnection, roomID, userID uuid.UUID) {
	defer conn.Close()

	for {
		select {
		case <-conn.closeChan:
			return
		default:
			messageType, message, err := conn.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.logger.Error("WebSocket read error",
						zap.Error(err),
						zap.String("room_id", roomID.String()),
						zap.String("user_id", userID.String()),
					)
				}
				return
			}

			if messageType != websocket.TextMessage {
				continue
			}

			var msg WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				h.logger.Error("Failed to unmarshal message",
					zap.Error(err),
					zap.String("room_id", roomID.String()),
					zap.String("user_id", userID.String()),
				)
				continue
			}

			switch msg.Type {
			case "message":
				// Создаем новое сообщение
				newMessage := &entities.Message{
					ID:        uuid.New(),
					RoomID:    roomID,
					UserID:    userID,
					Content:   msg.Content,
					Timestamp: time.Now(),
				}

				// Сериализуем сообщение для payload
				messageJSON, err := json.Marshal(newMessage)
				if err != nil {
					h.logger.Error("Failed to marshal message", zap.Error(err))
					continue
				}

				// Создаем событие нового сообщения
				messageEvent := &entities.Event{
					Type:      entities.EventNewMessage,
					RoomID:    roomID,
					UserID:    userID,
					Payload:   messageJSON,
					Timestamp: time.Now(),
				}

				// Отправляем сообщение через use case с событием
				if err := h.messageUC.Execute(context.Background(), sendmessage.MessageInput{
					RoomID:  roomID,
					UserID:  userID,
					Content: msg.Content,
					Event:   messageEvent,
				}); err != nil {
					h.logger.Error("Failed to handle message",
						zap.Error(err),
						zap.String("room_id", roomID.String()),
						zap.String("user_id", userID.String()),
					)

					// Отправляем уведомление об ошибке отправителю
					errorEvent := &entities.Event{
						Type:      entities.EventError,
						RoomID:    roomID,
						UserID:    userID,
						Timestamp: time.Now(),
						Payload:   []byte(`{"error":"Failed to send message"}`),
					}

					if err := conn.Send(mustMarshal(errorEvent)); err != nil {
						h.logger.Error("Failed to send error event", zap.Error(err))
					}
					continue
				}

			case "get_history":
				if err := h.handleHistoryRequest(conn, roomID, userID, msg.Limit, msg.Offset); err != nil {
					h.logger.Error("Failed to handle history request",
						zap.Error(err),
						zap.String("room_id", roomID.String()),
						zap.String("user_id", userID.String()),
					)
				}
			}
		}
	}
}

func (h *WebSocketHandler) handleHistoryRequest(conn *WebSocketConnection, roomID, userID uuid.UUID, limit, offset int) error {
	response, err := h.getMessagesUC.Execute(context.Background(), getmessages.MessagesInput{
		RoomID: roomID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return errors.Wrap(err, "failed to get messages history")
	}

	historyJSON, err := json.Marshal(response)
	if err != nil {
		return errors.Wrap(err, "failed to marshal history response")
	}

	historyEvent := &entities.Event{
		Type:      entities.EventMessageHistory,
		RoomID:    roomID,
		UserID:    userID,
		Payload:   historyJSON,
		Timestamp: time.Now(),
	}

	eventJSON, err := json.Marshal(historyEvent)
	if err != nil {
		return errors.Wrap(err, "failed to marshal history event")
	}

	return conn.Send(eventJSON)
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
	}
	return data
}
