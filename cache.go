package cache_go

import (
	"context"
	"time"
)

//go:generate mockgen -destination=mocks/mock_CacheRepo.go -package=mocks . CacheRepo
type CacheRepo interface {
	Store(ctx context.Context, key string, value []byte, exp time.Duration) error
	StoreWithoutTTL(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
	Increment(ctx context.Context, key string) (int64, error)
	IncrementWithTTL(ctx context.Context, key string, exp time.Duration) (int64, error)
	LPush(ctx context.Context, key string, value []byte) error
	LRange(ctx context.Context, key string, start int64, end int64) ([]string, error)
	LTrim(ctx context.Context, key string, start int64, end int64) error
	LRem(ctx context.Context, key string, count int64, value []byte) error
	KeysByPattern(ctx context.Context, pattern string) ([]string, error)
	ValuesByKeys(ctx context.Context, keys []string) ([]interface{}, error)
	Close() error
	Ping(ctx context.Context) error
}
