package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(opt *redis.Options) (CacheClient, error) {
	client := redis.NewClient(opt)

	err := client.Ping(context.Background()).Err()
	if err != nil {
		logger.Error("Unable to connect to redis, error: %+v", err)
		return nil, err
	}

	return &redisClient{
		client: client,
	}, nil
}

func (r *redisClient) Get(ctx context.Context, key string) (any, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) Scan(ctx context.Context, cursor uint64, pattern string, count int64) ([]string, uint64, error) {
	return r.client.Scan(ctx, cursor, pattern, count).Result()
}

func (r *redisClient) Set(ctx context.Context, key string, value any) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *redisClient) SetWithTTL(ctx context.Context, key string, value any, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisClient) Close() error {
	return r.client.Close()
}

func (r *redisClient) Ping() error {
	return r.client.Ping(context.Background()).Err()
}
