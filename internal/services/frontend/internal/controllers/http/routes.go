package http

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	tokenmanager "github.com/HexArch/go-chat/internal/services/frontend/internal/services/token-manager"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (c *Controller) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	if err := c.loadTemplates(); err != nil {
		c.logger.Fatal("Failed to load templates", zap.Error(err))
	}

	// Public.
	router.HandleFunc("/", c.handleHome).Methods("GET")
	router.HandleFunc("/login", c.handleLoginPage).Methods("GET")
	router.HandleFunc("/login", c.handleLogin).Methods("POST")
	router.HandleFunc("/register", c.handleRegisterPage).Methods("GET")
	router.HandleFunc("/register", c.handleRegister).Methods("POST")

	// Protected.
	router.HandleFunc("/rooms", c.requireAuth(c.handleOwnRoomsList)).Methods("GET")
	router.HandleFunc("/rooms/all", c.requireAuth(c.handleRoomsList)).Methods("GET")
	router.HandleFunc("/rooms/create", c.requireAuth(c.handleRoomCreate)).Methods("GET", "POST")
	router.HandleFunc("/rooms/search", c.requireAuth(c.handleRoomSearch)).Methods("GET")
	router.HandleFunc("/rooms/{id}", c.requireAuth(c.handleRoomView)).Methods("GET")
	router.HandleFunc("/rooms/{id}/delete", c.requireAuth(c.handleRoomDelete)).Methods("POST")
	router.HandleFunc("/logout", c.requireAuth(c.handleLogout)).Methods("POST")
	router.HandleFunc("/profile/edit", c.requireAuth(c.handleProfileEdit)).Methods("GET", "POST")
	router.HandleFunc("/profile", c.requireAuth(c.handleProfile)).Methods("GET")

	return router
}

func (c *Controller) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := c.store.Get(r, c.sessionName)
		if err != nil {
			c.logger.Error("Failed to get session", zap.Error(err))
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		sessionID, ok := session.Values["session_id"].(string)
		if !ok || sessionID == "" {
			c.logger.Debug("No session ID found",
				zap.Bool("is_new", session.IsNew))
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session.ID = sessionID

		c.logger.Debug("Checking auth",
			zap.String("session_id", session.ID),
			zap.Bool("is_new", session.IsNew))

		data, ok := session.Values["session_data"].(*tokenmanager.SessionData)
		if !ok || data == nil {
			c.logger.Debug("No session data",
				zap.String("session_id", session.ID),
			)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if data.AccessToken == "" {
			c.logger.Debug("No access token in session",
				zap.String("session_id", session.ID),
			)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "http_request", r)
		ctx = context.WithValue(ctx, "http_response", w)
		ctx = contextWithToken(ctx, data.AccessToken)

		if err := session.Save(r, w); err != nil {
			c.logger.Error("Failed to save session in middleware",
				zap.Error(err),
				zap.String("session_id", session.ID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
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
			"add": func(a, b int) int {
				return a + b
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

// contextWithToken injects the access token into the context.
func contextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, entities.ContextKeyAccessToken, token)
}

// contextWithToken injects the access token into the context.
func contextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, entities.ContextKeyUserID, userID.String())
}
