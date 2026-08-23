package service

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
	dbgroup "github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const apiKeyAccountBindingStatusActive = "active"

var (
	ErrCafeActivationPending = infraerrors.Conflict("CAFE_ACTIVATION_PENDING", "cafe room round is not ready for activation")
	ErrCafeActivationFailed  = infraerrors.Conflict("CAFE_ACTIVATION_FAILED", "cafe room activation could not be completed safely")
)

// CafeRoundActivation is the small post-payment boundary used by GroupBuyService.
// It deliberately owns only Cafe Round state and never invokes aggregate fulfillment.
type CafeRoundActivation interface {
	ActivateRound(ctx context.Context, roundID int64) error
}

// CafeRoomActivationService owns paid-full Room activation and its durable Key/
// Binding facts. Gateway pinning intentionally remains outside this service.
type CafeRoomActivationService struct {
	entClient   *dbent.Client
	apiKeySvc   *APIKeyService
	apiKeyRepo  APIKeyRepository
	now         func() time.Time
	generateKey func() (string, error)
}

func NewCafeRoomActivationService(entClient *dbent.Client, apiKeySvc *APIKeyService, apiKeyRepo APIKeyRepository) *CafeRoomActivationService {
	return &CafeRoomActivationService{
		entClient:  entClient,
		apiKeySvc:  apiKeySvc,
		apiKeyRepo: apiKeyRepo,
		now:        time.Now,
	}
}

// ActivateRound is safe to call after every Cafe payment callback. The first
// transaction claims durable activation state; the second transaction either
// completes all entitlement facts or leaves the Round retryable as activating.
func (s *CafeRoomActivationService) ActivateRound(ctx context.Context, roundID int64) error {
	if s == nil || s.entClient == nil || s.apiKeySvc == nil || s.apiKeyRepo == nil || roundID <= 0 {
		return ErrCafeActivationFailed
	}
	claimed, err := s.claimActivation(ctx, roundID)
	if err != nil || !claimed {
		return err
	}
	if err := s.completeActivation(ctx, roundID); err != nil {
		if eventErr := s.recordActivationFailure(ctx, roundID, err); eventErr != nil {
			return ErrCafeActivationFailed.WithCause(fmt.Errorf("%w; record activation failure: %v", err, eventErr))
		}
		return ErrCafeActivationFailed.WithCause(err)
	}
	return nil
}

func (s *CafeRoomActivationService) recordActivationFailure(ctx context.Context, roundID int64, cause error) error {
	if cause == nil {
		return nil
	}
	return s.entClient.GroupBuyEvent.Create().
		SetRoundID(roundID).
		SetEventType("activation_failed").
		SetMessage(cause.Error()).
		SetMetadata(map[string]any{"error": cause.Error()}).
		Exec(ctx)
}

func (s *CafeRoomActivationService) claimActivation(ctx context.Context, roundID int64) (bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin cafe activation claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	round, err := s.cafeRoundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return false, ErrGroupBuyRoundNotFound
		}
		return false, fmt.Errorf("lock cafe round for activation: %w", err)
	}
	if round.CafeRoomID == nil {
		return false, ErrCafeRoundLifecycleDeferred
	}
	switch round.Status {
	case GroupBuyRoundStatusActive:
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	case GroupBuyRoundStatusActivating:
		if round.ActivationToken == nil || round.ActivatedAt == nil || round.EntitlementExpiresAt == nil {
			return false, ErrCafeActivationFailed
		}
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return true, nil
	case GroupBuyRoundStatusOpen:
		// Continue below after checking the full paid Seat set.
	case GroupBuyRoundStatusFailed, GroupBuyRoundStatusCancelled:
		return false, ErrCafeActivationFailed.WithMetadata(map[string]string{"round_status": round.Status})
	default:
		return false, ErrCafeActivationFailed.WithMetadata(map[string]string{"round_status": round.Status})
	}

	seats, err := s.cafeSeatForUpdate(tx.GroupBuySeat.Query().Where(
		groupbuyseat.RoundIDEQ(round.ID),
		groupbuyseat.StatusEQ(GroupBuySeatStatusPaid),
	)).All(txCtx)
	if err != nil {
		return false, fmt.Errorf("lock paid cafe seats: %w", err)
	}
	if !cafeRoundPaidFull(round, seats) {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return false, ErrCafeActivationPending
	}
	plan, err := s.cafePlanForUpdate(tx.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(round.PlanID), groupbuyplan.DeletedAtIsNil())).Only(txCtx)
	if err != nil {
		return false, fmt.Errorf("lock cafe activation plan: %w", err)
	}
	if plan.ValidityDays <= 0 {
		return false, ErrCafePlanInvalid
	}
	token, err := s.newKeyValue()
	if err != nil {
		return false, fmt.Errorf("generate cafe activation token: %w", err)
	}
	now := s.now()
	expiresAt := now.AddDate(0, 0, plan.ValidityDays)
	if _, err := tx.GroupBuyRound.Update().
		Where(groupbuyround.IDEQ(round.ID), groupbuyround.StatusEQ(GroupBuyRoundStatusOpen)).
		SetStatus(GroupBuyRoundStatusActivating).
		SetActivationToken(token).
		SetActivatedAt(now).
		SetEntitlementExpiresAt(expiresAt).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("claim cafe round activation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit cafe activation claim: %w", err)
	}
	return true, nil
}

