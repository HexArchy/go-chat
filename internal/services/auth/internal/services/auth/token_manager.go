package auth

import (
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

type TokenManager interface {
	GenerateAccessToken(user *entities.User, expiresAt time.Time) (string, error)
	GenerateRefreshToken(user *entities.User, expiresAt time.Time) (string, error)
	ValidateAccessToken(tokenString string) (*TokenClaims, error)
	ValidateRefreshToken(tokenString string) (*TokenClaims, error)
}

type jwtManager struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJWTManager(accessSecret, refreshSecret []byte) TokenManager {
	return &jwtManager{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

func (m *jwtManager) GenerateAccessToken(user *entities.User, expiresAt time.Time) (string, error) {
	claims := &TokenClaims{
		UserID:      user.ID,
		Email:       user.Email,
		Username:    user.Username,
		Permissions: user.Permissions,
		TokenType:   "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.accessSecret)
}

func (m *jwtManager) GenerateRefreshToken(user *entities.User, expiresAt time.Time) (string, error) {
	claims := &TokenClaims{
		UserID:    user.ID,
		TokenType: "refresh",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.refreshSecret)
}

func (m *jwtManager) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	return m.validateToken(tokenString, m.accessSecret, "access")
}

func (m *jwtManager) ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	return m.validateToken(tokenString, m.refreshSecret, "refresh")
}

func (m *jwtManager) validateToken(tokenString string, secret []byte, tokenType string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.TokenType != tokenType {
		return nil, errors.Errorf("invalid token type: expected %s, got %s", tokenType, claims.TokenType)
	}

	return claims, nil
}
