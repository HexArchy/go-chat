package website

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client provides access to the website service.
type Client struct {
	logger       *zap.Logger
	conn         *grpc.ClientConn
	client       website.RoomServiceClient
	serviceToken string
}

// NewClient creates a new website service client.
func NewClient(logger *zap.Logger, address string, serviceToken string) (*Client, error) {
	if serviceToken == "" {
		return nil, errors.New("service token is required")
	}

	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to website service")
	}

	return &Client{
		logger:       logger,
		conn:         conn,
		client:       website.NewRoomServiceClient(conn),
		serviceToken: serviceToken,
	}, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// createServiceContext creates a context with service token.
func (c *Client) createServiceContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + c.serviceToken,
	}))
}

// RoomExists checks if a room exists.
func (c *Client) RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	ctx = c.createServiceContext(ctx)

	_, err := c.client.GetRoom(ctx, &website.GetRoomRequest{
		RoomId: roomID.String(),
	})
	if err != nil {
		return false, nil
	}

	return true, nil
}
