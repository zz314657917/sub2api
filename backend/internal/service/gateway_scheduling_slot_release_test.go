//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type slotReleaseHydrationCache struct {
	*snapshotHydrationCache
	err error
}

func (c *slotReleaseHydrationCache) GetAccount(ctx context.Context, id int64) (*Account, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.snapshotHydrationCache.GetAccount(ctx, id)
}

type slotReleaseAccountRepo struct {
	AccountRepository
	account *Account
	err     error
	calls   int
}

func (r *slotReleaseAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	if r.account != nil && r.account.ID == id {
		return r.account, nil
	}
	return nil, nil
}

func TestGatewayServiceNewSelectionResultHydrationFailureReleasesSlot(t *testing.T) {
	fallbackErr := errors.New("injected repository error")
	for _, tc := range []struct {
		name     string
		cacheErr error
		repoErr  error
	}{
		{name: "missing account"},
		{name: "repository error", repoErr: fallbackErr},
		{name: "cache error with missing fallback", cacheErr: errors.New("injected cache error")},
		{name: "cache error with repository error", cacheErr: errors.New("injected cache error"), repoErr: fallbackErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := &slotReleaseHydrationCache{snapshotHydrationCache: &snapshotHydrationCache{}, err: tc.cacheErr}
			repo := &slotReleaseAccountRepo{err: tc.repoErr}
			svc := &GatewayService{schedulerSnapshot: NewSchedulerSnapshotService(cache, nil, repo, nil, nil)}
			released := 0
			result, err := svc.newSelectionResult(context.Background(), &Account{ID: 9}, true, func() { released++ }, nil)
			require.Error(t, err)
			require.Nil(t, result)
			require.Equal(t, 1, repo.calls)
			require.Equal(t, 1, released)
		})
	}
}

func TestGatewayServiceNewSelectionResultSuccessfulHydrationRetainsOwnership(t *testing.T) {
	for _, tc := range []struct {
		name      string
		cacheHit  bool
		cacheErr  error
		repoCalls int
	}{
		{name: "cache hit", cacheHit: true},
		{name: "cache miss uses repository", repoCalls: 1},
		{name: "cache error uses repository", cacheErr: errors.New("injected cache error"), repoCalls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hydrated := &Account{ID: 9, Credentials: map[string]any{"api_key": "hydrated-key"}}
			cache := &slotReleaseHydrationCache{snapshotHydrationCache: &snapshotHydrationCache{}, err: tc.cacheErr}
			if tc.cacheHit {
				cache.accounts = map[int64]*Account{9: hydrated}
			}
			repo := &slotReleaseAccountRepo{account: hydrated}
			svc := &GatewayService{schedulerSnapshot: NewSchedulerSnapshotService(cache, nil, repo, nil, nil)}
			released := 0
			wait := &AccountWaitPlan{}
			result, err := svc.newSelectionResult(context.Background(), &Account{ID: 9}, true, func() { released++ }, wait)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, "hydrated-key", result.Account.GetCredential("api_key"))
			require.True(t, result.Acquired)
			require.Same(t, wait, result.WaitPlan)
			require.Equal(t, tc.repoCalls, repo.calls)
			require.Zero(t, released)
			require.NotNil(t, result.ReleaseFunc)
			result.ReleaseFunc()
			require.Equal(t, 1, released)
		})
	}
}

func TestGatewayServiceNewSelectionResultFailureWithoutOwnedRelease(t *testing.T) {
	cache := &snapshotHydrationCache{}
	repo := &slotReleaseAccountRepo{}
	svc := &GatewayService{schedulerSnapshot: NewSchedulerSnapshotService(cache, nil, repo, nil, nil)}
	result, err := svc.newSelectionResult(context.Background(), &Account{ID: 9}, true, nil, nil)
	require.Error(t, err)
	require.Nil(t, result)
	released := 0
	result, err = svc.newSelectionResult(context.Background(), &Account{ID: 9}, false, func() { released++ }, nil)
	require.Error(t, err)
	require.Nil(t, result)
	require.Zero(t, released)
}
