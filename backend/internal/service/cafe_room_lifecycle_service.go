package service

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/caferoundmembership"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyevent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyrefund"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	cafeRoomLifecycleBatchSize = 50
	cafeActivationMaxAttempts  = 3
)

var ErrCafeLifecycleUnavailable = infraerrors.Conflict("CAFE_LIFECYCLE_UNAVAILABLE", "cafe room lifecycle service is unavailable")

// CafeRoomLifecycleService coordinates the Cafe-only paths that cannot enter
// the legacy GroupBuy lifecycle: unfilled timeout/refund and activating retry.
// Active entitlement expiry continues to be owned by CafeRoomExpiryService.
type CafeRoomLifecycleService struct {
	entClient    *dbent.Client
	groupBuy     *GroupBuyService
	activation   CafeRoundActivation
	activeExpiry *CafeRoomExpiryService
	now          func() time.Time
}

func NewCafeRoomLifecycleService(
	entClient *dbent.Client,
	groupBuy *GroupBuyService,
	activation CafeRoundActivation,
	activeExpiry *CafeRoomExpiryService,
) *CafeRoomLifecycleService {
	return &CafeRoomLifecycleService{
		entClient:    entClient,
		groupBuy:     groupBuy,
		activation:   activation,
		activeExpiry: activeExpiry,
		now:          time.Now,
	}
}

// RunCafeLifecycle keeps all Cafe transitions behind one ticker entry so the
// ordinary GroupBuy timeout and refund jobs cannot mutate Cafe rounds.
func (s *CafeRoomLifecycleService) RunCafeLifecycle(ctx context.Context) (int, error) {
	if s == nil || s.entClient == nil || s.groupBuy == nil || s.activation == nil || s.activeExpiry == nil {
		return 0, ErrCafeLifecycleUnavailable
	}

	updated := 0
	var firstErr error
	record := func(count int, err error) {
		updated += count
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	count, err := s.activeExpiry.ExpireCafeRounds(ctx)
	record(count, err)
	count, err = s.expireUnfilledCafeRounds(ctx, s.now())
	record(count, err)
	count, err = s.expireAwaitingAccountCafeRounds(ctx, s.now())
	record(count, err)
	count, err = s.processCafeRefunds(ctx)
	record(count, err)
	count, err = s.releaseFailedCafeOrderSeats(ctx)
	record(count, err)
	count, err = s.reconcileCafePendingProviderRefunds(ctx)
	record(count, err)
	count, err = s.finalizeCafeRefundedRounds(ctx, s.now())
	record(count, err)
	count, err = s.compensateCafeActivation(ctx, s.now())
	record(count, err)

	return updated, firstErr
}

func (s *CafeRoomLifecycleService) expireAwaitingAccountCafeRounds(ctx context.Context, now time.Time) (int, error) {
	rounds, err := s.entClient.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDNotNil(),
		groupbuyround.CafeFulfillmentVersionEQ("membership_share"),
		groupbuyround.StatusEQ(GroupBuyRoundStatusAwaitingAccount),
		groupbuyround.FulfillmentDeadlineAtNotNil(),
		groupbuyround.FulfillmentDeadlineAtLTE(now),
	).Order(dbent.Asc(groupbuyround.FieldFulfillmentDeadlineAt), dbent.Asc(groupbuyround.FieldID)).Limit(cafeRoomLifecycleBatchSize).All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list expired cafe fulfillment rounds: %w", err)
	}
	updated := 0
	var firstErr error
	for _, round := range rounds {
		changed, expireErr := s.expireAwaitingAccountCafeRound(ctx, round.ID, now)
		if changed {
			updated++
		}
		if expireErr != nil && firstErr == nil {
			firstErr = expireErr
		}
	}
	return updated, firstErr
}

