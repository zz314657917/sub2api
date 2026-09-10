package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/ent/caferoundmembership"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCafeReservationCancelReleasesExactOrderlessBatchAndIsRetrySafe(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_reservation_cancel")
	now := time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC)
	fixture := newCafeRoomOrderFixture(t, ctx, client, now, 10)
	fixture.orderService.settings = cafePublicSettingsStub{enabled: true}
	first, err := fixture.orderService.ReserveShares(ctx, CafeRoomReservationInput{UserID: fixture.user.ID, RoomID: fixture.room.ID, ShareCount: 6, AgreementAccepted: true})
	require.NoError(t, err)
	other := createGroupBuyTestUser(t, ctx, client, "cafe-cancel-other@example.com")
	second, err := fixture.orderService.ReserveShares(ctx, CafeRoomReservationInput{UserID: other.ID, RoomID: fixture.room.ID, ShareCount: 4, AgreementAccepted: true})
	require.NoError(t, err)
	require.Equal(t, CafeRoundStatusAwaitingPayment, second.Status)

	cancelled, err := fixture.orderService.CancelReservation(ctx, CafeRoomReservationCancelInput{UserID: other.ID, RoomID: fixture.room.ID, ReservationID: second.ReservationID})
	require.NoError(t, err)
	require.True(t, cancelled.Cancelled)
	require.Equal(t, CafeRoundStatusReserving, cancelled.Status)
	require.Equal(t, 6, cancelled.ReservedShares)

	seat, err := client.GroupBuySeat.Get(ctx, second.ReservationID)
	require.NoError(t, err)
	require.Equal(t, GroupBuySeatStatusReleased, seat.Status)
	require.Nil(t, seat.OrderID)
	membership, err := client.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(fixture.round.ID), caferoundmembership.UserIDEQ(other.ID)).Only(ctx)
	require.NoError(t, err)
	require.Zero(t, membership.ReservedShares)
	round, err := client.GroupBuyRound.Get(ctx, fixture.round.ID)
	require.NoError(t, err)
	require.Equal(t, 6, round.ReservedShares)
	require.Equal(t, CafeRoundStatusReserving, round.Status)

	retry, err := fixture.orderService.CancelReservation(ctx, CafeRoomReservationCancelInput{UserID: other.ID, RoomID: fixture.room.ID, ReservationID: second.ReservationID})
	require.NoError(t, err)
	require.False(t, retry.Cancelled)
	live, err := client.GroupBuySeat.Query().Where(groupbuyseat.IDEQ(first.ReservationID), groupbuyseat.StatusEQ(GroupBuySeatStatusLocked)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, first.ReservationID, live.ID)
}

func TestCafeReservationCancelRejectsCrossUserAndOrderAttachedBatches(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_reservation_cancel_guard")
	now := time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC)
	fixture := newCafeRoomOrderFixture(t, ctx, client, now, 1)
	fixture.orderService.settings = cafePublicSettingsStub{enabled: true}
	reservation, err := fixture.orderService.ReserveShares(ctx, CafeRoomReservationInput{UserID: fixture.user.ID, RoomID: fixture.room.ID, ShareCount: 1, AgreementAccepted: true})
	require.NoError(t, err)
	other := createGroupBuyTestUser(t, ctx, client, "cafe-cancel-attacker@example.com")
	_, err = fixture.orderService.CancelReservation(ctx, CafeRoomReservationCancelInput{UserID: other.ID, RoomID: fixture.room.ID, ReservationID: reservation.ReservationID})
	require.ErrorIs(t, err, ErrCafeReservationNotCancellable)

	_, _, err = fixture.orderService.lockSeatAndCreateOrder(ctx, CreateOrderRequest{UserID: fixture.user.ID, PaymentType: payment.TypeAlipay}, fixture.room.ID, 1, &PaymentConfig{MaxPendingOrders: 3, OrderTimeoutMin: 30}, 0, fixture.plan.PricePerShare, nil)
	require.NoError(t, err)
	_, err = fixture.orderService.CancelReservation(ctx, CafeRoomReservationCancelInput{UserID: fixture.user.ID, RoomID: fixture.room.ID, ReservationID: reservation.ReservationID})
	require.ErrorIs(t, err, ErrCafeReservationNotCancellable)
}

