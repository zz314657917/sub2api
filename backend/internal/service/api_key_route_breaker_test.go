package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

type apiKeyRouteBreakerCacheStub struct {
	APIKeyCache
	acquire func(context.Context, APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error)
	release int
}

func TestAPIKeyRouteBreakerSkipperMemoizesHalfOpenLease(t *testing.T) {
	defaultGroupID := int64(1)
	key := &APIKey{
		ID: 4, GroupID: &defaultGroupID,
		Group:             &Group{ID: defaultGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, ModelMatchPatterns: []string{"gpt-*"}},
		MultiGroupRoutes:  []domain.APIKeyMultiGroupRoute{{GroupID: defaultGroupID, Priority: 1, Enabled: true}},
		RouteBreakerLease: nil,
	}
	cache := &apiKeyRouteBreakerCacheStub{}
	acquires := 0
	cache.acquire = func(context.Context, APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
		acquires++
		return &APIKeyRouteBreakerLease{Generation: 8, ProbeToken: 2, HalfOpen: true}, nil
	}
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
	resolved := svc.ResolveForModelRequest(context.Background(), key, "/v1/chat/completions", "", "gpt-5.6", false)
	require.NotNil(t, resolved)
	require.Equal(t, 1, acquires)
	require.NotNil(t, resolved.RouteBreakerLease)
	require.True(t, resolved.RouteBreakerLease.HalfOpen)
}

func (s *apiKeyRouteBreakerCacheStub) AcquireAPIKeyRouteBreaker(ctx context.Context, key APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
	return s.acquire(ctx, key)
}

func (s *apiKeyRouteBreakerCacheStub) IsRouteGroupCooling(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (s *apiKeyRouteBreakerCacheStub) RecordAPIKeyRouteBreakerSuccess(context.Context, APIKeyRouteBreakerLease) error {
	return nil
}

func (s *apiKeyRouteBreakerCacheStub) RecordAPIKeyRouteBreakerFailure(context.Context, APIKeyRouteBreakerLease) error {
	return nil
}

func (s *apiKeyRouteBreakerCacheStub) ReleaseAPIKeyRouteBreakerProbe(context.Context, APIKeyRouteBreakerLease) error {
	s.release++
	return nil
}

func TestAPIKeyRouteBreakerRedisFailureFailsOpen(t *testing.T) {
	defaultGroupID := int64(1)
	fallbackGroupID := int64(2)
	key := &APIKey{
		ID:      3,
		GroupID: &defaultGroupID,
		Group:   &Group{ID: defaultGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, ModelMatchPatterns: []string{"gpt-*"}},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: defaultGroupID, Priority: 1, Enabled: true},
			{GroupID: fallbackGroupID, Priority: 2, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{{ID: fallbackGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, ModelMatchPatterns: []string{"gpt-*"}}},
	}
	cache := &apiKeyRouteBreakerCacheStub{acquire: func(context.Context, APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
		return nil, errors.New("redis unavailable")
	}}
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)

	resolved := svc.ResolveForModelRequest(context.Background(), key, "/v1/chat/completions", "", "gpt-5.6", false)
	require.NotNil(t, resolved)
	require.Equal(t, defaultGroupID, *resolved.GroupID)
}

func TestAPIKeyRouteBreakerDoesNotFallbackToSkippedDefaultGroup(t *testing.T) {
	defaultGroupID := int64(1)
	key := &APIKey{
		ID:               3,
		GroupID:          &defaultGroupID,
		Group:            &Group{ID: defaultGroupID, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, ModelMatchPatterns: []string{"gpt-*"}},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: defaultGroupID, Priority: 1, Enabled: true}},
	}
	cache := &apiKeyRouteBreakerCacheStub{acquire: func(context.Context, APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
		return nil, nil
	}}
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)

	resolved := svc.ResolveForModelRequest(context.Background(), key, "/v1/chat/completions", "", "gpt-5.6", false)
	require.Nil(t, resolved)
}
