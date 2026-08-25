package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/apikeyaccountbinding"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
	"github.com/Wei-Shaw/sub2api/ent/caferoundmembership"
	dbgroup "github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const cafeRoomMigrationBatchSize = 100

const (
	cafeBindingStatusMigrated                   = "migrated"
	cafeMigrationIssueRoomAccountMismatch       = "room_account_mismatch"
	cafeMigrationIssueRoundAccountMismatch      = "round_account_mismatch"
	cafeMigrationIssueSeatBindingMismatch       = "seat_binding_mismatch"
	cafeMigrationIssueMembershipBindingMismatch = "membership_binding_mismatch"
	cafeMigrationIssueManagedKeyMismatch        = "managed_key_mismatch"
	cafeMigrationIssueAccountMultipleLiveRounds = "account_multiple_live_rounds"
	cafeMigrationActionRetryActivation          = "retry_activation"
	cafeMigrationActionExpireOverdue            = "expire_overdue"
	cafeMigrationActionManualInvestigation      = "manual_investigate"
)

var (
	ErrCafeMigrationUnavailable    = infraerrors.Conflict("CAFE_MIGRATION_UNAVAILABLE", "cafe room migration service is unavailable")
	ErrCafeMigrationInvalid        = infraerrors.BadRequest("CAFE_MIGRATION_INVALID", "cafe room migration request is invalid")
	ErrCafeMigrationInconsistent   = infraerrors.Conflict("CAFE_MIGRATION_INCONSISTENT", "cafe room migration facts are inconsistent")
	ErrCafeMigrationAccountInvalid = infraerrors.Conflict("CAFE_MIGRATION_ACCOUNT_INVALID", "target account is not compatible with the cafe room")
)

type cafeMigrationCacheInvalidator interface {
	InvalidateAuthCacheByKey(ctx context.Context, key string)
}

// CafeRoomMigrationService contains the write boundary for emergency account
// migration and the read-only consistency audit. Public admin authorization is
// intentionally outside this service and remains a separate Sprint.
type CafeRoomMigrationService struct {
	entClient        *dbent.Client
	cacheInvalidator cafeMigrationCacheInvalidator
	now              func() time.Time
}

type CafeRoomMigrationResult struct {
	RoundID          int64 `json:"round_id"`
	RoomID           int64 `json:"room_id"`
	OldAccountID     int64 `json:"old_account_id"`
	NewAccountID     int64 `json:"new_account_id"`
	MigratedBindings int   `json:"migrated_bindings"`
	Noop             bool  `json:"noop"`
}

type CafeConsistencyIssue struct {
	Code      string `json:"code"`
	RoundID   int64  `json:"round_id,omitempty"`
	RoomID    int64  `json:"room_id,omitempty"`
	SeatID    int64  `json:"seat_id,omitempty"`
	KeyID     int64  `json:"key_id,omitempty"`
	BindingID int64  `json:"binding_id,omitempty"`
	Detail    string `json:"detail"`
}

type CafeConsistencyReport struct {
	Issues []CafeConsistencyIssue `json:"issues"`
}

type CafeRepairSuggestion struct {
	Action    string `json:"action"`
	RoundID   int64  `json:"round_id,omitempty"`
	RoomID    int64  `json:"room_id,omitempty"`
	SeatID    int64  `json:"seat_id,omitempty"`
	IssueCode string `json:"issue_code,omitempty"`
	Detail    string `json:"detail"`
}

type CafeDryRunRepairPlan struct {
	Report      *CafeConsistencyReport `json:"report"`
	Suggestions []CafeRepairSuggestion `json:"suggestions"`
}

func NewCafeRoomMigrationService(entClient *dbent.Client, cacheInvalidator cafeMigrationCacheInvalidator) *CafeRoomMigrationService {
	return &CafeRoomMigrationService{entClient: entClient, cacheInvalidator: cacheInvalidator, now: time.Now}
}

