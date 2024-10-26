package profile

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type EditProfileUseCase interface {
	Execute(ctx context.Context, userID string, updates map[string]interface{}) error
}

type editProfileUseCase struct {
	authClient *auth.Client
	logger     *zap.Logger
}

func NewEditProfileUseCase(authClient *auth.Client, logger *zap.Logger) EditProfileUseCase {
	return &editProfileUseCase{
		authClient: authClient,
		logger:     logger,
	}
}

func (uc *editProfileUseCase) Execute(ctx context.Context, userID string, updates map[string]interface{}) error {
	uc.logger.Debug("EditProfileUseCase: updating user profile",
		zap.String("user_id", userID))

	if userID == "" {
		return errors.New("user ID is required")
	}

	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	err := uc.authClient.UpdateUser(ctx, userID, updates)
	if err != nil {
		uc.logger.Error("EditProfileUseCase: failed to update profile", zap.Error(err))
		return errors.Wrap(err, "failed to update profile")
	}

	uc.logger.Debug("EditProfileUseCase: user profile updated successfully")
	return nil
}
