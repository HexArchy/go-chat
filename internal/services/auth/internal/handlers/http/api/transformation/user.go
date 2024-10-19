package transformation

import (
	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/models"
)

func MapUserToResponse(user entities.User) *models.UserResponse {
	return &models.UserResponse{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Nickname: user.Nickname,
		Age:      int64(user.Age),
		Bio:      user.Bio,
		Roles:    mapRolesToStrings(user.Roles),
	}
}
