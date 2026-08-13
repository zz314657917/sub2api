package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type groupPlatformCacheRepoStub struct {
	AdminGroupRepository
	group      *Group
	updateErr  error
	getLiteErr error
	events     *[]string
}

func (s *groupPlatformCacheRepoStub) GetByID(_ context.Context, _ int64) (*Group, error) {
	return s.group, nil
}

func (s *groupPlatformCacheRepoStub) Update(_ context.Context, _ *Group) error {
	*s.events = append(*s.events, "update")
	return s.updateErr
}

func (s *groupPlatformCacheRepoStub) GetByIDLite(_ context.Context, _ int64) (*Group, error) {
	*s.events = append(*s.events, "copy")
	return nil, s.getLiteErr
}

type channelCacheInvalidatorStub struct {
	events *[]string
	calls  int
}

func (s *channelCacheInvalidatorStub) InvalidateCache() {
	s.calls++
	*s.events = append(*s.events, "invalidate")
}

func newAdminServiceForGroupPlatformCacheTest(group *Group, events *[]string) (*adminServiceImpl, *groupPlatformCacheRepoStub, *channelCacheInvalidatorStub) {
	repo := &groupPlatformCacheRepoStub{group: group, events: events}
	cache := &channelCacheInvalidatorStub{events: events}
	return &adminServiceImpl{groupRepo: repo, channelCacheInvalidator: cache}, repo, cache
}

func TestAdminServiceUpdateGroupChannelCacheSuccessChanged(t *testing.T) {
	events := make([]string, 0, 2)
	svc, _, cache := newAdminServiceForGroupPlatformCacheTest(&Group{ID: 1, Platform: PlatformAnthropic, SubscriptionType: SubscriptionTypeStandard}, &events)

	_, err := svc.UpdateGroup(context.Background(), 1, &UpdateGroupInput{Platform: PlatformOpenAI})

	require.NoError(t, err)
	require.Equal(t, 1, cache.calls)
	require.Equal(t, []string{"update", "invalidate"}, events)
}

func TestAdminServiceUpdateGroupChannelCacheSameOrOmitted(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input *UpdateGroupInput
	}{
		{name: "same", input: &UpdateGroupInput{Platform: PlatformAnthropic}},
		{name: "omitted", input: &UpdateGroupInput{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := make([]string, 0, 1)
			svc, _, cache := newAdminServiceForGroupPlatformCacheTest(&Group{ID: 1, Platform: PlatformAnthropic, SubscriptionType: SubscriptionTypeStandard}, &events)

			_, err := svc.UpdateGroup(context.Background(), 1, tc.input)

			require.NoError(t, err)
			require.Zero(t, cache.calls)
			require.Equal(t, []string{"update"}, events)
		})
	}
}

func TestAdminServiceUpdateGroupChannelCachePersistenceFailure(t *testing.T) {
	events := make([]string, 0, 1)
	svc, repo, cache := newAdminServiceForGroupPlatformCacheTest(&Group{ID: 1, Platform: PlatformAnthropic, SubscriptionType: SubscriptionTypeStandard}, &events)
	repo.updateErr = errors.New("update failed")

	_, err := svc.UpdateGroup(context.Background(), 1, &UpdateGroupInput{Platform: PlatformOpenAI})

	require.EqualError(t, err, "update failed")
	require.Zero(t, cache.calls)
	require.Equal(t, []string{"update"}, events)
}

func TestAdminServiceUpdateGroupChannelCacheInvalidatesBeforeCopyFailure(t *testing.T) {
	events := make([]string, 0, 3)
	svc, repo, cache := newAdminServiceForGroupPlatformCacheTest(&Group{ID: 1, Platform: PlatformAnthropic, SubscriptionType: SubscriptionTypeStandard}, &events)
	repo.getLiteErr = errors.New("source group missing")

	_, err := svc.UpdateGroup(context.Background(), 1, &UpdateGroupInput{
		Platform:                 PlatformOpenAI,
		CopyAccountsFromGroupIDs: []int64{2},
	})

	require.EqualError(t, err, "source group 2 not found: source group missing")
	require.Equal(t, 1, cache.calls)
	require.Equal(t, []string{"update", "invalidate", "copy"}, events)
}
