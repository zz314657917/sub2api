package service

import (
	"context"
	"log"
	"strings"
	"time"
)

const (
	leaderboardStaleRetention     = 24 * time.Hour
	leaderboardRefreshTimeout     = 30 * time.Second
	leaderboardRetryInitial       = 5 * time.Second
	leaderboardRetryMaximum       = 5 * time.Minute
	leaderboardEmptyBadgeCacheTTL = 30 * time.Second
)

type leaderboardRefresh struct {
	done    chan struct{}
	err     error
	retryAt time.Time
	backoff time.Duration
}

// cachedLeaderboard serves a stale snapshot immediately while one bounded query
// refreshes it. On a cold cache, callers wait independently: cancelling one HTTP
// request must not cancel the shared query or poison other callers' results.
// read receives an expiry cutoff, allowing a bounded stale window without
// extending the snapshot's original expiry or its reported generation time.
func cachedLeaderboard[T any](ctx context.Context, s *UsageService, key string, allowStale bool,
	read func(time.Time) (T, bool), load func(context.Context) (T, error), store func(T, time.Time),
) (T, error) {
	var zero T
	for {
		now := time.Now()
		s.leaderboardRefreshMu.Lock()
		if value, ok := read(now); ok {
			s.leaderboardRefreshMu.Unlock()
			return value, nil
		}
		stale, hasStale := zero, false
		if allowStale {
			stale, hasStale = read(now.Add(-leaderboardStaleRetention))
		}
		if err := ctx.Err(); err != nil {
			s.leaderboardRefreshMu.Unlock()
			return zero, err
		}
		if s.leaderboardRefreshes == nil {
			s.leaderboardRefreshes = make(map[string]*leaderboardRefresh)
		}
		for oldKey, old := range s.leaderboardRefreshes {
			if old.done == nil && now.After(old.retryAt.Add(leaderboardStaleRetention)) {
				delete(s.leaderboardRefreshes, oldKey)
			}
		}
		refresh := s.leaderboardRefreshes[key]
		if refresh != nil && refresh.done == nil && now.Before(refresh.retryAt) {
			err := refresh.err
			s.leaderboardRefreshMu.Unlock()
			if hasStale {
				return stale, nil
			}
			return zero, err
		}
		if refresh == nil {
			refresh = &leaderboardRefresh{}
			s.leaderboardRefreshes[key] = refresh
		}
		if refresh.done == nil {
			refresh.done = make(chan struct{})
			queryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), leaderboardRefreshTimeout)
			go func(refresh *leaderboardRefresh) {
				defer cancel()
				value, err := load(queryCtx)
				s.leaderboardRefreshMu.Lock()
				done := refresh.done
				if s.leaderboardRefreshes[key] == refresh {
					if err == nil {
						store(value, time.Now())
						delete(s.leaderboardRefreshes, key)
					} else {
						refresh.backoff *= 2
						if refresh.backoff < leaderboardRetryInitial {
							refresh.backoff = leaderboardRetryInitial
						}
						if refresh.backoff > leaderboardRetryMaximum {
							refresh.backoff = leaderboardRetryMaximum
						}
						refresh.retryAt = time.Now().Add(refresh.backoff)
						refresh.err = err
						kind, _, _ := strings.Cut(key, ":")
						log.Printf("leaderboard: %s refresh failed; retry in %s: %v", kind, refresh.backoff, err)
					}
				}
				refresh.done = nil
				close(done)
				s.leaderboardRefreshMu.Unlock()
			}(refresh)
		}
		done := refresh.done
		s.leaderboardRefreshMu.Unlock()
		if hasStale {
			return stale, nil
		}
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-done:
			// Re-read the cache: invalidation may have discarded this refresh.
		}
	}
}