func (s *CafeRoomLifecycleService) expireAwaitingAccountCafeRound(ctx context.Context, roundID int64, now time.Time) (bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin cafe fulfillment timeout: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	round, err := s.roundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("lock cafe fulfillment timeout round %d: %w", roundID, err)
	}
	if round.CafeFulfillmentVersion != "membership_share" || round.Status != GroupBuyRoundStatusAwaitingAccount || round.FulfillmentDeadlineAt == nil || now.Before(*round.FulfillmentDeadlineAt) {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}
	reason := "cafe account fulfillment deadline expired"
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).SetStatus(GroupBuyRoundStatusRefunding).SetClosedAt(now).SetCloseReason(reason).SetUpdatedAt(now).Save(txCtx); err != nil {
		return false, fmt.Errorf("mark cafe round refunding: %w", err)
	}
	if _, err := tx.GroupBuySeat.Update().Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusPaid)).SetStatus(GroupBuySeatStatusRefundPending).SetRefundNote(reason).SetUpdatedAt(now).Save(txCtx); err != nil {
		return false, fmt.Errorf("queue cafe fulfillment refunds: %w", err)
	}
	if _, err := tx.CafeRoundMembership.Update().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.PaidSharesGT(0)).SetStatus(GroupBuyRoundStatusRefunding).SetUpdatedAt(now).Save(txCtx); err != nil {
		return false, fmt.Errorf("mark cafe memberships refunding: %w", err)
	}
	if err := s.groupBuy.createEventTxStrict(txCtx, tx.Client(), &groupBuyEventInput{RoundID: &round.ID, PlanID: &round.PlanID, EventType: groupBuyEventRefundQueued, Message: "像素网吧成团后未在时限内配号，已进入全额退款", Metadata: map[string]any{"cafe_room_id": *round.CafeRoomID, "reason": reason}}); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit cafe fulfillment timeout: %w", err)
	}
	return true, nil
}

func (s *CafeRoomLifecycleService) expireUnfilledCafeRounds(ctx context.Context, now time.Time) (int, error) {
	rounds, err := s.entClient.GroupBuyRound.Query().
		Where(
			groupbuyround.CafeRoomIDNotNil(),
			groupbuyround.StatusIn(GroupBuyRoundStatusOpen, CafeRoundStatusReserving, CafeRoundStatusAwaitingPayment),
			groupbuyround.DeadlineAtLTE(now),
		).
		Order(dbent.Asc(groupbuyround.FieldDeadlineAt), dbent.Asc(groupbuyround.FieldID)).
		Limit(cafeRoomLifecycleBatchSize).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list timed out cafe rounds: %w", err)
	}

	closed := 0
	var firstErr error
	for _, round := range rounds {
		changed, closeErr := s.expireUnfilledCafeRound(ctx, round.ID, now)
		if closeErr != nil && firstErr == nil {
			firstErr = closeErr
		}
		if changed {
			closed++
		}
	}
	return closed, firstErr
}

func (s *CafeRoomLifecycleService) expireUnfilledCafeRound(ctx context.Context, roundID int64, now time.Time) (bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin cafe timeout transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	round, err := s.roundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("lock cafe timeout round %d: %w", roundID, err)
	}
	if !isCafeRoundTimedOutUnfilled(round, now) {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}

	seats, err := s.seatForUpdate(tx.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(round.ID))).All(txCtx)
	if err != nil {
		return false, fmt.Errorf("lock cafe timeout seats for round %d: %w", round.ID, err)
	}
	if len(seats) == 0 && round.PaidSeats > 0 {
		return false, fmt.Errorf("cafe timeout round %d has paid seats but no seat records", round.ID)
	}

	reason := "cafe round deadline reached before full seats"
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusFailed).
		SetClosedAt(now).
		SetCloseReason(reason).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("mark cafe round %d failed: %w", round.ID, err)
	}
	if _, err := tx.GroupBuySeat.Update().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusPaid)).
		SetStatus(GroupBuySeatStatusRefundPending).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("queue cafe paid seats for refund in round %d: %w", round.ID, err)
	}
	if _, err := tx.GroupBuySeat.Update().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusLocked)).
		SetStatus(GroupBuySeatStatusReleased).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("release cafe locked seats in round %d: %w", round.ID, err)
	}
	if round.CafeFulfillmentVersion == "membership_share" {
		if _, err := tx.CafeRoundMembership.Update().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.ReservedSharesGT(0)).SetReservedShares(0).SetUpdatedAt(now).Save(txCtx); err != nil {
			return false, fmt.Errorf("clear cafe reserved membership shares in round %d: %w", round.ID, err)
		}
		if _, err := tx.CafeRoundMembership.Update().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.PaidSharesEQ(0), caferoundmembership.StatusEQ(GroupBuySeatStatusLocked)).SetStatus(GroupBuySeatStatusReleased).SetUpdatedAt(now).Save(txCtx); err != nil {
			return false, fmt.Errorf("release empty cafe memberships in round %d: %w", round.ID, err)
		}
	}
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).SetReservedShares(0).SetReservedSeats(0).SetUpdatedAt(now).Save(txCtx); err != nil {
		return false, fmt.Errorf("clear reserved cafe round counters %d: %w", round.ID, err)
	}
	if err := s.groupBuy.createEventTxStrict(txCtx, tx.Client(), &groupBuyEventInput{
		RoundID:   &round.ID,
		PlanID:    &round.PlanID,
		EventType: groupBuyEventRoundFailed,
		Message:   "像素网吧房间未在截止时间前满员，已进入退款处理",
		Metadata: map[string]any{
			"cafe_room_id": *round.CafeRoomID,
			"reason":       reason,
		},
	}); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit cafe timeout round %d: %w", round.ID, err)
	}
	return true, nil
}

