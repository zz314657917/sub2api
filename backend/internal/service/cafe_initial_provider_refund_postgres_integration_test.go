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
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCafeInitialProviderRefundPostgresIntegration(t *testing.T) {
	client := newCafeRoomOrderPostgresIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Date(2026, 8, 3, 22, 15, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	plan, err := client.GroupBuyPlan.UpdateOneID(fixture.plan.ID).
		SetRefundMode(GroupBuyRefundModeProviderRefund).
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	user, err := client.User.Get(ctx, seat.UserID)
	require.NoError(t, err)

	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeStripe).
		SetName("cafe-initial-refund-provider").
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
		SetAmount(plan.PricePerSeat).
		SetPayAmount(plan.PricePerSeat).
		SetFeeRate(0).
		SetRechargeCode("cafe-initial-provider-refund").
		SetOutTradeNo("cafe_initial_provider_refund_1").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("cafe-initial-provider-trade").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusCompleted).
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
		SetStatus(GroupBuySeatStatusRefundPending).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRound.UpdateOneID(fixture.round.ID).
		SetStatus(GroupBuyRoundStatusFailed).
		SetClosedAt(now).
		Save(ctx)
	require.NoError(t, err)

	provider := &cafeInitialProviderRefundStub{
		key:            payment.TypeStripe,
		refundResponse: &payment.RefundResponse{RefundID: "cafe-initial-refund-1", Status: payment.ProviderStatusPending},
		queryResponse:  &payment.RefundResponse{RefundID: "cafe-initial-refund-1", Status: payment.ProviderStatusSuccess},
	}
	originalFactory := createPaymentProviderFromInstance
	createPaymentProviderFromInstance = func(providerKey, factoryInstanceID string, _ map[string]string) (payment.Provider, error) {
		require.Equal(t, payment.TypeStripe, providerKey)
		require.Equal(t, instanceID, factoryInstanceID)
		return provider, nil
	}
	t.Cleanup(func() { createPaymentProviderFromInstance = originalFactory })

	paymentSvc := &PaymentService{
		entClient:    client,
		loadBalancer: cafeInitialProviderRefundLoadBalancer{},
	}
	groupBuySvc := newGroupBuyTestService(client, nil, nil)
	groupBuySvc.refundSvc = paymentSvc
	groupBuySvc.now = func() time.Time { return now }
	expirySvc := NewCafeRoomExpiryService(client, &cafeExpiryCacheInvalidatorStub{})
	lifecycle := NewCafeRoomLifecycleService(client, groupBuySvc, fixture.service, expirySvc)

	processed, err := lifecycle.processCafeRefunds(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	assertCafeInitialProviderRefundPending(t, ctx, client, order.ID, seat.ID)
	require.Len(t, provider.refundRequests, 1)
	require.Empty(t, provider.queryRequests)
	require.Equal(t, payment.RefundRequest{
		TradeNo: order.PaymentTradeNo,
		OrderID: order.OutTradeNo,
		Amount:  payment.FormatAmountForCurrency(plan.PricePerSeat, PaymentOrderCurrency(order)),
		Reason:  "Token拼拼拼 未满份原路退款",
	}, provider.refundRequests[0])
	require.Equal(t, 1, cafeInitialRefundAuditCount(t, ctx, client, order.ID, "REFUND_PENDING"))
	require.Equal(t, 1, cafeInitialGroupBuyRefundCount(t, ctx, client, seat.ID))

	pendingReplayProcessed, err := lifecycle.processCafeRefunds(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, pendingReplayProcessed)
	assertCafeInitialProviderRefundPending(t, ctx, client, order.ID, seat.ID)
	require.Len(t, provider.refundRequests, 1)
	require.Empty(t, provider.queryRequests)
	require.Equal(t, 1, cafeInitialRefundAuditCount(t, ctx, client, order.ID, "REFUND_PENDING"))
	require.Equal(t, 1, cafeInitialGroupBuyRefundCount(t, ctx, client, seat.ID))

	finalized, err := lifecycle.reconcileCafePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, finalized)
	assertCafeInitialProviderRefundSucceeded(t, ctx, client, order.ID, seat.ID)
	require.Len(t, provider.refundRequests, 1)
	require.Len(t, provider.queryRequests, 1)
	require.Equal(t, "cafe-initial-refund-1", provider.queryRequests[0].RefundID)
	require.Equal(t, 1, cafeInitialGroupBuyRefundCount(t, ctx, client, seat.ID))
	require.Equal(t, 1, cafeInitialRefundProcessedEventCount(t, ctx, client, seat.ID))

	terminalReplayProcessed, err := lifecycle.processCafeRefunds(ctx)
	require.NoError(t, err)
	require.Zero(t, terminalReplayProcessed)
	terminalReplayFinalized, err := lifecycle.reconcileCafePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Zero(t, terminalReplayFinalized)
	assertCafeInitialProviderRefundSucceeded(t, ctx, client, order.ID, seat.ID)
	require.Len(t, provider.refundRequests, 1)
	require.Len(t, provider.queryRequests, 1)
	require.Equal(t, 1, cafeInitialGroupBuyRefundCount(t, ctx, client, seat.ID))
	require.Equal(t, 1, cafeInitialRefundProcessedEventCount(t, ctx, client, seat.ID))
}

