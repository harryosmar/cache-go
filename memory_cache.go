package cache_go

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CacheItem struct {
	value     []byte
	expiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]CacheItem
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]CacheItem),
	}
}

func (m *MemoryCache) Store(ctx context.Context, key string, value []byte, exp time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt := time.Now().Add(exp)
	m.items[key] = CacheItem{
		value:     value,
		expiresAt: expiresAt,
	}
	return nil
}

func (m *MemoryCache) StoreWithoutTTL(ctx context.Context, key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[key] = CacheItem{
		value:     value,
		expiresAt: time.Time{}, // Zero time means no expiration
	}
	return nil
}

func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.items[key]
	if !exists {
		return nil, false, nil
	}

	// Check if item has expired
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		delete(m.items, key)
		return nil, false, nil
	}

	return item.value, true, nil
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.items, key)
	return nil
}

func (m *MemoryCache) Increment(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var val int64 = 0
	if item, exists := m.items[key]; exists {
		val = bytesToInt64(item.value)
	}

	val++
	m.items[key] = CacheItem{
		value:     int64ToBytes(val),
		expiresAt: time.Time{},
	}

	return val, nil
}

func (m *MemoryCache) IncrementWithTTL(ctx context.Context, key string, exp time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var val int64 = 0
	if item, exists := m.items[key]; exists {
		val = bytesToInt64(item.value)
	}

	val++
	m.items[key] = CacheItem{
		value:     int64ToBytes(val),
		expiresAt: time.Now().Add(exp),
	}

	return val, nil
}

func (m *MemoryCache) LPush(ctx context.Context, key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var values []string
	if item, exists := m.items[key]; exists {
		values = strings.Split(string(item.value), ",")
		if len(values) == 1 && values[0] == "" {
			values = []string{}
		}
	}

	values = append([]string{string(value)}, values...)
	m.items[key] = CacheItem{
		value:     []byte(strings.Join(values, ",")),
		expiresAt: time.Time{},
	}

	return nil
}

func (m *MemoryCache) LRange(ctx context.Context, key string, start int64, end int64) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.items[key]
	if !exists {
		return []string{}, nil
	}

	values := strings.Split(string(item.value), ",")
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

func (m *MemoryCache) LTrim(ctx context.Context, key string, start int64, end int64) error {
	values, err := m.LRange(ctx, key, start, end)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(values) == 0 {
		delete(m.items, key)
		return nil
	}

	m.items[key] = CacheItem{
		value:     []byte(strings.Join(values, ",")),
		expiresAt: time.Time{},
	}

	return nil
}

func (m *MemoryCache) LRem(ctx context.Context, key string, count int64, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[key]
	if !exists {
		return nil
	}

	values := strings.Split(string(item.value), ",")
	if len(values) == 1 && values[0] == "" {
		return nil
	}

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

	if len(result) == 0 {
		delete(m.items, key)
		return nil
	}

	m.items[key] = CacheItem{
		value:     []byte(strings.Join(result, ",")),
		expiresAt: time.Time{},
	}

	return nil
}

func (m *MemoryCache) KeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matches []string
	for key := range m.items {
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %v", err)
		}
		if matched {
			matches = append(matches, key)
		}
	}

	return matches, nil
}

func (m *MemoryCache) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear all items
	m.items = make(map[string]CacheItem)
	return nil
}

func (m *MemoryCache) Ping(ctx context.Context) error {
	return nil // In-memory cache is always available
}

// Helper functions for converting between int64 and []byte
func int64ToBytes(n int64) []byte {
	return []byte(fmt.Sprintf("%d", n))
}

func bytesToInt64(b []byte) int64 {
	var n int64
	fmt.Sscanf(string(b), "%d", &n)
	return n
}
