package auth

import (
	"context"
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ValidateResponse struct {
	UserID      uuid.UUID
	Permissions []string
}

type Client struct {
	logger       *zap.Logger
	conn         *grpc.ClientConn
	client       auth.AuthServiceClient
	serviceToken string
	mutex        sync.RWMutex
}

func NewClient(logger *zap.Logger, address string, serviceToken string) (*Client, error) {
	if serviceToken == "" {
		return nil, errors.New("service token is required")
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to auth service")
	}

	return &Client{
		logger:       logger,
		conn:         conn,
		client:       auth.NewAuthServiceClient(conn),
		serviceToken: serviceToken,
		mutex:        sync.RWMutex{},
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// ValidateToken validates access token and returns user info.
func (c *Client) ValidateToken(ctx context.Context, token string) (*ValidateResponse, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	}))

	resp, err := c.client.ValidateToken(ctx, &auth.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate token")
	}

	userID, err := uuid.Parse(resp.User.Id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user ID format")
	}

	return &ValidateResponse{
		UserID:      userID,
		Permissions: resp.Permissions,
	}, nil
}

// CheckUserExists verifies if user exists using service token.
func (c *Client) CheckUserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + c.serviceToken,
	}))

	_, err := c.client.GetUser(ctx, &auth.GetUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return false, nil // User not found or other error.
	}

	return true, nil
}
