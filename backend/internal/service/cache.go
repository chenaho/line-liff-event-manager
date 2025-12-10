package service

import (
	"sync"
	"time"
)

type CachedEventStatus struct {
	Data      map[string]interface{}
	ExpiresAt time.Time
}

type CacheService struct {
	cache map[string]*CachedEventStatus
	mutex sync.RWMutex
	ttl   time.Duration
}

func NewCacheService(ttl time.Duration) *CacheService {
	cache := &CacheService{
		cache: make(map[string]*CachedEventStatus),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

func (c *CacheService) Get(key string) (map[string]interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	cached, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	return cached.Data, true
}

func (c *CacheService) Set(key string, data map[string]interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &CachedEventStatus{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *CacheService) Invalidate(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, key)
}

func (c *CacheService) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, cached := range c.cache {
			if now.After(cached.ExpiresAt) {
				delete(c.cache, key)
			}
		}
		c.mutex.Unlock()
	}
}
