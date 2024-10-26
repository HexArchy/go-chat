package tokenmanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	sessionKeyAccessToken        = "access_token"
	sessionKeyRefreshToken       = "refresh_token"
	sessionKeyAccessTokenExpiry  = "access_token_expiry"
	sessionKeyRefreshTokenExpiry = "refresh_token_expiry"
	sessionKeyUserID             = "userID"
)

type SessionData struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	AccessTokenExpiry  string `json:"access_token_expiry"`
	RefreshTokenExpiry string `json:"refresh_token_expiry"`
	UserID             string `json:"user_id"`
}

// TokenManager defines the interface for managing tokens via sessions.
type TokenManager interface {
	// GetAccessToken retrieves the access token from the session.
	GetAccessToken(ctx context.Context, session *sessions.Session) (string, error)
	// RefreshAccessToken refreshes the access token using the refresh token stored in the session.
	RefreshAccessToken(ctx context.Context, session *sessions.Session, r *http.Request, w http.ResponseWriter) (string, error)
	// SetTokens sets the access and refresh tokens in the session.
	SetTokens(ctx context.Context, session *sessions.Session, tokens *auth.TokenResponse, r *http.Request, w http.ResponseWriter) error
	// GetUserID retrieves the user ID from the session.
	GetUserID(session *sessions.Session) (uuid.UUID, error)
}

type tokenManager struct {
	authClient   *auth.Client
	logger       *zap.Logger
	sessionStore sessions.Store
	sessionName  string
}

func NewTokenManager(authClient *auth.Client, logger *zap.Logger, sessionStore sessions.Store, sessionName string) TokenManager {
	return &tokenManager{
		authClient:   authClient,
		logger:       logger,
		sessionStore: sessionStore,
		sessionName:  sessionName,
	}
}

func (tm *tokenManager) GetAccessToken(ctx context.Context, session *sessions.Session) (string, error) {
	data, ok := session.Values["session_data"].(*SessionData)
	if !ok || data == nil {
		tm.logger.Warn("No session data found")
		return "", status.Error(codes.Unauthenticated, "no session data found")
	}

	if data.AccessToken == "" {
		tm.logger.Warn("No access token found in session")
		return "", status.Error(codes.Unauthenticated, "no access token in session")
	}

	tm.logger.Debug("Retrieved token from session",
		zap.String("session_id", session.ID),
	)

	expiry, err := time.Parse(time.RFC3339, data.AccessTokenExpiry)
	if err != nil {
		tm.logger.Error("Failed to parse access token expiry", zap.Error(err))
		return "", status.Error(codes.Internal, "invalid access token expiry format")
	}

	// Check if token needs refresh (expires in less than 1 minute).
	if time.Until(expiry) < time.Minute {
		tm.logger.Info("Access token expired or expiring soon, refreshing")

		r, ok := ctx.Value("http_request").(*http.Request)
		if !ok {
			return "", status.Error(codes.Internal, "no http request in context")
		}
		w, ok := ctx.Value("http_response").(http.ResponseWriter)
		if !ok {
			return "", status.Error(codes.Internal, "no http response in context")
		}

		return tm.RefreshAccessToken(ctx, session, r, w)
	}

	return data.AccessToken, nil
}

func (tm *tokenManager) GetUserID(session *sessions.Session) (uuid.UUID, error) {
	data, ok := session.Values["session_data"].(*SessionData)
	if !ok || data == nil {
		return uuid.Nil, status.Error(codes.Unauthenticated, "no session data found")
	}

	if data.UserID == "" {
		return uuid.Nil, status.Error(codes.Unauthenticated, "no user ID in session")
	}

	userID, err := uuid.Parse(data.UserID)
	if err != nil {
		return uuid.Nil, status.Error(codes.Internal, "invalid user ID format")
	}

	return userID, nil
}

func (tm *tokenManager) RefreshAccessToken(ctx context.Context, session *sessions.Session, r *http.Request, w http.ResponseWriter) (string, error) {
	refreshToken, ok := session.Values[sessionKeyRefreshToken].(string)
	if !ok || refreshToken == "" {
		tm.logger.Warn("No refresh token found in session")
		return "", status.Error(codes.Unauthenticated, "no refresh token in session")
	}

	tokenResp, err := tm.authClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		tm.logger.Error("Failed to refresh token", zap.Error(err))
		return "", err
	}

	if err := tm.SetTokens(ctx, session, tokenResp, r, w); err != nil {
		tm.logger.Error("Failed to update session with new tokens", zap.Error(err))
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (tm *tokenManager) SetTokens(ctx context.Context, session *sessions.Session, tokens *auth.TokenResponse, r *http.Request, w http.ResponseWriter) error {
	userID, err := extractUserIDFromToken(tokens.AccessToken)
	if err != nil {
		tm.logger.Error("Failed to extract user ID from token", zap.Error(err))
		return status.Error(codes.Internal, "failed to extract user ID from token")
	}

	data := &SessionData{
		AccessToken:        tokens.AccessToken,
		RefreshToken:       tokens.RefreshToken,
		AccessTokenExpiry:  tokens.AccessTokenExpiresAt.Format(time.RFC3339),
		RefreshTokenExpiry: tokens.RefreshTokenExpiresAt.Format(time.RFC3339),
		UserID:             userID.String(),
	}

	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	session.Values["session_id"] = session.ID
	session.Values["session_data"] = data

	if err := session.Save(r, w); err != nil {
		tm.logger.Error("Failed to save session",
			zap.Error(err),
			zap.String("userID", userID.String()),
		)
		return status.Error(codes.Internal, "failed to save session")
	}

	tm.logger.Debug("Tokens successfully saved to session",
		zap.String("userID", userID.String()),
		zap.String("session_id", session.ID),
		zap.Bool("is_new", session.IsNew),
	)
	return nil
}

func (tm *tokenManager) ClearTokens(session *sessions.Session, r *http.Request, w http.ResponseWriter) error {
	delete(session.Values, sessionKeyAccessToken)
	delete(session.Values, sessionKeyRefreshToken)
	delete(session.Values, sessionKeyAccessTokenExpiry)
	delete(session.Values, sessionKeyRefreshTokenExpiry)
	delete(session.Values, sessionKeyUserID)

	if err := session.Save(r, w); err != nil {
		tm.logger.Error("Failed to save session after clearing tokens", zap.Error(err))
		return status.Error(codes.Internal, "failed to save session")
	}

	return nil
}

func extractUserIDFromToken(tokenString string) (uuid.UUID, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return uuid.Nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id not found in token claims")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user_id: %w", err)
	}

	return userID, nil
}
