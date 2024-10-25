package auth

import (
	"context"
	"fmt"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn         *grpc.ClientConn
	client       auth.AuthServiceClient
	serviceToken string
}

func NewClient(address string, serviceToken string) (*Client, error) {
	if serviceToken == "" {
		return nil, errors.New("service token is required")
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection")
	}

	return &Client{
		conn:         conn,
		client:       auth.NewAuthServiceClient(conn),
		serviceToken: serviceToken,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) createAuthContext(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", c.serviceToken),
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *Client) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
	user, err := c.GetUser(ctx, "")
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to validate token")
	}

	userID, err := uuid.Parse(user.Id)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to parse uuid")
	}

	return userID, nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*auth.User, error) {
	req := &auth.GetUserRequest{
		UserId: userID,
	}

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	return resp, nil
}

func (c *Client) CheckUserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	ctx = c.createAuthContext(ctx)

	user, err := c.GetUser(ctx, userID.String())
	if err != nil {
		return false, errors.Wrap(err, "failed to check user existence")
	}

	if user == nil {
		return false, nil
	}

	return true, nil
}
