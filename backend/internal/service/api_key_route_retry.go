package service

import (
	"context"
	"errors"
	"time"
)

// ErrAPIKeyRouteTemporarilyUnavailable means compatible routes exist, but the
// ordinary resolver could not select one within the bounded retry window.
var ErrAPIKeyRouteTemporarilyUnavailable = errors.New("API key group routes temporarily unavailable")

// ResolveForModelRequestWithRetry is the pre-forward gateway entry point.
// ResolveForModelRequest remains a single selection for its existing callers.
func (s *APIKeyService) ResolveForModelRequestWithRetry(ctx context.Context, apiKey *APIKey, path, forcePlatform, requestedModel string, imageIntent bool) (*APIKey, error) {
	return s.resolveForModelRequestWithRetry(ctx, apiKey, path, forcePlatform, requestedModel, imageIntent, waitAPIKeyRouteRetry)
}

func (s *APIKeyService) resolveForModelRequestWithRetry(ctx context.Context, apiKey *APIKey, path, forcePlatform, requestedModel string, imageIntent bool, wait func(context.Context, time.Duration) error) (*APIKey, error) {
	delays := [...]time.Duration{time.Second, 2 * time.Second, 4 * time.Second}
	for attempt := 0; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		resolved := s.resolveAPIKeyRouteRetryAttempt(ctx, apiKey, path, forcePlatform, requestedModel, imageIntent)
		if err := ctx.Err(); err != nil {
			s.ReleaseUnusedAPIKeyRouteBreakerProbe(ctx, resolved)
			return nil, err
		}
		if resolved != nil {
			return resolved, nil
		}
		if !hasEligibleAPIKeyModelRoute(apiKey, path, forcePlatform, requestedModel, imageIntent) {
			return nil, nil
		}
		if attempt == len(delays) {
			return nil, ErrAPIKeyRouteTemporarilyUnavailable
		}
		if err := wait(ctx, delays[attempt]); err != nil {
			return nil, err
		}
	}
}

// resolveAPIKeyRouteRetryAttempt reuses the ordinary pure resolver, cooldown
// checks, per-selection breaker memoization and group-state refresh. Its only
// lifecycle difference is that unused probes are released even if the request
// was cancelled during acquisition. The existing single-selection entry point
// retains its behavior for non-gateway callers.
func (s *APIKeyService) resolveAPIKeyRouteRetryAttempt(ctx context.Context, apiKey *APIKey, path, forcePlatform, requestedModel string, imageIntent bool) *APIKey {
	if apiKey == nil || apiKey.PinnedAccountID > 0 {
		return s.ResolveForModelRequest(ctx, apiKey, path, forcePlatform, requestedModel, imageIntent)
	}
	decisions := make(map[int64]bool)
	leases := make(map[int64]*APIKeyRouteBreakerLease)
	resolved := apiKey.ResolveForModelRequestWithGroupSkipper(path, forcePlatform, requestedModel, imageIntent, func(groupID int64) bool {
		if ctx.Err() != nil {
			return true
		}
		if skip, ok := decisions[groupID]; ok {
			return skip
		}
		skip := apiKey.IsRouteGroupUnavailable(groupID) ||
			(s != nil && s.isRouteGroupCooling(ctx, apiKey, groupID))
		if !skip {
			lease, allowed := s.acquireRouteBreaker(ctx, groupID, path, requestedModel, imageIntent)
			skip = !allowed
			if lease != nil {
				leases[groupID] = lease
			}
		}
		decisions[groupID] = skip
		return skip
	})
	if len(leases) > 0 {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		s.releaseUnselectedRouteBreakerProbes(cleanup, leases, resolved)
		cancel()
	}
	resolved = s.attachRouteBreakerLease(resolved, leases)
	return s.refreshResolvedAPIKeyGroupState(ctx, apiKey, resolved)
}

// hasEligibleAPIKeyModelRoute deliberately performs no cache/breaker reads or
// acquisitions. Ignoring only temporary health restrictions must still leave
// an eligible selection under the existing pure routing compatibility rules.
func hasEligibleAPIKeyModelRoute(apiKey *APIKey, path, forcePlatform, requestedModel string, imageIntent bool) bool {
	if apiKey == nil || apiKey.PinnedAccountID > 0 || len(apiKey.MultiGroupRoutes) == 0 {
		return false
	}
	groups := apiKey.routeGroupsByID()
	resolved := apiKey.ResolveForModelRequestWithGroupSkipper(path, forcePlatform, requestedModel, imageIntent, func(groupID int64) bool {
		if apiKey.IsRouteGroupUnavailable(groupID) {
			return true
		}
		group := groups[groupID]
		return group != nil && !group.IsSubscriptionType() && apiKey.User != nil &&
			!apiKey.User.CanBindGroup(groupID, group.IsExclusive)
	})
	return resolved != nil
}

func waitAPIKeyRouteRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
	}
}

// ReleaseUnusedAPIKeyRouteBreakerProbe abandons a pre-forward selection. The
// caller owns exactly-once invocation; cleanup cannot inherit cancellation or
// run indefinitely. It never records provider health or changes cooldown.
func (s *APIKeyService) ReleaseUnusedAPIKeyRouteBreakerProbe(ctx context.Context, apiKey *APIKey) {
	if apiKey == nil || apiKey.RouteBreakerLease == nil || !apiKey.RouteBreakerLease.HalfOpen {
		return
	}
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	s.ReleaseAPIKeyRouteBreakerProbe(cleanup, apiKey)
}
