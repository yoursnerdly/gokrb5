// Package service provides server side integrations for Kerberos authentication.
package service

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/jcmturner/gokrb5/v8/rcache"
	"github.com/jcmturner/gokrb5/v8/types"
)

// Replay cache is required as specified in RFC 4120 section 3.2.3

var (
	globalCacheMu  sync.Mutex
	globalCacheSet bool

	// replayCache is the process-wide replay cache.
	replayCache rcache.Cache
)

var (
	// ErrReplayCacheAlreadyInUse indicates SetReplayCache was called after the replay cache was initialized.
	ErrReplayCacheAlreadyInUse = errors.New("service: replay cache already in use")

	// ErrReplayCacheNil indicates SetReplayCache was called with a nil cache implementation.
	ErrReplayCacheNil = errors.New("service: replay cache is nil")
)

// SetReplayCache configures the process-wide replay cache.
// It must be called before any call to GetReplayCache.
func SetReplayCache(rc rcache.Cache) error {
	globalCacheMu.Lock()
	defer globalCacheMu.Unlock()

	if rc == nil {
		return ErrReplayCacheNil
	}
	if globalCacheSet {
		return ErrReplayCacheAlreadyInUse
	}

	replayCache = rc
	globalCacheSet = true
	return nil
}

// GetReplayCache returns the process-wide replay cache.
// If SetReplayCache was not called, a default in-memory cache is lazily created.
func GetReplayCache(maxClockSkew time.Duration) rcache.Cache {
	globalCacheMu.Lock()
	defer globalCacheMu.Unlock()

	if !globalCacheSet {
		replayCache = rcache.New(maxClockSkew)
		globalCacheSet = true
	}

	return replayCache
}

// replayCacheKey derives the replay cache lookup key from Kerberos types.
func replayCacheKey(sname, cname types.PrincipalName, cTime time.Time) string {
	client := url.PathEscape(cname.PrincipalNameString())
	service := url.PathEscape(sname.PrincipalNameString())
	return fmt.Sprintf("%s/%s/%s", client, service, strconv.FormatInt(cTime.UnixNano(), 10))
}