func (s *CafeRoomMigrationService) MigrateActiveRound(ctx context.Context, roundID, newAccountID int64, reason string) (*CafeRoomMigrationResult, error) {
	if s == nil || s.entClient == nil || s.cacheInvalidator == nil {
		return nil, ErrCafeMigrationUnavailable
	}
	if roundID <= 0 || newAccountID <= 0 {
		return nil, ErrCafeMigrationInvalid
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin cafe migration transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	round, err := s.roundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrGroupBuyRoundNotFound
		}
		return nil, fmt.Errorf("lock cafe migration round %d: %w", roundID, err)
	}
	if round.CafeRoomID == nil || round.Status != GroupBuyRoundStatusActive || round.AssignedAccountID == nil {
		return nil, ErrCafeMigrationInvalid.WithMetadata(map[string]string{"round_id": fmt.Sprint(roundID), "reason": "round is not an active cafe round"})
	}
	room, plan, _, targetAccount, err := s.loadMigrationFacts(txCtx, tx, round, newAccountID)
	if err != nil {
		return nil, err
	}
	var facts []cafeMigrationFact
	if round.CafeFulfillmentVersion == "membership_share" {
		facts, err = s.loadActiveMembershipMigrationFacts(txCtx, tx, round, cafeRoundTargetGroupID(round, plan), room.ID)
	} else {
		var seats []*dbent.GroupBuySeat
		seats, facts, err = s.loadActiveMigrationFacts(txCtx, tx, round, plan.TargetGroupID, room.ID)
		if err == nil && len(seats) != cafeRoundSeatCount(round) {
			err = cafeMigrationInconsistent(round.ID, room.ID, 0, 0, 0, "active seat count does not match cafe round")
		}
	}
	if err != nil {
		return nil, err
	}

	result := &CafeRoomMigrationResult{
		RoundID:      round.ID,
		RoomID:       room.ID,
		OldAccountID: *round.AssignedAccountID,
		NewAccountID: targetAccount.ID,
	}
	if targetAccount.ID == *round.AssignedAccountID {
		result.Noop = true
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit cafe migration no-op: %w", err)
		}
		return result, nil
	}

	now := s.now()
	keys := make([]string, 0, len(facts))
	for _, fact := range facts {
		old := fact.binding
		if _, err := tx.APIKeyAccountBinding.UpdateOneID(old.ID).
			SetStatus(cafeBindingStatusMigrated).
			SetMigratedAt(now).
			Save(txCtx); err != nil {
			return nil, fmt.Errorf("mark cafe binding %d migrated: %w", old.ID, err)
		}
		builder := tx.APIKeyAccountBinding.Create().
			SetAPIKeyID(old.APIKeyID).
			SetUserID(old.UserID).
			SetGroupID(old.GroupID).
			SetAccountID(targetAccount.ID).
			SetCafeRoomID(old.CafeRoomID).
			SetRoundID(old.RoundID).
			SetStatus(apiKeyAccountBindingStatusActive).
			SetStrictMode(old.StrictMode).
			SetStartsAt(old.StartsAt).
			SetExpiresAt(old.ExpiresAt)
		if fact.membership != nil {
			builder.SetMembershipID(fact.membership.ID)
		} else if fact.seat != nil {
			builder.SetSeatID(fact.seat.ID)
		} else {
			return nil, cafeMigrationInconsistent(round.ID, room.ID, 0, old.APIKeyID, old.ID, "migration binding has no owner")
		}
		created, err := builder.Save(txCtx)
		if err != nil {
			return nil, fmt.Errorf("create replacement cafe binding: %w", err)
		}
		if _, err := tx.APIKeyAccountBinding.UpdateOneID(old.ID).
			SetReplacedByBindingID(created.ID).
			SetUpdatedAt(now).
			Save(txCtx); err != nil {
			return nil, fmt.Errorf("link migrated cafe binding %d: %w", old.ID, err)
		}
		keys = append(keys, fact.key.Key)
		result.MigratedBindings++
	}
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).
		SetAssignedAccountID(targetAccount.ID).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return nil, fmt.Errorf("update cafe round %d account: %w", round.ID, err)
	}
	if round.CafeFulfillmentVersion != "membership_share" {
		if _, err := tx.CafeRoom.UpdateOneID(room.ID).
			SetAccountID(targetAccount.ID).
			SetUpdatedAt(now).
			Save(txCtx); err != nil {
			return nil, fmt.Errorf("update cafe room %d account: %w", room.ID, err)
		}
	}
	message := strings.TrimSpace(reason)
	if message == "" {
		message = "像素网吧包间紧急迁移账号"
	}
	if err := s.createMigrationEvent(txCtx, tx, round, room.ID, *round.AssignedAccountID, targetAccount.ID, result.MigratedBindings, message); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit cafe migration round %d: %w", round.ID, err)
	}
	for _, key := range keys {
		s.cacheInvalidator.InvalidateAuthCacheByKey(ctx, key)
	}
	return result, nil
}

