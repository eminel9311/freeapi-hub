package cache

import (
	"context"
	"sync"
	"time"
)

// Cache là interface chung. TUẦN 5: bạn sẽ thêm Redis implementation.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// MemoryCache là implementation in-memory đơn giản.
// TUẦN 5: bạn sẽ refactor để dùng generics: MemoryCache[T any].
type MemoryCache struct {
	mu    sync.RWMutex // RWMutex: nhiều reader cùng lúc, chỉ 1 writer
	items map[string]item
}

type item struct {
	value    []byte
	expireAt time.Time
}

func NewMemory() *MemoryCache {
	c := &MemoryCache{items: make(map[string]item)}

	// Background goroutine cleanup expired keys mỗi 5 phút.
	// TUẦN 3 sau khi học goroutine, bạn sẽ hiểu pattern này.
	go c.cleanup()

	return c
}

func (c *MemoryCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it, ok := c.items[key]
	if !ok {
		return nil, false, nil
	}
	if time.Now().After(it.expireAt) {
		return nil, false, nil // expired
	}
	return it.value, true, nil
}

func (c *MemoryCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		value:    value,
		expireAt: time.Now().Add(ttl),
	}
	return nil
}

func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.items {
			if now.After(v.expireAt) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
