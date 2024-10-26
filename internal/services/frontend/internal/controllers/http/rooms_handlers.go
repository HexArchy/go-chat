package http

import (
	"encoding/json"
	"net/http"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// handleRoomsList displays the list of rooms owned by the user.
func (c *Controller) handleRoomsList(w http.ResponseWriter, r *http.Request) {
	// Retrieve tokens from session
	session, err := c.store.Get(r, c.sessionName)
	if err != nil {
		c.logger.Error("Failed to get session", zap.Error(err))
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	userID, err := c.tokenManager.GetUserID(session)
	if err != nil {
		c.logger.Error("Failed to get user ID", zap.Error(err))
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ctx := r.Context()
	ctx = WithHTTPContext(ctx, r, w)

	accessToken, err := c.tokenManager.GetAccessToken(ctx, session)
	if err != nil {
		c.logger.Error("Failed to get access token", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Inject token into context.
	ctx = contextWithToken(ctx, accessToken)

	// Execute ListRoomsUseCase.
	rooms, err := c.listRoomsUseCase.Execute(ctx, userID.String())
	if err != nil {
		c.logger.Error("Failed to list rooms", zap.Error(err))
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	c.render(w, "rooms.tmpl", map[string]interface{}{
		"Title": "My Rooms",
		"Rooms": rooms,
	})
}

// handleRoomCreate handles creating a new room.
func (c *Controller) handleRoomCreate(w http.ResponseWriter, r *http.Request) {
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

	// Execute GetProfileUseCase to get current user ID.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get user profile", zap.Error(err))
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		c.render(w, "room_create.tmpl", map[string]interface{}{
			"Title": "Create Room",
			"User":  user,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		c.render(w, "room_create.tmpl", map[string]interface{}{
			"Title": "Create Room",
			"User":  user,
			"Error": "Room name is required",
		})
		return
	}

	// Execute CreateRoomUseCase.
	room, err := c.createRoomUseCase.Execute(ctx, name, user.ID.String())
	if err != nil {
		c.logger.Error("Failed to create room", zap.Error(err))
		c.render(w, "room_create.tmpl", map[string]interface{}{
			"Title": "Create Room",
			"User":  user,
			"Error": "Failed to create room: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/rooms/"+room.ID.String(), http.StatusSeeOther)
}

// handleRoomSearch handles searching for rooms.
func (c *Controller) handleRoomSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = WithHTTPContext(ctx, r, w)

	// Retrieve tokens from session
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

	query := r.URL.Query().Get("q")
	var rooms []*entities.Room

	if query != "" {
		if len(query) < 6 {
			user, _ := c.getProfileUseCase.Execute(ctx)
			c.render(w, "room_search.tmpl", map[string]interface{}{
				"Title": "Search Rooms",
				"User":  user,
				"Error": "Search query must be at least 6 characters long",
			})
			return
		}

		// Execute SearchRoomsUseCase.
		rooms, err = c.searchRoomsUseCase.Execute(ctx, query, 20, 0)
		if err != nil {
			c.logger.Error("Failed to search rooms", zap.Error(err))
			http.Error(w, "Failed to search rooms", http.StatusInternalServerError)
			return
		}
	}

	// Execute GetProfileUseCase to get current user.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get user profile", zap.Error(err))
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	c.render(w, "room_search.tmpl", map[string]interface{}{
		"Title": "Search Rooms",
		"User":  user,
		"Query": query,
		"Rooms": rooms,
	})
}

// handleRoomView displays a specific room and its messages.
func (c *Controller) handleRoomView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

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

	// Execute ViewRoomUseCase.
	room, messages, err := c.viewRoomUseCase.Execute(ctx, roomID)
	if err != nil {
		c.logger.Error("Failed to view room", zap.Error(err))
		http.Error(w, "Room not found or failed to retrieve messages", http.StatusNotFound)
		return
	}

	// Execute GetProfileUseCase to get current user.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get user profile", zap.Error(err))
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		c.logger.Error("Failed to marshal messages", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	c.render(w, "room_view.tmpl", map[string]interface{}{
		"Title":        room.Name,
		"User":         user,
		"Room":         room,
		"Messages":     string(messagesJSON),
		"MessagesData": messages,
	})
}

// handleRoomDelete handles deleting a room.
func (c *Controller) handleRoomDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

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

	// Execute GetProfileUseCase to get current user.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get user profile", zap.Error(err))
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	// Execute DeleteRoomUseCase.
	err = c.deleteRoomUseCase.Execute(ctx, roomID, user.ID.String())
	if err != nil {
		c.logger.Error("Failed to delete room", zap.Error(err))
		http.Error(w, "Failed to delete room: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/rooms", http.StatusSeeOther)
}

// handleWebSocket manages WebSocket connections for a room.
func (c *Controller) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

	ctx := r.Context()
	ctx = WithHTTPContext(ctx, r, w)

	// Retrieve tokens from session.
	session, err := c.store.Get(r, c.sessionName)
	if err != nil {
		c.logger.Error("Failed to get session", zap.Error(err))
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	accessToken, ok := session.Values[c.tokenKey].(string)
	if !ok || accessToken == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Inject token into context.
	ctx = contextWithToken(ctx, accessToken)

	// Execute GetProfileUseCase to get current user.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get user profile", zap.Error(err))
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	// Inject userID into context.
	ctx = contextWithUserID(ctx, user.ID)

	// Upgrade the HTTP connection to a WebSocket.
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// Execute ManageWebSocketUseCase.
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		c.logger.Error("Invalid room ID", zap.Error(err))
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	err = c.manageWebSocketUseCase.Connect(ctx, roomUUID, conn)
	if err != nil {
		c.logger.Error("Failed to manage WebSocket connection", zap.Error(err))
		return
	}
}
