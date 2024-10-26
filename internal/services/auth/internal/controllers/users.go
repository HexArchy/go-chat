package controllers

import (
	"context"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RegisterUser handles user registration.
func (s *AuthServiceServer) RegisterUser(ctx context.Context, req *auth.RegisterUserRequest) (*emptypb.Empty, error) {
	s.logger.Debug("Register user request received", zap.String("email", req.Email))

	user := &entities.User{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
		Phone:    req.Phone,
		Age:      int(req.Age),
		Bio:      req.Bio,
	}

	if err := s.createUserUC.Execute(ctx, user); err != nil {
		s.logger.Error("User registration failed", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	s.logger.Info("User registered successfully", zap.String("email", req.Email))
	return &emptypb.Empty{}, nil
}

// GetUser handles retrieving a single user.
func (s *AuthServiceServer) GetUser(ctx context.Context, req *auth.GetUserRequest) (*auth.User, error) {
	s.logger.Debug("Get user request received", zap.String("userID", req.UserId))

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	user, err := s.getUserUC.Execute(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	return mapUserToProto(user), nil
}

// GetUsers handles retrieving a list of users with pagination.
func (s *AuthServiceServer) GetUsers(ctx context.Context, req *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
	s.logger.Debug("Get users request received",
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset))

	limit := int(req.Limit)
	offset := int(req.Offset)

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	users, err := s.getUsersUC.Execute(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get users", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	response := &auth.GetUsersResponse{
		Users: make([]*auth.User, len(users)),
	}

	for i, user := range users {
		response.Users[i] = mapUserToProto(user)
	}

	return response, nil
}

// UpdateUser handles user updates.
func (s *AuthServiceServer) UpdateUser(ctx context.Context, req *auth.UpdateUserRequest) (*emptypb.Empty, error) {
	s.logger.Debug("Update user request received", zap.String("userID", req.UserId))

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	// Get requester ID from context (set by middleware).
	requesterID, err := getUserIDFromContext(ctx)
	if err != nil {
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
		Permissions: stringsToPermissions(req.Permissions),
	}

	if err := s.updateUserUC.Execute(ctx, requesterID, user); err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	s.logger.Info("User updated successfully", zap.String("userID", req.UserId))
	return &emptypb.Empty{}, nil
}

// DeleteUser handles user deletion.
func (s *AuthServiceServer) DeleteUser(ctx context.Context, req *auth.DeleteUserRequest) (*emptypb.Empty, error) {
	s.logger.Debug("Delete user request received", zap.String("userID", req.UserId))

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	requesterID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if err := s.deleteUserUC.Execute(ctx, requesterID, userID); err != nil {
		s.logger.Error("Failed to delete user", zap.Error(err))
		return nil, mapErrorToStatus(err)
	}

	s.logger.Info("User deleted successfully", zap.String("userID", req.UserId))
	return &emptypb.Empty{}, nil
}

// Helper functions for context and permissions.
func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, status.Error(codes.Internal, "user ID not found in context")
	}
	return userID, nil
}

func permissionsToStrings(perms []entities.Permission) []string {
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = string(p)
	}
	return result
}

func stringsToPermissions(perms []string) []entities.Permission {
	result := make([]entities.Permission, len(perms))
	for i, p := range perms {
		result[i] = entities.Permission(p)
	}
	return result
}
