package cache_go

import (
	"context"
	"testing"
	"time"
)

func setupTestMemcache(t *testing.T) *MemcacheRepo {
	cache := NewMemcacheRepo("localhost:11211")
	ctx := context.Background()
	err := cache.Ping(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to Memcache: %v", err)
	}
	return cache
}

func TestMemcacheRepo_BasicOperations(t *testing.T) {
	cache := setupTestMemcache(t)
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

func TestMemcacheRepo_Increment(t *testing.T) {
	cache := setupTestMemcache(t)
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

	// Cleanup
	_ = cache.Delete(ctx, key)
}

func TestMemcacheRepo_ListOperations(t *testing.T) {
	cache := setupTestMemcache(t)
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

	// Cleanup
	_ = cache.Delete(ctx, key)
}

func TestMemcacheRepo_KeysByPattern(t *testing.T) {
	cache := setupTestMemcache(t)
	ctx := context.Background()

	_, err := cache.KeysByPattern(ctx, "test:*")
	if err == nil {
		t.Error("KeysByPattern should return error for Memcache")
	}
}
