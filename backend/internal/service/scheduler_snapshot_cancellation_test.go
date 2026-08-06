package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type schedulerCancellationCache struct {
	SchedulerCache
	cancel             context.CancelFunc
	cancelOnSnapshot   bool
	cancelOnGetAccount bool
	setSnapshotCalls   int
}

func (c *schedulerCancellationCache) GetSnapshot(ctx context.Context, _ SchedulerBucket) ([]*Account, bool, error) {
	if c.cancelOnSnapshot {
		c.cancel()
		return nil, false, ctx.Err()
	}
	return nil, false, nil
}

func (c *schedulerCancellationCache) SetSnapshot(ctx context.Context, _ SchedulerBucket, _ []Account) error {
	c.setSnapshotCalls++
	return ctx.Err()
}

func (c *schedulerCancellationCache) GetAccount(ctx context.Context, _ int64) (*Account, error) {
	if c.cancelOnGetAccount {
		c.cancel()
		return nil, ctx.Err()
	}
	return nil, nil
}

type schedulerCancellationAccountRepo struct {
	AccountRepository
	cancel       context.CancelFunc
	listCalls    int
	getByIDCalls int
}

func (r *schedulerCancellationAccountRepo) ListSchedulableUngroupedByPlatform(ctx context.Context, _ string) ([]Account, error) {
	r.listCalls++
	if r.cancel != nil {
		r.cancel()
	}
	return nil, ctx.Err()
}

func (r *schedulerCancellationAccountRepo) GetByID(ctx context.Context, _ int64) (*Account, error) {
	r.getByIDCalls++
	if r.cancel != nil {
		r.cancel()
	}
	return nil, ctx.Err()
}

func TestSchedulerSnapshotListStopsAfterRequestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cache := &schedulerCancellationCache{cancel: cancel, cancelOnSnapshot: true}
	repo := &schedulerCancellationAccountRepo{}
	svc := NewSchedulerSnapshotService(cache, nil, repo, nil, nil)

	accounts, useMixed, err := svc.ListSchedulableAccounts(ctx, nil, PlatformOpenAI, false)

	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, accounts)
	require.False(t, useMixed)
	require.Zero(t, repo.listCalls, "canceled requests must not fall back to the database")
	require.Zero(t, cache.setSnapshotCalls, "canceled requests must not publish a cache snapshot")
}

func TestSchedulerSnapshotListDoesNotPublishAfterDatabaseCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cache := &schedulerCancellationCache{}
	repo := &schedulerCancellationAccountRepo{cancel: cancel}
	svc := NewSchedulerSnapshotService(cache, nil, repo, nil, nil)

	accounts, useMixed, err := svc.ListSchedulableAccounts(ctx, nil, PlatformOpenAI, false)

	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, accounts)
	require.False(t, useMixed)
	require.Equal(t, 1, repo.listCalls)
	require.Zero(t, cache.setSnapshotCalls, "canceled database fallbacks must not publish a cache snapshot")
}

func TestSchedulerSnapshotGetAccountStopsAfterRequestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cache := &schedulerCancellationCache{cancel: cancel, cancelOnGetAccount: true}
	repo := &schedulerCancellationAccountRepo{}
	svc := NewSchedulerSnapshotService(cache, nil, repo, nil, nil)

	account, err := svc.GetAccount(ctx, 42)

	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, account)
	require.Zero(t, repo.getByIDCalls, "canceled requests must not fall back to the database")
}
