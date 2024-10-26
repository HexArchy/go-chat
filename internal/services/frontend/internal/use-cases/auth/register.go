package auth

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// RegisterUseCase defines the interface for user registration.
type RegisterUseCase interface {
	Execute(ctx context.Context, email, password, username, phone string, age int32, bio string) error
}

// registerUseCase is the concrete implementation of RegisterUseCase.
type registerUseCase struct {
	authClient *auth.Client
	logger     *zap.Logger
}

// NewRegisterUseCase creates a new instance of RegisterUseCase.
func NewRegisterUseCase(authClient *auth.Client, logger *zap.Logger) RegisterUseCase {
	return &registerUseCase{
		authClient: authClient,
		logger:     logger,
	}
}

// Execute handles the user registration process.
func (uc *registerUseCase) Execute(ctx context.Context, email, password, username, phone string, age int32, bio string) error {
	uc.logger.Debug("RegisterUseCase: executing user registration",
		zap.String("email", email),
		zap.String("username", username))

	// Input validation can be added here.
	if email == "" || password == "" || username == "" {
		return errors.New("email, password, and username are required")
	}

	// Call the auth client to register the user.
	err := uc.authClient.Register(ctx, email, password, username, phone, age, bio)
	if err != nil {
		uc.logger.Error("RegisterUseCase: registration failed", zap.Error(err))
		return errors.Wrap(err, "user registration failed")
	}

	uc.logger.Debug("RegisterUseCase: user registered successfully")
	return nil
}
