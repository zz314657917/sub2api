//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type studioBridgeGatewayAPIKeyRepoStub struct {
	authRepoStub
	keys []APIKey
}

func (s *studioBridgeGatewayAPIKeyRepoStub) ListByUserID(_ context.Context, userID int64, params pagination.PaginationParams, _ APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	out := make([]APIKey, 0, len(s.keys))
	for _, key := range s.keys {
		if key.UserID == userID {
			out = append(out, key)
		}
	}
	if params.PageSize > 0 && len(out) > params.PageSize {
		out = out[:params.PageSize]
	}
	return out, &pagination.PaginationResult{Total: int64(len(out)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (s *studioBridgeGatewayAPIKeyRepoStub) CountByUserID(_ context.Context, userID int64) (int64, error) {
	count := int64(0)
	for _, key := range s.keys {
		if key.UserID == userID {
			count++
		}
	}
	return count, nil
}

func TestAPIKeyService_BuildStudioBridgeGatewayAPIKey_UsesDefaultKeyGroup(t *testing.T) {
	user := &User{ID: 7, Email: "u@example.com", Status: StatusActive}
	defaultGroupID := int64(10)
	repo := &studioBridgeGatewayAPIKeyRepoStub{keys: []APIKey{
		{
			ID:                  1,
			UserID:              7,
			Key:                 "sk-default",
			Name:                DefaultAPIKeyName,
			Status:              StatusActive,
			GroupID:             &defaultGroupID,
			Group:               &Group{ID: defaultGroupID, Name: "default", Status: StatusActive, Platform: PlatformOpenAI, Hydrated: true},
			AccountPoolStrategy: AccountPoolStrategyPrivateFirst,
		},
	}}
	svc := NewAPIKeyService(repo, &userRepoStub{user: user}, &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		10: {ID: 10, Name: "default", Status: StatusActive, Platform: PlatformOpenAI, Hydrated: true},
		99: {ID: 99, Name: "fallback", Status: StatusActive, Platform: PlatformAnthropic, Hydrated: true},
	}}, nil, nil, nil, &config.Config{})

	apiKey, err := svc.BuildStudioBridgeGatewayAPIKey(context.Background(), 7, 99, "/v1/chat/completions")

	require.NoError(t, err)
	require.Equal(t, int64(1), apiKey.ID)
	require.Equal(t, "sk-default", apiKey.Key)
	require.NotNil(t, apiKey.GroupID)
	require.Equal(t, int64(10), *apiKey.GroupID)
	require.Equal(t, AccountPoolStrategyPrivateFirst, apiKey.AccountPoolStrategy)
}

func TestAPIKeyService_BuildStudioBridgeGatewayAPIKey_UsesFallbackGroupWhenDefaultUngrouped(t *testing.T) {
	user := &User{ID: 7, Email: "u@example.com", Status: StatusActive}
	repo := &studioBridgeGatewayAPIKeyRepoStub{keys: []APIKey{
		{
			ID:                  1,
			UserID:              7,
			Key:                 "sk-default",
			Name:                DefaultAPIKeyName,
			Status:              StatusActive,
			AccountPoolStrategy: AccountPoolStrategySharedOnly,
		},
	}}
	svc := NewAPIKeyService(repo, &userRepoStub{user: user}, &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		99: {ID: 99, Name: "fallback", Status: StatusActive, Platform: PlatformOpenAI, Hydrated: true},
	}}, nil, nil, nil, &config.Config{})

	apiKey, err := svc.BuildStudioBridgeGatewayAPIKey(context.Background(), 7, 99, "/v1/chat/completions")

	require.NoError(t, err)
	require.Equal(t, int64(1), apiKey.ID)
	require.NotNil(t, apiKey.GroupID)
	require.Equal(t, int64(99), *apiKey.GroupID)
	require.Equal(t, "fallback", apiKey.Group.Name)
}

func TestAPIKeyService_BuildStudioBridgeGatewayAPIKey_RoutesDefaultKeyByModel(t *testing.T) {
	user := &User{ID: 7, Email: "u@example.com", Status: StatusActive}
	defaultGroupID := int64(10)
	repo := &studioBridgeGatewayAPIKeyRepoStub{keys: []APIKey{
		{
			ID:      1,
			UserID:  7,
			Key:     "sk-default",
			Name:    DefaultAPIKeyName,
			Status:  StatusActive,
			GroupID: &defaultGroupID,
			Group:   &Group{ID: defaultGroupID, Name: "default", Status: StatusActive, Platform: PlatformAnthropic, RoutingScope: GroupRoutingScopeInference, Hydrated: true},
			MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
				{GroupID: 20, Enabled: true, Priority: 1, TextOnly: true, ModelPatterns: []string{"gpt-*"}},
			},
			MultiGroupRouteGroups: []*Group{{ID: 20, Name: "openai", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference, Hydrated: true}},
			AccountPoolStrategy:   AccountPoolStrategySharedOnly,
		},
	}}
	svc := NewAPIKeyService(repo, &userRepoStub{user: user}, &groupRepoStubForStudioBridgeGateway{groups: map[int64]*Group{
		10: {ID: 10, Name: "default", Status: StatusActive, Platform: PlatformAnthropic, RoutingScope: GroupRoutingScopeInference, Hydrated: true},
		20: {ID: 20, Name: "openai", Status: StatusActive, Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference, Hydrated: true},
	}}, nil, nil, nil, &config.Config{})

	apiKey, err := svc.BuildStudioBridgeGatewayAPIKey(context.Background(), 7, 99, "/v1/chat/completions")
	require.NoError(t, err)
	require.NotNil(t, apiKey.GroupID)
	require.Equal(t, int64(20), *apiKey.GroupID)

	routed := svc.ResolveForModelRequest(context.Background(), apiKey, "/v1/chat/completions", "", "gpt-5.5", false)
	require.NotNil(t, routed.GroupID)
	require.Equal(t, int64(20), *routed.GroupID)
	require.Equal(t, "openai", routed.Group.Name)
}

type groupRepoStubForStudioBridgeGateway struct {
	groupRepoStubForGroupUpdate
	groups map[int64]*Group
}

func (s *groupRepoStubForStudioBridgeGateway) GetByID(_ context.Context, id int64) (*Group, error) {
	group := s.groups[id]
	if group == nil {
		return nil, ErrGroupNotFound
	}
	clone := *group
	return &clone, nil
}
