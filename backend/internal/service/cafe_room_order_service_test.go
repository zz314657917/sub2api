package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCafeRoomOrderLocksSeatRejectsDuplicateUserAndReusesExpiredSeat(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_order_lock_reuse")
	now := time.Date(2026, 8, 3, 13, 0, 0, 0, time.UTC)
	fixture := newCafeRoomOrderFixture(t, ctx, client, now, 2)
	orderSvc := fixture.orderService
	cfg := &PaymentConfig{MaxPendingOrders: 3, OrderTimeoutMin: 30}

	first, round, err := orderSvc.lockSeatAndCreateOrder(ctx, CreateOrderRequest{UserID: fixture.user.ID, PaymentType: payment.TypeAlipay, ClientIP: "127.0.0.1", SrcHost: "api.example.com"}, fixture.room.ID, 1, cfg, 0, fixture.plan.PricePerShare, nil)
	require.NoError(t, err)
	require.Equal(t, fixture.round.ID, round.ID)
	seat, err := client.GroupBuySeat.Query().Where(groupbuyseat.OrderIDEQ(first.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusLocked, seat.Status)
	require.Equal(t, 1, *seat.SeatNo)

	_, _, err = orderSvc.lockSeatAndCreateOrder(ctx, CreateOrderRequest{UserID: fixture.user.ID, PaymentType: payment.TypeAlipay, ClientIP: "127.0.0.1", SrcHost: "api.example.com"}, fixture.room.ID, 2, cfg, 0, fixture.plan.PricePerShare, nil)
	require.ErrorIs(t, err, ErrCafeSeatHeld)

	first, err = client.PaymentOrder.UpdateOneID(first.ID).SetStatus(OrderStatusPending).Save(ctx)
	require.NoError(t, err)
	expiredAt := now.Add(-time.Minute)
	_, err = client.GroupBuySeat.UpdateOneID(seat.ID).SetLockedUntil(expiredAt).Save(ctx)
	require.NoError(t, err)
	secondUser := createGroupBuyTestUser(t, ctx, client, "cafe-reuse@example.com")
	second, _, err := orderSvc.lockSeatAndCreateOrder(ctx, CreateOrderRequest{UserID: secondUser.ID, PaymentType: payment.TypeAlipay, ClientIP: "127.0.0.1", SrcHost: "api.example.com"}, fixture.room.ID, 1, cfg, 0, fixture.plan.PricePerShare, nil)
	require.NoError(t, err)
	reused, err := client.GroupBuySeat.Query().Where(groupbuyseat.OrderIDEQ(second.ID)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, *reused.SeatNo)
	released, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusReleased, released.Status)
}

func TestCafeRoomOrderCapsDistinctLiveRooms(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_order_cap")
	now := time.Date(2026, 8, 3, 13, 0, 0, 0, time.UTC)
	fixture := newCafeRoomOrderFixture(t, ctx, client, now, 1)
	for index := 0; index < cafeRoomConcurrentRoomCap; index++ {
		room, round := createCafeRoomOrderRoom(t, ctx, client, fixture.plan.ID, fixture.accountID, now, 1, index+10)
		_, err := client.GroupBuySeat.Create().
			SetRoundID(round.ID).
			SetPlanID(fixture.plan.ID).
			SetUserID(fixture.user.ID).
			SetSeatNo(1).
			SetStatus(GroupBuySeatStatusPaid).
			SetShareCount(1).
			Save(ctx)
		require.NoError(t, err)
		require.NotZero(t, room.ID)
	}

	_, _, err := fixture.orderService.lockSeatAndCreateOrder(ctx, CreateOrderRequest{UserID: fixture.user.ID, PaymentType: payment.TypeAlipay, ClientIP: "127.0.0.1", SrcHost: "api.example.com"}, fixture.room.ID, 1, &PaymentConfig{MaxPendingOrders: 5, OrderTimeoutMin: 30}, 0, fixture.plan.PricePerShare, nil)
	require.ErrorIs(t, err, ErrCafeRoomCapExceeded)
}

