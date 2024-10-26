package controllers

import (
	"context"
	"strings"

	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/validatetoken"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userKeyType string

const (
	userIDKey    userKeyType = "userID"
	bearerPrefix             = "Bearer "
)

var publicMethods = map[string]bool{
	"/auth.AuthService/Login":         true,
	"/auth.AuthService/RegisterUser":  true,
	"/auth.AuthService/RefreshToken":  true,
	"/auth.AuthService/ValidateToken": true,
}

func AuthInterceptor(logger *zap.Logger, validateTokenUC *validatetoken.UseCase, serviceToken string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for public methods.
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		token, err := extractToken(ctx)
		if err != nil {
			logger.Debug("Failed to extract token",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return nil, err
		}

		// Check service token.
		if token == serviceToken {
			return handler(ctx, req)
		}

		// Validate user token.
		user, err := validateTokenUC.Execute(ctx, token)
		if err != nil {
			logger.Debug("Token validation failed",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Add user ID to context.
		newCtx := context.WithValue(ctx, userIDKey, user.ID)
		return handler(newCtx, req)
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
	if !strings.HasPrefix(token, bearerPrefix) {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(token, bearerPrefix), nil
}
