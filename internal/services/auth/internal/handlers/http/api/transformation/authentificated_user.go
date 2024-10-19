package transformation

import (
	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/models"
	"github.com/go-openapi/strfmt"
)

func MapAuthenticatedUserToResponse(user entities.AuthenticatedUser) *models.AuthenticatedUserResponse {
	return &models.AuthenticatedUserResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		Name:      user.Name,
		Nickname:  user.Nickname,
		Age:       int64(user.Age),
		Bio:       user.Bio,
		Roles:     mapRolesToStrings(user.Roles),
		ExpiresAt: strfmt.DateTime(user.ExpiresAt),
		TokenPair: &models.TokenPair{
			AccessToken:  user.TokenPair.AccessToken,
			RefreshToken: user.TokenPair.RefreshToken,
		},
	}
}

func mapRolesToStrings(roles []entities.Role) []string {
	stringRoles := make([]string, len(roles))
	for i, role := range roles {
		stringRoles[i] = string(role)
	}
	return stringRoles
}
