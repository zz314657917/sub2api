package service

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/caferoundmembership"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	apiKeyAccountBindingStatusExpired = "expired"
	cafeRoomExpiryBatchSize           = 50
)

var (
	ErrCafeExpiryUnavailable  = infraerrors.Conflict("CAFE_EXPIRY_UNAVAILABLE", "cafe room expiry service is unavailable")
	ErrCafeExpiryInconsistent = infraerrors.Conflict("CAFE_EXPIRY_INCONSISTENT", "cafe room expiry facts are inconsistent")
)

type cafeRoomAuthCacheInvalidator interface {
	InvalidateAuthCacheByKey(ctx context.Context, key string)
}

// CafeRoomExpiryService owns only active Cafe entitlement expiry. Refunds,
// activation compensation and account migration remain separate lifecycles.
type CafeRoomExpiryService struct {
	entClient        *dbent.Client
	cacheInvalidator cafeRoomAuthCacheInvalidator
	now              func() time.Time
}

func NewCafeRoomExpiryService(entClient *dbent.Client, cacheInvalidator cafeRoomAuthCacheInvalidator) *CafeRoomExpiryService {
	return &CafeRoomExpiryService{
		entClient:        entClient,
		cacheInvalidator: cacheInvalidator,
		now:              time.Now,
	}
}

// ExpireCafeRounds claims due Cafe Rounds individually. A malformed Room is
// rolled back without preventing other due Rooms from being reclaimed.
func (s *CafeRoomExpiryService) ExpireCafeRounds(ctx context.Context) (int, error) {
	if s == nil || s.entClient == nil || s.cacheInvalidator == nil {
		return 0, ErrCafeExpiryUnavailable
	}
	now := s.now()
	rounds, err := s.entClient.GroupBuyRound.Query().
		Where(
			groupbuyround.CafeRoomIDNotNil(),
			groupbuyround.StatusEQ(GroupBuyRoundStatusActive),
			groupbuyround.EntitlementExpiresAtNotNil(),
			groupbuyround.EntitlementExpiresAtLTE(now),
		).
		Order(dbent.Asc(groupbuyround.FieldEntitlementExpiresAt), dbent.Asc(groupbuyround.FieldID)).
		Limit(cafeRoomExpiryBatchSize).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list due cafe rounds: %w", err)
	}

	expired := 0
	var firstErr error
	for _, round := range rounds {
		keys, completed, err := s.expireRound(ctx, round.ID, now)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if !completed {
			continue
		}
		expired++
		for _, key := range keys {
			s.cacheInvalidator.InvalidateAuthCacheByKey(ctx, key)
		}
	}
	return expired, firstErr
}

func (s *CafeRoomExpiryService) expireRound(ctx context.Context, roundID int64, now time.Time) ([]string, bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin cafe expiry transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	round, err := s.roundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("lock cafe expiry round %d: %w", roundID, err)
	}
	if !isCafeRoundDue(round, now) {
		if err := tx.Commit(); err != nil {
			return nil, false, err
		}
		return nil, false, nil
	}
	if round.CafeFulfillmentVersion == "membership_share" {
		return s.expireMembershipRound(txCtx, tx, round, now)
	}
	if round.CafeFulfillmentVersion != "" && round.CafeFulfillmentVersion != "legacy_seat" {
		return nil, false, cafeExpiryInconsistent(round.ID, 0, 0, 0, "unsupported cafe fulfillment version")
	}
	plan, err := s.planForUpdate(tx.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(round.PlanID), groupbuyplan.DeletedAtIsNil())).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, false, cafeExpiryInconsistent(round.ID, 0, 0, 0, "round plan is missing")
		}
		return nil, false, fmt.Errorf("lock cafe expiry plan for round %d: %w", round.ID, err)
	}
	if plan.FulfillmentMode != CafeRoomFulfillmentMode || plan.TargetGroupID <= 0 {
		return nil, false, cafeExpiryInconsistent(round.ID, 0, 0, 0, "round plan is not a valid cafe plan")
	}

	seats, err := s.seatForUpdate(tx.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(round.ID))).
		Order(dbent.Asc(groupbuyseat.FieldID)).
		All(txCtx)
	if err != nil {
		return nil, false, fmt.Errorf("lock cafe expiry seats for round %d: %w", round.ID, err)
	}
	if len(seats) == 0 {
		return nil, false, cafeExpiryInconsistent(round.ID, 0, 0, 0, "round has no seats")
	}

	keys := make([]string, 0, len(seats))
	for _, seat := range seats {
		key, binding, err := s.loadExpiryFacts(txCtx, tx, round, seat, plan.TargetGroupID, now)
		if err != nil {
			return nil, false, err
		}
		if _, err := tx.APIKeyAccountBinding.UpdateOneID(binding.ID).
			SetStatus(apiKeyAccountBindingStatusExpired).
			SetUpdatedAt(now).
			Save(txCtx); err != nil {
			return nil, false, fmt.Errorf("expire cafe binding %d: %w", binding.ID, err)
		}
		if _, err := tx.APIKey.UpdateOneID(key.ID).
			SetStatus(StatusAPIKeyExpired).
			SetUpdatedAt(now).
			Save(txCtx); err != nil {
			return nil, false, fmt.Errorf("expire cafe managed key %d: %w", key.ID, err)
		}
		if _, err := tx.GroupBuySeat.UpdateOneID(seat.ID).
			SetStatus(GroupBuySeatStatusExpired).
			SetUpdatedAt(now).
			Save(txCtx); err != nil {
			return nil, false, fmt.Errorf("expire cafe seat %d: %w", seat.ID, err)
		}
		keys = append(keys, key.Key)
	}
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusCompleted).
		SetCompletedAt(now).
		SetCloseReason("cafe entitlement expired").
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return nil, false, fmt.Errorf("complete cafe round %d: %w", round.ID, err)
	}
	if err := tx.GroupBuyEvent.Create().
		SetPlanID(round.PlanID).
		SetRoundID(round.ID).
		SetEventType(groupBuyEventRoundCompleted).
		SetMessage("像素网吧房间权益已到期").
		SetMetadata(map[string]any{"cafe_room_id": *round.CafeRoomID, "seat_count": len(seats)}).
		Exec(txCtx); err != nil {
		return nil, false, fmt.Errorf("record cafe round completion: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit cafe expiry round %d: %w", round.ID, err)
	}
	return keys, true, nil
}

