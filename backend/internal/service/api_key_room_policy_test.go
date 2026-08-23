package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyServiceRejectsRoomManagedGroupsForOrdinaryPaths(t *testing.T) {
	ctx := context.Background()
	roomGroup := &Group{
		ID:               17,
		AccessMode:       GroupAccessModeRoomManaged,
		SubscriptionType: SubscriptionTypeSubscription,
		Status:           StatusActive,
	}

	require.ErrorIs(t, validateOrdinaryAPIKeyGroup(roomGroup), ErrGroupNotAllowed)
	svc := &APIKeyService{}
	require.False(t, svc.canUserBindGroup(ctx, &User{ID: 42}, roomGroup))
	require.False(t, svc.canUserBindGroupInternal(&User{ID: 42}, roomGroup, map[int64]bool{roomGroup.ID: true}))

	normalGroup := &Group{ID: 18, Status: StatusActive}
	require.NoError(t, validateOrdinaryAPIKeyGroup(normalGroup))
	require.True(t, svc.canUserBindGroup(ctx, &User{ID: 42}, normalGroup))
}

func TestAPIKeyServiceRejectsRoomManagedRouteWhenPermissionCheckIsSkipped(t *testing.T) {
	ctx := context.Background()
	roomGroup := &Group{ID: 27, AccessMode: GroupAccessModeRoomManaged}
	svc := &APIKeyService{groupRepo: &roomPolicyGroupRepoStub{group: roomGroup}}
	route := []domain.APIKeyMultiGroupRoute{{GroupID: roomGroup.ID}}

	require.ErrorIs(t, svc.validateAPIKeyRouteGroups(ctx, nil, route, true, nil), ErrGroupNotAllowed)
}

func TestAPIKeyServiceCreateRejectsRoomManagedGroupWhenPermissionCheckIsSkipped(t *testing.T) {
	ctx := context.Background()
	roomGroup := &Group{ID: 37, AccessMode: GroupAccessModeRoomManaged}
	groupID := roomGroup.ID
	svc := &APIKeyService{
		apiKeyRepo: &roomPolicyAPIKeyRepoStub{},
		userRepo:   &s91APIKeyUserRepo{user: &User{ID: 42, Status: StatusActive}},
		groupRepo:  &roomPolicyGroupRepoStub{group: roomGroup},
	}

	_, err := svc.Create(ctx, 42, CreateAPIKeyRequest{
		GroupID:                  &groupID,
		AccountPoolStrategy:      AccountPoolStrategySharedOnly,
		SkipGroupPermissionCheck: true,
	})
	require.ErrorIs(t, err, ErrGroupNotAllowed)
}

func TestAPIKeyServiceUpdateRejectsLoadedRoomManagedGroupWhenExistingBindingIsUnhydrated(t *testing.T) {
	ctx := context.Background()
	groupID := int64(47)
	svc := &APIKeyService{
		apiKeyRepo: &roomPolicyAPIKeyRepoStub{key: &APIKey{ID: 9, UserID: 42, GroupID: &groupID}},
		userRepo:   &s91APIKeyUserRepo{user: &User{ID: 42, Status: StatusActive}},
		groupRepo:  &roomPolicyGroupRepoStub{group: &Group{ID: groupID, AccessMode: GroupAccessModeRoomManaged}},
	}

	_, err := svc.Update(ctx, 9, 42, UpdateAPIKeyRequest{GroupID: &groupID})
	require.ErrorIs(t, err, ErrGroupNotAllowed)
}

func TestAPIKeyServiceUpdateRejectsPreservedRoomManagedRoute(t *testing.T) {
	ctx := context.Background()
	groupID := int64(57)
	svc := &APIKeyService{
		apiKeyRepo: &roomPolicyAPIKeyRepoStub{key: &APIKey{
			ID: 9, UserID: 42,
			MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: groupID}},
		}},
		userRepo:  &s91APIKeyUserRepo{user: &User{ID: 42, Status: StatusActive}},
		groupRepo: &roomPolicyGroupRepoStub{group: &Group{ID: groupID, AccessMode: GroupAccessModeRoomManaged}},
	}

	_, err := svc.Update(ctx, 9, 42, UpdateAPIKeyRequest{
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: groupID}},
	})
	require.ErrorIs(t, err, ErrGroupNotAllowed)
}

type roomPolicyGroupRepoStub struct {
	GroupRepository
	group *Group
}

func (s *roomPolicyGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	clone := *s.group
	return &clone, nil
}

type roomPolicyAPIKeyRepoStub struct {
	APIKeyRepository
	key *APIKey
}

func (s *roomPolicyAPIKeyRepoStub) GetByID(context.Context, int64) (*APIKey, error) {
	clone := *s.key
	return &clone, nil
}
