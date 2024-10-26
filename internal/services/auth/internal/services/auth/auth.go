package auth

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type AuthService interface {
	// Login authenticates user and returns token pair.
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	// Refresh validates refresh token and returns new token pair.
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
	// Validate validates access token and returns user info.
	Validate(ctx context.Context, accessToken string) (*entities.User, error)
	// Revoke invalidates all user's tokens.
	Revoke(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	userStorage  UserStorage
	tokenStorage TokenStorage
	tokenManager TokenManager
	tokenTTL     TokenTTL
}

func NewService(deps Deps) AuthService {
	return &service{
		userStorage:  deps.UserStorage,
		tokenStorage: deps.TokenStorage,
		tokenManager: NewJWTManager([]byte(deps.Secrets.AccessTokenSecret), []byte(deps.Secrets.RefreshTokenSecret)),
		tokenTTL:     deps.TokenTTL,
	}
}

func (s *service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if err := comparePasswords(user.Password, password); err != nil {
		return nil, entities.ErrInvalidCredentials
	}

	return s.createTokenPair(ctx, user)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Verify refresh token.
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "invalid refresh token")
	}

	// Check if token exists in storage.
	storedToken, err := s.tokenStorage.GetToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "token not found")
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, entities.ErrRefreshTokenExpired
	}

	// Get user.
	user, err := s.userStorage.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	// Invalidate old refresh token.
	if err := s.tokenStorage.DeleteToken(ctx, refreshToken); err != nil {
		return nil, errors.Wrap(err, "failed to delete old token")
	}

	// Create new token pair.
	return s.createTokenPair(ctx, user)
}

func (s *service) Validate(ctx context.Context, accessToken string) (*entities.User, error) {
	claims, err := s.tokenManager.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, errors.Wrap(err, "invalid access token")
	}

	user, err := s.userStorage.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	return user, nil
}

func (s *service) Revoke(ctx context.Context, userID uuid.UUID) error {
	return s.tokenStorage.DeleteTokensByUserID(ctx, userID)
}

func (s *service) createTokenPair(ctx context.Context, user *entities.User) (*TokenPair, error) {
	now := time.Now()
	accessTokenExp := now.Add(s.tokenTTL.AccessTokenTTL)
	refreshTokenExp := now.Add(s.tokenTTL.RefreshTokenTTL)

	// Generate access token.
	accessToken, err := s.tokenManager.GenerateAccessToken(user, accessTokenExp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate access token")
	}

	// Generate refresh token.
	refreshToken, err := s.tokenManager.GenerateRefreshToken(user, refreshTokenExp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate refresh token")
	}

	// Store refresh token.
	token := &entities.Token{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: refreshTokenExp,
	}

	if err := s.tokenStorage.CreateToken(ctx, token); err != nil {
		return nil, errors.Wrap(err, "failed to store refresh token")
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessTokenExp,
		RefreshTokenExpiresAt: refreshTokenExp,
	}, nil
}
