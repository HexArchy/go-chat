package http

import (
	"net/http"
	"strconv"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// handleRoomsList displays a paginated list of all rooms.
func (c *Controller) handleRoomsList(w http.ResponseWriter, r *http.Request) {
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

	// Parse pagination query parameters.
	limit := 100
	offset := 0

	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	// Execute ListRoomsUseCase to get all rooms with pagination.
	rooms, err := c.listRoomsUseCase.Execute(ctx, int32(limit), int32(offset))
	if err != nil {
		c.logger.Error("Failed to list rooms", zap.Error(err))
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	// Execute GetProfileUseCase to get current user.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get user profile", zap.Error(err))
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	c.render(w, "all_rooms.tmpl", map[string]interface{}{
		"Title":      "All Rooms",
		"Rooms":      rooms,
		"User":       user,
		"Limit":      limit,
		"Offset":     offset,
		"NextOffset": offset + limit,
		"PrevOffset": offset - limit,
		"CurrentTab": "all",
	})
}

// handleRoomsList displays the list of rooms owned by the user.
func (c *Controller) handleOwnRoomsList(w http.ResponseWriter, r *http.Request) {
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

	// Execute ListOwnRoomsUseCase.
	rooms, err := c.listOwnRoomsUseCase.Execute(ctx, userID.String())
	if err != nil {
		c.logger.Error("Failed to list rooms", zap.Error(err))
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	// Execute GetProfileUseCase to get current user.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get user profile", zap.Error(err))
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	c.render(w, "rooms.tmpl", map[string]interface{}{
		"Title":      "My Rooms",
		"Rooms":      rooms,
		"User":       user,
		"CurrentTab": "my",
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

	query := r.URL.Query().Get("q")
	var rooms []*entities.Room

	if query != "" {
		if len(query) < 2 {
			user, _ := c.getProfileUseCase.Execute(ctx)
			c.render(w, "room_search.tmpl", map[string]interface{}{
				"Title": "Search Rooms",
				"User":  user,
				"Error": "Search query must be at least 2 characters long",
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