type cafeInitialProviderRefundLoadBalancer struct{}

func (cafeInitialProviderRefundLoadBalancer) GetInstanceConfig(context.Context, int64) (map[string]string, error) {
	return map[string]string{}, nil
}

func (cafeInitialProviderRefundLoadBalancer) SelectInstance(context.Context, string, payment.PaymentType, payment.Strategy, float64) (*payment.InstanceSelection, error) {
	return nil, fmt.Errorf("unexpected provider instance selection")
}

type cafeInitialProviderRefundStub struct {
	key            string
	refundResponse *payment.RefundResponse
	queryResponse  *payment.RefundResponse
	refundRequests []payment.RefundRequest
	queryRequests  []payment.RefundQueryRequest
}

func (p *cafeInitialProviderRefundStub) Name() string { return p.key }

func (p *cafeInitialProviderRefundStub) ProviderKey() string { return p.key }

func (p *cafeInitialProviderRefundStub) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeStripe}
}

func (p *cafeInitialProviderRefundStub) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	return nil, fmt.Errorf("unexpected payment creation")
}

func (p *cafeInitialProviderRefundStub) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	return nil, fmt.Errorf("unexpected payment order query")
}

func (p *cafeInitialProviderRefundStub) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	return nil, fmt.Errorf("unexpected payment notification verification")
}

func (p *cafeInitialProviderRefundStub) Refund(_ context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	p.refundRequests = append(p.refundRequests, req)
	return p.refundResponse, nil
}

func (p *cafeInitialProviderRefundStub) QueryRefund(_ context.Context, req payment.RefundQueryRequest) (*payment.RefundResponse, error) {
	p.queryRequests = append(p.queryRequests, req)
	return p.queryResponse, nil
}

func assertCafeInitialProviderRefundPending(t *testing.T, ctx context.Context, client *dbent.Client, orderID, seatID int64) {
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

func assertCafeInitialProviderRefundSucceeded(t *testing.T, ctx context.Context, client *dbent.Client, orderID, seatID int64) {
	t.Helper()
	order, err := client.PaymentOrder.Get(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, order.Status)
	seat, err := client.GroupBuySeat.Get(ctx, seatID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefunded, seat.Status)
	refund, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seatID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusSucceeded, refund.Status)
}

func cafeInitialRefundAuditCount(t *testing.T, ctx context.Context, client *dbent.Client, orderID int64, action string) int {
	t.Helper()
	count, err := client.PaymentAuditLog.Query().Where(
		paymentauditlog.OrderIDEQ(strconv.FormatInt(orderID, 10)),
		paymentauditlog.ActionEQ(action),
	).Count(ctx)
	require.NoError(t, err)
	return count
}

func cafeInitialGroupBuyRefundCount(t *testing.T, ctx context.Context, client *dbent.Client, seatID int64) int {
	t.Helper()
	count, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seatID)).Count(ctx)
	require.NoError(t, err)
	return count
}

func cafeInitialRefundProcessedEventCount(t *testing.T, ctx context.Context, client *dbent.Client, seatID int64) int {
	t.Helper()
	count, err := client.GroupBuyEvent.Query().Where(groupbuyevent.SeatIDEQ(seatID)).Count(ctx)
	require.NoError(t, err)
	return count
}