type cafeMigrationFact struct {
	seat       *dbent.GroupBuySeat
	membership *dbent.CafeRoundMembership
	binding    *dbent.APIKeyAccountBinding
	key        *dbent.APIKey
}

func (s *CafeRoomMigrationService) loadMigrationFacts(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound, newAccountID int64) (*dbent.CafeRoom, *dbent.GroupBuyPlan, *dbent.Group, *dbent.Account, error) {
	roomQuery := tx.CafeRoom.Query().Where(caferoom.IDEQ(*round.CafeRoomID), caferoom.DeletedAtIsNil())
	room, err := s.roomForUpdate(roomQuery).Only(ctx)
	if err != nil {
		return nil, nil, nil, nil, cafeMigrationInconsistent(round.ID, *round.CafeRoomID, 0, 0, 0, "cafe room is missing")
	}
	if room.Status != CafeRoomStatusEnabled && room.Status != CafeRoomStatusMaintenance {
		return nil, nil, nil, nil, cafeMigrationInconsistent(round.ID, room.ID, 0, 0, 0, "room status is inconsistent")
	}
	if round.CafeFulfillmentVersion != "membership_share" && (room.AccountID == nil || *room.AccountID != *round.AssignedAccountID) {
		return nil, nil, nil, nil, cafeMigrationInconsistent(round.ID, room.ID, 0, 0, 0, "room account or status is inconsistent")
	}
	planQuery := tx.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(round.PlanID), groupbuyplan.DeletedAtIsNil())
	plan, err := s.planForUpdate(planQuery).Only(ctx)
	if err != nil {
		return nil, nil, nil, nil, cafeMigrationInconsistent(round.ID, room.ID, 0, 0, 0, "cafe plan is missing")
	}
	if plan.FulfillmentMode != CafeRoomFulfillmentMode || plan.ID != room.PlanID || round.AssignedAccountID == nil || cafeRoundTargetGroupID(round, plan) <= 0 {
		return nil, nil, nil, nil, cafeMigrationInconsistent(round.ID, room.ID, 0, 0, 0, "round, room and plan are inconsistent")
	}
	targetGroupID := cafeRoundTargetGroupID(round, plan)
	groupQuery := tx.Group.Query().Where(dbgroup.IDEQ(targetGroupID))
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		groupQuery = groupQuery.ForUpdate()
	}
	targetGroup, err := groupQuery.Only(ctx)
	if err != nil || targetGroup.Status != StatusActive || targetGroup.AccessMode != CafeRoomGroupAccessMode {
		return nil, nil, nil, nil, ErrCafeMigrationAccountInvalid
	}
	accountQuery := tx.Account.Query().Where(account.IDEQ(newAccountID), account.StatusEQ(StatusActive)).WithGroups()
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		accountQuery = accountQuery.ForUpdate()
	}
	targetAccount, err := accountQuery.Only(ctx)
	if err != nil {
		return nil, nil, nil, nil, ErrCafeMigrationAccountInvalid
	}
	belongs := false
	for _, g := range targetAccount.Edges.Groups {
		if g.ID == targetGroup.ID {
			belongs = true
			break
		}
	}
	if !belongs {
		return nil, nil, nil, nil, ErrCafeMigrationAccountInvalid
	}
	if round.CafeFulfillmentVersion == "membership_share" {
		if round.PlatformSnapshot == nil || targetAccount.Platform != *round.PlatformSnapshot || targetAccount.Platform != PlatformOpenAI ||
			!cafeAccountTierMatches(cafeRoundSubscriptionTier(round), normalizeCafeAccountPlanType(cafeAccountCredentialString(targetAccount, "plan_type"))) {
			return nil, nil, nil, nil, ErrCafeMigrationAccountInvalid
		}
	} else {
		occupiedQuery := tx.CafeRoom.Query().Where(
			caferoom.AccountIDEQ(newAccountID),
			caferoom.StatusIn(CafeRoomStatusEnabled, CafeRoomStatusMaintenance),
			caferoom.IDNEQ(room.ID),
			caferoom.DeletedAtIsNil(),
		)
		if s.entClient.Driver().Dialect() != dialect.SQLite {
			occupiedQuery = occupiedQuery.ForUpdate()
		}
		occupied, err := occupiedQuery.Exist(ctx)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("check target cafe account assignment: %w", err)
		}
		if occupied {
			return nil, nil, nil, nil, ErrCafeMigrationAccountInvalid
		}
	}
	occupied, err := tx.GroupBuyRound.Query().Where(
		groupbuyround.IDNEQ(round.ID),
		groupbuyround.AssignedAccountIDEQ(newAccountID),
		groupbuyround.StatusIn(GroupBuyRoundStatusActivating, GroupBuyRoundStatusActive),
	).Exist(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("check target cafe round account assignment: %w", err)
	}
	if occupied {
		return nil, nil, nil, nil, ErrCafeMigrationAccountInvalid
	}
	return room, plan, targetGroup, targetAccount, nil
}

