package cache

import (
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/google/uuid"
)

type RoomEntry struct {
	Room      *entities.Room
	ExpiresAt time.Time
}

type RoomCache struct {
	mu          sync.RWMutex
	data        map[uuid.UUID]RoomEntry
	ttl         time.Duration
	cleanupTick time.Duration
	stopCleanup chan struct{}
}

func NewRoomCache(ttl time.Duration) *RoomCache {
	cache := &RoomCache{
		data:        make(map[uuid.UUID]RoomEntry),
		ttl:         ttl,
		cleanupTick: ttl / 2,
		stopCleanup: make(chan struct{}),
	}

	go cache.startCleanup()
	return cache
}

func (c *RoomCache) Get(roomID uuid.UUID) (*entities.Room, bool) {
	c.mu.RLock()
	entry, exists := c.data[roomID]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.Delete(roomID)
		return nil, false
	}

	return entry.Room, true
}

func (c *RoomCache) Set(roomID uuid.UUID, room *entities.Room) {
	c.mu.Lock()
	c.data[roomID] = RoomEntry{
		Room:      room,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func (c *RoomCache) Delete(roomID uuid.UUID) {
	c.mu.Lock()
	delete(c.data, roomID)
	c.mu.Unlock()
}

func (c *RoomCache) Clear() {
	c.mu.Lock()
	c.data = make(map[uuid.UUID]RoomEntry)
	c.mu.Unlock()
}

func (c *RoomCache) Size() int {
	c.mu.RLock()
	size := len(c.data)
	c.mu.RUnlock()
	return size
}

func (c *RoomCache) Close() {
	close(c.stopCleanup)
}

func (c *RoomCache) startCleanup() {
	ticker := time.NewTicker(c.cleanupTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *RoomCache) cleanup() {
	now := time.Now()

	c.mu.Lock()
	for roomID, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, roomID)
		}
	}
	c.mu.Unlock()
}
