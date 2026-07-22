package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

type s99GroupRepo struct {
	GroupRepository
	groups map[int64]*Group
}

func (r *s99GroupRepo) GetByID(_ context.Context, id int64) (*Group, error) {
	if group := r.groups[id]; group != nil {
		return group, nil
	}
	return nil, ErrGroupNotFound
}

type s99UserSubRepo struct {
	UserSubscriptionRepository
	err error
}

func (r *s99UserSubRepo) GetActiveByUserIDAndGroupID(context.Context, int64, int64) (*UserSubscription, error) {
	return nil, r.err
}

func TestS99ResolveForModelRequestSkipsUnavailableHigherPriorityRoute(t *testing.T) {
	primaryID := int64(11)
	fallbackID := int64(22)
	key := &APIKey{
		GroupID: &primaryID,
		Group: &Group{
			ID: primaryID, Status: StatusActive, Hydrated: true,
			Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"gpt-*"},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: primaryID, Priority: 1, Weight: 1, Enabled: true},
			{GroupID: fallbackID, Priority: 2, Weight: 1, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{{
			ID: fallbackID, Status: StatusActive, Hydrated: true,
			Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"gpt-*"},
		}},
		UnavailableRouteGroupIDs: map[int64]struct{}{primaryID: {}},
	}

	resolved := (&APIKeyService{}).ResolveForModelRequest(context.Background(), key, "/v1/responses", "", "gpt-5.4", false)

	require.NotNil(t, resolved)
	require.NotNil(t, resolved.GroupID)
	require.Equal(t, fallbackID, *resolved.GroupID)
}

func TestS99ResolveForRequestFailsClosedWhenEveryEnabledRouteIsUnavailable(t *testing.T) {
	groupID := int64(11)
	key := &APIKey{
		GroupID: &groupID,
		Group: &Group{
			ID: groupID, Status: StatusActive, Hydrated: true,
			Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID: groupID, Priority: 1, Weight: 1, Enabled: true,
		}},
		UnavailableRouteGroupIDs: map[int64]struct{}{groupID: {}},
	}

	require.Nil(t, (&APIKeyService{}).ResolveForRequest(context.Background(), key, "/v1/models", ""))
}

func TestS99ModelCatalogCanUseAvailableRouteFromAnotherScope(t *testing.T) {
	expiredGroupID := int64(11)
	imageGroupID := int64(22)
	key := &APIKey{
		GroupID: &expiredGroupID,
		Group: &Group{
			ID: expiredGroupID, Status: StatusActive, Hydrated: true,
			Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeInference,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: expiredGroupID, Priority: 1, Weight: 1, Enabled: true},
			{GroupID: imageGroupID, Priority: 2, Weight: 1, Enabled: true},
		},
		MultiGroupRouteGroups: []*Group{{
			ID: imageGroupID, Status: StatusActive, Hydrated: true,
			Platform: PlatformOpenAI, RoutingScope: GroupRoutingScopeImage,
			AllowImageGeneration: true,
		}},
		UnavailableRouteGroupIDs: map[int64]struct{}{expiredGroupID: {}},
	}

	resolved := (&APIKeyService{}).ResolveForRequest(context.Background(), key, "/v1/models", "")

	require.NotNil(t, resolved)
	require.NotNil(t, resolved.GroupID)
	require.Equal(t, imageGroupID, *resolved.GroupID)
}

func TestS99UpdatePreservesExistingUnavailableBindings(t *testing.T) {
	baseGroupID := int64(10)
	routeGroupID := int64(20)
	repo := &s87APIKeyRepo{key: &APIKey{
		ID: 9, UserID: 7, Key: "s99-key", Status: StatusActive,
		GroupID: &baseGroupID,
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID: routeGroupID, Priority: 1, Weight: 1, Enabled: true,
		}},
	}}
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   &s91APIKeyUserRepo{user: &User{ID: 7, Status: StatusActive}},
	}

	updated, err := svc.Update(context.Background(), 9, 7, UpdateAPIKeyRequest{
		GroupID: &baseGroupID,
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID: routeGroupID, Priority: 1, Weight: 1, Enabled: true,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, routeGroupID, updated.MultiGroupRoutes[0].GroupID)
}

func TestS99UpdateRejectsNewUnavailableSubscriptionRoute(t *testing.T) {
	existingGroupID := int64(20)
	newGroupID := int64(30)
	repo := &s87APIKeyRepo{key: &APIKey{
		ID: 9, UserID: 7, Key: "s99-key", Status: StatusActive,
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID: existingGroupID, Priority: 1, Weight: 1, Enabled: true,
		}},
	}}
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   &s91APIKeyUserRepo{user: &User{ID: 7, Status: StatusActive}},
		groupRepo: &s99GroupRepo{groups: map[int64]*Group{
			newGroupID: {ID: newGroupID, SubscriptionType: SubscriptionTypeSubscription},
		}},
		userSubRepo: &s99UserSubRepo{err: ErrSubscriptionNotFound},
	}

	_, err := svc.Update(context.Background(), 9, 7, UpdateAPIKeyRequest{
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: existingGroupID, Priority: 1, Weight: 1, Enabled: true},
			{GroupID: newGroupID, Priority: 2, Weight: 1, Enabled: true},
		},
	})

	require.ErrorIs(t, err, ErrGroupNotAllowed)
	require.Equal(t, 0, repo.updateCalls)
}
