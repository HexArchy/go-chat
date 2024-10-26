package controllers

import (
	"context"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	createuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/create-user"
	deleteuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/delete-user"
	getuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/get-user"
	getusers "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/get-users"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/login"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/logout"
	refreshtoken "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/refresh-token"
	updateuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/update-user"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/validatetoken"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthServiceServer struct {
	logger *zap.Logger
	auth.UnimplementedAuthServiceServer
	createUserUC    *createuser.UseCase
	loginUC         *login.UseCase
	refreshTokenUC  *refreshtoken.UseCase
	validateTokenUC *validatetoken.UseCase
	logoutUC        *logout.UseCase
	getUserUC       *getuser.UseCase
	getUsersUC      *getusers.UseCase
	updateUserUC    *updateuser.UseCase
	deleteUserUC    *deleteuser.UseCase
}

func NewAuthServiceServer(
	logger *zap.Logger,
	createUserUC *createuser.UseCase,
	loginUC *login.UseCase,
	refreshTokenUC *refreshtoken.UseCase,
	validateTokenUC *validatetoken.UseCase,
	logoutUC *logout.UseCase,
	getUserUC *getuser.UseCase,
	getUsersUC *getusers.UseCase,
	updateUserUC *updateuser.UseCase,
	deleteUserUC *deleteuser.UseCase,
) *AuthServiceServer {
	return &AuthServiceServer{
		logger:          logger,
		createUserUC:    createUserUC,
		loginUC:         loginUC,
		refreshTokenUC:  refreshTokenUC,
		validateTokenUC: validateTokenUC,
		logoutUC:        logoutUC,
		getUserUC:       getUserUC,
		getUsersUC:      getUsersUC,
		updateUserUC:    updateUserUC,
		deleteUserUC:    deleteUserUC,
	}
}

func (s *AuthServiceServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	s.logger.Debug("Login request received", zap.String("email", req.Email))

	result, err := s.loginUC.Execute(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.Error("Login failed", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	return &auth.LoginResponse{
		AccessToken:           result.AccessToken,
		RefreshToken:          result.RefreshToken,
		AccessTokenExpiresAt:  timestamppb.New(result.AccessTokenExpiresAt),
		RefreshTokenExpiresAt: timestamppb.New(result.RefreshTokenExpiresAt),
	}, nil
}

func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	s.logger.Debug("Refresh token request received")

	result, err := s.refreshTokenUC.Execute(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error("Token refresh failed", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	return &auth.RefreshTokenResponse{
		AccessToken:           result.AccessToken,
		RefreshToken:          result.RefreshToken,
		AccessTokenExpiresAt:  timestamppb.New(result.AccessTokenExpiresAt),
		RefreshTokenExpiresAt: timestamppb.New(result.RefreshTokenExpiresAt),
	}, nil
}

func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	s.logger.Debug("Validate token request received")

	user, err := s.validateTokenUC.Execute(ctx, req.Token)
	if err != nil {
		s.logger.Error("Token validation failed", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	return &auth.ValidateTokenResponse{
		User: mapUserToProto(user),
	}, nil
}

func (s *AuthServiceServer) Logout(ctx context.Context, req *auth.LogoutRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	s.logger.Debug("Logout request received", zap.String("userID", userID.String()))

	if err := s.logoutUC.Execute(ctx, userID); err != nil {
		s.logger.Error("Logout failed", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

// Helper functions.
func mapUserToProto(user *entities.User) *auth.User {
	return &auth.User{
		Id:          user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		Phone:       user.Phone,
		Age:         int32(user.Age),
		Bio:         user.Bio,
		Permissions: permissionsToStrings(user.Permissions),
		CreatedAt:   timestamppb.New(user.CreatedAt),
		UpdatedAt:   timestamppb.New(user.UpdatedAt),
	}
}

func mapErrorToStatus(err error) error {
	switch {
	case errors.Is(err, entities.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, entities.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, entities.ErrRefreshTokenExpired):
		return status.Error(codes.Unauthenticated, "refresh token expired")
	case errors.Is(err, entities.ErrTokenNotFound):
		return status.Error(codes.NotFound, "token not found")
	case errors.Is(err, entities.ErrInvalidTokenClaims):
		return status.Error(codes.InvalidArgument, "invalid token claims")
	case errors.Is(err, entities.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, "permission denied")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
