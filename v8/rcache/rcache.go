// Package rcache defines the replay-cache provider contract used by service
// authentication paths.
//
// This package exists to decouple replay-detection storage from Kerberos
// verification logic. The service package derives replay keys and interprets
// duplicate-key behavior, while rcache implementations only need to provide an
// add-if-absent key set with TTL.
//
// If no provider is configured, service lazily uses New (MemCache). To use a
// non-default backend (for example Redis), implement Cache and configure it via
// service.SetReplayCache(customCache) during process startup, before any
// authentication occurs.
//
// A custom implementation should:
//   - perform Add as an atomic check-and-store operation.
//   - return ErrAlreadyExists (or a wrapped form) when the key already exists
//     and has not expired.
//   - honor context cancellation and timeouts in Add.
//   - make Close idempotent and release resources.
package rcache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrAlreadyExists indicates the key already exists and has not expired.
var ErrAlreadyExists = errors.New("rcache: key already exists")

// Cache is an add-if-absent key set where each entry has a TTL.
type Cache interface {
	// Add records the key with the given expiry duration.
	// If the key already exists and has not expired it returns ErrAlreadyExists.
	Add(ctx context.Context, key string, expiry time.Duration) error

	// Close releases cache resources. It is safe to call multiple times.
	Close() error
}

// MemCache is an in-memory Cache implementation.
type MemCache struct {
	entries map[string]time.Time
	mux     sync.Mutex
	maxAge  time.Duration
	done    chan struct{}
	once    sync.Once
}

// New creates a new in-memory replay cache.
func New(maxAge time.Duration) *MemCache {
	if maxAge <= 0 {
		maxAge = time.Second
	}

	c := &MemCache{
		entries: make(map[string]time.Time),
		maxAge:  maxAge,
		done:    make(chan struct{}),
	}

	go c.cleanup()
	return c
}

// Add atomically checks and stores a key with its expiry.
func (c *MemCache) Add(_ context.Context, key string, expiry time.Duration) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	now := time.Now()
	if exp, ok := c.entries[key]; ok && now.Before(exp) {
		return ErrAlreadyExists
	}

	c.entries[key] = now.Add(expiry)
	return nil
}

func (c *MemCache) cleanup() {
	ticker := time.NewTicker(c.maxAge)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.clearExpired()
		case <-c.done:
			return
		}
	}
}

func (c *MemCache) clearExpired() {
	c.mux.Lock()
	defer c.mux.Unlock()

	now := time.Now()
	for key, exp := range c.entries {
		if !exp.After(now) {
			delete(c.entries, key)
		}
	}
}

// Close stops the background cleanup goroutine.
func (c *MemCache) Close() error {
	c.once.Do(func() {
		close(c.done)
	})
	return nil
}
