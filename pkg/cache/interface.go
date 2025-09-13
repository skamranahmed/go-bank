package cache

import (
	"context"
	"time"
)

type CacheClient interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) error
	SetWithTTL(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Ping() error
	Close() error
}
