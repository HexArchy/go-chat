package http

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/chat"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/config"
	tokenmanager "github.com/HexArch/go-chat/internal/services/frontend/internal/controllers/token-manager"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	sessionName = "chat-session"
	userKey     = "user"
	tokenKey    = "token"
)

type Controller struct {
	logger        *zap.Logger
	cfg           *config.Config
	authClient    *auth.Client
	chatClient    *chat.Client
	websiteClient *website.Client
	tokenManager  *tokenmanager.TokenManager
	store         sessions.Store
	templates     map[string]*template.Template
	upgrader      websocket.Upgrader
	roomClients   map[string]map[*websocket.Conn]bool
	mu            sync.RWMutex
}

func New(
	logger *zap.Logger,
	cfg *config.Config,
	authClient *auth.Client,
	chatClient *chat.Client,
	websiteClient *website.Client,
	store sessions.Store,
) *Controller {
	return &Controller{
		logger:        logger,
		cfg:           cfg,
		authClient:    authClient,
		chatClient:    chatClient,
		websiteClient: websiteClient,
		tokenManager:  tokenmanager.NewTokenManager(store, sessionName, logger, authClient),
		store:         store,
		templates:     make(map[string]*template.Template),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		roomClients: make(map[string]map[*websocket.Conn]bool),
	}
}
