package cache

import (
	"context"
	"time"
)

type CacheClient interface {
	Get(ctx context.Context, key string) (any, error)
	Scan(ctx context.Context, cursor uint64, pattern string, count int64) ([]string, uint64, error)
	Set(ctx context.Context, key string, value any) error
	SetWithTTL(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Ping() error
	Close() error
}