func (s *CafeRoomActivationService) completeActivation(ctx context.Context, roundID int64) error {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin cafe activation completion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	round, err := s.cafeRoundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).Only(txCtx)
	if err != nil {
		return fmt.Errorf("lock activating cafe round: %w", err)
	}
	if round.CafeRoomID == nil {
		return ErrCafeRoundLifecycleDeferred
	}
	if round.Status == GroupBuyRoundStatusActive {
		return tx.Commit()
	}
	if round.Status != GroupBuyRoundStatusActivating || round.AssignedAccountID == nil || round.ActivatedAt == nil || round.EntitlementExpiresAt == nil {
		return ErrCafeActivationFailed
	}

	room, plan, targetGroup, err := s.loadActivationFacts(txCtx, tx, round)
	if err != nil {
		return err
	}
	seats, err := s.cafeSeatForUpdate(tx.GroupBuySeat.Query().Where(
		groupbuyseat.RoundIDEQ(round.ID),
		groupbuyseat.StatusIn(GroupBuySeatStatusPaid, GroupBuySeatStatusActive),
	)).Order(dbent.Asc(groupbuyseat.FieldID)).All(txCtx)
	if err != nil {
		return fmt.Errorf("lock cafe seats for activation: %w", err)
	}
	if !cafeRoundPaidOrActiveFull(round, seats) {
		return ErrCafeActivationPending
	}

	managedKeys := make([]string, 0, len(seats))
	for _, seat := range seats {
		key, err := s.ensureManagedKey(txCtx, tx, room, plan, targetGroup.ID, seat, *round.EntitlementExpiresAt)
		if err != nil {
			return err
		}
		managedKeys = append(managedKeys, key.Key)
		if err := s.ensureBinding(txCtx, tx, key.ID, targetGroup.ID, *round.AssignedAccountID, room.ID, round, seat); err != nil {
			return err
		}
		if _, err := tx.GroupBuySeat.Update().
			Where(groupbuyseat.IDEQ(seat.ID), groupbuyseat.StatusIn(GroupBuySeatStatusPaid, GroupBuySeatStatusActive)).
			SetStatus(GroupBuySeatStatusActive).
			SetBoundAPIKeyID(key.ID).
			SetBoundAt(*round.ActivatedAt).
			SetActivatedAt(*round.ActivatedAt).
			SetExpiresAt(*round.EntitlementExpiresAt).
			SetUpdatedAt(s.now()).
			Save(txCtx); err != nil {
			return fmt.Errorf("activate cafe seat %d: %w", seat.ID, err)
		}
	}
	if _, err := tx.GroupBuyRound.Update().
		Where(groupbuyround.IDEQ(round.ID), groupbuyround.StatusEQ(GroupBuyRoundStatusActivating)).
		SetStatus(GroupBuyRoundStatusActive).
		SetClosedAt(*round.ActivatedAt).
		SetUpdatedAt(s.now()).
		Save(txCtx); err != nil {
		return fmt.Errorf("mark cafe round active: %w", err)
	}
	if err := tx.GroupBuyEvent.Create().
		SetPlanID(plan.ID).
		SetRoundID(round.ID).
		SetEventType(groupBuyEventRoundActivated).
		SetMessage("像素网吧房间已满员激活").
		SetMetadata(map[string]any{"seat_count": len(seats), "cafe_room_id": room.ID}).
		Exec(txCtx); err != nil {
		return fmt.Errorf("record cafe activation event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cafe activation completion: %w", err)
	}
	for _, key := range managedKeys {
		s.apiKeySvc.InvalidateAuthCacheByKey(ctx, key)
	}
	return nil
}

