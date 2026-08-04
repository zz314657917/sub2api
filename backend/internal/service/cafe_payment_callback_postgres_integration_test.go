//go:build integration

package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyevent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCafePaymentCallbackPaidFullPostgresIntegration(t *testing.T) {
	client := newCafeRoomOrderPostgresIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Date(2026, 8, 3, 21, 20, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	seat, err := client.GroupBuySeat.Query().
		Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).
		Only(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(seat.UserID).
		SetUserEmail("cafe-payment-callback@example.com").
		SetUserName("Cafe payment callback").
		SetAmount(fixture.plan.PricePerShare).
		SetPayAmount(fixture.plan.PricePerShare).
		SetRechargeCode("cafe-payment-callback-activation").
		SetOutTradeNo("cafe_payment_callback_activation_1").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetProviderKey(payment.TypeAlipay).
		SetOrderType(payment.OrderTypeGroupBuy).
		SetPlanID(fixture.plan.ID).
		SetStatus(OrderStatusPending).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("cafe-payment-test.invalid").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.GroupBuySeat.UpdateOneID(seat.ID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusLocked).
		SetLockedUntil(now.Add(time.Hour)).
		ClearPaidAt().
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRound.UpdateOneID(fixture.round.ID).
		SetPaidSeats(0).
		SetPaidShares(0).
		SetReservedSeats(1).
		SetReservedShares(1).
		Save(ctx)
	require.NoError(t, err)

	paymentSvc := &PaymentService{entClient: client}
	groupBuySvc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(fixture.groupID), nil)
	groupBuySvc.paymentSvc = paymentSvc
	groupBuySvc.now = func() time.Time { return now }
	groupBuySvc.SetCafeRoundActivation(fixture.service)
	paymentSvc.SetGroupBuyFulfillment(groupBuySvc)

	badAmountNotification := &payment.PaymentNotification{
		OrderID: order.OutTradeNo,
		TradeNo: "cafe-provider-trade-amount-mismatch",
		Amount:  fixture.plan.PricePerShare + 1,
		Status:  payment.NotificationStatusSuccess,
	}
	require.Error(t, paymentSvc.HandlePaymentNotification(ctx, badAmountNotification, payment.TypeAlipay))
	assertCafeCallbackStillPending(t, ctx, client, order.ID, seat.ID, fixture.round.ID)

	notification := &payment.PaymentNotification{
		OrderID: order.OutTradeNo,
		TradeNo: "cafe-provider-trade-success",
		Amount:  fixture.plan.PricePerShare,
		Status:  payment.NotificationStatusSuccess,
	}
	require.NoError(t, paymentSvc.HandlePaymentNotification(ctx, notification, payment.TypeAlipay))
	assertCafeCallbackActivated(t, ctx, client, fixture, order.ID, seat.ID)

	paymentAuditCount, err := client.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10))).Count(ctx)
	require.NoError(t, err)
	eventCount, err := client.GroupBuyEvent.Query().Where(groupbuyevent.RoundIDEQ(fixture.round.ID)).Count(ctx)
	require.NoError(t, err)
	require.NoError(t, paymentSvc.HandlePaymentNotification(ctx, notification, payment.TypeAlipay))

	assertCafeCallbackActivated(t, ctx, client, fixture, order.ID, seat.ID)
	replayedPaymentAuditCount, err := client.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10))).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, paymentAuditCount, replayedPaymentAuditCount)
	replayedEventCount, err := client.GroupBuyEvent.Query().Where(groupbuyevent.RoundIDEQ(fixture.round.ID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, eventCount, replayedEventCount)
}

func assertCafeCallbackStillPending(t *testing.T, ctx context.Context, client *dbent.Client, orderID, seatID, roundID int64) {
	t.Helper()
	order, err := client.PaymentOrder.Get(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusPending, order.Status)
	seat, err := client.GroupBuySeat.Get(ctx, seatID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusLocked, seat.Status)
	round, err := client.GroupBuyRound.Get(ctx, roundID)
	require.NoError(t, err)
	require.Equal(t, CafeRoundStatusOpen, round.Status)
	require.Equal(t, 1, round.ReservedSeats)
	require.Equal(t, 1, round.ReservedShares)
	require.Zero(t, round.PaidSeats)
	require.Zero(t, round.PaidShares)
	keyCount, err := client.APIKey.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, keyCount)
}

func assertCafeCallbackActivated(t *testing.T, ctx context.Context, client *dbent.Client, fixture cafeActivationFixture, orderID, seatID int64) {
	t.Helper()
	order, err := client.PaymentOrder.Get(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, order.Status)
	seat, err := client.GroupBuySeat.Get(ctx, seatID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusActive, seat.Status)
	require.NotNil(t, seat.BoundAPIKeyID)
	require.NotNil(t, seat.OrderID)
	require.Equal(t, orderID, *seat.OrderID)
	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, round.Status)
	require.Equal(t, 1, round.PaidSeats)
	require.Equal(t, 1, round.PaidShares)
	require.Zero(t, round.ReservedSeats)
	require.Zero(t, round.ReservedShares)

	keys, err := client.APIKey.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	key := keys[0]
	require.Equal(t, StatusAPIKeyActive, key.Status)
	require.Equal(t, APIKeyManagedSourceCafeRoomSeat, key.ManagedSourceType)
	require.NotNil(t, key.ManagedSourceID)
	require.Equal(t, seatID, *key.ManagedSourceID)

	bindings, err := client.APIKeyAccountBinding.Query().Where(
		apikeyaccountbinding.SeatIDEQ(seatID),
		apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive),
	).All(ctx)
	require.NoError(t, err)
	require.Len(t, bindings, 1)
	binding := bindings[0]
	require.Equal(t, key.ID, binding.APIKeyID)
	require.Equal(t, fixture.accountID, binding.AccountID)
	require.Equal(t, fixture.room.ID, binding.CafeRoomID)
	require.Equal(t, fixture.round.ID, binding.RoundID)
	require.True(t, binding.StrictMode)

	subscriptions, err := client.UserSubscription.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, subscriptions)
}
