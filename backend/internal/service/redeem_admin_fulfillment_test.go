package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

type fulfillmentRedeemCacheStub struct {
	count          int
	getCalls       int
	incrementCalls int
	acquireCalls   int
	releaseCalls   int
}

func (c *fulfillmentRedeemCacheStub) GetRedeemAttemptCount(context.Context, int64) (int, error) {
	c.getCalls++
	return c.count, nil
}
func (c *fulfillmentRedeemCacheStub) IncrementRedeemAttemptCount(context.Context, int64) error {
	c.incrementCalls++
	c.count++
	return nil
}
func (c *fulfillmentRedeemCacheStub) AcquireRedeemLock(context.Context, string, time.Duration) (bool, error) {
	c.acquireCalls++
	return true, nil
}
func (c *fulfillmentRedeemCacheStub) ReleaseRedeemLock(context.Context, string) error {
	c.releaseCalls++
	return nil
}

func TestPublicRedeemStillEnforcesFailureLimit(t *testing.T) {
	cache := &fulfillmentRedeemCacheStub{count: redeemMaxErrorsPerHour}
	svc := &RedeemService{cache: cache}
	result, err := svc.Redeem(context.Background(), 42, "PUBLIC-CODE")
	require.Nil(t, result)
	require.ErrorIs(t, err, ErrRedeemRateLimited)
	require.Equal(t, 1, cache.getCalls)
	require.Zero(t, cache.incrementCalls)
	require.Zero(t, cache.acquireCalls)
}

func TestPaymentAndAdminFulfillmentBypassFailureLimit(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*RedeemService) (*RedeemCode, error)
	}{
		{name: "payment", call: func(s *RedeemService) (*RedeemCode, error) {
			return s.redeemForPaymentFulfillment(context.Background(), 42, "PAY-EXPIRED")
		}},
		{name: "admin", call: func(s *RedeemService) (*RedeemCode, error) {
			return s.RedeemForAdminFulfillment(context.Background(), 42, "PAY-EXPIRED")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := &fulfillmentRedeemCacheStub{count: redeemMaxErrorsPerHour}
			repo := &redeemRejectRepo{code: RedeemCode{ID: 1, Code: "PAY-EXPIRED", Type: RedeemTypeBalance, Value: 10, Status: StatusExpired}}
			svc := NewRedeemService(repo, nil, nil, cache, nil, nil, nil, nil)
			result, err := tc.call(svc)
			require.Nil(t, result)
			require.ErrorIs(t, err, ErrRedeemCodeExpired)
			require.Zero(t, cache.getCalls)
			require.Zero(t, cache.incrementCalls)
			require.Equal(t, 1, cache.acquireCalls)
			require.Equal(t, 1, cache.releaseCalls)
		})
	}
}

func TestValidatePaymentRedeemCodeRejectsMismatches(t *testing.T) {
	userID := int64(42)
	otherUserID := int64(43)
	order := &dbent.PaymentOrder{ID: 7, UserID: userID, RechargeCode: "PAY-7-12345", Amount: 80}
	require.NoError(t, validatePaymentRedeemCode(order, &RedeemCode{Code: order.RechargeCode, Type: RedeemTypeBalance, Value: 80, Status: StatusUnused}))
	require.NoError(t, validatePaymentRedeemCode(order, &RedeemCode{Code: order.RechargeCode, Type: RedeemTypeBalance, Value: 80, Status: StatusUsed, UsedBy: &userID}))
	for _, tc := range []struct {
		name string
		code *RedeemCode
		want string
	}{
		{"foreign user", &RedeemCode{Code: order.RechargeCode, Type: RedeemTypeBalance, Value: 80, Status: StatusUsed, UsedBy: &otherUserID}, "user mismatch"},
		{"wrong code", &RedeemCode{Code: "OTHER", Type: RedeemTypeBalance, Value: 80, Status: StatusUnused}, "code mismatch"},
		{"wrong type", &RedeemCode{Code: order.RechargeCode, Type: RedeemTypeConcurrency, Value: 80, Status: StatusUnused}, "type mismatch"},
		{"wrong amount", &RedeemCode{Code: order.RechargeCode, Type: RedeemTypeBalance, Value: 79, Status: StatusUnused}, "amount mismatch"},
		{"invalid status", &RedeemCode{Code: order.RechargeCode, Type: RedeemTypeBalance, Value: 80, Status: StatusDisabled}, "invalid status"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorContains(t, validatePaymentRedeemCode(order, tc.code), tc.want)
		})
	}
}
