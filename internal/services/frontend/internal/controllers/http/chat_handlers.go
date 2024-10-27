package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// handleRoomView displays a specific room and its messages.
func (c *Controller) handleRoomView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = WithHTTPContext(ctx, r, w)

	vars := mux.Vars(r)
	roomID := vars["id"]

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
	room, err := c.viewRoomUseCase.Execute(ctx, roomID)
	if err != nil {
		c.logger.Error("Failed to view room", zap.Error(err))
		http.Error(w, "Room not found or failed to retrieve room details", http.StatusNotFound)
		return
	}

	// Execute GetProfileUseCase to get current user.
	user, err := c.getProfileUseCase.Execute(ctx)
	if err != nil {
		c.logger.Error("Failed to get user profile", zap.Error(err))
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	// Include the token in the user data passed to the template
	c.render(w, "room_view.tmpl", map[string]interface{}{
		"Title": room.Name,
		"User": map[string]interface{}{
			"ID":    user.ID.String(),
			"Token": accessToken,
		},
		"Room": map[string]interface{}{
			"ID":      room.ID.String(),
			"Name":    room.Name,
			"OwnerID": room.OwnerID.String(),
		},
	})
}
