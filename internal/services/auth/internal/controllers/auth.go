package controllers

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/HexArch/go-chat/internal/services/auth/internal/controllers/middleware"
	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/metrics"
	createuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/create-user"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/login"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/logout"
	refreshtoken "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/refresh-token"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/validatetoken"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthController struct {
	logger          *zap.Logger
	metrics         *metrics.AuthMetrics
	createUserUC    *createuser.UseCase
	loginUC         *login.UseCase
	refreshTokenUC  *refreshtoken.UseCase
	validateTokenUC *validatetoken.UseCase
	logoutUC        *logout.UseCase
	auth.UnimplementedAuthServiceServer
}

func NewAuthController(
	logger *zap.Logger,
	metrics *metrics.AuthMetrics,
	createUserUC *createuser.UseCase,
	loginUC *login.UseCase,
	refreshTokenUC *refreshtoken.UseCase,
	validateTokenUC *validatetoken.UseCase,
	logoutUC *logout.UseCase,
) *AuthController {
	return &AuthController{
		logger:          logger,
		metrics:         metrics,
		createUserUC:    createUserUC,
		loginUC:         loginUC,
		refreshTokenUC:  refreshTokenUC,
		validateTokenUC: validateTokenUC,
		logoutUC:        logoutUC,
	}
}

func (c *AuthController) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("auth", "login", time.Since(start).Seconds())
	}()

	c.logger.Debug("Login request received", zap.String("email", req.Email))

	result, err := c.loginUC.Execute(ctx, req.Email, req.Password)
	if err != nil {
		c.metrics.RecordError("login_failed")
		c.logger.Error("Login failed", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	c.metrics.RecordTokenValidation("login_success")
	return &auth.LoginResponse{
		AccessToken:           result.AccessToken,
		RefreshToken:          result.RefreshToken,
		AccessTokenExpiresAt:  timestamppb.New(result.AccessTokenExpiresAt),
		RefreshTokenExpiresAt: timestamppb.New(result.RefreshTokenExpiresAt),
	}, nil
}

func (c *AuthController) RegisterUser(ctx context.Context, req *auth.RegisterUserRequest) (*emptypb.Empty, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("auth", "register", time.Since(start).Seconds())
	}()

	c.logger.Debug("Register user request received", zap.String("email", req.Email))

	user := &entities.User{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
		Phone:    req.Phone,
		Age:      int(req.Age),
		Bio:      req.Bio,
	}

	if err := c.createUserUC.Execute(ctx, user); err != nil {
		c.metrics.RecordError("registration_failed")
		c.logger.Error("User registration failed", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (c *AuthController) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("auth", "refresh_token", time.Since(start).Seconds())
	}()

	result, err := c.refreshTokenUC.Execute(ctx, req.RefreshToken)
	if err != nil {
		c.metrics.RecordError("refresh_token_failed")
		c.logger.Error("Token refresh failed", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	c.metrics.RecordTokenValidation("refresh_success")
	return &auth.RefreshTokenResponse{
		AccessToken:           result.AccessToken,
		RefreshToken:          result.RefreshToken,
		AccessTokenExpiresAt:  timestamppb.New(result.AccessTokenExpiresAt),
		RefreshTokenExpiresAt: timestamppb.New(result.RefreshTokenExpiresAt),
	}, nil
}

func (c *AuthController) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("auth", "validate_token", time.Since(start).Seconds())
	}()

	user, err := c.validateTokenUC.Execute(ctx, req.Token)
	if err != nil {
		c.metrics.RecordError("validate_token_failed")
		c.logger.Error("Token validation failed", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	c.metrics.RecordTokenValidation("validation_success")
	return &auth.ValidateTokenResponse{
		User: c.mapUserToProto(user),
	}, nil
}

func (c *AuthController) Logout(ctx context.Context, req *auth.LogoutRequest) (*emptypb.Empty, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("auth", "logout", time.Since(start).Seconds())
	}()

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		c.metrics.RecordError("logout_unauthorized")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if err := c.logoutUC.Execute(ctx, userID); err != nil {
		c.metrics.RecordError("logout_failed")
		c.logger.Error("Logout failed", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (c *AuthController) mapErrorToStatus(err error) error {
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

func (c *AuthController) mapUserToProto(user *entities.User) *auth.User {
	return &auth.User{
		Id:          user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		Phone:       user.Phone,
		Age:         int32(user.Age),
		Bio:         user.Bio,
		Permissions: c.permissionsToStrings(user.Permissions),
		CreatedAt:   timestamppb.New(user.CreatedAt),
		UpdatedAt:   timestamppb.New(user.UpdatedAt),
	}
}

func (c *AuthController) permissionsToStrings(perms []entities.Permission) []string {
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = string(p)
	}
	return result
}
