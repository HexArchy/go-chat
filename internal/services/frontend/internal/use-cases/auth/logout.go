package auth

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// LogoutUseCase defines the interface for user logout.
type LogoutUseCase interface {
	Execute(ctx context.Context) error
}

// logoutUseCase is the concrete implementation of LogoutUseCase.
type logoutUseCase struct {
	authClient *auth.Client
	logger     *zap.Logger
}

// NewLogoutUseCase creates a new instance of LogoutUseCase.
func NewLogoutUseCase(authClient *auth.Client, logger *zap.Logger) LogoutUseCase {
	return &logoutUseCase{
		authClient: authClient,
		logger:     logger,
	}
}

// Execute handles the user logout process.
func (uc *logoutUseCase) Execute(ctx context.Context) error {
	uc.logger.Debug("LogoutUseCase: executing user logout")

	// Call the auth client to logout the user.
	err := uc.authClient.Logout(ctx)
	if err != nil {
		uc.logger.Error("LogoutUseCase: logout failed", zap.Error(err))
		return errors.Wrap(err, "user logout failed")
	}

	uc.logger.Debug("LogoutUseCase: user logged out successfully")
	return nil
}
