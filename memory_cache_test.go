package cache_go

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	cache := NewMemoryCache()
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

	// Test expiration
	expKey := "exp_key"
	if err := cache.Store(ctx, expKey, value, time.Millisecond); err != nil {
		t.Errorf("Failed to store value with expiration: %v", err)
	}
	time.Sleep(time.Millisecond * 2)
	_, exists, err = cache.Get(ctx, expKey)
	if err != nil {
		t.Errorf("Failed to get expired value: %v", err)
	}
	if exists {
		t.Error("Key should have expired")
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

func TestMemoryCache_Increment(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "test_counter"

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

	// Test expiration
	expKey := "exp_counter"
	_, err = cache.IncrementWithTTL(ctx, expKey, time.Millisecond)
	if err != nil {
		t.Errorf("Failed to increment counter with expiration: %v", err)
	}
	time.Sleep(time.Millisecond * 2)
	val, err = cache.Increment(ctx, expKey)
	if err != nil {
		t.Errorf("Failed to increment expired counter: %v", err)
	}
	if val != 1 {
		t.Errorf("Increment after expiration should return 1, got %d", val)
	}
}

func TestMemoryCache_ListOperations(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	key := "test_list"
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
	if values[0] != string(value2) {
		t.Errorf("Expected first value to be %s, got %s", string(value2), values[0])
	}

	// Test negative indices
	values, err = cache.LRange(ctx, key, -2, -1)
	if err != nil {
		t.Errorf("Failed to get range with negative indices: %v", err)
	}
	if len(values) != 2 {
		t.Errorf("Expected 2 values with negative indices, got %d", len(values))
	}

	// Test LRem
	if err := cache.LRem(ctx, key, 1, value1); err != nil {
		t.Errorf("Failed to remove value: %v", err)
	}

	values, err = cache.LRange(ctx, key, 0, -1)
	if err != nil {
		t.Errorf("Failed to get range after remove: %v", err)
	}
	if len(values) != 1 {
		t.Errorf("Expected 1 value after remove, got %d", len(values))
	}

	// Test LTrim
	if err := cache.LTrim(ctx, key, 0, 0); err != nil {
		t.Errorf("Failed to trim list: %v", err)
	}

	values, err = cache.LRange(ctx, key, 0, -1)
	if err != nil {
		t.Errorf("Failed to get range after trim: %v", err)
	}
	if len(values) != 1 {
		t.Errorf("Expected 1 value after trim, got %d", len(values))
	}
}

func TestMemoryCache_KeysByPattern(t *testing.T) {
	cache := NewMemoryCache()
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

	// Test invalid pattern
	_, err = cache.KeysByPattern(ctx, "[invalid")
	if err == nil {
		t.Error("Expected error for invalid pattern")
	}
}
