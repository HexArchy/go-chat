package http

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (c *Controller) RegisterRoutes(r *mux.Router) {
	if err := c.loadTemplates(); err != nil {
		c.logger.Fatal("Failed to load templates", zap.Error(err))
	}

	// Public routes.
	r.HandleFunc("/", c.handleHome).Methods("GET")
	r.HandleFunc("/register", c.handleRegisterPage).Methods("GET")
	r.HandleFunc("/register", c.handleRegister).Methods("POST")
	r.HandleFunc("/login", c.handleLoginPage).Methods("GET")
	r.HandleFunc("/login", c.handleLogin).Methods("POST")

	// Protected routes.
	protected := r.NewRoute().Subrouter()
	protected.Use(c.authMiddleware)

	protected.HandleFunc("/logout", c.handleLogout).Methods("POST")
	protected.HandleFunc("/profile", c.handleProfile).Methods("GET")
	protected.HandleFunc("/profile/edit", c.handleProfileEdit).Methods("GET", "POST")

	// Rooms.
	protected.HandleFunc("/rooms", c.handleRoomsList).Methods("GET")
	protected.HandleFunc("/rooms/create", c.handleRoomCreate).Methods("GET", "POST")
	protected.HandleFunc("/rooms/search", c.handleRoomSearch).Methods("GET")
	protected.HandleFunc("/rooms/{id}", c.handleRoomView).Methods("GET")
	protected.HandleFunc("/rooms/{id}/delete", c.handleRoomDelete).Methods("POST")
	protected.HandleFunc("/rooms/{id}/ws", c.handleWebSocket)
}

func (c *Controller) loadTemplates() error {
	c.logger.Info("Loading templates from", zap.String("path", c.cfg.Handlers.HTTP.TemplatesPath))

	layoutFiles, err := filepath.Glob(filepath.Join(c.cfg.Handlers.HTTP.TemplatesPath, "layout/*.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read layout templates: %w", err)
	}
	c.logger.Debug("Found layout files", zap.Strings("files", layoutFiles))

	pageFiles, err := filepath.Glob(filepath.Join(c.cfg.Handlers.HTTP.TemplatesPath, "pages/*.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read page templates: %w", err)
	}
	c.logger.Debug("Found page files", zap.Strings("files", pageFiles))

	for _, page := range pageFiles {
		files := append(layoutFiles, page)
		name := filepath.Base(page)

		tmpl, err := template.New(name).Funcs(template.FuncMap{
			"formatDate": func(t time.Time) string {
				return t.Format("2006-01-02 15:04:05")
			},
			"json": func(v interface{}) string {
				b, err := json.Marshal(v)
				if err != nil {
					c.logger.Error("Failed to marshal JSON", zap.Error(err))
					return ""
				}
				return string(b)
			},
		}).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		c.templates[name] = tmpl
		c.logger.Debug("Loaded template", zap.String("name", name))
	}

	return nil
}

func (c *Controller) render(w http.ResponseWriter, name string, data interface{}) {
	c.logger.Debug("Rendering template",
		zap.String("name", name),
		zap.Any("data", data),
		zap.Int("templates_count", len(c.templates)))

	tmpl, ok := c.templates[name]
	if !ok {
		c.logger.Error("Template not found",
			zap.String("name", name),
			zap.Strings("available_templates", mapKeys(c.templates)))
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		c.logger.Error("Failed to execute template",
			zap.String("name", name),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) authMiddleware(next http.Handler) http.Handler {
	return c.tokenManager.Middleware(next)
}

func (c *Controller) getToken(r *http.Request) string {
	if token := r.Context().Value(tokenKey); token != nil {
		return token.(string)
	}
	return ""
}

func mapKeys(m map[string]*template.Template) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
