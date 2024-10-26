package http

import (
	"net/http"

	"go.uber.org/zap"
)

func (c *Controller) handleProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := c.tokenManager.GetToken(ctx)

	user, err := c.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusUnauthorized)
		return
	}

	c.render(w, "profile.tmpl", map[string]interface{}{
		"User": user,
	})
}

func (c *Controller) handleProfileEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := c.tokenManager.GetToken(ctx)

	user, err := c.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusUnauthorized)
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

	updates := map[string]interface{}{
		"email":    r.FormValue("email"),
		"username": r.FormValue("username"),
		"phone":    r.FormValue("phone"),
		"bio":      r.FormValue("bio"),
	}

	if password := r.FormValue("password"); password != "" {
		updates["password"] = password
	}

	if err := c.authClient.UpdateUser(ctx, token, user.ID.String(), updates); err != nil {
		c.logger.Error("Failed to update profile", zap.Error(err))
		c.render(w, "profile_edit.tmpl", map[string]interface{}{
			"User":  user,
			"Error": "Failed to update profile: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