func cafeRoundTargetGroupID(round *dbent.GroupBuyRound, plan *dbent.GroupBuyPlan) int64 {
	if round != nil && round.CafeFulfillmentVersion == "membership_share" && round.TargetGroupIDSnapshot != nil {
		return *round.TargetGroupIDSnapshot
	}
	if plan == nil {
		return 0
	}
	return plan.TargetGroupID
}

func (s *CafeRoomMigrationService) loadActiveMigrationFacts(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound, expectedGroupID, roomID int64) ([]*dbent.GroupBuySeat, []cafeMigrationFact, error) {
	seatQuery := tx.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusActive)).Order(dbent.Asc(groupbuyseat.FieldID))
	seats, err := s.seatForUpdate(seatQuery).All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("load active cafe migration seats: %w", err)
	}
	facts := make([]cafeMigrationFact, 0, len(seats))
	for _, seat := range seats {
		if seat.BoundAPIKeyID == nil || seat.ExpiresAt == nil || round.ActivatedAt == nil || round.EntitlementExpiresAt == nil || !seat.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
			return nil, nil, cafeMigrationInconsistent(round.ID, roomID, seat.ID, 0, 0, "active seat entitlement is inconsistent")
		}
		bindingQuery := tx.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.SeatIDEQ(seat.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive))
		binding, err := s.bindingForUpdate(bindingQuery).Only(ctx)
		if err != nil {
			return nil, nil, cafeMigrationInconsistent(round.ID, roomID, seat.ID, 0, 0, "active binding is missing or duplicated")
		}
		keyQuery := tx.APIKey.Query().Where(apikey.IDEQ(*seat.BoundAPIKeyID), apikey.DeletedAtIsNil())
		key, err := s.keyForUpdate(keyQuery).Only(ctx)
		if err != nil {
			return nil, nil, cafeMigrationInconsistent(round.ID, roomID, seat.ID, 0, binding.ID, "managed key is missing")
		}
		if binding.APIKeyID != key.ID || binding.UserID != seat.UserID || binding.GroupID != expectedGroupID || binding.AccountID != *round.AssignedAccountID || binding.CafeRoomID != roomID || binding.RoundID != round.ID || binding.SeatID == nil || *binding.SeatID != seat.ID || !binding.StrictMode || !binding.StartsAt.Equal(*round.ActivatedAt) || !binding.ExpiresAt.After(s.now()) || !binding.ExpiresAt.Equal(*round.EntitlementExpiresAt) || key.UserID != seat.UserID || key.GroupID == nil || *key.GroupID != expectedGroupID || key.ManagedSourceType != APIKeyManagedSourceCafeRoomSeat || key.ManagedSourceID == nil || *key.ManagedSourceID != seat.ID || key.ExpiresAt == nil || !key.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
			return nil, nil, cafeMigrationInconsistent(round.ID, roomID, seat.ID, key.ID, binding.ID, "binding or managed key does not match active round")
		}
		facts = append(facts, cafeMigrationFact{seat: seat, binding: binding, key: key})
	}
	return seats, facts, nil
}

