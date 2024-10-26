package middleware

import (
	"context"
	"strings"

	"github.com/HexArch/go-chat/internal/services/chat/internal/clients/auth"
	"github.com/google/uuid"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ContextKey string

const (
	TokenKey       ContextKey = "token"
	UserIDKey      ContextKey = "user_id"
	PermissionsKey ContextKey = "permissions"
)

type AuthMiddleware struct {
	logger     *zap.Logger
	authClient *auth.Client
}

func NewAuthMiddleware(logger *zap.Logger, authClient *auth.Client) *AuthMiddleware {
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
		token, err := extractToken(ctx)
		if err != nil {
			m.logger.Debug("Token extraction failed",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return nil, err
		}

		// Validate token
		validationResp, err := m.authClient.ValidateToken(ctx, token)
		if err != nil {
			m.logger.Error("Token validation failed",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Add auth info to context.
		ctx = context.WithValue(ctx, TokenKey, token)
		ctx = context.WithValue(ctx, UserIDKey, validationResp.UserID)
		ctx = context.WithValue(ctx, PermissionsKey, validationResp.Permissions)

		return handler(ctx, req)
	}
}

func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no metadata provided")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "no authorization header")
	}

	token := values[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(token, "Bearer "), nil
}

// Helper functions.
func GetTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(TokenKey).(string); ok {
		return token
	}
	return ""
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID, nil
	}
	return uuid.Nil, status.Error(codes.Internal, "user ID not found in context")
}

func GetPermissionsFromContext(ctx context.Context) ([]string, error) {
	if permissions, ok := ctx.Value(PermissionsKey).([]string); ok {
		return permissions, nil
	}
	return nil, status.Error(codes.Internal, "permissions not found in context")
}
