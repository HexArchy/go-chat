package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/HexArch/go-chat/internal/services/website/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/website/internal/metrics"
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

var publicEndpoints = map[string]bool{
	"/website.RoomService/SearchRooms": true,
	"/website.RoomService/GetRoom":     true,
	"/website.RoomService/GetAllRooms": true,
}

type AuthMiddleware struct {
	logger     *zap.Logger
	authClient *auth.AuthClient
	metrics    *metrics.WebsiteMetrics
}

func NewAuthMiddleware(
	logger *zap.Logger,
	authClient *auth.AuthClient,
	metrics *metrics.WebsiteMetrics,
) *AuthMiddleware {
	return &AuthMiddleware{
		logger:     logger,
		authClient: authClient,
		metrics:    metrics,
	}
}

func (m *AuthMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		m.metrics.IncActiveRequests()
		defer m.metrics.DecActiveRequests()

		// Skip auth for public endpoints.
		if publicEndpoints[info.FullMethod] {
			return handler(ctx, req)
		}

		token, err := m.extractToken(ctx)
		if err != nil {
			m.logger.Debug("Failed to extract token",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			m.metrics.RecordError("token_extraction_failed")
			return nil, err
		}

		// Validate token and get user info.
		validationResp, err := m.authClient.ValidateToken(ctx, token)
		if err != nil {
			m.logger.Error("Token validation failed",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			m.metrics.RecordError("token_validation_failed")
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Create new context with user info.
		newCtx := context.WithValue(ctx, UserIDKey, validationResp.UserID)
		newCtx = context.WithValue(newCtx, PermissionsKey, validationResp.Permissions)

		// Record request duration.
		defer func() {
			duration := time.Since(start).Seconds()
			m.metrics.RecordRequestDuration(info.FullMethod, "success", duration)
		}()

		return handler(newCtx, req)
	}
}

func (m *AuthMiddleware) extractToken(ctx context.Context) (string, error) {
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

// Helper functions.
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

func HasPermission(ctx context.Context, requiredPermission string) bool {
	permissions, err := GetPermissionsFromContext(ctx)
	if err != nil {
		return false
	}

	for _, p := range permissions {
		if p == requiredPermission {
			return true
		}
	}
	return false
}

func (m *AuthMiddleware) ValidateRoomOwner(ctx context.Context, roomOwnerID string) error {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if userID != roomOwnerID {
		m.metrics.RecordError("permission_denied")
		return status.Error(codes.PermissionDenied, "not room owner")
	}

	return nil
}
