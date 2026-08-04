package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/ent/groupbuyrefund"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCafeRoomLifecycleExpiresUnfilledRoundAndRefundsPaidSeat(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_lifecycle_timeout_refund")
	now := time.Date(2026, 8, 3, 18, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 2)

	paidSeat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID), groupbuyseat.SeatNoEQ(1)).Only(ctx)
	require.NoError(t, err)
	lockedSeat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID), groupbuyseat.SeatNoEQ(2)).Only(ctx)
	require.NoError(t, err)
	require.NoError(t, client.GroupBuySeat.UpdateOneID(lockedSeat.ID).
		SetStatus(GroupBuySeatStatusLocked).
		ClearPaidAt().
		Exec(ctx))
	require.NoError(t, client.GroupBuyRound.UpdateOneID(fixture.round.ID).
		SetPaidSeats(1).
		SetPaidShares(1).
		SetReservedSeats(1).
		SetReservedShares(1).
		SetDeadlineAt(now.Add(-time.Minute)).
		Exec(ctx))

	entUser, err := client.User.Get(ctx, paidSeat.UserID)
	require.NoError(t, err)
	userRepo := &groupBuyUserRepoStub{users: map[int64]*User{
		entUser.ID: {ID: entUser.ID, Email: entUser.Email, Username: entUser.Username, Status: entUser.Status},
	}}
	groupBuySvc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(fixture.groupID), nil)
	groupBuySvc.userRepo = userRepo
	groupBuySvc.now = func() time.Time { return now }
	expirySvc := NewCafeRoomExpiryService(client, &cafeExpiryCacheInvalidatorStub{})
	expirySvc.now = func() time.Time { return now }
	lifecycle := NewCafeRoomLifecycleService(client, groupBuySvc, fixture.service, expirySvc)
	lifecycle.now = func() time.Time { return now }

	updated, err := lifecycle.RunCafeLifecycle(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, updated)

	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusFailed, round.Status)
	require.NotNil(t, round.ClosedAt)
	refunded, err := client.GroupBuySeat.Get(ctx, paidSeat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefunded, refunded.Status)
	released, err := client.GroupBuySeat.Get(ctx, lockedSeat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusReleased, released.Status)
	require.Equal(t, 1, userRepo.balanceUpdates)
	require.InDelta(t, fixture.plan.PricePerSeat, userRepo.balanceTotal, 0.000001)
	refunds, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(paidSeat.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, refunds, 1)
	require.Equal(t, GroupBuyRefundStatusSucceeded, refunds[0].Status)
}

func TestCafeRoomLifecycleReconcilesCafeProviderRefundOutsideLegacyLifecycle(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_lifecycle_provider_reconcile")
	now := time.Date(2026, 8, 3, 18, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	user, err := client.User.Get(ctx, seat.UserID)
	require.NoError(t, err)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(fixture.plan.PricePerSeat).
		SetPayAmount(fixture.plan.PricePerSeat).
		SetRechargeCode("cafe-provider-refund").
		SetOutTradeNo("cafe_provider_refund_1").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("cafe-provider-trade").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusRefundPending).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.test").
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, client.GroupBuySeat.UpdateOneID(seat.ID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusRefundProcessing).
		Exec(ctx))
	require.NoError(t, client.GroupBuyRound.UpdateOneID(fixture.round.ID).
		SetStatus(GroupBuyRoundStatusFailed).
		SetClosedAt(now).
		Exec(ctx))
	_, err = client.GroupBuyRefund.Create().
		SetSeatID(seat.ID).
		SetOrderID(order.ID).
		SetUserID(seat.UserID).
		SetMode(GroupBuyRefundModeProviderRefund).
		SetStatus(GroupBuyRefundStatusPendingProvider).
		SetAmount(order.Amount).
		SetIdempotencyKey("cafe-provider-refund-seat").
		SetNote("cafe provider refund pending").
		Save(ctx)
	require.NoError(t, err)

	groupBuySvc := newGroupBuyTestService(client, nil, nil)
	refundStub := &groupBuyPaymentRefundStub{client: client, hasPendingAudit: true, queryStatus: OrderStatusRefunded}
	groupBuySvc.refundSvc = refundStub
	legacyFinalized, err := groupBuySvc.ReconcilePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Zero(t, legacyFinalized)
	require.Zero(t, refundStub.queryCalls)

	expirySvc := NewCafeRoomExpiryService(client, &cafeExpiryCacheInvalidatorStub{})
	expirySvc.now = func() time.Time { return now }
	lifecycle := NewCafeRoomLifecycleService(client, groupBuySvc, fixture.service, expirySvc)
	lifecycle.now = func() time.Time { return now }
	finalized, err := lifecycle.reconcileCafePendingProviderRefunds(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, finalized)
	require.Equal(t, 1, refundStub.queryCalls)

	refund, err := client.GroupBuyRefund.Query().Where(groupbuyrefund.SeatIDEQ(seat.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRefundStatusSucceeded, refund.Status)
	reloadedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusRefunded, reloadedSeat.Status)
}

