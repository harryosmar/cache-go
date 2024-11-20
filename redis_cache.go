package cache_go

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: client,
	}
}

func (c *RedisCache) Store(ctx context.Context, key string, value []byte, exp time.Duration) error {
	return c.client.Set(ctx, key, value, exp).Err()
}

func (c *RedisCache) StoreWithoutTTL(ctx context.Context, key string, value []byte) error {
	return c.client.Set(ctx, key, value, 0).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *RedisCache) IncrementWithTTL(ctx context.Context, key string, exp time.Duration) (int64, error) {
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, exp)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (c *RedisCache) LPush(ctx context.Context, key string, value []byte) error {
	return c.client.LPush(ctx, key, value).Err()
}

func (c *RedisCache) LRange(ctx context.Context, key string, start int64, end int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, end).Result()
}

func (c *RedisCache) LTrim(ctx context.Context, key string, start int64, end int64) error {
	return c.client.LTrim(ctx, key, start, end).Err()
}

func (c *RedisCache) LRem(ctx context.Context, key string, count int64, value []byte) error {
	return c.client.LRem(ctx, key, count, value).Err()
}

func (c *RedisCache) KeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return keys, err
	}

	return keys, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