func (s *CafeRoomExpiryService) expireMembershipRound(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound, now time.Time) ([]string, bool, error) {
	if round.CafeRoomID == nil || round.AssignedAccountID == nil || round.TargetGroupIDSnapshot == nil || *round.TargetGroupIDSnapshot <= 0 || round.ActivatedAt == nil || round.EntitlementExpiresAt == nil {
		return nil, false, cafeExpiryInconsistent(round.ID, 0, 0, 0, "membership round entitlement snapshot is incomplete")
	}
	memberships, err := s.membershipForUpdate(tx.CafeRoundMembership.Query().Where(
		caferoundmembership.RoundIDEQ(round.ID),
		caferoundmembership.StatusEQ(GroupBuySeatStatusActive),
	)).Order(dbent.Asc(caferoundmembership.FieldID)).All(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("lock cafe expiry memberships for round %d: %w", round.ID, err)
	}
	if len(memberships) == 0 {
		return nil, false, cafeExpiryInconsistent(round.ID, 0, 0, 0, "membership round has no active memberships")
	}
	batches, err := s.seatForUpdate(tx.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(round.ID))).
		Order(dbent.Asc(groupbuyseat.FieldID)).All(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("lock cafe expiry payment batches for round %d: %w", round.ID, err)
	}
	if len(batches) == 0 {
		return nil, false, cafeExpiryInconsistent(round.ID, 0, 0, 0, "membership round has no payment batches")
	}
	batchesByMembership := make(map[int64][]*dbent.GroupBuySeat, len(memberships))
	for _, batch := range batches {
		if batch.MembershipID == nil {
			return nil, false, cafeExpiryInconsistent(round.ID, batch.ID, 0, 0, "payment batch is missing membership")
		}
		batchesByMembership[*batch.MembershipID] = append(batchesByMembership[*batch.MembershipID], batch)
	}

	keys := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		key, binding, err := s.loadMembershipExpiryFacts(ctx, tx, round, membership, batchesByMembership[membership.ID], *round.TargetGroupIDSnapshot, now)
		if err != nil {
			return nil, false, err
		}
		if _, err := tx.APIKeyAccountBinding.UpdateOneID(binding.ID).SetStatus(apiKeyAccountBindingStatusExpired).SetUpdatedAt(now).Save(ctx); err != nil {
			return nil, false, fmt.Errorf("expire cafe membership binding %d: %w", binding.ID, err)
		}
		if _, err := tx.APIKey.UpdateOneID(key.ID).SetStatus(StatusAPIKeyExpired).SetUpdatedAt(now).Save(ctx); err != nil {
			return nil, false, fmt.Errorf("expire cafe membership managed key %d: %w", key.ID, err)
		}
		if _, err := tx.CafeRoundMembership.UpdateOneID(membership.ID).SetStatus(GroupBuySeatStatusExpired).SetUpdatedAt(now).Save(ctx); err != nil {
			return nil, false, fmt.Errorf("expire cafe membership %d: %w", membership.ID, err)
		}
		for _, batch := range batchesByMembership[membership.ID] {
			if _, err := tx.GroupBuySeat.UpdateOneID(batch.ID).SetStatus(GroupBuySeatStatusExpired).SetUpdatedAt(now).Save(ctx); err != nil {
				return nil, false, fmt.Errorf("expire cafe payment batch %d: %w", batch.ID, err)
			}
		}
		delete(batchesByMembership, membership.ID)
		keys = append(keys, key.Key)
	}
	if len(batchesByMembership) != 0 {
		return nil, false, cafeExpiryInconsistent(round.ID, 0, 0, 0, "payment batch references a non-active membership")
	}
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusCompleted).
		SetCompletedAt(now).
		SetCloseReason("cafe entitlement expired").
		SetUpdatedAt(now).
		Save(ctx); err != nil {
		return nil, false, fmt.Errorf("complete cafe membership round %d: %w", round.ID, err)
	}
	if err := tx.GroupBuyEvent.Create().SetPlanID(round.PlanID).SetRoundID(round.ID).SetEventType(groupBuyEventRoundCompleted).
		SetMessage("像素网吧房间权益已到期").
		SetMetadata(map[string]any{"cafe_room_id": *round.CafeRoomID, "membership_count": len(memberships), "payment_batch_count": len(batches)}).
		Exec(ctx); err != nil {
		return nil, false, fmt.Errorf("record cafe membership round completion: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit cafe membership expiry round %d: %w", round.ID, err)
	}
	return keys, true, nil
}

