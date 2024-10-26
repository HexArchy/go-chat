package auth

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// LoginUseCase defines the interface for user login.
type LoginUseCase interface {
	Execute(ctx context.Context, email, password string) (*entities.TokenResponse, error)
}

// loginUseCase is the concrete implementation of LoginUseCase.
type loginUseCase struct {
	authClient *auth.Client
	logger     *zap.Logger
}

// NewLoginUseCase creates a new instance of LoginUseCase.
func NewLoginUseCase(authClient *auth.Client, logger *zap.Logger) LoginUseCase {
	return &loginUseCase{
		authClient: authClient,
		logger:     logger,
	}
}

// Execute handles the user login process.
func (uc *loginUseCase) Execute(ctx context.Context, email, password string) (*entities.TokenResponse, error) {
	uc.logger.Debug("LoginUseCase: executing user login",
		zap.String("email", email))

	// Input validation can be added here.
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	// Call the auth client to login the user.
	tokenResp, err := uc.authClient.Login(ctx, email, password)
	if err != nil {
		uc.logger.Error("LoginUseCase: login failed", zap.Error(err))
		return nil, errors.Wrap(err, "user login failed")
	}

	// Convert auth client response to entities.TokenResponse
	tokens := &entities.TokenResponse{
		AccessToken:           tokenResp.AccessToken,
		RefreshToken:          tokenResp.RefreshToken,
		AccessTokenExpiresAt:  tokenResp.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenResp.RefreshTokenExpiresAt,
	}

	uc.logger.Debug("LoginUseCase: user logged in successfully",
		zap.String("accessToken", tokens.AccessToken[:10]+"..."),
		zap.String("refreshToken", tokens.RefreshToken[:10]+"..."))

	return tokens, nil
}
