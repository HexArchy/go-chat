package controllers

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	tokenKey  contextKey = "token"
	userIDKey contextKey = "userID"
)

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		var token string
		if values := md["authorization"]; len(values) > 0 {
			token = strings.TrimPrefix(values[0], "Bearer ")
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		ctx = context.WithValue(ctx, tokenKey, token)

		return handler(ctx, req)
	}
}

func getTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(tokenKey).(string); ok {
		return token
	}
	return ""
}
