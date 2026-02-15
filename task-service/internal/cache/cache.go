package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	InvalidateByPattern(ctx context.Context, pattern string) error
}

type CacheConfig struct {
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
	MaxItems        int
}