func (s *CafeRoomExpiryService) loadMembershipExpiryFacts(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound, membership *dbent.CafeRoundMembership, batches []*dbent.GroupBuySeat, expectedGroupID int64, now time.Time) (*dbent.APIKey, *dbent.APIKeyAccountBinding, error) {
	if membership == nil || membership.Status != GroupBuySeatStatusActive || membership.PaidShares <= 0 || membership.ReservedShares != 0 || membership.BoundAPIKeyID == nil || membership.ActivatedAt == nil || membership.ExpiresAt == nil ||
		!membership.ActivatedAt.Equal(*round.ActivatedAt) || !membership.ExpiresAt.Equal(*round.EntitlementExpiresAt) || len(batches) == 0 {
		return nil, nil, cafeExpiryInconsistent(round.ID, 0, 0, 0, "membership does not match active cafe entitlement")
	}
	paidShares := 0
	for _, batch := range batches {
		if batch.Status != GroupBuySeatStatusActive || batch.UserID != membership.UserID || batch.MembershipID == nil || *batch.MembershipID != membership.ID || batch.BoundAPIKeyID == nil || *batch.BoundAPIKeyID != *membership.BoundAPIKeyID || batch.ExpiresAt == nil || !batch.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
			return nil, nil, cafeExpiryInconsistent(round.ID, batch.ID, 0, 0, "payment batch does not match active membership")
		}
		paidShares += batch.ShareCount
	}
	if paidShares != membership.PaidShares {
		return nil, nil, cafeExpiryInconsistent(round.ID, 0, 0, 0, "payment batch shares do not match membership")
	}
	binding, err := s.bindingForUpdate(tx.APIKeyAccountBinding.Query().Where(
		apikeyaccountbinding.MembershipIDEQ(membership.ID),
		apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive),
	)).Only(ctx)
	if err != nil {
		return nil, nil, cafeExpiryInconsistent(round.ID, 0, 0, 0, "active membership binding is missing or duplicated")
	}
	key, err := s.keyForUpdate(tx.APIKey.Query().Where(apikey.IDEQ(*membership.BoundAPIKeyID), apikey.DeletedAtIsNil())).Only(ctx)
	if err != nil {
		return nil, nil, cafeExpiryInconsistent(round.ID, 0, 0, binding.ID, "membership managed key is missing")
	}
	if binding.APIKeyID != key.ID || binding.UserID != membership.UserID || binding.GroupID != expectedGroupID || binding.CafeRoomID != *round.CafeRoomID || binding.RoundID != round.ID || binding.AccountID != *round.AssignedAccountID ||
		binding.SeatID != nil || binding.MembershipID == nil || *binding.MembershipID != membership.ID || !binding.StrictMode || binding.ExpiresAt.After(now) || !binding.StartsAt.Equal(*round.ActivatedAt) || !binding.ExpiresAt.Equal(*round.EntitlementExpiresAt) ||
		key.UserID != membership.UserID || key.GroupID == nil || *key.GroupID != expectedGroupID || key.ManagedSourceType != APIKeyManagedSourceCafeRoomMembership || key.ManagedSourceID == nil || *key.ManagedSourceID != membership.ID || key.ExpiresAt == nil || !key.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
		return nil, nil, cafeExpiryInconsistent(round.ID, 0, key.ID, binding.ID, "membership binding or managed key does not match round")
	}
	return key, binding, nil
}