func (s *CafeRoomMigrationService) loadActiveMembershipMigrationFacts(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound, expectedGroupID, roomID int64) ([]cafeMigrationFact, error) {
	membershipQuery := tx.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.StatusEQ(GroupBuySeatStatusActive)).Order(dbent.Asc(caferoundmembership.FieldID))
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		membershipQuery = membershipQuery.ForUpdate()
	}
	memberships, err := membershipQuery.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load active cafe migration memberships: %w", err)
	}
	if len(memberships) == 0 || round.ActivatedAt == nil || round.EntitlementExpiresAt == nil || expectedGroupID <= 0 {
		return nil, cafeMigrationInconsistent(round.ID, roomID, 0, 0, 0, "active membership entitlement is incomplete")
	}
	facts := make([]cafeMigrationFact, 0, len(memberships))
	paidShares := 0
	for _, membership := range memberships {
		paidShares += membership.PaidShares
		if membership.PaidShares <= 0 || membership.ReservedShares != 0 || membership.BoundAPIKeyID == nil || membership.ActivatedAt == nil || membership.ExpiresAt == nil || !membership.ActivatedAt.Equal(*round.ActivatedAt) || !membership.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
			return nil, cafeMigrationInconsistent(round.ID, roomID, 0, 0, 0, "active membership entitlement is inconsistent")
		}
		binding, err := s.bindingForUpdate(tx.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.MembershipIDEQ(membership.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive))).Only(ctx)
		if err != nil {
			return nil, cafeMigrationInconsistent(round.ID, roomID, 0, 0, 0, "active membership binding is missing or duplicated")
		}
		key, err := s.keyForUpdate(tx.APIKey.Query().Where(apikey.IDEQ(*membership.BoundAPIKeyID), apikey.DeletedAtIsNil())).Only(ctx)
		if err != nil {
			return nil, cafeMigrationInconsistent(round.ID, roomID, 0, 0, binding.ID, "membership managed key is missing")
		}
		if binding.APIKeyID != key.ID || binding.UserID != membership.UserID || binding.GroupID != expectedGroupID || binding.AccountID != *round.AssignedAccountID || binding.CafeRoomID != roomID || binding.RoundID != round.ID || binding.SeatID != nil || binding.MembershipID == nil || *binding.MembershipID != membership.ID || !binding.StrictMode || !binding.StartsAt.Equal(*round.ActivatedAt) || !binding.ExpiresAt.After(s.now()) || !binding.ExpiresAt.Equal(*round.EntitlementExpiresAt) || key.UserID != membership.UserID || key.GroupID == nil || *key.GroupID != expectedGroupID || key.ManagedSourceType != APIKeyManagedSourceCafeRoomMembership || key.ManagedSourceID == nil || *key.ManagedSourceID != membership.ID || key.ExpiresAt == nil || !key.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
			return nil, cafeMigrationInconsistent(round.ID, roomID, 0, key.ID, binding.ID, "membership binding or managed key does not match active round")
		}
		facts = append(facts, cafeMigrationFact{membership: membership, binding: binding, key: key})
	}
	if paidShares != round.PaidShares || paidShares != round.TotalShares {
		return nil, cafeMigrationInconsistent(round.ID, roomID, 0, 0, 0, "active membership shares do not match cafe round")
	}
	return facts, nil
}

