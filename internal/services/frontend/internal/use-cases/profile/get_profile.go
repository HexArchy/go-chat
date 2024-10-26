package profile

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// GetProfileUseCase defines the interface for retrieving user profile.
type GetProfileUseCase interface {
	Execute(ctx context.Context) (*entities.User, error)
}

// getProfileUseCase is the concrete implementation of GetProfileUseCase.
type getProfileUseCase struct {
	authClient *auth.Client
	logger     *zap.Logger
}

// NewGetProfileUseCase creates a new instance of GetProfileUseCase.
func NewGetProfileUseCase(authClient *auth.Client, logger *zap.Logger) GetProfileUseCase {
	return &getProfileUseCase{
		authClient: authClient,
		logger:     logger,
	}
}

// Execute retrieves the user profile based on the access token.
func (uc *getProfileUseCase) Execute(ctx context.Context) (*entities.User, error) {
	uc.logger.Debug("GetProfileUseCase: retrieving user profile")

	// Call the auth client to validate the token and get user info.
	user, permissions, err := uc.authClient.ValidateToken(ctx)
	if err != nil {
		uc.logger.Error("GetProfileUseCase: token validation failed", zap.Error(err))
		return nil, errors.Wrap(err, "failed to validate token")
	}

	user.Permissions = permissions

	uc.logger.Debug("GetProfileUseCase: user profile retrieved successfully",
		zap.String("user_id", user.ID.String()))

	return user, nil
}
