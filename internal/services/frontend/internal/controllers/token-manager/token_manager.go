package tokenmanager

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

type tokenContextKey string

const (
	accessTokenKey  tokenContextKey = "accessToken"
	refreshTokenKey tokenContextKey = "refreshToken"
)

type TokenManager struct {
	mu            sync.RWMutex
	store         sessions.Store
	sessionName   string
	logger        *zap.Logger
	authClient    AuthClient
	refreshWindow time.Duration
}

func NewTokenManager(store sessions.Store, sessionName string, logger *zap.Logger, authClient AuthClient) *TokenManager {
	return &TokenManager{
		store:         store,
		sessionName:   sessionName,
		logger:        logger,
		authClient:    authClient,
		refreshWindow: 5 * time.Minute,
	}
}

func (tm *TokenManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := tm.store.Get(r, tm.sessionName)
		if err != nil {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		accessToken, ok := session.Values["accessToken"].(string)
		if !ok || accessToken == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if exp, err := tm.getTokenExpiration(accessToken); err != nil || time.Until(exp) < tm.refreshWindow {
			refreshToken, ok := session.Values["refreshToken"].(string)
			if !ok || refreshToken == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			newAccessToken, newRefreshToken, err := tm.authClient.RefreshToken(r.Context(), refreshToken)
			if err != nil {
				tm.logger.Error("Failed to refresh token", zap.Error(err))
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			session.Values["accessToken"] = newAccessToken
			session.Values["refreshToken"] = newRefreshToken
			if err := session.Save(r, w); err != nil {
				tm.logger.Error("Failed to save session", zap.Error(err))
				http.Error(w, "Failed to save session", http.StatusInternalServerError)
				return
			}

			accessToken = newAccessToken
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, accessTokenKey, accessToken)
		if refreshToken, ok := session.Values["refreshToken"].(string); ok {
			ctx = context.WithValue(ctx, refreshTokenKey, refreshToken)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (tm *TokenManager) getTokenExpiration(tokenString string) (time.Time, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}, errors.New("invalid token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, errors.New("missing expiration time")
	}

	return time.Unix(int64(exp), 0), nil
}

func (tm *TokenManager) GetToken(ctx context.Context) string {
	if token := ctx.Value(accessTokenKey); token != nil {
		return token.(string)
	}
	return ""
}

func (tm *TokenManager) GetRefreshToken(ctx context.Context) string {
	if token := ctx.Value(refreshTokenKey); token != nil {
		return token.(string)
	}
	return ""
}

func (tm *TokenManager) WithTokens(ctx context.Context) context.Context {
	if accessToken := tm.GetToken(ctx); accessToken != "" {
		ctx = context.WithValue(ctx, accessTokenKey, accessToken)
	}
	if refreshToken := tm.GetRefreshToken(ctx); refreshToken != "" {
		ctx = context.WithValue(ctx, refreshTokenKey, refreshToken)
	}
	return ctx
}
