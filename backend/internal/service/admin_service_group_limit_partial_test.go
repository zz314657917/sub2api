package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type groupRepoStubForPartialLimits struct {
	GroupRepository
	group   *Group
	updated *Group
}

func (s *groupRepoStubForPartialLimits) GetByID(_ context.Context, _ int64) (*Group, error) {
	return s.group, nil
}

func (s *groupRepoStubForPartialLimits) Update(_ context.Context, group *Group) error {
	s.updated = group
	return nil
}

func TestAdminService_UpdateGroup_LimitFieldsPartialUpdate(t *testing.T) {
	newGroup := func() *Group {
		daily, weekly, monthly := 10.0, 20.0, 30.0
		return &Group{
			ID:              1,
			Name:            "existing-group",
			Platform:        PlatformOpenAI,
			Status:          StatusActive,
			DailyLimitUSD:   &daily,
			WeeklyLimitUSD:  &weekly,
			MonthlyLimitUSD: &monthly,
		}
	}

	t.Run("omitted limits are preserved", func(t *testing.T) {
		repo := &groupRepoStubForPartialLimits{group: newGroup()}
		svc := &adminServiceImpl{groupRepo: repo}

		description := "updated"
		group, err := svc.UpdateGroup(context.Background(), repo.group.ID, &UpdateGroupInput{Description: &description})
		require.NoError(t, err)
		require.Equal(t, 10.0, *group.DailyLimitUSD)
		require.Equal(t, 20.0, *group.WeeklyLimitUSD)
		require.Equal(t, 30.0, *group.MonthlyLimitUSD)
	})

	t.Run("only provided limits change", func(t *testing.T) {
		repo := &groupRepoStubForPartialLimits{group: newGroup()}
		svc := &adminServiceImpl{groupRepo: repo}

		newDaily, unlimited := 15.0, -1.0
		group, err := svc.UpdateGroup(context.Background(), repo.group.ID, &UpdateGroupInput{
			DailyLimitUSD:  &newDaily,
			WeeklyLimitUSD: &unlimited,
		})
		require.NoError(t, err)
		require.Equal(t, 15.0, *group.DailyLimitUSD)
		require.Nil(t, group.WeeklyLimitUSD)
		require.Equal(t, 30.0, *group.MonthlyLimitUSD)
	})
}

func TestAdminService_UpdateGroup_RoomManagedLimitInvariant(t *testing.T) {
	newGroup := func() *Group {
		daily, weekly, monthly := 10.0, 20.0, 30.0
		return &Group{
			ID:               1,
			Name:             "room-managed-group",
			Platform:         PlatformOpenAI,
			Status:           StatusActive,
			SubscriptionType: SubscriptionTypeSubscription,
			AccessMode:       GroupAccessModeRoomManaged,
			DailyLimitUSD:    &daily,
			WeeklyLimitUSD:   &weekly,
			MonthlyLimitUSD:  &monthly,
		}
	}

	tests := []struct {
		name  string
		input *UpdateGroupInput
	}{
		{name: "omitted limits", input: &UpdateGroupInput{}},
		func() struct {
			name  string
			input *UpdateGroupInput
		} {
			daily, weekly, monthly := 1.0, 2.0, 3.0
			return struct {
				name  string
				input *UpdateGroupInput
			}{
				name:  "provided limits",
				input: &UpdateGroupInput{DailyLimitUSD: &daily, WeeklyLimitUSD: &weekly, MonthlyLimitUSD: &monthly},
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &groupRepoStubForPartialLimits{group: newGroup()}
			svc := &adminServiceImpl{groupRepo: repo}

			group, err := svc.UpdateGroup(context.Background(), repo.group.ID, test.input)
			require.NoError(t, err)
			require.Nil(t, group.DailyLimitUSD)
			require.Nil(t, group.WeeklyLimitUSD)
			require.Nil(t, group.MonthlyLimitUSD)
		})
	}
}
