package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jcmturner/gokrb5/v8/types"
)

type stubReplayCache struct {
	addErr     error
	addCalls   int
	lastKey    string
	lastExpiry time.Duration
}

func (s *stubReplayCache) Add(_ context.Context, key string, expiry time.Duration) error {
	s.addCalls++
	s.lastKey = key
	s.lastExpiry = expiry
	return s.addErr
}

func (s *stubReplayCache) Close() error {
	return nil
}

func resetReplayCacheForTest(t *testing.T) {
	t.Helper()

	globalCacheMu.Lock()
	current := replayCache
	replayCache = nil
	globalCacheSet = false
	globalCacheMu.Unlock()

	if current != nil {
		_ = current.Close()
	}
}

func TestSetReplayCacheRejectsNil(t *testing.T) {
	resetReplayCacheForTest(t)
	t.Cleanup(func() {
		resetReplayCacheForTest(t)
	})

	err := SetReplayCache(nil)
	if err == nil {
		t.Fatal("expected SetReplayCache(nil) to return an error")
	}
	if err != ErrReplayCacheNil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetReplayCacheRejectsAfterCacheInUse(t *testing.T) {
	resetReplayCacheForTest(t)
	t.Cleanup(func() {
		resetReplayCacheForTest(t)
	})

	_ = GetReplayCache(time.Second)
	err := SetReplayCache(&stubReplayCache{})
	if err == nil {
		t.Fatal("expected SetReplayCache to fail after GetReplayCache")
	}
	if err != ErrReplayCacheAlreadyInUse {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetReplayCacheBeforeUse(t *testing.T) {
	resetReplayCacheForTest(t)
	t.Cleanup(func() {
		resetReplayCacheForTest(t)
	})

	rc := &stubReplayCache{}
	if err := SetReplayCache(rc); err != nil {
		t.Fatalf("expected SetReplayCache to succeed: %v", err)
	}

	got := GetReplayCache(time.Second)
	if got != rc {
		t.Fatal("expected GetReplayCache to return configured cache")
	}
}

func TestGetReplayCacheLazyInitReturnsSameInstance(t *testing.T) {
	resetReplayCacheForTest(t)
	t.Cleanup(func() {
		resetReplayCacheForTest(t)
	})

	rc1 := GetReplayCache(time.Second)
	rc2 := GetReplayCache(2 * time.Second)

	if rc1 == nil {
		t.Fatal("expected cache to be initialized")
	}
	if rc1 != rc2 {
		t.Fatal("expected GetReplayCache to return the same singleton instance")
	}
}

func TestReplayCacheKeyEscapesAndSeparatesComponents(t *testing.T) {
	cTime := time.Unix(1, 0)

	key1 := replayCacheKey(
		types.PrincipalName{NameString: []string{"c"}},
		types.PrincipalName{NameString: []string{"a/b"}},
		cTime,
	)
	key2 := replayCacheKey(
		types.PrincipalName{NameString: []string{"b/c"}},
		types.PrincipalName{NameString: []string{"a"}},
		cTime,
	)

	if key1 == key2 {
		t.Fatalf("expected keys to differ, got %q", key1)
	}
	if strings.Count(key1, "/") != 2 {
		t.Fatalf("expected key to contain two separators, got %q", key1)
	}
	if !strings.Contains(key1, "%2F") {
		t.Fatalf("expected escaped slash in key, got %q", key1)
	}
}
