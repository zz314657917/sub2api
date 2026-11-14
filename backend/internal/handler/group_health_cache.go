package handler

import (
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"golang.org/x/sync/singleflight"
)

const groupHealthCacheTTL = time.Hour
const groupHealthCacheMaxEntries = 64

type groupHealthCacheEntry struct {
	groups    []usagestats.GroupStat
	expiresAt time.Time
}

// Shared by all users of this handler instance. Date ranges are normalized to
// UTC so equivalent windows share an entry. Process restarts clear the cache.
type groupHealthCache struct {
	mu    sync.Mutex
	items map[string]groupHealthCacheEntry
	sf    singleflight.Group
}

func (c *groupHealthCache) get(key string) ([]usagestats.GroupStat, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.items[key]
	if !ok || !time.Now().Before(entry.expiresAt) {
		delete(c.items, key)
		return nil, false
	}
	return entry.groups, true
}

func (c *groupHealthCache) getOrLoad(start, end time.Time, load func() ([]usagestats.GroupStat, error)) ([]usagestats.GroupStat, error) {
	key := start.UTC().Format(time.RFC3339Nano) + "/" + end.UTC().Format(time.RFC3339Nano)
	if groups, ok := c.get(key); ok {
		return groups, nil
	}
	value, err, _ := c.sf.Do(key, func() (any, error) {
		if groups, ok := c.get(key); ok {
			return groups, nil
		}
		groups, err := load()
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.items == nil {
			c.items = make(map[string]groupHealthCacheEntry)
		}
		now := time.Now()
		var oldestKey string
		var oldestExpiry time.Time
		for k, entry := range c.items {
			if !now.Before(entry.expiresAt) {
				delete(c.items, k)
				continue
			}
			if oldestKey == "" || entry.expiresAt.Before(oldestExpiry) {
				oldestKey, oldestExpiry = k, entry.expiresAt
			}
		}
		if len(c.items) >= groupHealthCacheMaxEntries {
			delete(c.items, oldestKey)
		}
		c.items[key] = groupHealthCacheEntry{groups: groups, expiresAt: now.Add(groupHealthCacheTTL)}
		return groups, nil
	})
	if err != nil {
		return nil, err
	}
	return value.([]usagestats.GroupStat), nil
}
