package service

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
	"github.com/Wei-Shaw/sub2api/ent/caferoundmembership"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrCafeReservationNotCancellable = infraerrors.Conflict("CAFE_RESERVATION_NOT_CANCELLABLE", "cafe reservation is not cancellable")

type CafeRoomReservationCancelInput struct {
	UserID        int64
	RoomID        int64
	ReservationID int64
}

type CafeRoomReservationCancelResponse struct {
	RoomID         int64  `json:"room_id"`
	RoundID        int64  `json:"round_id"`
	ReservationID  int64  `json:"reservation_id"`
	Cancelled      bool   `json:"cancelled"`
	Status         string `json:"status"`
	ReservedShares int    `json:"reserved_shares"`
}

// CancelReservation releases exactly one orderless reservation batch. It uses
// the same room -> plan -> round -> seat lock order as reservation and payment
// creation, so a stale client cannot cancel a batch that has become payable.
func (s *CafeRoomOrderService) CancelReservation(ctx context.Context, input CafeRoomReservationCancelInput) (*CafeRoomReservationCancelResponse, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	if input.UserID <= 0 || input.RoomID <= 0 || input.ReservationID <= 0 {
		return nil, infraerrors.BadRequest("CAFE_INVALID_RESERVATION", "room and reservation are required")
	}
	if s.groupBuySvc == nil || s.entClient == nil {
		return nil, infraerrors.InternalServer("CAFE_ORDER_SERVICE_UNAVAILABLE", "cafe room order service is unavailable")
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin cafe reservation cancellation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	room, err := s.cafeRoomForUpdate(tx.CafeRoom.Query().Where(caferoom.IDEQ(input.RoomID), caferoom.DeletedAtIsNil())).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrCafeRoomNotFound
		}
		return nil, fmt.Errorf("lock cafe cancellation room: %w", err)
	}
	plan, err := s.groupBuySvc.groupBuyPlanForUpdate(tx.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(room.PlanID), groupbuyplan.DeletedAtIsNil()).WithTargetGroup()).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrCafePlanNotFound
		}
		return nil, fmt.Errorf("lock cafe cancellation plan: %w", err)
	}
	if !isCafeOperationalPlanEntity(plan) {
		return nil, ErrCafeReservationNotCancellable
	}
	round, err := s.groupBuySvc.groupBuyRoundForUpdate(tx.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDEQ(room.ID),
		groupbuyround.StatusIn(CafeRoundStatusOpen, CafeRoundStatusReserving, CafeRoundStatusAwaitingPayment),
	)).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrCafeReservationNotCancellable
		}
		return nil, fmt.Errorf("lock cafe cancellation round: %w", err)
	}
	if round.PlanID != plan.ID || round.CafeFulfillmentVersion != "membership_share" {
		return nil, ErrCafeReservationNotCancellable
	}

	// Read the immutable seat identity before locking the membership. The actual
	// state is reloaded with FOR UPDATE below after the membership lock, matching
	// ReserveShares' membership -> seat ordering and avoiding a lock inversion.
	candidate, err := tx.GroupBuySeat.Query().Where(
		groupbuyseat.IDEQ(input.ReservationID), groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.UserIDEQ(input.UserID),
	).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrCafeReservationNotCancellable
		}
		return nil, fmt.Errorf("lock cafe reservation batch: %w", err)
	}
	if candidate.MembershipID == nil {
		return nil, ErrCafeReservationNotCancellable
	}

	membershipQuery := tx.CafeRoundMembership.Query().Where(caferoundmembership.IDEQ(*candidate.MembershipID), caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.UserIDEQ(input.UserID))
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		membershipQuery = membershipQuery.ForUpdate()
	}
	membership, err := membershipQuery.Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrCafeReservationNotCancellable
		}
		return nil, fmt.Errorf("lock cafe cancellation membership: %w", err)
	}
	seat, err := s.groupBuySvc.groupBuySeatForUpdate(tx.GroupBuySeat.Query().Where(
		groupbuyseat.IDEQ(input.ReservationID), groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.UserIDEQ(input.UserID),
	)).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrCafeReservationNotCancellable
		}
		return nil, fmt.Errorf("relock cafe reservation batch: %w", err)
	}
	if seat.Status == GroupBuySeatStatusReleased && seat.OrderID == nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit idempotent cafe cancellation: %w", err)
		}
		return &CafeRoomReservationCancelResponse{RoomID: room.ID, RoundID: round.ID, ReservationID: seat.ID, Cancelled: false, Status: round.Status, ReservedShares: round.ReservedShares}, nil
	}
	if seat.Status != GroupBuySeatStatusLocked || seat.OrderID != nil || seat.PaidAt != nil || seat.MembershipID == nil || *seat.MembershipID != membership.ID || seat.ShareCount <= 0 {
		return nil, ErrCafeReservationNotCancellable
	}
	if membership.ReservedShares < seat.ShareCount {
		return nil, ErrCafeReservationNotCancellable
	}

	now := s.now()
	if err := tx.GroupBuySeat.UpdateOneID(seat.ID).SetStatus(GroupBuySeatStatusReleased).SetUpdatedAt(now).Exec(txCtx); err != nil {
		return nil, fmt.Errorf("release cafe reservation batch: %w", err)
	}
	if _, err := tx.CafeRoundMembership.UpdateOneID(membership.ID).AddReservedShares(-seat.ShareCount).SetUpdatedAt(now).Save(txCtx); err != nil {
		return nil, fmt.Errorf("release cafe membership shares: %w", err)
	}
	status := round.Status
	if round.Status == CafeRoundStatusAwaitingPayment && round.PaidShares+round.ReservedShares-seat.ShareCount < round.TotalShares {
		status = CafeRoundStatusReserving
	}
	round, err = tx.GroupBuyRound.UpdateOneID(round.ID).AddReservedShares(-seat.ShareCount).AddReservedSeats(-seat.ShareCount).SetStatus(status).SetUpdatedAt(now).Save(txCtx)
	if err != nil {
		return nil, fmt.Errorf("release cafe round shares: %w", err)
	}
	s.groupBuySvc.createEventTx(txCtx, tx.Client(), &groupBuyEventInput{PlanID: &plan.ID, RoundID: &round.ID, SeatID: &seat.ID, UserID: &input.UserID, EventType: groupBuyEventSharesReleased, Message: "用户取消像素网吧预约份额", Metadata: map[string]any{"share_count": seat.ShareCount, "reservation": true, "cancelled": true}})
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit cafe reservation cancellation: %w", err)
	}

	s.notifyCafeReservationCancelled(ctx, room.ID, round, seat.ID)
	return &CafeRoomReservationCancelResponse{RoomID: room.ID, RoundID: round.ID, ReservationID: seat.ID, Cancelled: true, Status: status, ReservedShares: round.ReservedShares}, nil
}

func (s *CafeRoomOrderService) notifyCafeReservationCancelled(ctx context.Context, roomID int64, round *dbent.GroupBuyRound, reservationID int64) {
	if s.paymentSvc == nil || s.paymentSvc.systemTicketSvc == nil || round == nil {
		return
	}
	members, err := s.entClient.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(round.ID)).All(ctx)
	if err != nil {
		return
	}
	for _, member := range members {
		event := NewCafeReservationChangedSystemTicketNotification(member.UserID, round.ID, round.Status, map[string]any{
			"room_id": roomID, "reserved_shares": round.ReservedShares, "total_shares": round.TotalShares,
			"reservation_id": reservationID, "cancelled": true,
		})
		s.paymentSvc.systemTicketSvc.NotifyEventBestEffort(ctx, "service.cafe", member.UserID, event)
	}
}
