package cache_go

import (
	"context"
	"time"
)

type NocacheRepo struct {
}

func NewNocacheRepo() *NocacheRepo {
	return &NocacheRepo{}
}

func (n NocacheRepo) Store(ctx context.Context, key string, value []byte, exp time.Duration) error {
	return nil
}

func (n NocacheRepo) StoreWithoutTTL(ctx context.Context, key string, value []byte) error {
	return nil
}

func (n NocacheRepo) Get(ctx context.Context, key string) ([]byte, bool, error) {
	return nil, false, nil
}

func (n NocacheRepo) Delete(ctx context.Context, key string) error {
	return nil
}

func (n NocacheRepo) Increment(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (n NocacheRepo) IncrementWithTTL(ctx context.Context, key string, exp time.Duration) (int64, error) {
	return 0, nil
}

func (n NocacheRepo) LPush(ctx context.Context, key string, value []byte) error {
	return nil
}

func (n NocacheRepo) LRange(ctx context.Context, key string, start int64, end int64) ([]string, error) {
	return []string{}, nil
}

func (n NocacheRepo) LTrim(ctx context.Context, key string, start int64, end int64) error {
	return nil
}

func (n NocacheRepo) LRem(ctx context.Context, key string, count int64, value []byte) error {
	return nil
}

func (n NocacheRepo) KeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	return []string{}, nil
}

func (n NocacheRepo) ValuesByKeys(ctx context.Context, keys []string) ([]interface{}, error) {
	return []interface{}{}, nil
}

func (n NocacheRepo) Close() error {
	return nil
}

func (n NocacheRepo) Ping(ctx context.Context) error {
	return nil
}