func (s *CafeRoomMigrationService) createMigrationEvent(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound, roomID, oldAccountID, newAccountID int64, bindingCount int, reason string) error {
	return tx.GroupBuyEvent.Create().
		SetPlanID(round.PlanID).
		SetRoundID(round.ID).
		SetEventType("account_migrated").
		SetMessage(reason).
		SetMetadata(map[string]any{"cafe_room_id": roomID, "old_account_id": oldAccountID, "new_account_id": newAccountID, "binding_count": bindingCount}).
		Exec(ctx)
}

func (s *CafeRoomMigrationService) CheckConsistency(ctx context.Context) (*CafeConsistencyReport, error) {
	if s == nil || s.entClient == nil {
		return nil, ErrCafeMigrationUnavailable
	}
	report := &CafeConsistencyReport{Issues: []CafeConsistencyIssue{}}
	rounds, err := s.entClient.GroupBuyRound.Query().Where(groupbuyround.CafeRoomIDNotNil(), groupbuyround.StatusIn(GroupBuyRoundStatusOpen, GroupBuyRoundStatusActivating, GroupBuyRoundStatusActive)).Order(dbent.Asc(groupbuyround.FieldID)).Limit(cafeRoomMigrationBatchSize).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cafe consistency rounds: %w", err)
	}
	accountRounds := make(map[int64][]int64)
	for _, round := range rounds {
		if round.AssignedAccountID != nil {
			accountRounds[*round.AssignedAccountID] = append(accountRounds[*round.AssignedAccountID], round.ID)
		}
		room, roomErr := s.entClient.CafeRoom.Query().Where(caferoom.IDEQ(*round.CafeRoomID), caferoom.DeletedAtIsNil()).Only(ctx)
		if round.CafeFulfillmentVersion == "membership_share" {
			if roomErr != nil {
				report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueRoomAccountMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, Detail: "cafe room is missing"})
				continue
			}
			if round.Status == GroupBuyRoundStatusActive {
				if round.AssignedAccountID == nil || round.TargetGroupIDSnapshot == nil || *round.TargetGroupIDSnapshot <= 0 || round.ActivatedAt == nil || round.EntitlementExpiresAt == nil {
					report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueRoundAccountMismatch, RoundID: round.ID, RoomID: room.ID, Detail: "membership round activation snapshot is incomplete"})
					continue
				}
				if err := s.appendMembershipConsistencyIssues(ctx, report, round); err != nil {
					return nil, err
				}
			}
			continue
		}
		if roomErr != nil || room.AccountID == nil || round.AssignedAccountID == nil || *room.AccountID != *round.AssignedAccountID {
			report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueRoomAccountMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, Detail: "room account does not match round assignment"})
		}
		if round.Status != GroupBuyRoundStatusActive {
			continue
		}
		plan, planErr := s.entClient.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(round.PlanID), groupbuyplan.DeletedAtIsNil()).Only(ctx)
		if planErr != nil || plan.FulfillmentMode != CafeRoomFulfillmentMode || plan.TargetGroupID <= 0 {
			report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueRoundAccountMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, Detail: "cafe plan is missing"})
			continue
		}
		seats, seatErr := s.entClient.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusActive)).Order(dbent.Asc(groupbuyseat.FieldID)).All(ctx)
		if seatErr != nil {
			return nil, fmt.Errorf("list cafe consistency seats for round %d: %w", round.ID, seatErr)
		}
		if len(seats) != cafeRoundSeatCount(round) {
			report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueSeatBindingMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, Detail: "active seat count does not match cafe round"})
		}
		for _, seat := range seats {
			bindings, bindingErr := s.entClient.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.SeatIDEQ(seat.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).All(ctx)
			if bindingErr != nil {
				return nil, fmt.Errorf("list cafe consistency bindings for seat %d: %w", seat.ID, bindingErr)
			}
			if len(bindings) != 1 {
				report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueSeatBindingMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, SeatID: seat.ID, Detail: "active seat does not have exactly one active binding"})
				continue
			}
			binding := bindings[0]
			if seat.BoundAPIKeyID == nil || binding.APIKeyID != *seat.BoundAPIKeyID || binding.UserID != seat.UserID || binding.RoundID != round.ID || binding.CafeRoomID != *round.CafeRoomID || binding.SeatID == nil || *binding.SeatID != seat.ID || round.AssignedAccountID == nil || binding.AccountID != *round.AssignedAccountID || binding.GroupID != plan.TargetGroupID || !binding.StrictMode || round.ActivatedAt == nil || !binding.StartsAt.Equal(*round.ActivatedAt) || round.EntitlementExpiresAt == nil || !binding.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
				report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueSeatBindingMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, SeatID: seat.ID, BindingID: binding.ID, Detail: "active binding does not match seat, room or round"})
				continue
			}
			key, keyErr := s.entClient.APIKey.Query().Where(apikey.IDEQ(binding.APIKeyID), apikey.DeletedAtIsNil()).Only(ctx)
			if keyErr != nil || key.ManagedSourceType != APIKeyManagedSourceCafeRoomSeat || key.ManagedSourceID == nil || *key.ManagedSourceID != seat.ID || key.UserID != seat.UserID || key.GroupID == nil || *key.GroupID != plan.TargetGroupID || key.ExpiresAt == nil || round.EntitlementExpiresAt == nil || !key.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
				report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueManagedKeyMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, SeatID: seat.ID, KeyID: binding.APIKeyID, BindingID: binding.ID, Detail: "managed key does not match active cafe binding"})
			}
		}
	}
	for accountID, roundIDs := range accountRounds {
		if len(roundIDs) > 1 {
			sort.Slice(roundIDs, func(i, j int) bool { return roundIDs[i] < roundIDs[j] })
			report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueAccountMultipleLiveRounds, Detail: fmt.Sprintf("account %d is assigned to multiple live cafe rounds", accountID)})
		}
	}
	return report, nil
}

