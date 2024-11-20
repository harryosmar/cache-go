package cache_go

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcacheRepo struct {
	client *memcache.Client
}

func NewMemcacheRepo(server string) *MemcacheRepo {
	return &MemcacheRepo{
		client: memcache.New(server),
	}
}

func (m *MemcacheRepo) Store(ctx context.Context, key string, value []byte, exp time.Duration) error {
	return m.client.Set(&memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(exp.Seconds()),
	})
}

func (m *MemcacheRepo) StoreWithoutTTL(ctx context.Context, key string, value []byte) error {
	return m.client.Set(&memcache.Item{
		Key:   key,
		Value: value,
	})
}

func (m *MemcacheRepo) Get(ctx context.Context, key string) ([]byte, bool, error) {
	item, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return item.Value, true, nil
}

func (m *MemcacheRepo) Delete(ctx context.Context, key string) error {
	err := m.client.Delete(key)
	if err == memcache.ErrCacheMiss {
		return nil
	}
	return err
}

func (m *MemcacheRepo) Increment(ctx context.Context, key string) (int64, error) {
	// Initialize with 0 if key doesn't exist
	_, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		err = m.client.Set(&memcache.Item{
			Key:   key,
			Value: []byte("0"),
		})
		if err != nil {
			return 0, err
		}
	}

	newVal, err := m.client.Increment(key, 1)
	if err != nil {
		return 0, err
	}
	return int64(newVal), nil
}

func (m *MemcacheRepo) IncrementWithTTL(ctx context.Context, key string, exp time.Duration) (int64, error) {
	val, err := m.Increment(ctx, key)
	if err != nil {
		return 0, err
	}

	// Update TTL
	err = m.client.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(strconv.FormatInt(val, 10)),
		Expiration: int32(exp.Seconds()),
	})
	if err != nil {
		return 0, err
	}

	return val, nil
}

func (m *MemcacheRepo) LPush(ctx context.Context, key string, value []byte) error {
	return m.listOp(ctx, key, func(values []string) []string {
		return append([]string{string(value)}, values...)
	})
}

func (m *MemcacheRepo) LRange(ctx context.Context, key string, start int64, end int64) ([]string, error) {
	item, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	values := strings.Split(string(item.Value), ",")
	if len(values) == 1 && values[0] == "" {
		return []string{}, nil
	}

	// Handle negative indices
	length := int64(len(values))
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}
	if end < 0 {
		end = length + end
		if end < 0 {
			end = 0
		}
	}
	if end >= length {
		end = length - 1
	}
	if start > end {
		return []string{}, nil
	}

	return values[start : end+1], nil
}

func (m *MemcacheRepo) LTrim(ctx context.Context, key string, start int64, end int64) error {
	values, err := m.LRange(ctx, key, start, end)
	if err != nil {
		return err
	}

	if len(values) == 0 {
		return m.Delete(ctx, key)
	}

	return m.client.Set(&memcache.Item{
		Key:   key,
		Value: []byte(strings.Join(values, ",")),
	})
}

func (m *MemcacheRepo) LRem(ctx context.Context, key string, count int64, value []byte) error {
	return m.listOp(ctx, key, func(values []string) []string {
		targetValue := string(value)
		var result []string
		removed := int64(0)

		if count > 0 {
			// Remove first N occurrences
			for _, v := range values {
				if v == targetValue && removed < count {
					removed++
					continue
				}
				result = append(result, v)
			}
		} else if count < 0 {
			// Remove last N occurrences
			count = -count
			for i := len(values) - 1; i >= 0; i-- {
				if values[i] == targetValue && removed < count {
					removed++
					continue
				}
				result = append([]string{values[i]}, result...)
			}
		} else {
			// Remove all occurrences
			for _, v := range values {
				if v != targetValue {
					result = append(result, v)
				}
			}
		}

		return result
	})
}

func (m *MemcacheRepo) KeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	// Memcache doesn't support pattern-based key search
	// This is a limitation of the memcache protocol
	return nil, fmt.Errorf("pattern-based key search is not supported in memcache")
}

func (m *MemcacheRepo) Close() error {
	// Close any open connections
	return nil
}

func (m *MemcacheRepo) Ping(ctx context.Context) error {
	// Try to get a non-existent key to check connection
	_, err := m.client.Get("__ping__")
	if err == memcache.ErrCacheMiss {
		return nil
	}
	return err
}

// Helper function for list operations
func (m *MemcacheRepo) listOp(ctx context.Context, key string, op func([]string) []string) error {
	item, err := m.client.Get(key)
	values := []string{}

	if err != nil && err != memcache.ErrCacheMiss {
		return err
	}

	if err == nil && len(item.Value) > 0 {
		values = strings.Split(string(item.Value), ",")
		if len(values) == 1 && values[0] == "" {
			values = []string{}
		}
	}

	values = op(values)

	if len(values) == 0 {
		return m.Delete(ctx, key)
	}

	return m.client.Set(&memcache.Item{
		Key:   key,
		Value: []byte(strings.Join(values, ",")),
	})
}
