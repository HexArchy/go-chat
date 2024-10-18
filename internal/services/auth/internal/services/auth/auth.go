package auth

import (
	"context"

	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/rbac"
	"github.com/pkg/errors"
)

type AuthService struct {
	storage            AuthStorage
	accessTokenSecret  []byte
	refreshTokenSecret []byte
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
	rbac               *rbac.RBAC
}

func NewAuthService(storage AuthStorage, rbac *rbac.RBAC, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		storage:            storage,
		accessTokenSecret:  []byte(accessSecret),
		refreshTokenSecret: []byte(refreshSecret),
		accessTokenTTL:     accessTTL,
		refreshTokenTTL:    refreshTTL,
		rbac:               rbac,
	}
}

func (a *AuthService) RegisterUser(ctx context.Context, user *entities.User) error {
	salt, err := a.generateSalt()
	if err != nil {
		return errors.Wrap(err, "generate salt")
	}

	hashedPassword, err := a.hashPassword(user.Password, salt)
	if err != nil {
		return errors.Wrap(err, "hash password")
	}

	user.Password = hashedPassword
	user.Roles = []entities.Role{entities.RoleUser}

	if err := a.storage.CreateUser(ctx, user); err != nil {
		return errors.Wrap(err, "create user")
	}

	return nil
}

func (a *AuthService) AuthenticateUser(ctx context.Context, email, password string) (entities.AuthenticatedUser, error) {
	user, err := a.storage.GetUserByEmail(ctx, email)
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "get user by email")
	}

	if user == nil {
		return entities.AuthenticatedUser{}, entities.ErrInvalidCredantials
	}

	if err := a.verifyPassword(password, user.Password); err != nil {
		return entities.AuthenticatedUser{}, entities.ErrInvalidCredantials
	}

	accessToken, err := a.generateToken(user, a.accessTokenSecret, a.accessTokenTTL, "access")
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "generate access token")
	}

	refreshToken, err := a.generateToken(user, a.refreshTokenSecret, a.refreshTokenTTL, "refresh")
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "generate refresh token")
	}

	if err := a.storage.StoreRefreshToken(ctx, user.ID, refreshToken, a.refreshTokenTTL); err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "store refresh token")
	}

	return entities.AuthenticatedUser{
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Age:       user.Age,
		Bio:       user.Bio,
		Roles:     user.Roles,
		ExpiresAt: time.Now().Add(a.accessTokenTTL),
		TokenPair: entities.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

func (a *AuthService) ValidateToken(ctx context.Context, tokenString, tokenType string) (*entities.AuthenticatedUser, error) {
	var secret []byte
	switch tokenType {
	case "access":
		secret = a.accessTokenSecret
	case "refresh":
		secret = a.refreshTokenSecret
	default:
		return nil, entities.ErrInvalidTokenType
	}

	claims, err := a.validateToken(tokenString, secret)
	if err != nil {
		return nil, err
	}

	if tokenType == "refresh" {
		isValid, err := a.storage.ValidateRefreshToken(ctx, claims["user_id"].(string), tokenString)
		if err != nil {
			return nil, errors.Wrap(err, "validate refresh token")
		}
		if !isValid {
			return nil, entities.ErrInvalidRefreshToken
		}
	}

	return &entities.AuthenticatedUser{
		UserID:    claims["user_id"].(string),
		Name:      claims["name"].(string),
		Email:     claims["email"].(string),
		Nickname:  claims["nickname"].(string),
		Age:       int(claims["age"].(float64)),
		Bio:       claims["bio"].(string),
		Roles:     a.convertToRoles(claims["roles"]),
		ExpiresAt: time.Unix(int64(claims["exp"].(float64)), 0),
	}, nil
}

func (a *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (entities.AuthenticatedUser, error) {
	authUser, err := a.ValidateToken(ctx, refreshToken, "refresh")
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "validate refresh token")
	}

	user, err := a.storage.GetUserByID(ctx, authUser.UserID)
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "get user")
	}

	newAccessToken, err := a.generateToken(user, a.accessTokenSecret, a.accessTokenTTL, "access")
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "generate new access token")
	}

	newRefreshToken, err := a.generateToken(user, a.refreshTokenSecret, a.refreshTokenTTL, "refresh")
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "generate new refresh token")
	}

	if err := a.storage.RevokeRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "revoke refresh token")
	}

	if err := a.storage.StoreRefreshToken(ctx, user.ID, newRefreshToken, a.refreshTokenTTL); err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "store refresh token")
	}

	return entities.AuthenticatedUser{
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Age:       user.Age,
		Bio:       user.Bio,
		Roles:     user.Roles,
		ExpiresAt: time.Now().Add(a.accessTokenTTL),
		TokenPair: entities.TokenPair{
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		},
	}, nil
}

func (a *AuthService) LogoutUser(ctx context.Context, userID, refreshToken string) error {
	return a.storage.RevokeRefreshToken(ctx, userID, refreshToken)
}

func (a *AuthService) AssignRole(ctx context.Context, userID string, role entities.Role) error {
	user, err := a.storage.GetUserByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "get user by id")
	}

	if !a.rbac.IsValidRole(role) {
		return entities.ErrInvalidRole
	}

	user.AddRole(role)
	if err := a.storage.UpdateUser(ctx, user); err != nil {
		return errors.Wrap(err, "update user")
	}

	return nil
}

func (a *AuthService) RemoveRole(ctx context.Context, userID string, role entities.Role) error {
	user, err := a.storage.GetUserByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "get user by id")
	}

	user.RemoveRole(role)
	if err := a.storage.UpdateUser(ctx, user); err != nil {
		return errors.Wrap(err, "update user")
	}

	return nil
}

func (a *AuthService) CheckPermission(user *entities.AuthenticatedUser, action, resource string) bool {
	return a.rbac.CheckPermission(user.Roles, action, resource)
}

func (a *AuthService) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	if err := a.storage.RevokeAllRefreshTokens(ctx, userID); err != nil {
		return errors.Wrap(err, "failed to revoke all refresh tokens")
	}
	return nil
}
