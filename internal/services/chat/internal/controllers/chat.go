package controllers

import (
	"context"
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/chat"
	"github.com/HexArch/go-chat/internal/services/chat/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/connect"
	"github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/disconnect"
	getmessages "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/get-messages"
	sendmessage "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/send-message"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WebSocketConnection реализует интерфейс ChatConnection
type WebSocketConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// WSMessage представляет структуру сообщения WebSocket
type WSMessage struct {
	Type    string `json:"type"`
	RoomID  string `json:"room_id"`
	Content string `json:"content,omitempty"`
	Token   string `json:"token"`
}

// ChatServiceServer реализует gRPC сервер чата
type ChatServiceServer struct {
	chat.UnimplementedChatServiceServer
	logger        *zap.Logger
	connectUC     *connect.UseCase
	disconnectUC  *disconnect.UseCase
	sendMessageUC *sendmessage.UseCase
	getMessagesUC *getmessages.UseCase
	authClient    *auth.Client
	activeConnMu  sync.RWMutex
	activeConns   map[uuid.UUID]*WebSocketConnection
}

func NewChatServiceServer(
	logger *zap.Logger,
	connectUC *connect.UseCase,
	disconnectUC *disconnect.UseCase,
	sendMessageUC *sendmessage.UseCase,
	getMessagesUC *getmessages.UseCase,
	authClient *auth.Client,
) *ChatServiceServer {
	return &ChatServiceServer{
		logger:        logger,
		connectUC:     connectUC,
		disconnectUC:  disconnectUC,
		sendMessageUC: sendMessageUC,
		getMessagesUC: getMessagesUC,
		authClient:    authClient,
		activeConns:   make(map[uuid.UUID]*WebSocketConnection),
	}
}

func (w *WebSocketConnection) SendEvent(event *entities.ChatEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.conn.WriteJSON(event)
}

func (w *WebSocketConnection) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.conn.Close()
}

func (s *ChatServiceServer) GetMessages(ctx context.Context, req *chat.GetMessagesRequest) (*chat.GetMessagesResponse, error) {
	token := getTokenFromContext(ctx)
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid room ID")
	}

	messages, err := s.getMessagesUC.Execute(ctx, token, roomID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	response := &chat.GetMessagesResponse{
		Messages: make([]*chat.Message, len(messages)),
	}

	for i, msg := range messages {
		response.Messages[i] = &chat.Message{
			Id:        msg.ID.String(),
			RoomId:    msg.RoomID.String(),
			UserId:    msg.UserID.String(),
			Content:   msg.Content,
			CreatedAt: timestamppb.New(msg.CreatedAt),
		}
	}

	return response, nil
}

func (s *ChatServiceServer) HandleWebSocket(ctx context.Context, wsConn *websocket.Conn) {
	defer wsConn.Close()

	var initMsg WSMessage
	if err := wsConn.ReadJSON(&initMsg); err != nil {
		s.logger.Error("Failed to read init message", zap.Error(err))
		return
	}

	userID, err := s.getUserIDFromToken(initMsg.Token)
	if err != nil {
		s.logger.Error("Invalid token",
			zap.Error(err),
			zap.String("token", initMsg.Token),
		)

		wsConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "Invalid authentication token",
		})
		return
	}

	roomID, err := uuid.Parse(initMsg.RoomID)
	if err != nil {
		s.logger.Error("Invalid room ID", zap.Error(err))
		wsConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "Invalid room ID",
		})
		return
	}

	conn := &WebSocketConnection{conn: wsConn}

	if err := s.connectUC.Execute(ctx, initMsg.Token, roomID, conn); err != nil {
		s.logger.Error("Failed to connect to chat",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("room_id", roomID.String()),
		)
		wsConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "Failed to connect to chat",
		})
		return
	}

	s.activeConnMu.Lock()
	s.activeConns[userID] = conn
	s.activeConnMu.Unlock()

	defer func() {
		s.activeConnMu.Lock()
		delete(s.activeConns, userID)
		s.activeConnMu.Unlock()

		if err := s.disconnectUC.Execute(ctx, initMsg.Token, roomID); err != nil {
			s.logger.Error("Failed to disconnect from chat",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("room_id", roomID.String()),
			)
		}
	}()

	wsConn.WriteJSON(map[string]string{
		"type":    "connected",
		"user_id": userID.String(),
		"room_id": roomID.String(),
	})

	for {
		var msg WSMessage
		if err := wsConn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket error",
					zap.Error(err),
					zap.String("user_id", userID.String()),
					zap.String("room_id", roomID.String()),
				)
			}
			break
		}

		currentUserID, err := s.getUserIDFromToken(msg.Token)
		if err != nil || currentUserID != userID {
			s.logger.Error("Invalid token in message",
				zap.Error(err),
				zap.String("user_id", userID.String()),
			)
			wsConn.WriteJSON(map[string]string{
				"type":    "error",
				"message": "Invalid authentication token",
			})
			continue
		}

		switch msg.Type {
		case "message":
			if err := s.sendMessageUC.Execute(ctx, msg.Token, roomID, msg.Content); err != nil {
				s.logger.Error("Failed to send message",
					zap.Error(err),
					zap.String("user_id", userID.String()),
					zap.String("room_id", roomID.String()),
				)
				wsConn.WriteJSON(map[string]string{
					"type":    "error",
					"message": "Failed to send message",
				})
				continue
			}
		}
	}
}

func (s *ChatServiceServer) getUserIDFromToken(token string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID, err := s.authClient.ValidateToken(ctx, token)
	if err != nil {
		s.logger.Error("Failed to validate token",
			zap.Error(err),
			zap.String("token", token),
		)
		return uuid.Nil, errors.Wrap(err, "failed to validate token")
	}

	return userID, nil
}

func (s *ChatServiceServer) broadcastEvent(ctx context.Context, event *entities.ChatEvent) {
	s.activeConnMu.RLock()
	defer s.activeConnMu.RUnlock()

	for userID, conn := range s.activeConns {
		go func(userID uuid.UUID, conn *WebSocketConnection) {
			if err := conn.SendEvent(event); err != nil {
				s.logger.Error("Failed to send event",
					zap.Error(err),
					zap.String("user_id", userID.String()),
					zap.String("room_id", event.RoomID.String()),
				)
			}
		}(userID, conn)
	}
}