func (s *CafeRoomLifecycleService) processCafeRefunds(ctx context.Context) (int, error) {
	seats, err := s.entClient.GroupBuySeat.Query().
		Where(
			groupbuyseat.StatusIn(GroupBuySeatStatusRefundPending, GroupBuySeatStatusRefundProcessing),
			groupbuyseat.HasRoundWith(
				groupbuyround.CafeRoomIDNotNil(),
				groupbuyround.StatusIn(GroupBuyRoundStatusFailed, GroupBuyRoundStatusRefunding),
			),
		).
		WithPlan().
		WithOrder().
		Order(dbent.Asc(groupbuyseat.FieldUpdatedAt), dbent.Asc(groupbuyseat.FieldID)).
		Limit(cafeRoomLifecycleBatchSize).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list pending cafe refunds: %w", err)
	}

	processed := 0
	var firstErr error
	for _, seat := range seats {
		if seat.Edges.Plan == nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("cafe refund seat %d has no plan", seat.ID)
			}
			continue
		}
		_, refundErr := s.groupBuy.processSeatRefund(ctx, seat.Edges.Plan, seat)
		if refundErr != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("process cafe refund for seat %d: %w", seat.ID, refundErr)
			}
			continue
		}
		processed++
	}
	return processed, firstErr
}

// releaseFailedCafeOrderSeats is the durable fallback for a provider failure
// that occurs after the order/seat transaction has committed. The request path
// attempts the release immediately; this scan makes a transient cleanup error
// recoverable on the next lifecycle tick instead of leaving a ghost seat.
func (s *CafeRoomLifecycleService) releaseFailedCafeOrderSeats(ctx context.Context) (int, error) {
	seats, err := s.entClient.GroupBuySeat.Query().
		Where(
			groupbuyseat.StatusEQ(GroupBuySeatStatusLocked),
			groupbuyseat.HasRoundWith(groupbuyround.CafeRoomIDNotNil()),
			groupbuyseat.HasOrderWith(paymentorder.StatusIn(OrderStatusFailed, OrderStatusExpired)),
		).
		Order(dbent.Asc(groupbuyseat.FieldUpdatedAt), dbent.Asc(groupbuyseat.FieldID)).
		Limit(cafeRoomLifecycleBatchSize).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list failed cafe order seats: %w", err)
	}
	released := 0
	var firstErr error
	for _, seat := range seats {
		if seat.OrderID == nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("failed cafe order seat %d has no order", seat.ID)
			}
			continue
		}
		if err := s.groupBuy.ReleaseGroupBuySeatForOrder(ctx, *seat.OrderID, "payment provider create failed"); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("release failed cafe order seat %d: %w", seat.ID, err)
			}
			continue
		}
		released++
	}
	return released, firstErr
}

func (s *CafeRoomLifecycleService) reconcileCafePendingProviderRefunds(ctx context.Context) (int, error) {
	refunds, err := s.entClient.GroupBuyRefund.Query().
		Where(
			groupbuyrefund.StatusEQ(GroupBuyRefundStatusPendingProvider),
			groupbuyrefund.HasSeatWith(groupbuyseat.HasRoundWith(
				groupbuyround.CafeRoomIDNotNil(),
				groupbuyround.StatusIn(GroupBuyRoundStatusFailed, GroupBuyRoundStatusRefunding),
			)),
		).
		WithOrder().
		WithSeat().
		Order(dbent.Asc(groupbuyrefund.FieldUpdatedAt), dbent.Asc(groupbuyrefund.FieldID)).
		Limit(cafeRoomLifecycleBatchSize).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list pending provider cafe refunds: %w", err)
	}

	finalized := 0
	var firstErr error
	for _, refund := range refunds {
		changed, reconcileErr := s.groupBuy.reconcilePendingProviderRefund(ctx, refund)
		if changed {
			finalized++
		}
		if reconcileErr != nil && firstErr == nil {
			firstErr = fmt.Errorf("reconcile cafe provider refund %d: %w", refund.ID, reconcileErr)
		}
	}
	return finalized, firstErr
}