func (s *CafeRoomActivationService) loadActivationFacts(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound) (*dbent.CafeRoom, *dbent.GroupBuyPlan, *dbent.Group, error) {
	room, err := s.cafeRoomForUpdate(tx.CafeRoom.Query().Where(caferoom.IDEQ(*round.CafeRoomID), caferoom.DeletedAtIsNil())).Only(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lock cafe room for activation: %w", err)
	}
	plan, err := s.cafePlanForUpdate(tx.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(round.PlanID), groupbuyplan.DeletedAtIsNil())).Only(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lock cafe plan for activation: %w", err)
	}
	if room.Status != CafeRoomStatusEnabled || room.PlanID != plan.ID || room.AccountID == nil || *room.AccountID != *round.AssignedAccountID ||
		plan.Status != GroupBuyPlanStatusActive || plan.FulfillmentMode != CafeRoomFulfillmentMode || !plan.AutoCreateRoomKey || plan.ValidityDays <= 0 {
		return nil, nil, nil, ErrCafeActivationFailed
	}
	groupQuery := tx.Group.Query().Where(dbgroup.IDEQ(plan.TargetGroupID))
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		groupQuery = groupQuery.ForUpdate()
	}
	targetGroup, err := groupQuery.Only(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lock cafe target group: %w", err)
	}
	if targetGroup.Status != StatusActive || targetGroup.AccessMode != CafeRoomGroupAccessMode {
		return nil, nil, nil, ErrCafeGroupInvalid
	}
	accountQuery := tx.Account.Query().Where(account.IDEQ(*round.AssignedAccountID), account.StatusEQ(StatusActive)).WithGroups()
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		accountQuery = accountQuery.ForUpdate()
	}
	assignedAccount, err := accountQuery.Only(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lock cafe assigned account: %w", err)
	}
	belongsToGroup := false
	for _, accountGroup := range assignedAccount.Edges.Groups {
		if accountGroup.ID == targetGroup.ID {
			belongsToGroup = true
			break
		}
	}
	if assignedAccount.Platform != targetGroup.Platform || !belongsToGroup {
		return nil, nil, nil, ErrCafeAccountIncompatible
	}
	return room, plan, targetGroup, nil
}

func (s *CafeRoomActivationService) ensureManagedKey(ctx context.Context, tx *dbent.Tx, room *dbent.CafeRoom, plan *dbent.GroupBuyPlan, groupID int64, seat *dbent.GroupBuySeat, expiresAt time.Time) (*dbent.APIKey, error) {
	query := tx.APIKey.Query().Where(
		apikey.ManagedSourceTypeEQ(APIKeyManagedSourceCafeRoomSeat),
		apikey.ManagedSourceIDEQ(seat.ID),
		apikey.DeletedAtIsNil(),
	)
	key, err := query.Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return nil, fmt.Errorf("load managed cafe key: %w", err)
	}
	if dbent.IsNotFound(err) {
		rawKey, err := s.newKeyValue()
		if err != nil {
			return nil, fmt.Errorf("generate managed cafe key: %w", err)
		}
		managedSourceID := seat.ID
		created := &APIKey{
			UserID:              seat.UserID,
			Key:                 rawKey,
			Name:                cafeManagedKeyName(room, seat),
			GroupID:             &groupID,
			AccountPoolStrategy: AccountPoolStrategySharedOnly,
			Status:              StatusAPIKeyActive,
			Quota:               plan.RoomKeyQuotaUsd,
			ExpiresAt:           &expiresAt,
			RateLimit5h:         plan.RoomKeyRateLimit5h,
			RateLimit1d:         plan.RoomKeyRateLimit1d,
			RateLimit7d:         plan.RoomKeyRateLimit7d,
			ManagedSourceType:   APIKeyManagedSourceCafeRoomSeat,
			ManagedSourceID:     &managedSourceID,
		}
		if err := s.apiKeyRepo.Create(ctx, created); err != nil {
			// A concurrent retry can win the managed-source unique index.
			key, err = query.Only(ctx)
			if err != nil {
				return nil, fmt.Errorf("create managed cafe key: %w", err)
			}
		} else {
			key, err = tx.APIKey.Query().Where(apikey.IDEQ(created.ID)).Only(ctx)
			if err != nil {
				return nil, fmt.Errorf("reload managed cafe key: %w", err)
			}
		}
	}
	if key.UserID != seat.UserID || key.GroupID == nil || *key.GroupID != groupID || key.ManagedSourceType != APIKeyManagedSourceCafeRoomSeat || key.ManagedSourceID == nil || *key.ManagedSourceID != seat.ID {
		return nil, ErrCafeActivationFailed
	}
	if _, err := tx.APIKey.UpdateOneID(key.ID).
		SetStatus(StatusAPIKeyActive).
		SetGroupID(groupID).
		SetAccountPoolStrategy(AccountPoolStrategySharedOnly).
		SetMultiGroupRoutes([]domain.APIKeyMultiGroupRoute{}).
		SetQuota(plan.RoomKeyQuotaUsd).
		SetRateLimit5h(plan.RoomKeyRateLimit5h).
		SetRateLimit1d(plan.RoomKeyRateLimit1d).
		SetRateLimit7d(plan.RoomKeyRateLimit7d).
		SetExpiresAt(expiresAt).
		SetUpdatedAt(s.now()).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("apply managed cafe key policy: %w", err)
	}
	return key, nil
}