func (s *CafeRoomMigrationService) appendMembershipConsistencyIssues(ctx context.Context, report *CafeConsistencyReport, round *dbent.GroupBuyRound) error {
	memberships, err := s.entClient.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.StatusEQ(GroupBuySeatStatusActive)).Order(dbent.Asc(caferoundmembership.FieldID)).All(ctx)
	if err != nil {
		return fmt.Errorf("list cafe consistency memberships for round %d: %w", round.ID, err)
	}
	paidShares := 0
	for _, membership := range memberships {
		paidShares += membership.PaidShares
		bindings, bindingErr := s.entClient.APIKeyAccountBinding.Query().Where(apikeyaccountbinding.MembershipIDEQ(membership.ID), apikeyaccountbinding.StatusEQ(apiKeyAccountBindingStatusActive)).All(ctx)
		if bindingErr != nil {
			return fmt.Errorf("list cafe consistency bindings for membership %d: %w", membership.ID, bindingErr)
		}
		if len(bindings) != 1 {
			report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueMembershipBindingMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, Detail: fmt.Sprintf("active membership %d does not have exactly one active binding", membership.ID)})
			continue
		}
		binding := bindings[0]
		if membership.PaidShares <= 0 || membership.ReservedShares != 0 || membership.BoundAPIKeyID == nil || membership.ActivatedAt == nil || !membership.ActivatedAt.Equal(*round.ActivatedAt) || membership.ExpiresAt == nil || !membership.ExpiresAt.Equal(*round.EntitlementExpiresAt) ||
			binding.APIKeyID != *membership.BoundAPIKeyID || binding.UserID != membership.UserID || binding.RoundID != round.ID || binding.CafeRoomID != *round.CafeRoomID || binding.SeatID != nil || binding.MembershipID == nil || *binding.MembershipID != membership.ID || binding.AccountID != *round.AssignedAccountID || binding.GroupID != *round.TargetGroupIDSnapshot || !binding.StrictMode || !binding.StartsAt.Equal(*round.ActivatedAt) || !binding.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
			report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueMembershipBindingMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, BindingID: binding.ID, Detail: fmt.Sprintf("active binding does not match membership %d or round", membership.ID)})
			continue
		}
		key, keyErr := s.entClient.APIKey.Query().Where(apikey.IDEQ(binding.APIKeyID), apikey.DeletedAtIsNil()).Only(ctx)
		if keyErr != nil || key.ManagedSourceType != APIKeyManagedSourceCafeRoomMembership || key.ManagedSourceID == nil || *key.ManagedSourceID != membership.ID || key.UserID != membership.UserID || key.GroupID == nil || *key.GroupID != *round.TargetGroupIDSnapshot || key.ExpiresAt == nil || !key.ExpiresAt.Equal(*round.EntitlementExpiresAt) {
			report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueManagedKeyMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, KeyID: binding.APIKeyID, BindingID: binding.ID, Detail: fmt.Sprintf("managed key does not match cafe membership %d", membership.ID)})
		}
	}
	if len(memberships) == 0 || paidShares != round.PaidShares || paidShares != round.TotalShares {
		report.Issues = append(report.Issues, CafeConsistencyIssue{Code: cafeMigrationIssueMembershipBindingMismatch, RoundID: round.ID, RoomID: *round.CafeRoomID, Detail: "active membership shares do not match cafe round"})
	}
	return nil
}

