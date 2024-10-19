package usecases

import (
	authService "github.com/HexArch/go-chat/internal/services/auth/internal/services/auth"
	authStorage "github.com/HexArch/go-chat/internal/services/auth/internal/services/auth/storage"
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/rbac"
	deleteuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/delete-user"
	loginuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/login-user"
	refreshtoken "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/refresh-token"
	registeruser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/register-user"
	updateuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/update-user"
)

type UseCases struct {
	RegisterUser *registeruser.UseCase
	LoginUser    *loginuser.UseCase
	RefreshToken *refreshtoken.UseCase
	UpdateUser   *updateuser.UseCase
	DeleteUser   *deleteuser.UseCase
}

func NewUseCases(authService *authService.AuthService, authStorage *authStorage.Storage, rbac *rbac.RBAC) *UseCases {
	return &UseCases{
		RegisterUser: registeruser.New(authService, authStorage),
		LoginUser:    loginuser.New(authService),
		RefreshToken: refreshtoken.New(authService),
		UpdateUser:   updateuser.New(authService, authStorage, rbac),
		DeleteUser:   deleteuser.New(authService, authStorage, rbac),
	}
}
