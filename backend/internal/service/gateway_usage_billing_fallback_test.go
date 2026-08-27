package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fallbackBillingUserRepoStub struct {
	UserRepository

	err   error
	calls int
}

func (s *fallbackBillingUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	s.calls++
	return s.err
}

type fallbackBillingCacheStub struct {
	BillingCache

	deductCalls     int
	invalidateCalls int
	lastUserID      int64
}

type usageBillingRepositoryErrorStub struct {
	UsageBillingRepository
	err error
}

func (s *usageBillingRepositoryErrorStub) Apply(context.Context, *UsageBillingCommand) (*UsageBillingApplyResult, error) {
	return nil, s.err
}

func (s *fallbackBillingCacheStub) DeductUserBalance(_ context.Context, userID int64, _ float64) error {
	s.deductCalls++
	s.lastUserID = userID
	return nil
}

func (s *fallbackBillingCacheStub) InvalidateUserBalance(_ context.Context, userID int64) error {
	s.invalidateCalls++
	s.lastUserID = userID
	return nil
}

func fallbackBillingParams() *postUsageBillingParams {
	return &postUsageBillingParams{
		Cost:    &CostBreakdown{TotalCost: 1, ActualCost: 1},
		User:    &User{ID: 101},
		APIKey:  &APIKey{ID: 202},
		Account: &Account{ID: 303},
	}
}

func TestApplyUsageBillingFallbackPropagatesBalanceFailureAndInvalidatesCache(t *testing.T) {
	userRepo := &fallbackBillingUserRepoStub{err: ErrInsufficientBalance}
	cache := &fallbackBillingCacheStub{}

	applied, err := applyUsageBilling(context.Background(), "req-fallback-failure", nil, fallbackBillingParams(), &billingDeps{
		userRepo:            userRepo,
		billingCacheService: &BillingCacheService{cache: cache},
	}, nil)

	require.ErrorIs(t, err, ErrInsufficientBalance)
	require.False(t, applied)
	require.Equal(t, 1, userRepo.calls)
	require.Equal(t, 1, cache.invalidateCalls)
	require.Equal(t, int64(101), cache.lastUserID)
	require.Zero(t, cache.deductCalls)
}

func TestApplyUsageBillingFallbackInvalidatesBalanceCacheAfterSuccess(t *testing.T) {
	userRepo := &fallbackBillingUserRepoStub{}
	cache := &fallbackBillingCacheStub{}

	applied, err := applyUsageBilling(context.Background(), "req-fallback-success", nil, fallbackBillingParams(), &billingDeps{
		userRepo:            userRepo,
		billingCacheService: &BillingCacheService{cache: cache},
	}, nil)

	require.NoError(t, err)
	require.True(t, applied)
	require.Equal(t, 1, userRepo.calls)
	require.Zero(t, cache.deductCalls)
	require.Equal(t, 1, cache.invalidateCalls)
	require.Equal(t, int64(101), cache.lastUserID)
}

func TestApplyUsageBillingFallbackFailsClosedWhenWelfareWalletIsActive(t *testing.T) {
	userRepo := &fallbackBillingUserRepoStub{}
	cache := &fallbackBillingCacheStub{}

	applied, err := applyUsageBillingWithNewUserTrialOverage(
		context.Background(),
		"req-fallback-welfare",
		nil,
		fallbackBillingParams(),
		&billingDeps{
			userRepo:            userRepo,
			billingCacheService: &BillingCacheService{cache: cache},
		},
		nil,
		&WelfareService{},
	)

	require.ErrorIs(t, err, ErrBillingServiceUnavailable)
	require.False(t, applied)
	require.Zero(t, userRepo.calls)
	require.Zero(t, cache.deductCalls)
	require.Zero(t, cache.invalidateCalls)
}

func TestApplyUsageBillingRejectsInvalidFallbackCommand(t *testing.T) {
	userRepo := &fallbackBillingUserRepoStub{}

	applied, err := applyUsageBilling(context.Background(), "req-invalid", nil, &postUsageBillingParams{
		User: &User{ID: 101},
	}, &billingDeps{userRepo: userRepo}, nil)

	require.ErrorIs(t, err, ErrUsageBillingCommandInvalid)
	require.False(t, applied)
	require.Zero(t, userRepo.calls)
}

func TestApplyUsageBillingFallbackRejectsMissingRepositoryDependency(t *testing.T) {
	applied, err := applyUsageBilling(
		context.Background(),
		"req-missing-user-repo",
		nil,
		fallbackBillingParams(),
		&billingDeps{},
		nil,
	)

	require.ErrorIs(t, err, ErrBillingServiceUnavailable)
	require.False(t, applied)
}

func TestApplyUsageBillingInvalidatesBalanceCacheWhenUnifiedDeductionIsRejected(t *testing.T) {
	cache := &fallbackBillingCacheStub{}

	applied, err := applyUsageBilling(
		context.Background(),
		"req-unified-insufficient",
		nil,
		fallbackBillingParams(),
		&billingDeps{billingCacheService: &BillingCacheService{cache: cache}},
		&usageBillingRepositoryErrorStub{err: ErrInsufficientBalance},
	)

	require.ErrorIs(t, err, ErrInsufficientBalance)
	require.False(t, applied)
	require.Equal(t, 1, cache.invalidateCalls)
	require.Equal(t, int64(101), cache.lastUserID)
}
