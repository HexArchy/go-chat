package auth

import (
	"context"
	"sync"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	serviceTokenHeader = "X-Service-Token"
)

type AuthClient struct {
	logger       *zap.Logger
	client       auth.AuthServiceClient
	conn         *grpc.ClientConn
	serviceToken string
	mutex        sync.RWMutex
}

type ValidateResponse struct {
	UserID      string
	Permissions []string
}

func NewAuthClient(logger *zap.Logger, authServiceAddr string, serviceToken string) (*AuthClient, error) {
	conn, err := grpc.NewClient(
		authServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to AuthService")
	}

	return &AuthClient{
		logger:       logger,
		client:       auth.NewAuthServiceClient(conn),
		conn:         conn,
		serviceToken: serviceToken,
		mutex:        sync.RWMutex{},
	}, nil
}

// ValidateToken validates the access token using auth service.
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*ValidateResponse, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Create context with service token.
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + token,
	})
	ctxWithMetadata := metadata.NewOutgoingContext(ctx, md)

	resp, err := c.client.ValidateToken(ctxWithMetadata, &auth.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate token")
	}

	return &ValidateResponse{
		UserID:      resp.User.Id,
		Permissions: resp.Permissions,
	}, nil
}

// Close closes the gRPC connection.
func (c *AuthClient) Close() error {
	return c.conn.Close()
}
