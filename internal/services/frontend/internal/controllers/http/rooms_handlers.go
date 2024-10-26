package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (c *Controller) handleRoomsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := c.tokenManager.GetToken(ctx)

	user, err := c.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusUnauthorized)
		return
	}

	rooms, err := c.websiteClient.GetOwnerRooms(ctx, token, user.ID.String())
	if err != nil {
		c.logger.Error("Failed to get rooms", zap.Error(err))
		http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
		return
	}

	c.render(w, "rooms.tmpl", map[string]interface{}{
		"Title": "My Rooms",
		"User":  user,
		"Rooms": rooms,
	})
}

func (c *Controller) handleRoomCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := c.tokenManager.GetToken(ctx)

	user, err := c.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusUnauthorized)
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

	room, err := c.websiteClient.CreateRoom(ctx, token, name, user.ID.String())
	if err != nil {
		c.render(w, "room_create.tmpl", map[string]interface{}{
			"Title": "Create Room",
			"User":  user,
			"Error": "Failed to create room: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/rooms/"+room.ID.String(), http.StatusSeeOther)
}

func (c *Controller) handleRoomSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := c.tokenManager.GetToken(ctx)

	user, err := c.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	var rooms []*entities.Room

	if query != "" {
		if len(query) < 6 {
			c.render(w, "room_search.tmpl", map[string]interface{}{
				"Title": "Search Rooms",
				"User":  user,
				"Error": "Search query must be at least 6 characters long",
			})
			return
		}

		rooms, err = c.websiteClient.SearchRooms(ctx, token, query, 20, 0)
		if err != nil {
			c.logger.Error("Failed to search rooms", zap.Error(err))
			http.Error(w, "Failed to search rooms", http.StatusInternalServerError)
			return
		}
	}

	c.render(w, "room_search.tmpl", map[string]interface{}{
		"Title": "Search Rooms",
		"User":  user,
		"Query": query,
		"Rooms": rooms,
	})
}

func (c *Controller) handleRoomView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

	ctx := r.Context()
	token := c.tokenManager.GetToken(ctx)

	user, err := c.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusUnauthorized)
		return
	}

	room, err := c.websiteClient.GetRoom(ctx, token, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	messages, err := c.chatClient.GetMessages(ctx, token, room.ID, 50, 0)
	if err != nil {
		c.logger.Error("Failed to get messages", zap.Error(err))
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
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

func (c *Controller) handleRoomDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

	ctx := r.Context()
	token := c.tokenManager.GetToken(ctx)

	user, err := c.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusUnauthorized)
		return
	}

	room, err := c.websiteClient.GetRoom(ctx, token, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if room.OwnerID != user.ID.String() {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if err := c.websiteClient.DeleteRoom(ctx, token, roomID, user.ID.String()); err != nil {
		c.logger.Error("Failed to delete room", zap.Error(err))
		http.Error(w, "Failed to delete room", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/rooms", http.StatusSeeOther)
}

func (c *Controller) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

	ctx := r.Context()
	token := c.getToken(r)

	user, err := c.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "Failed to validate token", http.StatusUnauthorized)
		return
	}

	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()

	c.mu.Lock()
	if c.roomClients[roomID] == nil {
		c.roomClients[roomID] = make(map[*websocket.Conn]bool)
	}
	c.roomClients[roomID][conn] = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.roomClients[roomID], conn)
		if len(c.roomClients[roomID]) == 0 {
			delete(c.roomClients, roomID)
		}
		c.mu.Unlock()
	}()

	c.broadcastUserEvent(roomID, user.ID.String(), "join")

	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		c.logger.Error("Invalid room ID", zap.Error(err))
		return
	}

	eventChan, errChan, err := c.chatClient.Connect(ctx, token, roomUUID, user.ID)
	if err != nil {
		c.logger.Error("Failed to connect to chat", zap.Error(err))
		return
	}

	go func() {
		defer func() {
			c.broadcastUserEvent(roomID, user.ID.String(), "leave")
		}()

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("WebSocket read error", zap.Error(err))
				}
				return
			}

			if messageType == websocket.TextMessage {
				var message struct {
					Content string `json:"content"`
				}
				if err := json.Unmarshal(p, &message); err != nil {
					continue
				}

				if err := c.chatClient.SendMessage(ctx, token, roomUUID, user.ID, message.Content); err != nil {
					c.logger.Error("Failed to send message", zap.Error(err))
				}
			}
		}
	}()

	for {
		select {
		case event := <-eventChan:
			data := map[string]interface{}{
				"type":      event.Type,
				"userId":    event.UserID.String(),
				"timestamp": event.Timestamp,
			}

			if event.Message != nil {
				data["message"] = map[string]interface{}{
					"id":        event.Message.ID.String(),
					"content":   event.Message.Content,
					"userId":    event.Message.UserID.String(),
					"createdAt": event.Message.CreatedAt,
				}
			}

			if err := conn.WriteJSON(data); err != nil {
				return
			}

		case err := <-errChan:
			c.logger.Error("Chat connection error", zap.Error(err))
			return

		case <-ctx.Done():
			return
		}
	}
}

func (c *Controller) broadcastUserEvent(roomID, userID, eventType string) {
	event := map[string]interface{}{
		"type":      eventType,
		"userId":    userID,
		"timestamp": time.Now(),
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	for conn := range c.roomClients[roomID] {
		if err := conn.WriteJSON(event); err != nil {
			c.logger.Error("Failed to broadcast event", zap.Error(err))
			conn.Close()
			delete(c.roomClients[roomID], conn)
		}
	}
}
