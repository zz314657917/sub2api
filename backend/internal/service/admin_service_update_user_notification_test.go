//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type userGroupRateRepoStubForUpdateUserNotification struct {
	byUser map[int64]map[int64]float64
	synced map[int64]map[int64]*float64
}

func (s *userGroupRateRepoStubForUpdateUserNotification) GetByUserID(_ context.Context, userID int64) (map[int64]float64, error) {
	out := make(map[int64]float64)
	for groupID, rate := range s.byUser[userID] {
		out[groupID] = rate
	}
	return out, nil
}

func (s *userGroupRateRepoStubForUpdateUserNotification) GetByUserAndGroup(_ context.Context, _, _ int64) (*float64, error) {
	panic("unexpected GetByUserAndGroup call")
}

func (s *userGroupRateRepoStubForUpdateUserNotification) GetRPMOverrideByUserAndGroup(_ context.Context, _, _ int64) (*int, error) {
	panic("unexpected GetRPMOverrideByUserAndGroup call")
}

func (s *userGroupRateRepoStubForUpdateUserNotification) GetByGroupID(_ context.Context, _ int64) ([]UserGroupRateEntry, error) {
	panic("unexpected GetByGroupID call")
}

func (s *userGroupRateRepoStubForUpdateUserNotification) SyncUserGroupRates(_ context.Context, userID int64, rates map[int64]*float64) error {
	if s.synced == nil {
		s.synced = make(map[int64]map[int64]*float64)
	}
	s.synced[userID] = rates
	return nil
}

func (s *userGroupRateRepoStubForUpdateUserNotification) SyncGroupRateMultipliers(_ context.Context, _ int64, _ []GroupRateMultiplierInput) error {
	panic("unexpected SyncGroupRateMultipliers call")
}

func (s *userGroupRateRepoStubForUpdateUserNotification) SyncGroupRPMOverrides(_ context.Context, _ int64, _ []GroupRPMOverrideInput) error {
	panic("unexpected SyncGroupRPMOverrides call")
}

func (s *userGroupRateRepoStubForUpdateUserNotification) ClearGroupRPMOverrides(_ context.Context, _ int64) error {
	panic("unexpected ClearGroupRPMOverrides call")
}

func (s *userGroupRateRepoStubForUpdateUserNotification) DeleteByGroupID(_ context.Context, _ int64) error {
	panic("unexpected DeleteByGroupID call")
}

func (s *userGroupRateRepoStubForUpdateUserNotification) DeleteByUserID(_ context.Context, _ int64) error {
	panic("unexpected DeleteByUserID call")
}

func TestAdminService_UpdateUser_SameGroupConfigDoesNotNotify(t *testing.T) {
	ticketRepo := newFakeTicketRepo()
	userRepo := &userRepoStub{user: &User{
		ID:            42,
		Email:         "user@example.com",
		AllowedGroups: []int64{1, 2},
		RPMLimit:      60,
	}}
	groupRateRepo := &userGroupRateRepoStubForUpdateUserNotification{
		byUser: map[int64]map[int64]float64{
			42: {7: 0.08},
		},
	}
	svc := &adminServiceImpl{
		userRepo:          userRepo,
		userGroupRateRepo: groupRateRepo,
		redeemCodeRepo:    &redeemRepoStub{},
	}
	svc.SetSystemTicketService(NewSystemTicketService(ticketRepo))

	rate := 0.08
	sameGroupsDifferentOrder := []int64{2, 1}
	sameRPM := 60
	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{
		AllowedGroups: &sameGroupsDifferentOrder,
		GroupRates: map[int64]*float64{
			7: &rate,
		},
		RPMLimit: &sameRPM,
	})

	require.NoError(t, err)
	require.Equal(t, sameGroupsDifferentOrder, updated.AllowedGroups)
	require.Zero(t, ticketRepo.systemByUser[42], "原样保存可用分组、专属倍率和 RPM 时不应创建系统通知")
}

func TestAdminService_UpdateUser_AllowedGroupChangeNotifiesExplicitly(t *testing.T) {
	ticketRepo := newFakeTicketRepo()
	userRepo := &userRepoStub{user: &User{
		ID:            42,
		Email:         "user@example.com",
		AllowedGroups: []int64{1},
		RPMLimit:      60,
	}}
	groupRateRepo := &userGroupRateRepoStubForUpdateUserNotification{
		byUser: map[int64]map[int64]float64{
			42: {7: 0.08},
		},
	}
	svc := &adminServiceImpl{
		userRepo:          userRepo,
		userGroupRateRepo: groupRateRepo,
		redeemCodeRepo:    &redeemRepoStub{},
	}
	svc.SetSystemTicketService(NewSystemTicketService(ticketRepo))

	rate := 0.08
	newGroups := []int64{1, 2}
	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{
		AllowedGroups: &newGroups,
		GroupRates: map[int64]*float64{
			7: &rate,
		},
	})

	require.NoError(t, err)
	got := requireSystemTicketNotification(t, ticketRepo, 42, SystemTicketEventGroupChanged, "")
	require.Contains(t, got.Message.Content, "可用分组：增加 #2")
	require.NotContains(t, got.Message.Content, "专属倍率已更新为 0.08x")
}
