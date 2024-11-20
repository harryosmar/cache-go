package cache_go

import (
	"context"
	"testing"
	"time"
)

func setupTestRedis(t *testing.T) *RedisCache {
	cache := NewRedisCache("localhost:6379", "", 0)
	ctx := context.Background()
	err := cache.Ping(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
	return cache
}

func cleanupTestRedis(t *testing.T, cache *RedisCache) {
	if err := cache.Close(); err != nil {
		t.Errorf("Failed to close Redis connection: %v", err)
	}
}

func TestRedisCache_BasicOperations(t *testing.T) {
	cache := setupTestRedis(t)
	defer cleanupTestRedis(t, cache)
	ctx := context.Background()

	// Test Store and Get
	key := "test_key"
	value := []byte("test_value")
	if err := cache.Store(ctx, key, value, time.Minute); err != nil {
		t.Errorf("Failed to store value: %v", err)
	}

	got, exists, err := cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if !exists {
		t.Error("Key should exist")
	}
	if string(got) != string(value) {
		t.Errorf("Got %s, want %s", string(got), string(value))
	}

	// Test Delete
	if err := cache.Delete(ctx, key); err != nil {
		t.Errorf("Failed to delete key: %v", err)
	}

	_, exists, err = cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get value after delete: %v", err)
	}
	if exists {
		t.Error("Key should not exist after deletion")
	}
}

func TestRedisCache_Increment(t *testing.T) {
	cache := setupTestRedis(t)
	defer cleanupTestRedis(t, cache)
	ctx := context.Background()

	key := "test_counter"
	// Clean up any existing key first
	_ = cache.Delete(ctx, key)
	
	// Test Increment
	val, err := cache.Increment(ctx, key)
	if err != nil {
		t.Errorf("Failed to increment counter: %v", err)
	}
	if val != 1 {
		t.Errorf("First increment should return 1, got %d", val)
	}

	// Test IncrementWithTTL
	val, err = cache.IncrementWithTTL(ctx, key, time.Minute)
	if err != nil {
		t.Errorf("Failed to increment counter with TTL: %v", err)
	}
	if val != 2 {
		t.Errorf("Second increment should return 2, got %d", val)
	}
}

func TestRedisCache_ListOperations(t *testing.T) {
	cache := setupTestRedis(t)
	defer cleanupTestRedis(t, cache)
	ctx := context.Background()

	key := "test_list"
	// Clean up any existing key first
	_ = cache.Delete(ctx, key)
	
	value1 := []byte("value1")
	value2 := []byte("value2")

	// Test LPush
	if err := cache.LPush(ctx, key, value1); err != nil {
		t.Errorf("Failed to push value1: %v", err)
	}
	if err := cache.LPush(ctx, key, value2); err != nil {
		t.Errorf("Failed to push value2: %v", err)
	}

	// Test LRange
	values, err := cache.LRange(ctx, key, 0, -1)
	if err != nil {
		t.Errorf("Failed to get range: %v", err)
	}
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	// Test LRem
	if err := cache.LRem(ctx, key, 1, value1); err != nil {
		t.Errorf("Failed to remove value: %v", err)
	}

	// Test LTrim
	if err := cache.LTrim(ctx, key, 0, 0); err != nil {
		t.Errorf("Failed to trim list: %v", err)
	}
}

func TestRedisCache_KeysByPattern(t *testing.T) {
	cache := setupTestRedis(t)
	defer cleanupTestRedis(t, cache)
	ctx := context.Background()

	// Store some test keys
	testKeys := []string{"test:1", "test:2", "other:1"}
	for _, key := range testKeys {
		if err := cache.StoreWithoutTTL(ctx, key, []byte("value")); err != nil {
			t.Errorf("Failed to store test key %s: %v", key, err)
		}
	}

	// Test pattern search
	keys, err := cache.KeysByPattern(ctx, "test:*")
	if err != nil {
		t.Errorf("Failed to search keys by pattern: %v", err)
	}

	// Should find 2 keys with prefix "test:"
	matchCount := 0
	for _, key := range keys {
		for _, testKey := range testKeys {
			if key == testKey && testKey[:5] == "test:" {
				matchCount++
			}
		}
	}
	if matchCount != 2 {
		t.Errorf("Expected to find 2 keys matching pattern 'test:*', found %d", matchCount)
	}

	// Cleanup test keys
	for _, key := range testKeys {
		if err := cache.Delete(ctx, key); err != nil {
			t.Errorf("Failed to cleanup test key %s: %v", key, err)
		}
	}
}
