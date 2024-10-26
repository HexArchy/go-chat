package http

import (
	"net/http"
	"strconv"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// handleHome renders the home page.
func (c *Controller) handleHome(w http.ResponseWriter, r *http.Request) {
	c.render(w, "home.tmpl", nil)
}

// handleRegisterPage renders the registration page.
func (c *Controller) handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	c.render(w, "register.tmpl", nil)
}

// handleRegister processes user registration.
func (c *Controller) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	username := r.FormValue("username")
	phone := r.FormValue("phone")
	ageStr := r.FormValue("age")
	bio := r.FormValue("bio")

	var age int64
	var err error
	if ageStr != "" {
		age, err = strconv.ParseInt(ageStr, 10, 32)
		if err != nil {
			http.Error(w, "Failed to parse age", http.StatusBadRequest)
			return
		}
	}

	ctx := r.Context()
	err = c.registerUseCase.Execute(ctx, email, password, username, phone, int32(age), bio)
	if err != nil {
		c.logger.Error("Registration failed", zap.Error(err))
		c.render(w, "register.tmpl", map[string]interface{}{
			"Error": "Registration failed: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// handleLoginPage renders the login page.
func (c *Controller) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	c.render(w, "login.tmpl", nil)
}

// handleLogin processes user login.
func (c *Controller) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	ctx := r.Context()
	ctx = WithHTTPContext(ctx, r, w)
	tokenResponse, err := c.loginUseCase.Execute(ctx, email, password)
	if err != nil {
		c.logger.Error("Login failed", zap.Error(err))
		c.render(w, "login.tmpl", map[string]interface{}{
			"Error": "Login failed: " + err.Error(),
		})
		return
	}

	// Store tokens in session.
	session, err := c.store.Get(r, c.sessionName)
	if err != nil {
		c.logger.Error("Failed to get session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session.ID == "" {
		session.ID = uuid.New().String()
		session.Values["session_id"] = session.ID
	}

	c.logger.Debug("Got session",
		zap.String("session_id", session.ID),
		zap.Bool("is_new", session.IsNew))

	authTokenResponse := &auth.TokenResponse{
		AccessToken:           tokenResponse.AccessToken,
		RefreshToken:          tokenResponse.RefreshToken,
		AccessTokenExpiresAt:  tokenResponse.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenResponse.RefreshTokenExpiresAt,
	}

	// Use token manager to set tokens.
	if err := c.tokenManager.SetTokens(ctx, session, authTokenResponse, r, w); err != nil {
		c.logger.Error("Failed to set tokens", zap.Error(err))
		http.Error(w, "Failed to set session", http.StatusInternalServerError)
		return
	}

	c.logger.Debug("Login successful",
		zap.String("email", email))

	http.Redirect(w, r, "/rooms", http.StatusSeeOther)
}

// handleLogout logs out the user.
func (c *Controller) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = WithHTTPContext(ctx, r, w)

	session, err := c.store.Get(r, c.sessionName)
	if err != nil {
		c.logger.Error("Failed to get session", zap.Error(err))
		// Even if session retrieval fails, attempt to clear it.
	}

	accessToken, ok := session.Values[c.tokenKey].(string)
	if !ok || accessToken == "" {
		// No token, redirect to home.
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Execute LogoutUseCase.
	err = c.logoutUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to logout", zap.Error(err))
		// Proceed to clear session regardless of logout failure.
	}

	// Clear session.
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		c.logger.Error("Failed to clear session", zap.Error(err))
		http.Error(w, "Failed to clear session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
