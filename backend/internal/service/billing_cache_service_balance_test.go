package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type balanceEligibilityCacheStub struct {
	billingCacheWorkerStub
	balance         float64
	invalidateCalls atomic.Int64
}

func (s *balanceEligibilityCacheStub) GetUserBalance(context.Context, int64) (float64, error) {
	return s.balance, nil
}

func (s *balanceEligibilityCacheStub) InvalidateUserBalance(context.Context, int64) error {
	s.invalidateCalls.Add(1)
	return nil
}

func TestCheckBillingEligibility_RejectsBalanceBelowMinimumReserve(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 0.0000005}
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.000001
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg)
	t.Cleanup(svc.Stop)

	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil)
	require.ErrorIs(t, err, ErrInsufficientBalance)
}

func TestCheckBillingEligibility_AllowsBalanceAtMinimumReserve(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 0.000001}
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.000001
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg)
	t.Cleanup(svc.Stop)

	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil)
	require.NoError(t, err)
}

func TestCheckBillingEligibility_PreservesVoucherFallbackBelowReserve(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 0.0000005}
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.000001
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg)
	t.Cleanup(svc.Stop)
	svc.SetVoucherBalanceProvider(balanceEligibilityVoucherStub{available: 0.01})

	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil)
	require.NoError(t, err)
}

type balanceEligibilityVoucherStub struct {
	available float64
}

func (s balanceEligibilityVoucherStub) GetVoucherAvailableAmount(context.Context, int64) (float64, error) {
	return s.available, nil
}

func TestCheckBillingEligibility_RejectsNegativeBalanceWithoutVoucher(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: -0.01}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{})
	t.Cleanup(svc.Stop)

	err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInsufficientBalance))
}
