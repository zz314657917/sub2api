//go:build integration

package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyevent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyrefund"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCafeProviderPendingRefundPostgresIntegration(t *testing.T) {
	client := newCafeRoomOrderPostgresIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Date(2026, 8, 3, 21, 30, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	user, err := client.User.Get(ctx, seat.UserID)
	require.NoError(t, err)

	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeStripe).
		SetName("cafe-refund-query-provider").
		SetConfig("{}").
		SetSupportedTypes(payment.TypeStripe).
		SetEnabled(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)
	instanceID := strconv.FormatInt(instance.ID, 10)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(fixture.plan.PricePerSeat).
		SetPayAmount(fixture.plan.PricePerSeat).
		SetFeeRate(0).
		SetRechargeCode("cafe-provider-pending-refund").
		SetOutTradeNo("cafe_provider_pending_refund_1").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("cafe-provider-trade").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusRefundPending).
		SetRefundAmount(fixture.plan.PricePerSeat).
		SetRefundReason("cafe provider pending refund").
		SetExpiresAt(now.Add(time.Hour)).
		SetPaidAt(now).
		SetClientIP("127.0.0.1").
		SetSrcHost("cafe-refund-test.invalid").
		SetProviderInstanceID(instanceID).
		SetProviderKey(payment.TypeStripe).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuySeat.UpdateOneID(seat.ID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusRefundProcessing).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRound.UpdateOneID(fixture.round.ID).
		SetStatus(GroupBuyRoundStatusFailed).
		SetClosedAt(now).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRefund.Create().
		SetSeatID(seat.ID).
		SetOrderID(order.ID).
		SetUserID(seat.UserID).
		SetMode(GroupBuyRefundModeProviderRefund).
		SetStatus(GroupBuyRefundStatusPendingProvider).
		SetAmount(order.Amount).
		SetIdempotencyKey("cafe-provider-pending-refund-seat").
		SetNote("cafe provider refund pending").
		Save(ctx)
	require.NoError(t, err)

	paymentSvc := &PaymentService{
		entClient:    client,
		loadBalancer: cafeProviderRefundLoadBalancer{},
	}
	paymentSvc.writeAuditLog(ctx, order.ID, "REFUND_PENDING", "test", map[string]any{
		"refundID":            "cafe-refund-pending-1",
		"deductionRequested":  false,
		"deductionType":       payment.DeductionTypeNone,
		"deductionRollbackOK": true,
	})
	provider := &cafeProviderRefundQueryStub{
		key:           payment.TypeStripe,
		queryResponse: &payment.RefundResponse{RefundID: "cafe-refund-pending-1", Status: payment.ProviderStatusPending},
	}
	originalFactory := createPaymentProviderFromInstance
	createPaymentProviderFromInstance = func(providerKey, factoryInstanceID string, _ map[string]string) (payment.Provider, error) {
		require.Equal(t, payment.TypeStripe, providerKey)
		require.Equal(t, instanceID, factoryInstanceID)
		return provider, nil
	}
	t.Cleanup(func() { createPaymentProviderFromInstance = originalFactory })

	groupBuySvc := newGroupBuyTestService(client, nil, nil)
	groupBuySvc.refundSvc = paymentSvc
	groupBuySvc.now = func() time.Time { return now }
	legacyFinalized, err := groupBuySvc.ReconcilePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Zero(t, legacyFinalized)
	require.Empty(t, provider.queryRequests)

	expirySvc := NewCafeRoomExpiryService(client, &cafeExpiryCacheInvalidatorStub{})
	lifecycle := NewCafeRoomLifecycleService(client, groupBuySvc, fixture.service, expirySvc)
	pendingFinalized, err := lifecycle.reconcileCafePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Zero(t, pendingFinalized)
	assertCafeProviderRefundPending(t, ctx, client, order.ID, seat.ID)
	require.Len(t, provider.queryRequests, 1)
	require.Equal(t, "cafe-refund-pending-1", provider.queryRequests[0].RefundID)

	provider.queryResponse = &payment.RefundResponse{RefundID: "cafe-refund-pending-1", Status: payment.ProviderStatusSuccess}
	finalized, err := lifecycle.reconcileCafePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, finalized)
	assertCafeProviderRefundSucceeded(t, ctx, client, order.ID, seat.ID)
	eventCount, err := client.GroupBuyEvent.Query().Where(groupbuyevent.SeatIDEQ(seat.ID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, eventCount)

	replayedFinalized, err := lifecycle.reconcileCafePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Zero(t, replayedFinalized)
	assertCafeProviderRefundSucceeded(t, ctx, client, order.ID, seat.ID)
	replayedEventCount, err := client.GroupBuyEvent.Query().Where(groupbuyevent.SeatIDEQ(seat.ID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, eventCount, replayedEventCount)
	require.Len(t, provider.queryRequests, 2)
}

type cafeProviderRefundLoadBalancer struct{}

func (cafeProviderRefundLoadBalancer) GetInstanceConfig(context.Context, int64) (map[string]string, error) {
	return map[string]string{}, nil
}

func (cafeProviderRefundLoadBalancer) SelectInstance(context.Context, string, payment.PaymentType, payment.Strategy, float64) (*payment.InstanceSelection, error) {
	return nil, fmt.Errorf("unexpected provider instance selection")
}

type cafeProviderRefundQueryStub struct {
	key           string
	queryResponse *payment.RefundResponse
	queryRequests []payment.RefundQueryRequest
}

func (p *cafeProviderRefundQueryStub) Name() string { return p.key }

func (p *cafeProviderRefundQueryStub) ProviderKey() string { return p.key }

func (p *cafeProviderRefundQueryStub) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeStripe}
}

func (p *cafeProviderRefundQueryStub) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	return nil, fmt.Errorf("unexpected payment creation")
}

func (p *cafeProviderRefundQueryStub) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	return nil, fmt.Errorf("unexpected payment order query")
}

func (p *cafeProviderRefundQueryStub) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	return nil, fmt.Errorf("unexpected payment notification verification")
}

func (p *cafeProviderRefundQueryStub) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("unexpected provider refund execution")
}

func (p *cafeProviderRefundQueryStub) QueryRefund(_ context.Context, req payment.RefundQueryRequest) (*payment.RefundResponse, error) {
	p.queryRequests = append(p.queryRequests, req)
	return p.queryResponse, nil
}

func assertCafeProviderRefundPending(t *testing.T, ctx context.Context, client *dbent.Client, orderID, seatID int64) {
	t.Helper()
	order, err := client.PaymentOrder.Get(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefundPending, order.Status)
	seat, err := client.GroupBuySeat.Get(ctx, seatID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefundProcessing, seat.Status)
	refund, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seatID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusPendingProvider, refund.Status)
}

func assertCafeProviderRefundSucceeded(t *testing.T, ctx context.Context, client *dbent.Client, orderID, seatID int64) {
	t.Helper()
	order, err := client.PaymentOrder.Get(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, order.Status)
	seat, err := client.GroupBuySeat.Get(ctx, seatID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefunded, seat.Status)
	refunds, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seatID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, refunds, 1)
	require.Equal(t, GroupBuyRefundStatusSucceeded, refunds[0].Status)
}
