package controllers

import (
	"context"
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/chat"
	"github.com/HexArch/go-chat/internal/services/chat/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/chat/internal/controllers/middleware"
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/connect"
	"github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/disconnect"
	getmessages "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/get-messages"
	sendmessage "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/send-message"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	writeTimeout = 10 * time.Second
	readTimeout  = 60 * time.Second
	pingPeriod   = 30 * time.Second
)

type WSMessage struct {
	Type    string `json:"type"`
	RoomID  string `json:"room_id"`
	Content string `json:"content,omitempty"`
	Token   string `json:"token"`
}

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

func (s *ChatServiceServer) GetMessages(ctx context.Context, req *chat.GetMessagesRequest) (*chat.GetMessagesResponse, error) {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid room ID")
	}

	messages, err := s.getMessagesUC.Execute(ctx, roomID, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("Failed to get messages",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("room_id", roomID.String()),
		)
		return nil, status.Error(codes.Internal, "failed to get messages")
	}

	return &chat.GetMessagesResponse{
		Messages: messagesToProto(messages),
	}, nil
}

func (s *ChatServiceServer) HandleWebSocket(ctx context.Context, wsConn *websocket.Conn) {
	var initMsg WSMessage
	if err := wsConn.ReadJSON(&initMsg); err != nil {
		s.logger.Error("Failed to read init message", zap.Error(err))
		return
	}

	// Validate token.
	resp, err := s.authClient.ValidateToken(ctx, initMsg.Token)
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

	conn := &WebSocketConnection{
		conn:      wsConn,
		writeChan: make(chan *entities.ChatEvent, 100),
		done:      make(chan struct{}),
	}

	go conn.writeLoop()

	if err := s.connectUC.Execute(ctx, resp.UserID, roomID, conn); err != nil {
		s.logger.Error("Failed to connect to chat",
			zap.Error(err),
			zap.String("user_id", resp.UserID.String()),
			zap.String("room_id", roomID.String()),
		)
		wsConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "Failed to connect to chat",
		})
		return
	}

	s.activeConnMu.Lock()
	s.activeConns[resp.UserID] = conn
	s.activeConnMu.Unlock()

	defer func() {
		close(conn.done)
		s.activeConnMu.Lock()
		delete(s.activeConns, resp.UserID)
		s.activeConnMu.Unlock()

		if err := s.disconnectUC.Execute(ctx, resp.UserID, roomID); err != nil {
			s.logger.Error("Failed to disconnect from chat",
				zap.Error(err),
				zap.String("user_id", resp.UserID.String()),
				zap.String("room_id", roomID.String()),
			)
		}
	}()

	wsConn.WriteJSON(map[string]string{
		"type":    "connected",
		"user_id": resp.UserID.String(),
		"room_id": roomID.String(),
	})

	wsConn.SetReadLimit(1024 * 1024) // 1MB
	wsConn.SetReadDeadline(time.Now().Add(readTimeout))
	wsConn.SetPongHandler(func(string) error {
		wsConn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	for {
		var msg WSMessage
		if err := wsConn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket error",
					zap.Error(err),
					zap.String("user_id", resp.UserID.String()),
					zap.String("room_id", roomID.String()),
				)
			}
			break
		}

		// Revalidate token
		currentResp, err := s.authClient.ValidateToken(ctx, msg.Token)
		if err != nil || currentResp.UserID != resp.UserID {
			s.logger.Error("Invalid token in message",
				zap.Error(err),
				zap.String("user_id", resp.UserID.String()),
			)
			wsConn.WriteJSON(map[string]string{
				"type":    "error",
				"message": "Invalid authentication token",
			})
			continue
		}

		switch msg.Type {
		case "message":
			if err := s.sendMessageUC.Execute(ctx, currentResp.UserID, roomID, msg.Content); err != nil {
				s.logger.Error("Failed to send message",
					zap.Error(err),
					zap.String("user_id", resp.UserID.String()),
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

func (s *ChatServiceServer) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*emptypb.Empty, error) {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid room ID")
	}

	if err := s.sendMessageUC.Execute(ctx, userID, roomID, req.Content); err != nil {
		s.logger.Error("Failed to send message",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("room_id", roomID.String()),
		)
		return nil, status.Error(codes.Internal, "failed to send message")
	}

	return &emptypb.Empty{}, nil
}

func (s *ChatServiceServer) broadcastEvent(event *entities.ChatEvent) {
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

func messagesToProto(messages []*entities.Message) []*chat.Message {
	result := make([]*chat.Message, len(messages))
	for i, msg := range messages {
		result[i] = &chat.Message{
			Id:        msg.ID.String(),
			RoomId:    msg.RoomID.String(),
			UserId:    msg.UserID.String(),
			Content:   msg.Content,
			CreatedAt: timestamppb.New(msg.CreatedAt),
		}
	}
	return result
}
