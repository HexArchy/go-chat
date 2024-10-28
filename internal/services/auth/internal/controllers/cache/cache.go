// internal/services/auth/internal/cache/cache.go
package cache

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// TokenEntry представляет запись в кэше для токена
type TokenEntry struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// TokenCache представляет потокобезопасный кэш для токенов
type TokenCache struct {
	mu          sync.RWMutex
	data        map[string]TokenEntry
	ttl         time.Duration
	cleanupTick time.Duration
	stopCleanup chan struct{}
}

// NewTokenCache создает новый экземпляр TokenCache
func NewTokenCache(ttl time.Duration) *TokenCache {
	cache := &TokenCache{
		data:        make(map[string]TokenEntry),
		ttl:         ttl,
		cleanupTick: ttl / 2,
		stopCleanup: make(chan struct{}),
	}

	go cache.startCleanup()
	return cache
}

// Get получает UserID из кэша по токену
func (c *TokenCache) Get(token string) (uuid.UUID, bool) {
	c.mu.RLock()
	entry, exists := c.data[token]
	c.mu.RUnlock()

	if !exists {
		return uuid.Nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.Delete(token)
		return uuid.Nil, false
	}

	return entry.UserID, true
}

// Set добавляет или обновляет запись в кэше
func (c *TokenCache) Set(token string, userID uuid.UUID) {
	c.mu.Lock()
	c.data[token] = TokenEntry{
		UserID:    userID,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// Delete удаляет запись из кэша
func (c *TokenCache) Delete(token string) {
	c.mu.Lock()
	delete(c.data, token)
	c.mu.Unlock()
}

// Clear очищает весь кэш
func (c *TokenCache) Clear() {
	c.mu.Lock()
	c.data = make(map[string]TokenEntry)
	c.mu.Unlock()
}

// Size возвращает текущий размер кэша
func (c *TokenCache) Size() int {
	c.mu.RLock()
	size := len(c.data)
	c.mu.RUnlock()
	return size
}

// Close останавливает фоновую очистку кэша
func (c *TokenCache) Close() {
	close(c.stopCleanup)
}

// startCleanup запускает фоновую очистку устаревших записей
func (c *TokenCache) startCleanup() {
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

// cleanup удаляет устаревшие записи из кэша
func (c *TokenCache) cleanup() {
	now := time.Now()

	c.mu.Lock()
	for token, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, token)
		}
	}
	c.mu.Unlock()
}

// UserCache представляет кэш для данных пользователей
type UserCache struct {
	mu          sync.RWMutex
	data        map[uuid.UUID]interface{}
	ttl         time.Duration
	cleanupTick time.Duration
	stopCleanup chan struct{}
}

// NewUserCache создает новый экземпляр UserCache
func NewUserCache(ttl time.Duration) *UserCache {
	cache := &UserCache{
		data:        make(map[uuid.UUID]interface{}),
		ttl:         ttl,
		cleanupTick: ttl / 2,
		stopCleanup: make(chan struct{}),
	}

	go cache.startCleanup()
	return cache
}

type userEntry struct {
	data      interface{}
	expiresAt time.Time
}

// Get получает данные пользователя из кэша
func (c *UserCache) Get(userID uuid.UUID) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.data[userID].(userEntry)
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.Delete(userID)
		return nil, false
	}

	return entry.data, true
}

// Set добавляет или обновляет данные пользователя в кэше
func (c *UserCache) Set(userID uuid.UUID, data interface{}) {
	c.mu.Lock()
	c.data[userID] = userEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// Delete удаляет данные пользователя из кэша
func (c *UserCache) Delete(userID uuid.UUID) {
	c.mu.Lock()
	delete(c.data, userID)
	c.mu.Unlock()
}

// Clear очищает весь кэш пользователей
func (c *UserCache) Clear() {
	c.mu.Lock()
	c.data = make(map[uuid.UUID]interface{})
	c.mu.Unlock()
}

// Close останавливает фоновую очистку кэша
func (c *UserCache) Close() {
	close(c.stopCleanup)
}

// startCleanup запускает фоновую очистку устаревших записей
func (c *UserCache) startCleanup() {
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

// cleanup удаляет устаревшие записи из кэша
func (c *UserCache) cleanup() {
	now := time.Now()

	c.mu.Lock()
	for userID, entry := range c.data {
		if e, ok := entry.(userEntry); ok && now.After(e.expiresAt) {
			delete(c.data, userID)
		}
	}
	c.mu.Unlock()
}