func (s *CafeRoomLifecycleService) finalizeCafeRefundedRounds(ctx context.Context, now time.Time) (int, error) {
	rounds, err := s.entClient.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDNotNil(),
		groupbuyround.CafeFulfillmentVersionEQ("membership_share"),
		groupbuyround.StatusEQ(GroupBuyRoundStatusRefunding),
	).Order(dbent.Asc(groupbuyround.FieldUpdatedAt), dbent.Asc(groupbuyround.FieldID)).Limit(cafeRoomLifecycleBatchSize).All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list refunding cafe rounds: %w", err)
	}
	finalized := 0
	var firstErr error
	for _, round := range rounds {
		changed, finalizeErr := s.finalizeCafeRefundedRound(ctx, round.ID, now)
		if changed {
			finalized++
		}
		if finalizeErr != nil && firstErr == nil {
			firstErr = finalizeErr
		}
	}
	return finalized, firstErr
}

func (s *CafeRoomLifecycleService) finalizeCafeRefundedRound(ctx context.Context, roundID int64, now time.Time) (bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin cafe refund finalization: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	round, err := s.roundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("lock refunding cafe round: %w", err)
	}
	if round.Status != GroupBuyRoundStatusRefunding || round.CafeFulfillmentVersion != "membership_share" {
		return false, tx.Commit()
	}
	seats, err := s.seatForUpdate(tx.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusIn(GroupBuySeatStatusRefundPending, GroupBuySeatStatusRefundProcessing, GroupBuySeatStatusRefunded)).WithRefunds()).All(txCtx)
	if err != nil {
		return false, fmt.Errorf("lock cafe refund batches: %w", err)
	}
	if len(seats) == 0 {
		return false, tx.Commit()
	}
	for _, seat := range seats {
		if seat.Status != GroupBuySeatStatusRefunded || len(seat.Edges.Refunds) != 1 || seat.Edges.Refunds[0].Status != GroupBuyRefundStatusSucceeded {
			return false, tx.Commit()
		}
	}
	if _, err := tx.CafeRoundMembership.Update().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.PaidSharesGT(0)).SetStatus(GroupBuyRoundStatusRefunded).SetUpdatedAt(now).Save(txCtx); err != nil {
		return false, fmt.Errorf("mark cafe memberships refunded: %w", err)
	}
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).SetStatus(GroupBuyRoundStatusRefunded).SetUpdatedAt(now).Save(txCtx); err != nil {
		return false, fmt.Errorf("mark cafe round refunded: %w", err)
	}
	if err := s.groupBuy.createEventTxStrict(txCtx, tx.Client(), &groupBuyEventInput{RoundID: &round.ID, PlanID: &round.PlanID, EventType: groupBuyEventRefundProcessed, Message: "像素网吧轮次已完成全额退款"}); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit cafe refund finalization: %w", err)
	}
	return true, nil
}

