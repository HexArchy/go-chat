package tokenmanager

type SessionData struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	AccessTokenExpiry  string `json:"access_token_expiry"`
	RefreshTokenExpiry string `json:"refresh_token_expiry"`
	UserID             string `json:"user_id"`
}
