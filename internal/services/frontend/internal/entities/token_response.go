package entities

import "time"

// TokenResponse encapsulates the access and refresh tokens along with their expiry times.
type TokenResponse struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}