func (s *CafeRoomLifecycleService) compensateCafeActivation(ctx context.Context, now time.Time) (int, error) {
	rounds, err := s.entClient.GroupBuyRound.Query().
		Where(
			groupbuyround.CafeRoomIDNotNil(),
			groupbuyround.StatusIn(GroupBuyRoundStatusOpen, GroupBuyRoundStatusActivating),
		).
		Order(dbent.Asc(groupbuyround.FieldUpdatedAt), dbent.Asc(groupbuyround.FieldID)).
		Limit(cafeRoomLifecycleBatchSize).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list activating cafe rounds: %w", err)
	}

	retried := 0
	var firstErr error
	for _, round := range rounds {
		if round.CafeFulfillmentVersion == "membership_share" {
			if round.Status == GroupBuyRoundStatusOpen && isCafeRoundPaidFull(round) {
				if err := s.groupBuy.markCafeRoundAwaitingAccount(ctx, round.ID); err != nil {
					if firstErr == nil {
						firstErr = fmt.Errorf("recover cafe awaiting-account transition for round %d: %w", round.ID, err)
					}
					continue
				}
				retried++
			}
			continue
		}
		if round.Status == GroupBuyRoundStatusOpen {
			if !isCafeRoundPaidFull(round) {
				continue
			}
		} else if !isCafeRoundActivationRetryable(round, now) {
			changed, failErr := s.failCafeActivationRound(ctx, round.ID, now, "activation facts are missing or expired")
			if changed {
				retried++
			}
			if failErr != nil && firstErr == nil {
				firstErr = failErr
			} else if firstErr == nil {
				firstErr = ErrCafeActivationFailed.WithMetadata(map[string]string{"round_id": fmt.Sprint(round.ID), "reason": "activation facts are missing or expired"})
			}
			continue
		}
		failedAttempts, countErr := s.entClient.GroupBuyEvent.Query().Where(
			groupbuyevent.RoundIDEQ(round.ID),
			groupbuyevent.EventTypeEQ("activation_failed"),
		).Count(ctx)
		if countErr != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("count cafe activation failures for round %d: %w", round.ID, countErr)
			}
			continue
		}
		if failedAttempts >= cafeActivationMaxAttempts {
			changed, failErr := s.failCafeActivationRound(ctx, round.ID, now, "activation retry limit reached")
			if changed {
				retried++
			}
			if failErr != nil && firstErr == nil {
				firstErr = failErr
			} else if firstErr == nil {
				firstErr = ErrCafeActivationFailed.WithMetadata(map[string]string{"round_id": fmt.Sprint(round.ID), "reason": "activation retry limit reached"})
			}
			continue
		}
		if err := s.activation.ActivateRound(ctx, round.ID); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("retry cafe activation for round %d: %w", round.ID, err)
			}
			continue
		}
		retried++
	}
	return retried, firstErr
}

func (s *CafeRoomLifecycleService) failCafeActivationRound(ctx context.Context, roundID int64, now time.Time, reason string) (bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin cafe activation failure transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	round, err := s.roundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("lock cafe activation failure round %d: %w", roundID, err)
	}
	if round.Status != GroupBuyRoundStatusActivating || round.CafeRoomID == nil {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusFailed).
		SetClosedAt(now).
		SetCloseReason(reason).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("mark cafe activation round %d failed: %w", round.ID, err)
	}
	if _, err := tx.GroupBuySeat.Update().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusPaid)).
		SetStatus(GroupBuySeatStatusRefundPending).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("queue cafe activation refunds for round %d: %w", round.ID, err)
	}
	if _, err := tx.GroupBuySeat.Update().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusLocked)).
		SetStatus(GroupBuySeatStatusReleased).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("release cafe activation seats for round %d: %w", round.ID, err)
	}
	if err := s.groupBuy.createEventTxStrict(txCtx, tx.Client(), &groupBuyEventInput{
		RoundID:   &round.ID,
		PlanID:    &round.PlanID,
		EventType: groupBuyEventRoundFailed,
		Message:   "像素网吧激活失败，已进入退款处理",
		Metadata: map[string]any{
			"cafe_room_id": *round.CafeRoomID,
			"reason":       reason,
		},
	}); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit cafe activation failure round %d: %w", round.ID, err)
	}
	return true, nil
}

func (s *CafeRoomLifecycleService) roundForUpdate(q *dbent.GroupBuyRoundQuery) *dbent.GroupBuyRoundQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomLifecycleService) seatForUpdate(q *dbent.GroupBuySeatQuery) *dbent.GroupBuySeatQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func isCafeRoundTimedOutUnfilled(round *dbent.GroupBuyRound, now time.Time) bool {
	if round == nil || round.CafeRoomID == nil || (round.Status != GroupBuyRoundStatusOpen && round.Status != CafeRoundStatusReserving && round.Status != CafeRoundStatusAwaitingPayment) || round.DeadlineAt.After(now) {
		return false
	}
	return round.PaidSeats < cafeRoundSeatCount(round) || round.PaidShares < round.TotalShares
}

func isCafeRoundActivationRetryable(round *dbent.GroupBuyRound, now time.Time) bool {
	return round != nil && round.CafeRoomID != nil && round.Status == GroupBuyRoundStatusActivating &&
		round.ActivationToken != nil && round.ActivatedAt != nil && round.EntitlementExpiresAt != nil && round.EntitlementExpiresAt.After(now)
}

func isCafeRoundPaidFull(round *dbent.GroupBuyRound) bool {
	return round != nil && round.Status == GroupBuyRoundStatusOpen &&
		round.PaidSeats == cafeRoundSeatCount(round) && round.PaidShares == round.TotalShares
}
