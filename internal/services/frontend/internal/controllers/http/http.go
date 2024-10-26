package http

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/config"
	tokenmanager "github.com/HexArch/go-chat/internal/services/frontend/internal/services/token-manager"
	authUseCases "github.com/HexArch/go-chat/internal/services/frontend/internal/use-cases/auth"
	profileUseCases "github.com/HexArch/go-chat/internal/services/frontend/internal/use-cases/profile"
	roomsUseCases "github.com/HexArch/go-chat/internal/services/frontend/internal/use-cases/rooms"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Controller handles HTTP requests and interacts with use-cases.
type Controller struct {
	logger                 *zap.Logger
	cfg                    *config.Config
	loginUseCase           authUseCases.LoginUseCase
	registerUseCase        authUseCases.RegisterUseCase
	logoutUseCase          authUseCases.LogoutUseCase
	getProfileUseCase      profileUseCases.GetProfileUseCase
	editProfileUseCase     profileUseCases.EditProfileUseCase
	createRoomUseCase      roomsUseCases.CreateRoomUseCase
	deleteRoomUseCase      roomsUseCases.DeleteRoomUseCase
	listRoomsUseCase       roomsUseCases.ListRoomsUseCase
	listOwnRoomsUseCase    roomsUseCases.ListOwnRoomsUseCase
	searchRoomsUseCase     roomsUseCases.SearchRoomsUseCase
	viewRoomUseCase        roomsUseCases.ViewRoomUseCase
	manageWebSocketUseCase roomsUseCases.ManageWebSocketUseCase
	tokenManager           tokenmanager.TokenManager
	templates              map[string]*template.Template
	upgrader               websocket.Upgrader
	roomClients            map[string]map[*websocket.Conn]bool
	store                  sessions.Store
	sessionName            string
	tokenKey               string
	mu                     sync.RWMutex
}

// NewController creates a new Controller with all dependencies.
func NewController(
	logger *zap.Logger,
	cfg *config.Config,
	loginUseCase authUseCases.LoginUseCase,
	registerUseCase authUseCases.RegisterUseCase,
	logoutUseCase authUseCases.LogoutUseCase,
	getProfileUseCase profileUseCases.GetProfileUseCase,
	editProfileUseCase profileUseCases.EditProfileUseCase,
	createRoomUseCase roomsUseCases.CreateRoomUseCase,
	deleteRoomUseCase roomsUseCases.DeleteRoomUseCase,
	listRoomsUseCase roomsUseCases.ListRoomsUseCase,
	listOwnRoomsUseCase roomsUseCases.ListOwnRoomsUseCase,
	searchRoomsUseCase roomsUseCases.SearchRoomsUseCase,
	viewRoomUseCase roomsUseCases.ViewRoomUseCase,
	manageWebSocketUseCase roomsUseCases.ManageWebSocketUseCase,
	tokenManager tokenmanager.TokenManager,
	store sessions.Store,
	sessionName string,
	tokenKey string,
) *Controller {
	return &Controller{
		logger:                 logger,
		cfg:                    cfg,
		loginUseCase:           loginUseCase,
		registerUseCase:        registerUseCase,
		logoutUseCase:          logoutUseCase,
		getProfileUseCase:      getProfileUseCase,
		editProfileUseCase:     editProfileUseCase,
		createRoomUseCase:      createRoomUseCase,
		deleteRoomUseCase:      deleteRoomUseCase,
		listRoomsUseCase:       listRoomsUseCase,
		listOwnRoomsUseCase:    listOwnRoomsUseCase,
		searchRoomsUseCase:     searchRoomsUseCase,
		viewRoomUseCase:        viewRoomUseCase,
		manageWebSocketUseCase: manageWebSocketUseCase,
		tokenManager:           tokenManager,
		templates:              make(map[string]*template.Template),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		roomClients: make(map[string]map[*websocket.Conn]bool),
		store:       store,
		sessionName: sessionName,
		tokenKey:    tokenKey,
	}
}