func TestCafeRoomLifecycleRetriesValidActivatingRound(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_lifecycle_activation_retry")
	now := time.Date(2026, 8, 3, 18, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	failingRepo := &cafeActivationFailOnceAPIKeyRepo{
		cafeActivationAPIKeyRepo: fixture.service.apiKeyRepo.(*cafeActivationAPIKeyRepo),
		failNextCreate:           true,
	}
	fixture.service.apiKeyRepo = failingRepo
	require.ErrorIs(t, fixture.service.ActivateRound(ctx, fixture.round.ID), ErrCafeActivationFailed)

	groupBuySvc := newGroupBuyTestService(client, nil, nil)
	expirySvc := NewCafeRoomExpiryService(client, &cafeExpiryCacheInvalidatorStub{})
	expirySvc.now = func() time.Time { return now.Add(time.Minute) }
	lifecycle := NewCafeRoomLifecycleService(client, groupBuySvc, fixture.service, expirySvc)
	lifecycle.now = func() time.Time { return now.Add(time.Minute) }

	updated, err := lifecycle.RunCafeLifecycle(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, updated)
	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, round.Status)
}

func TestCafeRoomLifecycleCompensatesPaidFullOpenRound(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_lifecycle_open_full_retry")
	now := time.Date(2026, 8, 3, 18, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)

	groupBuySvc := newGroupBuyTestService(client, nil, nil)
	expirySvc := NewCafeRoomExpiryService(client, &cafeExpiryCacheInvalidatorStub{})
	expirySvc.now = func() time.Time { return now.Add(time.Minute) }
	lifecycle := NewCafeRoomLifecycleService(client, groupBuySvc, fixture.service, expirySvc)
	lifecycle.now = func() time.Time { return now.Add(time.Minute) }

	updated, err := lifecycle.RunCafeLifecycle(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, updated)
	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActive, round.Status)
}

func TestCafeRoomLifecycleLeavesInvalidActivatingRoundUntouched(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_lifecycle_invalid_activation")
	now := time.Date(2026, 8, 3, 18, 0, 0, 0, time.UTC)
	fixture := newCafeActivationFixture(t, ctx, client, now, 1)
	require.NoError(t, client.GroupBuyRound.UpdateOneID(fixture.round.ID).
		SetStatus(GroupBuyRoundStatusActivating).
		Exec(ctx))

	groupBuySvc := newGroupBuyTestService(client, nil, nil)
	expirySvc := NewCafeRoomExpiryService(client, &cafeExpiryCacheInvalidatorStub{})
	expirySvc.now = func() time.Time { return now }
	lifecycle := NewCafeRoomLifecycleService(client, groupBuySvc, fixture.service, expirySvc)
	lifecycle.now = func() time.Time { return now }

	updated, err := lifecycle.RunCafeLifecycle(ctx)
	require.ErrorIs(t, err, ErrCafeActivationFailed)
	require.Zero(t, updated)
	round, err := client.GroupBuyRound.Query().Where(groupbuyround.IDEQ(fixture.round.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuyRoundStatusActivating, round.Status)
}
