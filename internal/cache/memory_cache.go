package cache

import (
	"context"
	"strings"
	"sync"
	"time"
)

type MemoryCache struct {
	items sync.Map
	stop  chan struct{}
	mu    sync.RWMutex
}

type cacheItem struct {
	value      []byte
	expiration int64
}

func NewMemoryCache(config CacheConfig) *MemoryCache {
	cache := &MemoryCache{
		stop: make(chan struct{}),
	}
	
	go cache.startCleanup(config.CleanupInterval)
	return cache
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := c.items.Load(key)
	if !ok {
		return nil, nil
	}
	
	item := val.(*cacheItem)
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		c.items.Delete(key)
		return nil, nil
	}
	
	return item.value, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	expiration := int64(0)
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}
	
	c.items.Store(key, &cacheItem{
		value:      value,
		expiration: expiration,
	})
	
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.items.Delete(key)
	return nil
}

func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	val, ok := c.items.Load(key)
	if !ok {
		return false, nil
	}
	
	item := val.(*cacheItem)
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		c.items.Delete(key)
		return false, nil
	}
	
	return true, nil
}

func (c *MemoryCache) InvalidateByPattern(ctx context.Context, pattern string) error {
	keysToDelete := []string{}
	
	c.items.Range(func(key, value interface{}) bool {
		k := key.(string)
		if strings.Contains(k, pattern) {
			keysToDelete = append(keysToDelete, k)
		}
		return true
	})
	
	for _, key := range keysToDelete {
		c.items.Delete(key)
	}
	
	return nil
}

func (c *MemoryCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stop:
			return
		}
	}
}

func (c *MemoryCache) cleanup() {
	c.items.Range(func(key, value interface{}) bool {
		item := value.(*cacheItem)
		if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
			c.items.Delete(key)
		}
		return true
	})
}

func (c *MemoryCache) Stop() {
	close(c.stop)
}