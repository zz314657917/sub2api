package service

import "context"

// ResolveForFirstResponseAttempt is the finite-plan selector used by the
// route dispatcher.  It preserves the existing priority/weight/breaker
// selection while permanently excluding groups already attempted in this
// client request.
func (s *APIKeyService) ResolveForFirstResponseAttempt(ctx context.Context, apiKey *APIKey, path, forcePlatform, requestedModel string, attempted map[int64]struct{}) *APIKey {
	if apiKey == nil {
		return nil
	}
	clone := *apiKey
	unavailable := make(map[int64]struct{}, len(apiKey.UnavailableRouteGroupIDs)+len(attempted))
	for groupID := range apiKey.UnavailableRouteGroupIDs {
		unavailable[groupID] = struct{}{}
	}
	for groupID := range attempted {
		unavailable[groupID] = struct{}{}
	}
	clone.UnavailableRouteGroupIDs = unavailable
	return s.ResolveForModelRequest(ctx, &clone, path, forcePlatform, requestedModel, false)
}

// FirstResponseTimeoutSeconds returns the route-local opt-in value. A default
// group without a route deliberately reports zero; callers use the initiating
// route's value for their finite chain.
func (k *APIKey) FirstResponseTimeoutSeconds(groupID int64) int {
	if k == nil || groupID <= 0 {
		return 0
	}
	for _, route := range k.MultiGroupRoutes {
		if route.GroupID == groupID && route.Enabled {
			return route.FirstResponseTimeoutSeconds
		}
	}
	return 0
}

func (k *APIKey) LargestFirstResponseTimeoutSeconds() int {
	if k == nil {
		return 0
	}
	largest := 0
	for _, route := range k.MultiGroupRoutes {
		if route.Enabled && route.FirstResponseTimeoutSeconds > largest {
			largest = route.FirstResponseTimeoutSeconds
		}
	}
	return largest
}
