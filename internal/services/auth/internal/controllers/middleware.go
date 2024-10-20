package controllers

import (
	"context"
	"strings"

	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/validatetoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userKeyType string

const userIDKey userKeyType = "userID"

func AuthInterceptor(jwtSecret []byte, validateTokenUC *validatetoken.UseCase) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if info.FullMethod == "/auth.AuthService/Login" || info.FullMethod == "/auth.AuthService/RegisterUser" {
			return handler(ctx, req)
		}

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

		user, err := validateTokenUC.Execute(ctx, token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = context.WithValue(ctx, userIDKey, user.ID)

		return handler(ctx, req)
	}
}