func TestGroupBuyPaymentForCafeRoomStopsAtPaid(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_room_paid_only")
	now := time.Date(2026, 8, 3, 13, 0, 0, 0, time.UTC)
	fixture := newCafeRoomOrderFixture(t, ctx, client, now, 1)
	order, err := client.PaymentOrder.Create().
		SetUserID(fixture.user.ID).
		SetUserEmail(fixture.user.Email).
		SetUserName(fixture.user.Username).
		SetAmount(fixture.plan.PricePerShare).
		SetPayAmount(fixture.plan.PricePerShare).
		SetRechargeCode("cafe-paid-only").
		SetOutTradeNo("cafe_paid_only_1").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-cafe-paid-only").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusPaid).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)
	seat, err := client.GroupBuySeat.Create().
		SetRoundID(fixture.round.ID).
		SetPlanID(fixture.plan.ID).
		SetUserID(fixture.user.ID).
		SetOrderID(order.ID).
		SetSeatNo(1).
		SetStatus(GroupBuySeatStatusLocked).
		SetShareCount(1).
		SetLockedUntil(now.Add(time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.GroupBuyRound.UpdateOneID(fixture.round.ID).
		SetReservedShares(1).
		SetReservedSeats(1).
		Save(ctx)
	require.NoError(t, err)

	fixture.groupBuySvc.now = func() time.Time { return now }
	activation := &cafeRoundActivationStub{}
	fixture.groupBuySvc.SetCafeRoundActivation(activation)
	require.NoError(t, fixture.groupBuySvc.HandleGroupBuyOrderPaid(ctx, order.ID))
	require.Equal(t, []int64{fixture.round.ID}, activation.roundIDs)
	updatedSeat, err := client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusPaid, updatedSeat.Status)
	require.NoError(t, fixture.groupBuySvc.HandleGroupBuyOrderPaid(ctx, order.ID))
	require.Equal(t, []int64{fixture.round.ID, fixture.round.ID}, activation.roundIDs)
	updatedRound, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, CafeRoundStatusOpen, updatedRound.Status)
	require.Equal(t, 1, updatedRound.PaidSeats)
	require.Equal(t, 0, updatedRound.ReservedSeats)

	require.ErrorIs(t, fixture.groupBuySvc.AdminRetryActivation(ctx, fixture.round.ID), ErrCafeRoundLifecycleDeferred)
	closed, err := fixture.groupBuySvc.failTimedOutRound(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.False(t, closed)
	require.ErrorIs(t, fixture.groupBuySvc.AdminCloseRound(ctx, fixture.round.ID, "legacy close"), ErrCafeRoundLifecycleDeferred)
	_, err = fixture.groupBuySvc.AdminProcessRefunds(ctx, fixture.round.ID)
	require.ErrorIs(t, err, ErrCafeRoundLifecycleDeferred)

	updatedRound, err = client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, CafeRoundStatusOpen, updatedRound.Status)
	updatedSeat, err = client.GroupBuySeat.Get(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusPaid, updatedSeat.Status)
	count, err := client.UserSubscription.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
}

type cafeRoundActivationStub struct {
	roundIDs []int64
	err      error
}

func (s *cafeRoundActivationStub) ActivateRound(_ context.Context, roundID int64) error {
	s.roundIDs = append(s.roundIDs, roundID)
	return s.err
}

type cafeRoomOrderFixture struct {
	user         *User
	plan         *dbent.GroupBuyPlan
	room         *dbent.CafeRoom
	round        *dbent.GroupBuyRound
	accountID    int64
	orderService *CafeRoomOrderService
	groupBuySvc  *GroupBuyService
}

func newCafeRoomOrderFixture(t *testing.T, ctx context.Context, client *dbent.Client, now time.Time, totalSeats int) cafeRoomOrderFixture {
	t.Helper()
	user := createGroupBuyTestUser(t, ctx, client, "cafe-order@example.com")
	groupID := createGroupBuyTestGroup(t, ctx, client, 1, 100)
	_, err := client.Group.UpdateOneID(groupID).SetAccessMode(CafeRoomGroupAccessMode).Save(ctx)
	require.NoError(t, err)
	plan := createGroupBuyTestPlan(t, ctx, client, groupID, GroupBuyLaunchModeManual, totalSeats)
	plan, err = client.GroupBuyPlan.UpdateOneID(plan.ID).SetFulfillmentMode(CafeRoomFulfillmentMode).Save(ctx)
	require.NoError(t, err)
	account, err := client.Account.Create().
		SetName("cafe-order-account").
		SetPlatform(PlatformOpenAI).
		SetType("api_key").
		SetStatus(StatusActive).
		AddGroupIDs(groupID).
		Save(ctx)
	require.NoError(t, err)
	room, round := createCafeRoomOrderRoom(t, ctx, client, plan.ID, account.ID, now, totalSeats, 1)
	paymentSvc := &PaymentService{entClient: client, now: func() time.Time { return now }}
	groupBuySvc := newGroupBuyTestService(client, newGroupBuyGroupRepoStubWithGroup(groupID), nil)
	groupBuySvc.paymentSvc = paymentSvc
	groupBuySvc.now = func() time.Time { return now }
	return cafeRoomOrderFixture{
		user:         user,
		plan:         plan,
		room:         room,
		round:        round,
		accountID:    account.ID,
		orderService: &CafeRoomOrderService{entClient: client, paymentSvc: paymentSvc, groupBuySvc: groupBuySvc, now: func() time.Time { return now }},
		groupBuySvc:  groupBuySvc,
	}
}

func createCafeRoomOrderRoom(t *testing.T, ctx context.Context, client *dbent.Client, planID, accountID int64, now time.Time, totalSeats, sequence int) (*dbent.CafeRoom, *dbent.GroupBuyRound) {
	t.Helper()
	room, err := client.CafeRoom.Create().
		SetCode(fmt.Sprintf("CAFE-ORDER-%d", sequence)).
		SetName(fmt.Sprintf("Cafe order %d", sequence)).
		SetPlanID(planID).
		SetAccountID(accountID).
		SetStatus(CafeRoomStatusEnabled).
		Save(ctx)
	require.NoError(t, err)
	round, err := client.GroupBuyRound.Create().
		SetPlanID(planID).
		SetCafeRoomID(room.ID).
		SetAssignedAccountID(accountID).
		SetStatus(CafeRoundStatusOpen).
		SetTotalShares(totalSeats).
		SetTotalSeats(totalSeats).
		SetDeadlineAt(now.Add(time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	return room, round
}

func TestTranslateCafeSeatCreateErrorMapsDuplicate(t *testing.T) {
	err := translateCafeSeatCreateError(fmt.Errorf("duplicate key"))
	require.Equal(t, "CAFE_SEAT_UNAVAILABLE", infraerrors.Reason(err))
}
