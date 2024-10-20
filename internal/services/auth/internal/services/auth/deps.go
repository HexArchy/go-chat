package auth

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
)

type UserStorage interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
}

type TokenStorage interface {
	CreateToken(ctx context.Context, token *entities.Token) error
	GetToken(ctx context.Context, tokenString string) (*entities.Token, error)
	DeleteToken(ctx context.Context, tokenString string) error
	DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error
}

type Deps struct {
	UserStorage  UserStorage
	TokenStorage TokenStorage
}
