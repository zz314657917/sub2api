package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// APIKeyRouteBreakerKey isolates shared route health by group, route scope,
// and the exact normalized model. ModelDigest keeps Redis keys bounded.
type APIKeyRouteBreakerKey struct {
	GroupID      int64
	RoutingScope string
	ModelDigest  string
}

// APIKeyRouteBreakerLease binds a selected request to one state generation.
// HalfOpen leases additionally include an expiring probe token.
type APIKeyRouteBreakerLease struct {
	Key        APIKeyRouteBreakerKey
	Generation int64
	ProbeToken int64
	HalfOpen   bool
}

// APIKeyRouteBreakerCache is optional so existing APIKeyCache test doubles
// and deployments continue to fail open until the Redis capability is present.
type APIKeyRouteBreakerCache interface {
	AcquireAPIKeyRouteBreaker(ctx context.Context, key APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error)
	RecordAPIKeyRouteBreakerSuccess(ctx context.Context, lease APIKeyRouteBreakerLease) error
	RecordAPIKeyRouteBreakerFailure(ctx context.Context, lease APIKeyRouteBreakerLease) error
	ReleaseAPIKeyRouteBreakerProbe(ctx context.Context, lease APIKeyRouteBreakerLease) error
}

func NewAPIKeyRouteBreakerKey(groupID int64, routingScope, requestedModel string) (APIKeyRouteBreakerKey, bool) {
	model := strings.ToLower(strings.TrimSpace(requestedModel))
	if groupID <= 0 || model == "" {
		return APIKeyRouteBreakerKey{}, false
	}
	scope := NormalizeGroupRoutingScope(routingScope, false)
	if scope == "" {
		scope = GroupRoutingScopeInference
	}
	digest := sha256.Sum256([]byte(model))
	return APIKeyRouteBreakerKey{
		GroupID:      groupID,
		RoutingScope: scope,
		ModelDigest:  hex.EncodeToString(digest[:]),
	}, true
}

func (s *APIKeyService) acquireRouteBreaker(ctx context.Context, groupID int64, path, requestedModel string, imageIntent bool) (*APIKeyRouteBreakerLease, bool) {
	if s == nil {
		return nil, true
	}
	cache, ok := s.cache.(APIKeyRouteBreakerCache)
	if !ok || cache == nil {
		return nil, true
	}
	key, ok := NewAPIKeyRouteBreakerKey(groupID, RoutingScopeForRequest(path, requestedModel, imageIntent), requestedModel)
	if !ok {
		return nil, true
	}
	lease, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	if err != nil {
		return nil, true // Redis errors must never block routing.
	}
	return lease, lease != nil
}

func (s *APIKeyService) releaseUnselectedRouteBreakerProbes(ctx context.Context, leases map[int64]*APIKeyRouteBreakerLease, selected *APIKey) {
	if s == nil || len(leases) == 0 {
		return
	}
	cache, ok := s.cache.(APIKeyRouteBreakerCache)
	if !ok || cache == nil {
		return
	}
	selectedGroupID := int64(0)
	if selected != nil && selected.GroupID != nil {
		selectedGroupID = *selected.GroupID
	}
	for groupID, lease := range leases {
		if lease == nil || !lease.HalfOpen || groupID == selectedGroupID {
			continue
		}
		_ = cache.ReleaseAPIKeyRouteBreakerProbe(ctx, *lease)
	}
}

func (s *APIKeyService) attachRouteBreakerLease(apiKey *APIKey, leases map[int64]*APIKeyRouteBreakerLease) *APIKey {
	if apiKey == nil || apiKey.GroupID == nil {
		return apiKey
	}
	return apiKey.WithRouteBreakerLease(leases[*apiKey.GroupID])
}

func (s *APIKeyService) RecordAPIKeyRouteBreakerSuccess(ctx context.Context, apiKey *APIKey) {
	if s == nil || apiKey == nil || apiKey.RouteBreakerLease == nil {
		return
	}
	if cache, ok := s.cache.(APIKeyRouteBreakerCache); ok && cache != nil {
		_ = cache.RecordAPIKeyRouteBreakerSuccess(ctx, *apiKey.RouteBreakerLease)
	}
}

func (s *APIKeyService) RecordAPIKeyRouteBreakerFailure(ctx context.Context, apiKey *APIKey) {
	if s == nil || apiKey == nil || apiKey.RouteBreakerLease == nil {
		return
	}
	if cache, ok := s.cache.(APIKeyRouteBreakerCache); ok && cache != nil {
		_ = cache.RecordAPIKeyRouteBreakerFailure(ctx, *apiKey.RouteBreakerLease)
	}
}

func (s *APIKeyService) ReleaseAPIKeyRouteBreakerProbe(ctx context.Context, apiKey *APIKey) {
	if s == nil || apiKey == nil || apiKey.RouteBreakerLease == nil || !apiKey.RouteBreakerLease.HalfOpen {
		return
	}
	if cache, ok := s.cache.(APIKeyRouteBreakerCache); ok && cache != nil {
		_ = cache.ReleaseAPIKeyRouteBreakerProbe(ctx, *apiKey.RouteBreakerLease)
	}
}
