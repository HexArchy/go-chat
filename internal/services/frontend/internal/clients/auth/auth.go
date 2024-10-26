package auth

import (
	"context"
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TokenResponse struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

type Client struct {
	logger *zap.Logger
	client auth.AuthServiceClient
	conn   *grpc.ClientConn
	mutex  sync.RWMutex
}

func NewClient(logger *zap.Logger, address string) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to auth service")
	}

	return &Client{
		logger: logger,
		client: auth.NewAuthServiceClient(conn),
		conn:   conn,
		mutex:  sync.RWMutex{},
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Register creates a new user account.
func (c *Client) Register(ctx context.Context, email, password, username, phone string, age int32, bio string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := c.client.RegisterUser(ctx, &auth.RegisterUserRequest{
		Email:    email,
		Password: password,
		Username: username,
		Phone:    phone,
		Age:      age,
		Bio:      bio,
	})
	if err != nil {
		return errors.Wrap(err, "failed to register user")
	}
	return nil
}

// Login authenticates user and returns tokens.
func (c *Client) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.client.Login(ctx, &auth.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}

	return &TokenResponse{
		AccessToken:           resp.AccessToken,
		RefreshToken:          resp.RefreshToken,
		AccessTokenExpiresAt:  resp.AccessTokenExpiresAt.AsTime(),
		RefreshTokenExpiresAt: resp.RefreshTokenExpiresAt.AsTime(),
	}, nil
}

// RefreshToken refreshes the access token using refresh token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.client.RefreshToken(ctx, &auth.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh token")
	}

	return &TokenResponse{
		AccessToken:           resp.AccessToken,
		RefreshToken:          resp.RefreshToken,
		AccessTokenExpiresAt:  resp.AccessTokenExpiresAt.AsTime(),
		RefreshTokenExpiresAt: resp.RefreshTokenExpiresAt.AsTime(),
	}, nil
}

// ValidateToken validates the access token and returns user info.
func (c *Client) ValidateToken(ctx context.Context) (*entities.User, []string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token, err := entities.GetAccessTokenFromContext(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get access token")
	}

	resp, err := c.client.ValidateToken(ctx, &auth.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to validate token")
	}

	user, err := protoToUser(resp.User)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get user")
	}

	return user, resp.Permissions, nil
}

// GetUser returns user information by ID.
func (c *Client) GetUser(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &auth.GetUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	user, err := protoToUser(resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	return user, nil
}

// GetUser updates user information.
func (c *Client) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	c.logger.Debug("UpdateUser: sending update request",
		zap.String("user_id", userID))

	req := &auth.UpdateUserRequest{
		UserId:      userID,
		Email:       "",
		Password:    "",
		Username:    "",
		Phone:       "",
		Age:         0,
		Bio:         "",
		Permissions: []string{},
	}

	for key, value := range updates {
		switch key {
		case "email":
			if v, ok := value.(string); ok {
				req.Email = v
			}
		case "password":
			if v, ok := value.(string); ok {
				req.Password = v
			}
		case "username":
			if v, ok := value.(string); ok {
				req.Username = v
			}
		case "phone":
			if v, ok := value.(string); ok {
				req.Phone = v
			}
		case "age":
			if v, ok := value.(int32); ok {
				req.Age = v
			} else if v, ok := value.(int); ok {
				req.Age = int32(v)
			}
		case "bio":
			if v, ok := value.(string); ok {
				req.Bio = v
			}
		case "permissions":
			if v, ok := value.([]string); ok {
				req.Permissions = v
			}
		}
	}

	_, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		c.logger.Error("UpdateUser: failed to update user", zap.Error(err))
		return errors.Wrap(err, "failed to update user")
	}

	c.logger.Debug("UpdateUser: user updated successfully")
	return nil
}

// Logout logs out a user by invalidating the access token.
func (c *Client) Logout(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	c.logger.Debug("Logout: sending logout request")

	req := &auth.LogoutRequest{}

	_, err := c.client.Logout(ctx, req)
	if err != nil {
		c.logger.Error("Logout: logout failed", zap.Error(err))
		return errors.Wrap(err, "logout failed")
	}

	c.logger.Debug("Logout: logout successful")
	return nil
}

func protoToUser(u *auth.User) (*entities.User, error) {
	userID, err := uuid.Parse(u.Id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user ID")
	}

	return &entities.User{
		ID:          userID,
		Email:       u.Email,
		Username:    u.Username,
		Phone:       u.Phone,
		Age:         u.Age,
		Bio:         u.Bio,
		Permissions: u.Permissions,
		CreatedAt:   u.CreatedAt.AsTime(),
		UpdatedAt:   u.UpdatedAt.AsTime(),
	}, nil
}
