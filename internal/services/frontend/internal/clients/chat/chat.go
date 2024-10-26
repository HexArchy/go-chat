package chat

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/chat"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/shared"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	logger    *zap.Logger
	client    chat.ChatServiceClient
	conn      *grpc.ClientConn
	authInter *shared.AuthInterceptor
}

func NewClient(logger *zap.Logger, address string, authInter *shared.AuthInterceptor) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInter.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(authInter.StreamClientInterceptor()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to chat service")
	}

	return &Client{
		logger:    logger,
		client:    chat.NewChatServiceClient(conn),
		conn:      conn,
		authInter: authInter,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// GetMessages retrieves chat messages with retry and timeout.
func (c *Client) GetMessages(ctx context.Context, roomID uuid.UUID, limit, offset int32) ([]*entities.Message, error) {
	c.logger.Debug("GetMessages: retrieving messages",
		zap.String("room_id", roomID.String()),
		zap.Int32("limit", limit),
		zap.Int32("offset", offset))

	var messages []*entities.Message

	err := shared.RetryWithBackoff(ctx, c.logger, shared.DefaultRetryConfig(), func() error {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		resp, err := c.client.GetMessages(ctx, &chat.GetMessagesRequest{
			RoomId: roomID.String(),
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return errors.Wrap(err, "failed to get messages")
		}

		messages = make([]*entities.Message, len(resp.Messages))
		for i, m := range resp.Messages {
			msg, err := protoToMessage(m)
			if err != nil {
				return errors.Wrap(err, "failed to convert message")
			}
			messages[i] = msg
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// SendMessage sends a new message to the chat.
func (c *Client) SendMessage(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, content string) error {
	return shared.RetryWithBackoff(ctx, c.logger, shared.DefaultRetryConfig(), func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_, err := c.client.SendMessage(ctx, &chat.SendMessageRequest{
			RoomId:  roomID.String(),
			UserId:  userID.String(),
			Content: content,
		})
		return err
	})
}

// Connect establishes WebSocket connection to chat.
func (c *Client) Connect(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) (chan *entities.ChatEvent, chan error, context.CancelFunc, error) {
	c.logger.Debug("Connect: establishing chat connection",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()))

	ctx, cancel := context.WithCancel(ctx)

	eventChan := make(chan *entities.ChatEvent, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errChan)

		stream, err := c.client.Connect(ctx, &chat.WebSocketRequest{
			RoomId: roomID.String(),
			UserId: userID.String(),
		})
		if err != nil {
			errChan <- errors.Wrap(err, "failed to initiate chat stream")
			return
		}

		for {
			event, err := stream.Recv()
			if err != nil {
				errChan <- errors.Wrap(err, "error receiving chat event")
				return
			}

			chatEvent, err := protoToChatEvent(event)
			if err != nil {
				c.logger.Error("Failed to convert chat event",
					zap.Error(err),
					zap.String("room_id", roomID.String()),
					zap.String("user_id", userID.String()))
				continue
			}

			select {
			case eventChan <- chatEvent:
			case <-ctx.Done():
				return
			}
		}
	}()

	return eventChan, errChan, cancel, nil
}

// Helper functions for proto conversion.
func protoToMessage(m *chat.Message) (*entities.Message, error) {
	messageID, err := uuid.Parse(m.Id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid message ID")
	}

	roomID, err := uuid.Parse(m.RoomId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid room ID")
	}

	userID, err := uuid.Parse(m.UserId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user ID")
	}

	return &entities.Message{
		ID:        messageID,
		RoomID:    roomID,
		UserID:    userID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.AsTime(),
	}, nil
}

func protoToChatEvent(e *chat.ChatEvent) (*entities.ChatEvent, error) {
	roomID, err := uuid.Parse(e.RoomId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid room ID")
	}

	userID, err := uuid.Parse(e.UserId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user ID")
	}

	event := &entities.ChatEvent{
		RoomID:    roomID,
		UserID:    userID,
		Type:      protoToEventType(e.EventType),
		Timestamp: e.Timestamp.AsTime(),
	}

	if e.Message != nil {
		msg, err := protoToMessage(e.Message)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert message")
		}
		event.Message = msg
	}

	return event, nil
}

func protoToEventType(t chat.ChatEvent_EventType) entities.EventType {
	switch t {
	case chat.ChatEvent_USER_JOINED:
		return entities.EventTypeUserJoined
	case chat.ChatEvent_USER_LEFT:
		return entities.EventTypeUserLeft
	case chat.ChatEvent_NEW_MESSAGE:
		return entities.EventTypeNewMessage
	default:
		return entities.EventTypeUnknown
	}
}
