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
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthServiceServer struct {
	auth.UnimplementedAuthServiceServer
	createUserUC    *createuser.UseCase
	loginUC         *login.UseCase
	refreshTokenUC  *refreshtoken.UseCase
	getUserUC       *getuser.UseCase
	updateUserUC    *updateuser.UseCase
	validateTokenUC *validatetoken.UseCase
	logoutUC        *logout.UseCase
	getUsersUC      *getusers.UseCase
	deleteUserUC    *deleteuser.UseCase
}

func NewAuthServiceServer(
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

func (s *AuthServiceServer) RegisterUser(ctx context.Context, req *auth.RegisterUserRequest) (*emptypb.Empty, error) {
	user := &entities.User{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
		Phone:    req.Phone,
		Age:      int(req.Age),
		Bio:      req.Bio,
	}

	if err := s.createUserUC.Execute(ctx, user); err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *AuthServiceServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	accessToken, refreshToken, err := s.loginUC.Execute(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	newAccessToken, newRefreshToken, err := s.refreshTokenUC.Execute(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &auth.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthServiceServer) GetUser(ctx context.Context, req *auth.GetUserRequest) (*auth.User, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	user, err := s.getUserUC.Execute(ctx, userID)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

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
	}, nil
}

func (s *AuthServiceServer) UpdateUser(ctx context.Context, req *auth.UpdateUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
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

	requesterID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
	}

	if err := s.updateUserUC.Execute(ctx, requesterID, user); err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *AuthServiceServer) DeleteUser(ctx context.Context, req *auth.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	requesterID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
	}

	if err := s.deleteUserUC.Execute(ctx, requesterID, userID); err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

// Вспомогательные функции

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

func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return userID, nil
}

func mapErrorToStatus(err error) error {
	switch {
	case errors.Is(err, entities.ErrUserNotFound):
		return status.Errorf(codes.NotFound, err.Error())
	case errors.Is(err, entities.ErrInvalidCredentials):
		return status.Errorf(codes.Unauthenticated, err.Error())
	case errors.Is(err, entities.ErrPermissionDenied):
		return status.Errorf(codes.PermissionDenied, err.Error())
	case errors.Is(err, entities.ErrInvalidTokenClaims):
		return status.Errorf(codes.Unauthenticated, err.Error())
	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}

func (s *AuthServiceServer) Logout(ctx context.Context, req *auth.LogoutRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
	}

	if err := s.logoutUC.Execute(ctx, userID); err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *AuthServiceServer) GetUsers(ctx context.Context, req *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
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
		return nil, mapErrorToStatus(err)
	}

	var userList []*auth.User
	for _, user := range users {
		userList = append(userList, &auth.User{
			Id:          user.ID.String(),
			Email:       user.Email,
			Username:    user.Username,
			Phone:       user.Phone,
			Age:         int32(user.Age),
			Bio:         user.Bio,
			Permissions: permissionsToStrings(user.Permissions),
			CreatedAt:   timestamppb.New(user.CreatedAt),
			UpdatedAt:   timestamppb.New(user.UpdatedAt),
		})
	}

	return &auth.GetUsersResponse{Users: userList}, nil
}