func (s *CafeRoomExpiryService) loadExpiryFacts(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound, seat *dbent.GroupBuySeat, expectedGroupID int64, now time.Time) (*dbent.APIKey, *dbent.APIKeyAccountBinding, error) {
	if seat.Status != GroupBuySeatStatusActive || seat.BoundAPIKeyID == nil || round.CafeRoomID == nil || round.AssignedAccountID == nil ||
		round.EntitlementExpiresAt == nil || seat.ExpiresAt == nil || !seat.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
		return nil, nil, cafeExpiryInconsistent(round.ID, seat.ID, 0, 0, "seat does not match active cafe entitlement")
	}
	binding, err := s.bindingForUpdate(tx.APIKeyAccountBinding.Query().Where(
		apikeyaccountbinding.SeatIDEQ(seat.ID),
		apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive),
	)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, cafeExpiryInconsistent(round.ID, seat.ID, 0, 0, "active binding is missing")
		}
		return nil, nil, fmt.Errorf("lock cafe expiry binding for seat %d: %w", seat.ID, err)
	}
	key, err := s.keyForUpdate(tx.APIKey.Query().Where(apikey.IDEQ(*seat.BoundAPIKeyID), apikey.DeletedAtIsNil())).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, cafeExpiryInconsistent(round.ID, seat.ID, 0, binding.ID, "managed key is missing")
		}
		return nil, nil, fmt.Errorf("lock cafe expiry key for seat %d: %w", seat.ID, err)
	}
	if binding.APIKeyID != key.ID || binding.UserID != seat.UserID || binding.GroupID != expectedGroupID || binding.CafeRoomID != *round.CafeRoomID || binding.RoundID != round.ID ||
		binding.AccountID != *round.AssignedAccountID || !binding.StrictMode || binding.ExpiresAt.After(now) ||
		!binding.ExpiresAt.Equal(*round.EntitlementExpiresAt) || key.UserID != seat.UserID || key.GroupID == nil || *key.GroupID != binding.GroupID ||
		key.ManagedSourceType != APIKeyManagedSourceCafeRoomSeat || key.ManagedSourceID == nil || *key.ManagedSourceID != seat.ID ||
		key.ExpiresAt == nil || !key.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
		return nil, nil, cafeExpiryInconsistent(round.ID, seat.ID, key.ID, binding.ID, "binding or managed key does not match round")
	}
	return key, binding, nil
}

func (s *CafeRoomExpiryService) roundForUpdate(q *dbent.GroupBuyRoundQuery) *dbent.GroupBuyRoundQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomExpiryService) seatForUpdate(q *dbent.GroupBuySeatQuery) *dbent.GroupBuySeatQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomExpiryService) membershipForUpdate(q *dbent.CafeRoundMembershipQuery) *dbent.CafeRoundMembershipQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomExpiryService) planForUpdate(q *dbent.GroupBuyPlanQuery) *dbent.GroupBuyPlanQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomExpiryService) bindingForUpdate(q *dbent.APIKeyAccountBindingQuery) *dbent.APIKeyAccountBindingQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomExpiryService) keyForUpdate(q *dbent.APIKeyQuery) *dbent.APIKeyQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func isCafeRoundDue(round *dbent.GroupBuyRound, now time.Time) bool {
	return round != nil && round.CafeRoomID != nil && round.Status == GroupBuyRoundStatusActive &&
		round.EntitlementExpiresAt != nil && !round.EntitlementExpiresAt.After(now)
}

func cafeExpiryInconsistent(roundID, seatID, keyID, bindingID int64, reason string) error {
	return ErrCafeExpiryInconsistent.WithMetadata(map[string]string{
		"round_id":   fmt.Sprint(roundID),
		"seat_id":    fmt.Sprint(seatID),
		"key_id":     fmt.Sprint(keyID),
		"binding_id": fmt.Sprint(bindingID),
		"reason":     reason,
	})
}
