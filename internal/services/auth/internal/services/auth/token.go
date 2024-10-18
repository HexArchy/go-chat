package auth

import (
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

func (a *AuthService) generateToken(user *entities.User, secret []byte, ttl time.Duration, tokenType string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"name":     user.Name,
		"email":    user.Email,
		"nickname": user.Nickname,
		"age":      user.Age,
		"bio":      user.Bio,
		"roles":    user.Roles,
		"type":     tokenType,
		"iat":      now.Unix(),
		"exp":      now.Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (a *AuthService) validateToken(tokenString string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "parse token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, entities.ErrInvalidToken
}

func (a *AuthService) convertToRoles(v interface{}) []entities.Role {
	if v == nil {
		return nil
	}
	strRoles := v.([]string)
	roles := make([]entities.Role, len(strRoles))
	for i, role := range strRoles {
		roles[i] = entities.Role(role)
	}
	return roles
}
