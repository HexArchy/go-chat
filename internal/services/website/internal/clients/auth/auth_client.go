package auth

import (
	"context"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	client    auth.AuthServiceClient
	jwtSecret []byte
}

func NewAuthClient(authServiceAddr string, jwtSecret []byte) (*AuthClient, error) {
	conn, err := grpc.NewClient(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to AuthService")
	}

	client := auth.NewAuthServiceClient(conn)

	return &AuthClient{
		client:    client,
		jwtSecret: jwtSecret,
	}, nil
}

// ValidateToken parses the JWT token, validates it and retrieves the user information.
func (c *AuthClient) ValidateToken(ctx context.Context, tokenString string) (*auth.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return c.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.Wrap(err, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid token: user_id not found")
	}

	getUserRequest := &auth.GetUserRequest{
		UserId: userID,
	}

	user, err := c.client.GetUser(ctx, getUserRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user from AuthService")
	}

	return user, nil
}
