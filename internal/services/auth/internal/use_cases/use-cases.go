package usecases

import (
	authService "github.com/HexArch/go-chat/internal/services/auth/internal/services/auth"
	authStorage "github.com/HexArch/go-chat/internal/services/auth/internal/services/auth/storage"
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/rbac"
	deleteuser "github.com/HexArch/go-chat/internal/services/auth/internal/use_cases/delete_user"
	loginuser "github.com/HexArch/go-chat/internal/services/auth/internal/use_cases/login_user"
	refreshtoken "github.com/HexArch/go-chat/internal/services/auth/internal/use_cases/refresh_token"
	registeruser "github.com/HexArch/go-chat/internal/services/auth/internal/use_cases/register_user"
	updateuser "github.com/HexArch/go-chat/internal/services/auth/internal/use_cases/update_user"
)

type UseCases struct {
	registerUser *registeruser.UseCase
	loginUser    *loginuser.UseCase
	refreshToken *refreshtoken.UseCase
	updateUser   *updateuser.UseCase
	deleteUser   *deleteuser.UseCase
}

func NewUseCases(authService *authService.AuthService, authStorage *authStorage.Storage, rbac *rbac.RBAC) *UseCases {
	return &UseCases{
		registerUser: registeruser.New(authService, authStorage),
		loginUser:    loginuser.New(authService),
		refreshToken: refreshtoken.New(authService),
		updateUser:   updateuser.New(authService, authStorage, rbac),
		deleteUser:   deleteuser.New(authService, authStorage, rbac),
	}
}
