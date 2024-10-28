package controllers

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/HexArch/go-chat/internal/services/auth/internal/controllers/middleware"
	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/metrics"
	deleteuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/delete-user"
	getuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/get-user"
	getusers "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/get-users"
	updateuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/update-user"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UsersController struct {
	logger       *zap.Logger
	metrics      *metrics.AuthMetrics
	getUserUC    *getuser.UseCase
	getUsersUC   *getusers.UseCase
	updateUserUC *updateuser.UseCase
	deleteUserUC *deleteuser.UseCase
	auth.UnimplementedAuthServiceServer
}

func NewUsersController(
	logger *zap.Logger,
	metrics *metrics.AuthMetrics,
	getUserUC *getuser.UseCase,
	getUsersUC *getusers.UseCase,
	updateUserUC *updateuser.UseCase,
	deleteUserUC *deleteuser.UseCase,
) *UsersController {
	return &UsersController{
		logger:       logger,
		metrics:      metrics,
		getUserUC:    getUserUC,
		getUsersUC:   getUsersUC,
		updateUserUC: updateUserUC,
		deleteUserUC: deleteUserUC,
	}
}

func (c *UsersController) GetUser(ctx context.Context, req *auth.GetUserRequest) (*auth.User, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("users", "get_user", time.Since(start).Seconds())
	}()

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		c.metrics.RecordError("invalid_user_id")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	user, err := c.getUserUC.Execute(ctx, userID)
	if err != nil {
		c.metrics.RecordError("get_user_failed")
		c.logger.Error("Failed to get user", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	return c.mapUserToProto(user), nil
}

func (c *UsersController) GetUsers(ctx context.Context, req *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("users", "get_users", time.Since(start).Seconds())
	}()

	limit, offset := c.normalizePageParams(req.Limit, req.Offset)

	users, err := c.getUsersUC.Execute(ctx, limit, offset)
	if err != nil {
		c.metrics.RecordError("get_users_failed")
		c.logger.Error("Failed to get users", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	response := &auth.GetUsersResponse{
		Users: make([]*auth.User, len(users)),
	}

	for i, user := range users {
		response.Users[i] = c.mapUserToProto(user)
	}

	return response, nil
}

func (c *UsersController) UpdateUser(ctx context.Context, req *auth.UpdateUserRequest) (*emptypb.Empty, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("users", "update_user", time.Since(start).Seconds())
	}()

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		c.metrics.RecordError("invalid_user_id")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	requesterID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		c.metrics.RecordError("update_unauthorized")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	user := &entities.User{
		ID:          userID,
		Email:       req.Email,
		Password:    req.Password,
		Username:    req.Username,
		Phone:       req.Phone,
		Age:         int(req.Age),
		Bio:         req.Bio,
		Permissions: c.stringsToPermissions(req.Permissions),
	}

	if err := c.updateUserUC.Execute(ctx, requesterID, user); err != nil {
		c.metrics.RecordError("update_user_failed")
		c.logger.Error("Failed to update user", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (c *UsersController) DeleteUser(ctx context.Context, req *auth.DeleteUserRequest) (*emptypb.Empty, error) {
	start := time.Now()
	defer func() {
		c.metrics.RecordRequestDuration("users", "delete_user", time.Since(start).Seconds())
	}()

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		c.metrics.RecordError("invalid_user_id")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	requesterID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		c.metrics.RecordError("delete_unauthorized")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if err := c.deleteUserUC.Execute(ctx, requesterID, userID); err != nil {
		c.metrics.RecordError("delete_user_failed")
		c.logger.Error("Failed to delete user", zap.Error(err))
		return nil, c.mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

// Helper methods
func (c *UsersController) normalizePageParams(limit, offset int32) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return int(limit), int(offset)
}

func (c *UsersController) mapUserToProto(user *entities.User) *auth.User {
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

func (c *UsersController) mapErrorToStatus(err error) error {
	switch {
	case errors.Is(err, entities.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, entities.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, entities.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, "permission denied")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func (c *UsersController) permissionsToStrings(perms []entities.Permission) []string {
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = string(p)
	}
	return result
}

func (c *UsersController) stringsToPermissions(perms []string) []entities.Permission {
	result := make([]entities.Permission, len(perms))
	for i, p := range perms {
		result[i] = entities.Permission(p)
	}
	return result
}