func TestCafeReservationCancelReentryStillHonorsBuyerCap(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_reservation_cancel_buyer_cap")
	now := time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC)
	fixture := newCafeRoomOrderFixture(t, ctx, client, now, 4)
	fixture.orderService.settings = cafePublicSettingsStub{enabled: true}
	_, err := client.GroupBuyRound.UpdateOneID(fixture.round.ID).SetMaxBuyers(1).Save(ctx)
	require.NoError(t, err)
	first, err := fixture.orderService.ReserveShares(ctx, CafeRoomReservationInput{UserID: fixture.user.ID, RoomID: fixture.room.ID, ShareCount: 1, AgreementAccepted: true})
	require.NoError(t, err)
	_, err = fixture.orderService.CancelReservation(ctx, CafeRoomReservationCancelInput{UserID: fixture.user.ID, RoomID: fixture.room.ID, ReservationID: first.ReservationID})
	require.NoError(t, err)
	other := createGroupBuyTestUser(t, ctx, client, "cafe-cancel-cap@example.com")
	_, err = fixture.orderService.ReserveShares(ctx, CafeRoomReservationInput{UserID: other.ID, RoomID: fixture.room.ID, ShareCount: 1, AgreementAccepted: true})
	require.NoError(t, err)
	_, err = fixture.orderService.ReserveShares(ctx, CafeRoomReservationInput{UserID: fixture.user.ID, RoomID: fixture.room.ID, ShareCount: 1, AgreementAccepted: true})
	require.ErrorIs(t, err, ErrCafeBuyerLimit)
}

func TestCafeMyRoomsShowsOnlyEligibleReservationID(t *testing.T) {
	ctx := context.Background()
	client := newGroupBuyTestClient(t, "cafe_my_rooms_cancel_projection")
	now := time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC)
	fixture := newCafeRoomOrderFixture(t, ctx, client, now, 3)
	fixture.orderService.settings = cafePublicSettingsStub{enabled: true}
	reservation, err := fixture.orderService.ReserveShares(ctx, CafeRoomReservationInput{UserID: fixture.user.ID, RoomID: fixture.room.ID, ShareCount: 1, AgreementAccepted: true})
	require.NoError(t, err)
	public := NewCafePublicService(client, cafePublicSettingsStub{enabled: true})
	public.now = func() time.Time { return now }
	items, _, err := public.MyRooms(ctx, fixture.user.ID, CafeMyRoomsListParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, 1, items[0].ReservedShares)
	require.NotNil(t, items[0].CancellableReservationID)
	require.Equal(t, reservation.ReservationID, *items[0].CancellableReservationID)

	_, err = fixture.orderService.CancelReservation(ctx, CafeRoomReservationCancelInput{UserID: fixture.user.ID, RoomID: fixture.room.ID, ReservationID: reservation.ReservationID})
	require.NoError(t, err)
	items, _, err = public.MyRooms(ctx, fixture.user.ID, CafeMyRoomsListParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestCafeReservationCancelNotificationKeyIsBatchSpecific(t *testing.T) {
	first := NewCafeReservationChangedSystemTicketNotification(1, 2, CafeRoundStatusReserving, map[string]any{"reservation_id": int64(3), "reserved_shares": 4})
	second := NewCafeReservationChangedSystemTicketNotification(1, 2, CafeRoundStatusReserving, map[string]any{"reservation_id": int64(4), "reserved_shares": 4})
	require.NotEqual(t, first.EventKey, second.EventKey)
	require.Contains(t, first.EventKey, "reservation:3")
}
