package http

import (
	"net/http"

	"go.uber.org/zap"
)

func (c *Controller) handleHome(w http.ResponseWriter, r *http.Request) {
	c.render(w, "home.tmpl", nil)
}

func (c *Controller) handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	c.render(w, "register.tmpl", nil)
}

func (c *Controller) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	username := r.FormValue("username")

	ctx := r.Context()
	err := c.authClient.Register(ctx, email, password, username, "", 0, "")
	if err != nil {
		c.render(w, "register.tmpl", map[string]interface{}{
			"Error": "Registration failed: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *Controller) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	c.render(w, "login.tmpl", nil)
}

func (c *Controller) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	ctx := r.Context()
	accessToken, refreshToken, err := c.authClient.Login(ctx, email, password)
	if err != nil {
		c.render(w, "login.tmpl", map[string]interface{}{
			"Error": "Login failed: " + err.Error(),
		})
		return
	}

	session, _ := c.store.Get(r, sessionName)
	session.Values[tokenKey] = accessToken
	session.Values["refreshToken"] = refreshToken
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/rooms", http.StatusSeeOther)
}

func (c *Controller) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := c.tokenManager.GetToken(ctx)

	if err := c.authClient.Logout(ctx, token); err != nil {
		c.logger.Error("Failed to logout", zap.Error(err))
	}

	session, _ := c.store.Get(r, sessionName)
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
