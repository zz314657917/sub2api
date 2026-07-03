package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type billingAmountUserRepoStub struct {
	UserRepository

	user *User
	err  error
}

func (s *billingAmountUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.user != nil {
		return s.user, nil
	}
	return &User{}, nil
}

type voucherAmountProviderStub struct {
	available float64
	err       error
	calls     int
}

func (s *voucherAmountProviderStub) GetVoucherAvailableAmount(context.Context, int64) (float64, error) {
	s.calls++
	return s.available, s.err
}

func TestBillingCacheServiceCheckBalanceAmountEligibility_UsesVoucherAndBalance(t *testing.T) {
	provider := &voucherAmountProviderStub{available: 1}
	svc := NewBillingCacheService(nil, &billingAmountUserRepoStub{user: &User{ID: 7, Balance: 1}}, nil, nil, nil, nil, &config.Config{})
	t.Cleanup(svc.Stop)
	svc.SetVoucherBalanceProvider(provider)

	require.NoError(t, svc.CheckBalanceAmountEligibility(context.Background(), 7, 2))
	require.ErrorIs(t, svc.CheckBalanceAmountEligibility(context.Background(), 7, 3), ErrInsufficientBalance)
	require.Equal(t, 2, provider.calls)
}
