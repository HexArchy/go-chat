package http

import (
	"net/http"

	"go.uber.org/zap"
)

// handleProfile displays the user's profile.
func (c *Controller) handleProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = WithHTTPContext(ctx, r, w)

	// Retrieve tokens from session.
	session, err := c.store.Get(r, c.sessionName)
	if err != nil {
		c.logger.Error("Failed to get session", zap.Error(err))
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	accessToken, err := c.tokenManager.GetAccessToken(ctx, session)
	if err != nil {
		c.logger.Error("Failed to get access token", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Inject token into context.
	ctx = contextWithToken(ctx, accessToken)

	// Execute GetProfileUseCase.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get profile", zap.Error(err))
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	c.render(w, "profile.tmpl", map[string]interface{}{
		"User": user,
	})
}

// handleProfileEdit handles displaying and updating the user's profile.
func (c *Controller) handleProfileEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = WithHTTPContext(ctx, r, w)

	// Retrieve tokens from session.
	session, err := c.store.Get(r, c.sessionName)
	if err != nil {
		c.logger.Error("Failed to get session", zap.Error(err))
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	accessToken, err := c.tokenManager.GetAccessToken(ctx, session)
	if err != nil {
		c.logger.Error("Failed to get access token", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Inject token into context.
	ctx = contextWithToken(ctx, accessToken)

	// Execute GetProfileUseCase to get current user data.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get profile", zap.Error(err))
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		c.render(w, "profile_edit.tmpl", map[string]interface{}{
			"User": user,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})

	email := r.FormValue("email")
	if email != "" && email != user.Email {
		updates["email"] = email
	}

	username := r.FormValue("username")
	if username != "" && username != user.Username {
		updates["username"] = username
	}

	phone := r.FormValue("phone")
	if phone != "" && phone != user.Phone {
		updates["phone"] = phone
	}

	bio := r.FormValue("bio")
	if bio != "" && bio != user.Bio {
		updates["bio"] = bio
	}

	password := r.FormValue("password")
	if password != "" {
		updates["password"] = password
	}

	// Inject token into context.
	ctx = contextWithToken(ctx, accessToken)

	// Execute EditProfileUseCase
	err = c.editProfileUseCase.Execute(ctx, user.ID.String(), updates)
	if err != nil {
		c.logger.Error("Failed to update profile", zap.Error(err))
		c.render(w, "profile_edit.tmpl", map[string]interface{}{
			"User":  user,
			"Error": "Failed to update profile: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
