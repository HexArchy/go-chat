package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// setUserIDCookie sets the user ID in a secure, HTTP-only cookie.
func setUserIDCookie(w http.ResponseWriter, userID uuid.UUID, logger *zap.Logger) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

// getUserIDFromCookie retrieves the user ID from the cookie.
func getUserIDFromCookie(r *http.Request, logger *zap.Logger) (uuid.UUID, error) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		logger.Warn("User ID cookie not found", zap.Error(err))
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(cookie.Value)
	if err != nil {
		logger.Warn("Invalid user ID in cookie", zap.Error(err))
		return uuid.Nil, err
	}

	return userID, nil
}

// clearUserIDCookie removes the user ID cookie.
func clearUserIDCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (c *Controller) getUserIDFromSession(r *http.Request) (uuid.UUID, error) {
	session, err := c.store.Get(r, c.sessionName)
	if err != nil {
		return uuid.Nil, err
	}

	userIDStr, ok := session.Values["userID"].(string)
	if !ok {
		return uuid.Nil, errors.New("userID not found in session")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func WithHTTPContext(ctx context.Context, r *http.Request, w http.ResponseWriter) context.Context {
	ctx = context.WithValue(ctx, "http_request", r)
	return context.WithValue(ctx, "http_response", w)
}
