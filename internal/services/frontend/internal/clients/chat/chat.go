package chat

import (
	"context"
	"fmt"
	"io"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/chat"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	client chat.ChatServiceClient
	conn   *grpc.ClientConn
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection")
	}

	return &Client{
		client: chat.NewChatServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetMessages(ctx context.Context, token string, roomID uuid.UUID, limit, offset int32) ([]*entities.Message, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := c.client.GetMessages(ctx, &chat.GetMessagesRequest{
		RoomId: roomID.String(),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get messages")
	}

	messages := make([]*entities.Message, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		msg, err := protoToMessage(m)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert message")
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (c *Client) Connect(ctx context.Context, token string, roomID, userID uuid.UUID) (chan *entities.ChatEvent, chan error, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
	stream, err := c.client.Connect(ctx, &chat.WebSocketRequest{
		RoomId: roomID.String(),
		UserId: userID.String(),
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to connect to chat")
	}

	eventChan := make(chan *entities.ChatEvent)
	errorChan := make(chan error)

	go func() {
		defer close(eventChan)
		defer close(errorChan)

		for {
			event, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				select {
				case errorChan <- errors.Wrap(err, "error receiving message"):
				case <-ctx.Done():
				}
				return
			}

			chatEvent, err := protoToChatEvent(event)
			if err != nil {
				select {
				case errorChan <- errors.Wrap(err, "error converting event"):
				case <-ctx.Done():
				}
				continue
			}

			select {
			case eventChan <- chatEvent:
			case <-ctx.Done():
				return
			}
		}
	}()

	return eventChan, errorChan, nil
}

func (c *Client) SendMessage(ctx context.Context, token string, roomID, userID uuid.UUID, content string) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
	_, err := c.client.SendMessage(ctx, &chat.SendMessageRequest{
		RoomId:  roomID.String(),
		UserId:  userID.String(),
		Content: content,
	})
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}
	return nil
}

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
		return nil, errors.Wrap(err, "invalid room ID in event")
	}

	userID, err := uuid.Parse(e.UserId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user ID in event")
	}

	chatEvent := &entities.ChatEvent{
		RoomID:    roomID,
		UserID:    userID,
		Type:      protoToEventType(e.EventType),
		Timestamp: e.Timestamp.AsTime(),
	}

	if e.Message != nil {
		message, err := protoToMessage(e.Message)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert message in event")
		}
		chatEvent.Message = message
	}

	return chatEvent, nil
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

func WithAuthToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}
