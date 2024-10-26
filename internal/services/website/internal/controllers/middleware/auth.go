package middleware

import (
	"context"
	"strings"

	"github.com/HexArch/go-chat/internal/services/website/internal/clients/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ContextKey string

const (
	UserIDKey           ContextKey = "user_id"
	PermissionsKey      ContextKey = "permissions"
	AuthorizationHeader            = "authorization"
	BearerPrefix                   = "Bearer "
)

// publicEndpoints maps endpoints that don't require authentication.
var publicEndpoints = map[string]bool{
	"/website.RoomService/SearchRooms": true,
	"/website.RoomService/GetRoom":     true,
}

type AuthMiddleware struct {
	logger     *zap.Logger
	authClient *auth.AuthClient
}

func NewAuthMiddleware(logger *zap.Logger, authClient *auth.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		logger:     logger,
		authClient: authClient,
	}
}

func (m *AuthMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for public endpoints.
		if publicEndpoints[info.FullMethod] {
			return handler(ctx, req)
		}

		token, err := extractToken(ctx)
		if err != nil {
			m.logger.Debug("Failed to extract token",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return nil, err
		}

		// Validate token and get user info.
		validationResp, err := m.authClient.ValidateToken(ctx, token)
		if err != nil {
			m.logger.Error("Token validation failed",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Create new context with user info.
		newCtx := context.WithValue(ctx, UserIDKey, validationResp.UserID)
		newCtx = context.WithValue(newCtx, PermissionsKey, validationResp.Permissions)

		return handler(newCtx, req)
	}
}

func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no metadata provided")
	}

	values := md.Get(AuthorizationHeader)
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "no authorization header")
	}

	token := values[0]
	if !strings.HasPrefix(token, BearerPrefix) {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(token, BearerPrefix), nil
}

// Helper functions for controllers.
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", status.Error(codes.Internal, "user ID not found in context")
	}
	return userID, nil
}

func GetPermissionsFromContext(ctx context.Context) ([]string, error) {
	permissions, ok := ctx.Value(PermissionsKey).([]string)
	if !ok {
		return nil, status.Error(codes.Internal, "permissions not found in context")
	}
	return permissions, nil
}