func (s *CafeRoomActivationService) ensureBinding(ctx context.Context, tx *dbent.Tx, keyID, groupID, accountID, roomID int64, round *dbent.GroupBuyRound, seat *dbent.GroupBuySeat) error {
	query := tx.APIKeyAccountBinding.Query().Where(
		apikeyaccountbinding.SeatIDEQ(seat.ID),
		apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive),
	)
	binding, err := query.Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return fmt.Errorf("load cafe key binding: %w", err)
	}
	if dbent.IsNotFound(err) {
		binding, err = tx.APIKeyAccountBinding.Create().
			SetAPIKeyID(keyID).
			SetUserID(seat.UserID).
			SetGroupID(groupID).
			SetAccountID(accountID).
			SetCafeRoomID(roomID).
			SetRoundID(round.ID).
			SetSeatID(seat.ID).
			SetStatus(apiKeyAccountBindingStatusActive).
			SetStrictMode(true).
			SetStartsAt(*round.ActivatedAt).
			SetExpiresAt(*round.EntitlementExpiresAt).
			Save(ctx)
		if err != nil {
			binding, err = query.Only(ctx)
			if err != nil {
				return fmt.Errorf("create cafe key binding: %w", err)
			}
		}
	}
	if binding.APIKeyID != keyID || binding.UserID != seat.UserID || binding.GroupID != groupID || binding.AccountID != accountID ||
		binding.CafeRoomID != roomID || binding.RoundID != round.ID || !binding.StrictMode ||
		!binding.StartsAt.Equal(*round.ActivatedAt) || !binding.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
		return ErrCafeActivationFailed
	}
	return nil
}

func (s *CafeRoomActivationService) cafeRoundForUpdate(q *dbent.GroupBuyRoundQuery) *dbent.GroupBuyRoundQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomActivationService) cafeRoomForUpdate(q *dbent.CafeRoomQuery) *dbent.CafeRoomQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomActivationService) cafeSeatForUpdate(q *dbent.GroupBuySeatQuery) *dbent.GroupBuySeatQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomActivationService) cafePlanForUpdate(q *dbent.GroupBuyPlanQuery) *dbent.GroupBuyPlanQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomActivationService) newKeyValue() (string, error) {
	if s.generateKey != nil {
		return s.generateKey()
	}
	return s.apiKeySvc.GenerateKey()
}

func cafeRoundPaidFull(round *dbent.GroupBuyRound, seats []*dbent.GroupBuySeat) bool {
	expectedSeats := cafeRoundSeatCount(round)
	return expectedSeats > 0 &&
		round.PaidSeats == expectedSeats &&
		round.PaidShares == round.TotalShares &&
		len(seats) == expectedSeats
}

func cafeRoundPaidOrActiveFull(round *dbent.GroupBuyRound, seats []*dbent.GroupBuySeat) bool {
	expectedSeats := cafeRoundSeatCount(round)
	if expectedSeats <= 0 || round.PaidSeats != expectedSeats || round.PaidShares != round.TotalShares || len(seats) != expectedSeats {
		return false
	}
	for _, seat := range seats {
		if seat.Status != GroupBuySeatStatusPaid && seat.Status != GroupBuySeatStatusActive {
			return false
		}
	}
	return true
}

func cafeManagedKeyName(room *dbent.CafeRoom, seat *dbent.GroupBuySeat) string {
	roomName := "Cafe Room"
	if room != nil && room.Name != "" {
		roomName = room.Name
	}
	seatNo := 0
	if seat != nil && seat.SeatNo != nil {
		seatNo = *seat.SeatNo
	}
	return fmt.Sprintf("%s / Seat %d", roomName, seatNo)
}
