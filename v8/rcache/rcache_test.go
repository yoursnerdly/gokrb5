package rcache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemCacheAddAndDuplicate(t *testing.T) {
	t.Parallel()

	c := New(time.Second)
	t.Cleanup(func() {
		_ = c.Close()
	})

	if err := c.Add(context.Background(), "key", time.Second); err != nil {
		t.Fatalf("expected first add to succeed: %v", err)
	}

	err := c.Add(context.Background(), "key", time.Second)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got: %v", err)
	}

	if err = c.Add(context.Background(), "other-key", time.Second); err != nil {
		t.Fatalf("expected different key add to succeed: %v", err)
	}
}

func TestMemCacheExpiredEntryCanBeAddedAgain(t *testing.T) {
	t.Parallel()

	c := New(time.Hour)
	t.Cleanup(func() {
		_ = c.Close()
	})

	if err := c.Add(context.Background(), "key", 10*time.Millisecond); err != nil {
		t.Fatalf("expected first add to succeed: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if err := c.Add(context.Background(), "key", time.Second); err != nil {
		t.Fatalf("expected add to succeed after expiry: %v", err)
	}
}

func TestMemCacheCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	c := New(10 * time.Millisecond)

	if err := c.Close(); err != nil {
		t.Fatalf("expected first close to succeed: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("expected second close to succeed: %v", err)
	}

	select {
	case <-c.done:
		// expected
	default:
		t.Fatal("expected done channel to be closed")
	}
}
