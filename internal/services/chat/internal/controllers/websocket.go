package controllers

import (
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WebSocketConnection struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	writeChan chan *entities.ChatEvent
	done      chan struct{}
	logger    *zap.Logger
}

func NewWebSocketConnection(conn *websocket.Conn, logger *zap.Logger) *WebSocketConnection {
	ws := &WebSocketConnection{
		conn:      conn,
		writeChan: make(chan *entities.ChatEvent, 100),
		done:      make(chan struct{}),
		logger:    logger,
	}

	go ws.writeLoop()

	return ws
}

// SendEvent реализует ChatConnection.SendEvent
func (w *WebSocketConnection) SendEvent(event *entities.ChatEvent) error {
	select {
	case w.writeChan <- event:
		return nil
	case <-w.done:
		return entities.ErrConnectionClosed
	}
}

// Close реализует ChatConnection.Close
func (w *WebSocketConnection) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return nil
	default:
		close(w.done)
		return w.conn.Close()
	}
}

func (w *WebSocketConnection) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		w.Close()
	}()

	for {
		select {
		case event := <-w.writeChan:
			w.mu.Lock()
			w.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			err := w.conn.WriteJSON(event)
			w.mu.Unlock()

			if err != nil {
				w.logger.Error("Failed to write event", zap.Error(err))
				return
			}

		case <-ticker.C:
			w.mu.Lock()
			w.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := w.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				w.mu.Unlock()
				return
			}
			w.mu.Unlock()

		case <-w.done:
			return
		}
	}
}
