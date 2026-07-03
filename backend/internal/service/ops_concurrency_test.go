package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type opsStatsAccountRepoStub struct {
	AccountRepository

	accounts []Account

	statsCalls int
	listCalls  int

	gotPlatform string
	gotGroupID  *int64
}

func (r *opsStatsAccountRepoStub) ListOpsAccountsForStats(ctx context.Context, platformFilter string, groupIDFilter *int64) ([]Account, error) {
	r.statsCalls++
	r.gotPlatform = platformFilter
	if groupIDFilter != nil {
		id := *groupIDFilter
		r.gotGroupID = &id
	} else {
		r.gotGroupID = nil
	}
	return r.accounts, nil
}

type opsFallbackAccountRepoStub struct {
	AccountRepository

	accounts []Account

	listCalls    int
	gotPlatform  string
	gotGroupID   int64
	gotPageSize  int
	gotPageCount int
}

func (r *opsFallbackAccountRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, accountType, status, search string, groupID int64, privacyMode string) ([]Account, *pagination.PaginationResult, error) {
	r.listCalls++
	r.gotPlatform = platform
	r.gotGroupID = groupID
	r.gotPageSize = params.PageSize
	r.gotPageCount = params.Page
	return r.accounts, &pagination.PaginationResult{
		Total:    int64(len(r.accounts)),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    1,
	}, nil
}

func TestListAllAccountsForOpsUsesStatsRepositoryWithGroupFilter(t *testing.T) {
	groupID := int64(42)
	repo := &opsStatsAccountRepoStub{
		accounts: []Account{{ID: 1, Name: "stats-account", Platform: PlatformOpenAI}},
	}
	svc := NewOpsService(nil, nil, nil, repo, nil, nil, nil, nil, nil, nil, nil)

	accounts, err := svc.listAllAccountsForOps(context.Background(), PlatformOpenAI, &groupID)
	require.NoError(t, err)
	require.Equal(t, []Account{{ID: 1, Name: "stats-account", Platform: PlatformOpenAI}}, accounts)
	require.Equal(t, 1, repo.statsCalls)
	require.Equal(t, PlatformOpenAI, repo.gotPlatform)
	require.NotNil(t, repo.gotGroupID)
	require.Equal(t, groupID, *repo.gotGroupID)
}

func TestListAllAccountsForOpsFallbackPassesGroupFilter(t *testing.T) {
	groupID := int64(7)
	repo := &opsFallbackAccountRepoStub{
		accounts: []Account{{ID: 2, Name: "fallback-account", Platform: PlatformAnthropic}},
	}
	svc := NewOpsService(nil, nil, nil, repo, nil, nil, nil, nil, nil, nil, nil)

	accounts, err := svc.listAllAccountsForOps(context.Background(), PlatformAnthropic, &groupID)
	require.NoError(t, err)
	require.Equal(t, []Account{{ID: 2, Name: "fallback-account", Platform: PlatformAnthropic}}, accounts)
	require.Equal(t, 1, repo.listCalls)
	require.Equal(t, PlatformAnthropic, repo.gotPlatform)
	require.Equal(t, groupID, repo.gotGroupID)
	require.Equal(t, opsAccountsPageSize, repo.gotPageSize)
	require.Equal(t, 1, repo.gotPageCount)
}