func (s *CafeRoomMigrationService) PlanDryRunRepair(ctx context.Context) (*CafeDryRunRepairPlan, error) {
	report, err := s.CheckConsistency(ctx)
	if err != nil {
		return nil, err
	}
	plan := &CafeDryRunRepairPlan{Report: report, Suggestions: []CafeRepairSuggestion{}}
	rounds, err := s.entClient.GroupBuyRound.Query().Where(groupbuyround.CafeRoomIDNotNil(), groupbuyround.StatusIn(GroupBuyRoundStatusActivating, GroupBuyRoundStatusActive)).Order(dbent.Asc(groupbuyround.FieldID)).Limit(cafeRoomMigrationBatchSize).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cafe dry-run rounds: %w", err)
	}
	now := s.now()
	for _, round := range rounds {
		action := ""
		switch {
		case round.Status == GroupBuyRoundStatusActivating:
			action = cafeMigrationActionRetryActivation
		case round.Status == GroupBuyRoundStatusActive && round.EntitlementExpiresAt != nil && !round.EntitlementExpiresAt.After(now):
			action = cafeMigrationActionExpireOverdue
		default:
			continue
		}
		plan.Suggestions = append(plan.Suggestions, CafeRepairSuggestion{Action: action, RoundID: round.ID, RoomID: cafeMigrationValueOrZero(round.CafeRoomID), Detail: "dry-run only; no database or cache mutation"})
	}
	for _, issue := range report.Issues {
		plan.Suggestions = append(plan.Suggestions, CafeRepairSuggestion{Action: cafeMigrationActionManualInvestigation, RoundID: issue.RoundID, RoomID: issue.RoomID, SeatID: issue.SeatID, IssueCode: issue.Code, Detail: issue.Detail})
	}
	return plan, nil
}

func (s *CafeRoomMigrationService) roundForUpdate(q *dbent.GroupBuyRoundQuery) *dbent.GroupBuyRoundQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomMigrationService) roomForUpdate(q *dbent.CafeRoomQuery) *dbent.CafeRoomQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomMigrationService) planForUpdate(q *dbent.GroupBuyPlanQuery) *dbent.GroupBuyPlanQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomMigrationService) seatForUpdate(q *dbent.GroupBuySeatQuery) *dbent.GroupBuySeatQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomMigrationService) bindingForUpdate(q *dbent.APIKeyAccountBindingQuery) *dbent.APIKeyAccountBindingQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomMigrationService) keyForUpdate(q *dbent.APIKeyQuery) *dbent.APIKeyQuery {
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func cafeMigrationInconsistent(roundID, roomID, seatID, keyID, bindingID int64, detail string) error {
	return ErrCafeMigrationInconsistent.WithMetadata(map[string]string{"round_id": fmt.Sprint(roundID), "room_id": fmt.Sprint(roomID), "seat_id": fmt.Sprint(seatID), "key_id": fmt.Sprint(keyID), "binding_id": fmt.Sprint(bindingID), "detail": detail})
}

func cafeMigrationValueOrZero(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
