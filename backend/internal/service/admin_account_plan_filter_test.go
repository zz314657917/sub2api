package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type accountPlanFilterRepoStub struct {
	AccountRepository
	listData          []Account
	lastPlanType      string
	usedShareFilters  bool
	bulkUpdateIDs     []int64
	accountsByID      map[int64]*Account
	updatedAccountIDs []int64
}

func (s *accountPlanFilterRepoStub) ListWithPlanFilters(_ context.Context, params pagination.PaginationParams, _, _, _, _ string, _ int64, _, planType string) ([]Account, *pagination.PaginationResult, error) {
	s.lastPlanType = planType
	return s.listData, &pagination.PaginationResult{Total: int64(len(s.listData)), Page: params.Page, PageSize: params.PageSize}, nil
}

func (s *accountPlanFilterRepoStub) ListWithSharePlanFilters(_ context.Context, params pagination.PaginationParams, _, _, _, _ string, _ int64, _, planType string, _ *int64, _, _, _ string) ([]Account, *pagination.PaginationResult, error) {
	s.lastPlanType = planType
	s.usedShareFilters = true
	return s.listData, &pagination.PaginationResult{Total: int64(len(s.listData)), Page: params.Page, PageSize: params.PageSize}, nil
}

func (s *accountPlanFilterRepoStub) BulkUpdate(_ context.Context, ids []int64, _ AccountBulkUpdate) (int64, error) {
	s.bulkUpdateIDs = append([]int64(nil), ids...)
	return int64(len(ids)), nil
}

func (s *accountPlanFilterRepoStub) GetByID(_ context.Context, id int64) (*Account, error) {
	return s.accountsByID[id], nil
}

func (s *accountPlanFilterRepoStub) Update(_ context.Context, account *Account) error {
	s.updatedAccountIDs = append(s.updatedAccountIDs, account.ID)
	return nil
}

func TestAdminServiceListAccountsPropagatesNormalizedPlanType(t *testing.T) {
	repo := &accountPlanFilterRepoStub{listData: []Account{{ID: 7, Platform: PlatformOpenAI}}}
	svc := &adminServiceImpl{accountRepo: repo}

	accounts, total, err := svc.ListAccounts(context.Background(), 1, 20, "", "", "", "", 0, "", " K12 ", nil, "", "", "", "", "")
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	require.Equal(t, int64(1), total)
	require.Equal(t, AccountPlanTypeFilterK12, repo.lastPlanType)

	_, _, err = svc.ListAccounts(context.Background(), 1, 20, "", "", "", "", 0, "", "invalid-plan", nil, "", "", "", "", "")
	require.Error(t, err)
}

func TestAdminServiceBulkUpdatePropagatesPlanTypeFilter(t *testing.T) {
	repo := &accountPlanFilterRepoStub{listData: []Account{{ID: 7}, {ID: 11}}}
	svc := &adminServiceImpl{accountRepo: repo}
	schedulable := true

	result, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		Filters:     &BulkUpdateAccountFilters{PlanType: AccountPlanTypeFilterK12},
		Schedulable: &schedulable,
	})
	require.NoError(t, err)
	require.Equal(t, AccountPlanTypeFilterK12, repo.lastPlanType)
	require.Equal(t, []int64{7, 11}, repo.bulkUpdateIDs)
	require.Equal(t, 2, result.Success)
}

func TestAdminServiceBulkShareStatusPropagatesPlanTypeFilter(t *testing.T) {
	ownerID := int64(42)
	repo := &accountPlanFilterRepoStub{
		listData: []Account{{ID: 7}},
		accountsByID: map[int64]*Account{
			7: {ID: 7, OwnerUserID: &ownerID, ShareMode: AccountShareModePublic, ShareStatus: AccountShareStatusPendingReview},
		},
	}
	svc := &adminServiceImpl{accountRepo: repo}

	result, err := svc.BulkSetAccountShareStatus(context.Background(), nil, &BulkUpdateAccountFilters{
		PlanType:    AccountPlanTypeFilterK12,
		OwnerFilter: "user",
		ShareMode:   AccountShareModePublic,
		ShareStatus: AccountShareStatusPendingReview,
	}, AccountShareStatusActive)
	require.NoError(t, err)
	require.True(t, repo.usedShareFilters)
	require.Equal(t, AccountPlanTypeFilterK12, repo.lastPlanType)
	require.Equal(t, []int64{7}, repo.updatedAccountIDs)
	require.Equal(t, 1, result.Success)
}
