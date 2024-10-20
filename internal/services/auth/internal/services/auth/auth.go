package auth

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Service interface {
	Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
	ValidateToken(ctx context.Context, tokenString string) (*entities.User, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	userStorage     UserStorage
	tokenStorage    TokenStorage
	jwtSecret       []byte
	refreshSecret   []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(deps Deps, jwtSecret, refreshSecret []byte, accessTokenTTL, refreshTokenTTL time.Duration) Service {
	return &service{
		userStorage:     deps.UserStorage,
		tokenStorage:    deps.TokenStorage,
		jwtSecret:       jwtSecret,
		refreshSecret:   refreshSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *service) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get user by email")
	}

	if err := comparePasswords(user.Password, password); err != nil {
		return "", "", entities.ErrInvalidCredentials
	}

	accessToken, err := s.createJWTToken(user, s.jwtSecret, s.accessTokenTTL)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create access token")
	}

	refreshToken, err := s.createJWTToken(user, s.refreshSecret, s.refreshTokenTTL)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create refresh token")
	}

	token := &entities.Token{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.tokenStorage.CreateToken(ctx, token); err != nil {
		return "", "", errors.Wrap(err, "failed to store refresh token")
	}

	return accessToken, refreshToken, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := s.parseToken(refreshToken, s.refreshSecret)
	if err != nil {
		return "", "", errors.Wrap(err, "invalid refresh token")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", "", entities.ErrInvalidTokenClaims
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to uuid parse")
	}

	tokenEntity, err := s.tokenStorage.GetToken(ctx, refreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get refresh token from storage")
	}

	if tokenEntity.ExpiresAt.Before(time.Now()) {
		return "", "", entities.ErrRefreshTokenExpired
	}

	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get user by ID")
	}

	newAccessToken, err := s.createJWTToken(user, s.jwtSecret, s.accessTokenTTL)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create new access token")
	}

	newRefreshToken, err := s.createJWTToken(user, s.refreshSecret, s.refreshTokenTTL)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create new refresh token")
	}

	if err := s.tokenStorage.DeleteToken(ctx, refreshToken); err != nil {
		return "", "", errors.Wrap(err, "failed to delete old refresh token")
	}

	newTokenEntity := &entities.Token{
		Token:     newRefreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}
	if err := s.tokenStorage.CreateToken(ctx, newTokenEntity); err != nil {
		return "", "", errors.Wrap(err, "failed to store new refresh token")
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *service) ValidateToken(ctx context.Context, tokenString string) (*entities.User, error) {
	claims, err := s.parseToken(tokenString, s.jwtSecret)
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, entities.ErrInvalidTokenClaims
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to uuid parse")
	}

	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user by ID")
	}

	return user, nil
}

func (s *service) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.tokenStorage.DeleteTokensByUserID(ctx, userID); err != nil {
		return errors.Wrap(err, "failed to logout user")
	}
	return nil
}
