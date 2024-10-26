package auth

import (
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID      uuid.UUID             `json:"user_id"`
	Email       string                `json:"email"`
	Username    string                `json:"username"`
	Permissions []entities.Permission `json:"permissions"`
	TokenType   string                `json:"token_type"`
	jwt.StandardClaims
}

type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}
