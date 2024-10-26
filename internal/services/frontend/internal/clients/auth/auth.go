package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	client auth.AuthServiceClient
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
		client: auth.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Register creates a new user account.
func (c *Client) Register(ctx context.Context, email, password, username, phone string, age int32, bio string) error {
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

// Login authenticates user and returns access & refresh tokens.
func (c *Client) Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error) {
	resp, err := c.client.Login(ctx, &auth.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to login")
	}
	return resp.AccessToken, resp.RefreshToken, nil
}

// RefreshToken refreshes the access token using refresh token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error) {
	resp, err := c.client.RefreshToken(ctx, &auth.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to refresh token")
	}
	return resp.AccessToken, resp.RefreshToken, nil
}

// GetUser returns user information by ID. Requires authentication.
func (c *Client) GetUser(ctx context.Context, token, userID string) (*entities.User, error) {
	ctx = addTokenToContext(ctx, token)

	resp, err := c.client.GetUser(ctx, &auth.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	return protoToUser(resp), nil
}

// ValidateToken validates the access token and returns the user.
func (c *Client) ValidateToken(ctx context.Context, token string) (*entities.User, error) {
	userID, err := extractUserIDFromToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract user ID from token")
	}

	ctx = addTokenToContext(ctx, token)

	resp, err := c.client.GetUser(ctx, &auth.GetUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate token")
	}

	return protoToUser(resp), nil
}

// UpdateUser updates user information. Requires authentication.
func (c *Client) UpdateUser(ctx context.Context, token string, userID string, updates map[string]interface{}) error {
	ctx = addTokenToContext(ctx, token)

	req := &auth.UpdateUserRequest{
		UserId: userID,
	}

	for key, value := range updates {
		switch key {
		case "email":
			req.Email = value.(string)
		case "password":
			req.Password = value.(string)
		case "username":
			req.Username = value.(string)
		case "phone":
			req.Phone = value.(string)
		case "age":
			req.Age = int32(value.(int))
		case "bio":
			req.Bio = value.(string)
		case "permissions":
			req.Permissions = value.([]string)
		}
	}

	_, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}
	return nil
}

// Logout invalidates the user's refresh tokens. Requires authentication.
func (c *Client) Logout(ctx context.Context, token string) error {
	ctx = addTokenToContext(ctx, token)

	_, err := c.client.Logout(ctx, &auth.LogoutRequest{})
	if err != nil {
		return errors.Wrap(err, "failed to logout")
	}
	return nil
}

// GetUsers returns a list of users with pagination. Requires authentication.
func (c *Client) GetUsers(ctx context.Context, token string, limit, offset int32) ([]*entities.User, error) {
	ctx = addTokenToContext(ctx, token)

	resp, err := c.client.GetUsers(ctx, &auth.GetUsersRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get users")
	}

	users := make([]*entities.User, len(resp.Users))
	for i, u := range resp.Users {
		users[i] = protoToUser(u)
	}
	return users, nil
}

// Helper functions.
func addTokenToContext(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
}

func protoToUser(u *auth.User) *entities.User {
	userID, err := uuid.Parse(u.Id)
	if err != nil {
		return nil
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
	}
}

func extractUserIDFromToken(tokenString string) (uuid.UUID, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to parse token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("user_id not found in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "invalid user_id format in token")
	}

	return userID, nil
}
