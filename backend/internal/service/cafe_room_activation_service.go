package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
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
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const apiKeyAccountBindingStatusActive = "active"

var (
	ErrCafeActivationPending = infraerrors.Conflict("CAFE_ACTIVATION_PENDING", "cafe room round is not ready for activation")
	ErrCafeActivationFailed  = infraerrors.Conflict("CAFE_ACTIVATION_FAILED", "cafe room activation could not be completed safely")
)

type CafePendingRound struct {
	ID                    int64      `json:"id"`
	Status                string     `json:"status"`
	RoomID                int64      `json:"room_id"`
	RoomCode              string     `json:"room_code"`
	RoomName              string     `json:"room_name"`
	SubscriptionTier      string     `json:"subscription_tier"`
	PaidShares            int        `json:"paid_shares"`
	TotalShares           int        `json:"total_shares"`
	JoinedBuyers          int        `json:"joined_buyers"`
	MaxBuyers             int        `json:"max_buyers"`
	PaidFullAt            *time.Time `json:"paid_full_at,omitempty"`
	FulfillmentDeadlineAt *time.Time `json:"fulfillment_deadline_at,omitempty"`
}

type CafePendingRoundParams struct {
	Page     int
	PageSize int
	Search   string
}

type CafeRoundAccountOption struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	Status      string `json:"status"`
	PlanType    string `json:"plan_type"`
	EmailMasked string `json:"email_masked,omitempty"`
}

type CafeRoundAccountOptionParams struct {
	Page     int
	PageSize int
	Search   string
}

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

func (s *CafeRoomActivationService) ListPendingRounds(ctx context.Context, params CafePendingRoundParams) ([]CafePendingRound, *pagination.PaginationResult, error) {
	if s == nil || s.entClient == nil {
		return nil, nil, ErrCafeActivationFailed
	}
	page, pageSize := normalizeCafeFulfillmentPagination(params.Page, params.PageSize)
	q := s.entClient.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDNotNil(),
		groupbuyround.CafeFulfillmentVersionEQ("membership_share"),
		groupbuyround.StatusEQ(GroupBuyRoundStatusAwaitingAccount),
	)
	if search := strings.TrimSpace(params.Search); search != "" {
		q = q.Where(groupbuyround.HasCafeRoomWith(caferoom.Or(caferoom.CodeContainsFold(search), caferoom.NameContainsFold(search))))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count pending cafe rounds: %w", err)
	}
	rounds, err := q.WithCafeRoom().WithCafeMemberships().
		Order(dbent.Asc(groupbuyround.FieldFulfillmentDeadlineAt), dbent.Asc(groupbuyround.FieldID)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list pending cafe rounds: %w", err)
	}
	items := make([]CafePendingRound, 0, len(rounds))
	for _, round := range rounds {
		items = append(items, cafePendingRoundFromEntity(round))
	}
	return items, cafeFulfillmentPagination(page, pageSize, total), nil
}

func (s *CafeRoomActivationService) ListRoundAccountOptions(ctx context.Context, roundID int64, params CafeRoundAccountOptionParams) ([]CafeRoundAccountOption, *pagination.PaginationResult, error) {
	if s == nil || s.entClient == nil || roundID <= 0 {
		return nil, nil, ErrCafeRoundNotAwaitingAccount
	}
	round, err := s.entClient.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, ErrGroupBuyRoundNotFound
		}
		return nil, nil, fmt.Errorf("load pending cafe round: %w", err)
	}
	if err := validateCafeRoundAwaitingAccount(round, s.now()); err != nil {
		return nil, nil, err
	}
	if round.TargetGroupIDSnapshot == nil || *round.TargetGroupIDSnapshot <= 0 || round.PlatformSnapshot == nil {
		return nil, nil, ErrCafeActivationFailed
	}
	q := s.entClient.Account.Query().Where(
		account.DeletedAtIsNil(),
		account.StatusEQ(StatusActive),
		account.PlatformEQ(*round.PlatformSnapshot),
		account.HasGroupsWith(dbgroup.IDEQ(*round.TargetGroupIDSnapshot), dbgroup.DeletedAtIsNil()),
	).WithGroups()
	if search := strings.TrimSpace(params.Search); search != "" {
		q = q.Where(account.Or(
			account.NameContainsFold(search),
			account.PlatformContainsFold(search),
			func(selector *sql.Selector) {
				selector.Where(sqljson.StringContains(account.FieldCredentials, search, sqljson.Path("email")))
			},
		))
	}
	accounts, err := q.All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list cafe account options: %w", err)
	}
	inUse, err := s.liveCafeAccountIDs(ctx, round.ID)
	if err != nil {
		return nil, nil, err
	}
	filtered := make([]CafeRoundAccountOption, 0, len(accounts))
	search := strings.ToLower(strings.TrimSpace(params.Search))
	for _, item := range accounts {
		if _, used := inUse[item.ID]; used {
			continue
		}
		planType := normalizeCafeAccountPlanType(cafeAccountCredentialString(item, "plan_type"))
		if !cafeAccountTierMatches(cafeRoundSubscriptionTier(round), planType) {
			continue
		}
		email := cafeAccountCredentialString(item, "email")
		if search != "" && !strings.Contains(strings.ToLower(item.Name), search) && !strings.Contains(strings.ToLower(item.Platform), search) && !strings.Contains(strings.ToLower(email), search) {
			continue
		}
		filtered = append(filtered, CafeRoundAccountOption{ID: item.ID, Name: item.Name, Platform: item.Platform, Status: item.Status, PlanType: planType, EmailMasked: MaskEmail(email)})
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })
	page, pageSize := normalizeCafeFulfillmentPagination(params.Page, params.PageSize)
	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return filtered[start:end], cafeFulfillmentPagination(page, pageSize, total), nil
}

func (s *CafeRoomActivationService) AssignAccountAndActivateRound(ctx context.Context, roundID, accountID int64) (*CafePendingRound, error) {
	if s == nil || s.entClient == nil || s.apiKeySvc == nil || s.apiKeyRepo == nil || roundID <= 0 || accountID <= 0 {
		return nil, ErrCafeActivationFailed
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin cafe account assignment: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	now := s.now()
	round, err := s.cafeRoundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))).WithCafeRoom().WithCafeMemberships().Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrGroupBuyRoundNotFound
		}
		return nil, fmt.Errorf("lock pending cafe round: %w", err)
	}
	if round.Status == GroupBuyRoundStatusActive && round.AssignedAccountID != nil && *round.AssignedAccountID == accountID {
		item := cafePendingRoundFromEntity(round)
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &item, nil
	}
	if err := validateCafeRoundAwaitingAccount(round, now); err != nil {
		return nil, err
	}
	if round.Edges.CafeRoom == nil || round.TargetGroupIDSnapshot == nil || round.PlatformSnapshot == nil || round.ValidityDaysSnapshot == nil || *round.ValidityDaysSnapshot <= 0 {
		return nil, ErrCafeActivationFailed
	}
	accountQuery := tx.Account.Query().Where(account.IDEQ(accountID), account.DeletedAtIsNil()).WithGroups()
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		accountQuery = accountQuery.ForUpdate()
	}
	assigned, err := accountQuery.Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrCafeAccountIncompatible
		}
		return nil, fmt.Errorf("lock cafe assigned account: %w", err)
	}
	if assigned.Status != StatusActive || assigned.Platform != *round.PlatformSnapshot || assigned.Platform != PlatformOpenAI || !cafeAccountBelongsToGroup(assigned, *round.TargetGroupIDSnapshot) {
		return nil, ErrCafeAccountIncompatible
	}
	planType := normalizeCafeAccountPlanType(cafeAccountCredentialString(assigned, "plan_type"))
	if !cafeAccountTierMatches(cafeRoundSubscriptionTier(round), planType) {
		return nil, ErrCafeAccountTierMismatch
	}
	used, err := tx.GroupBuyRound.Query().Where(
		groupbuyround.IDNEQ(round.ID),
		groupbuyround.AssignedAccountIDEQ(accountID),
		groupbuyround.StatusIn(GroupBuyRoundStatusActivating, GroupBuyRoundStatusActive),
	).Exist(txCtx)
	if err != nil {
		return nil, fmt.Errorf("check cafe account assignment: %w", err)
	}
	if used {
		return nil, ErrCafeAccountAlreadyInUse
	}
	memberships := round.Edges.CafeMemberships
	paidShares := 0
	for _, membership := range memberships {
		if membership.PaidShares > 0 {
			paidShares += membership.PaidShares
		}
	}
	if paidShares != round.TotalShares || len(memberships) == 0 {
		return nil, ErrCafeActivationPending
	}
	projection := cafePendingRoundFromEntity(round)
	room := round.Edges.CafeRoom
	token, err := s.newKeyValue()
	if err != nil {
		return nil, fmt.Errorf("generate cafe activation token: %w", err)
	}
	expiresAt := now.AddDate(0, 0, *round.ValidityDaysSnapshot)
	round, err = tx.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusActivating).
		SetAssignedAccountID(accountID).
		SetActivationToken(token).
		ClearActivatedAt().
		ClearEntitlementExpiresAt().
		SetUpdatedAt(now).
		Save(txCtx)
	if err != nil {
		translated := translateCafeAccountAssignmentError(err)
		if infraerrors.Reason(translated) == "CAFE_ACCOUNT_ALREADY_IN_USE" {
			return nil, translated
		}
		return nil, ErrCafeActivationFailed.WithCause(translated)
	}
	managedKeys := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		if membership.PaidShares <= 0 {
			continue
		}
		key, err := s.createMembershipManagedKey(txCtx, tx, round, room, membership, *round.TargetGroupIDSnapshot, expiresAt)
		if err != nil {
			return nil, ErrCafeActivationFailed.WithCause(err)
		}
		managedKeys = append(managedKeys, key.Key)
		if _, err := tx.APIKeyAccountBinding.Create().
			SetAPIKeyID(key.ID).
			SetUserID(membership.UserID).
			SetGroupID(*round.TargetGroupIDSnapshot).
			SetAccountID(accountID).
			SetCafeRoomID(*round.CafeRoomID).
			SetRoundID(round.ID).
			SetMembershipID(membership.ID).
			SetStatus(apiKeyAccountBindingStatusActive).
			SetStrictMode(true).
			SetStartsAt(now).
			SetExpiresAt(expiresAt).
			Save(txCtx); err != nil {
			return nil, ErrCafeActivationFailed.WithCause(fmt.Errorf("create cafe membership binding: %w", err))
		}
		if _, err := tx.CafeRoundMembership.UpdateOneID(membership.ID).
			SetStatus(GroupBuySeatStatusActive).
			SetBoundAPIKeyID(key.ID).
			SetActivatedAt(now).
			SetExpiresAt(expiresAt).
			SetUpdatedAt(now).
			Save(txCtx); err != nil {
			return nil, ErrCafeActivationFailed.WithCause(fmt.Errorf("activate cafe membership %d: %w", membership.ID, err))
		}
		if _, err := tx.GroupBuySeat.Update().Where(
			groupbuyseat.MembershipIDEQ(membership.ID),
			groupbuyseat.StatusEQ(GroupBuySeatStatusPaid),
		).SetStatus(GroupBuySeatStatusActive).SetBoundAPIKeyID(key.ID).SetBoundAt(now).SetActivatedAt(now).SetExpiresAt(expiresAt).SetUpdatedAt(now).Save(txCtx); err != nil {
			return nil, ErrCafeActivationFailed.WithCause(fmt.Errorf("activate cafe membership payment batches: %w", err))
		}
	}
	round, err = tx.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusActive).
		SetActivatedAt(now).
		SetEntitlementExpiresAt(expiresAt).
		SetClosedAt(now).
		SetUpdatedAt(now).
		Save(txCtx)
	if err != nil {
		return nil, ErrCafeActivationFailed.WithCause(fmt.Errorf("activate cafe membership round: %w", err))
	}
	if err := tx.GroupBuyEvent.Create().SetRoundID(round.ID).SetPlanID(round.PlanID).SetEventType(groupBuyEventRoundActivated).SetMessage("像素网吧份额轮次配号激活").SetMetadata(map[string]any{"membership_count": len(managedKeys), "cafe_room_id": *round.CafeRoomID}).Exec(txCtx); err != nil {
		return nil, ErrCafeActivationFailed.WithCause(fmt.Errorf("record cafe membership activation: %w", err))
	}
	if err := tx.Commit(); err != nil {
		translated := translateCafeAccountAssignmentError(err)
		if infraerrors.Reason(translated) == "CAFE_ACCOUNT_ALREADY_IN_USE" {
			return nil, translated
		}
		return nil, ErrCafeActivationFailed.WithCause(translated)
	}
	for _, key := range managedKeys {
		s.apiKeySvc.InvalidateAuthCacheByKey(ctx, key)
	}
	projection.Status = GroupBuyRoundStatusActive
	return &projection, nil
}

func (s *CafeRoomActivationService) createMembershipManagedKey(ctx context.Context, tx *dbent.Tx, round *dbent.GroupBuyRound, room *dbent.CafeRoom, membership *dbent.CafeRoundMembership, groupID int64, expiresAt time.Time) (*dbent.APIKey, error) {
	if membership == nil || membership.PaidShares <= 0 {
		return nil, ErrCafeActivationFailed
	}
	rawKey, err := s.newKeyValue()
	if err != nil {
		return nil, fmt.Errorf("generate cafe membership key: %w", err)
	}
	sourceID := membership.ID
	created := &APIKey{
		UserID:              membership.UserID,
		Key:                 rawKey,
		Name:                cafeMembershipManagedKeyName(room),
		GroupID:             &groupID,
		AccountPoolStrategy: AccountPoolStrategySharedOnly,
		Status:              StatusAPIKeyActive,
		Quota:               cafeSnapshotLimit(round.QuotaPerShareSnapshot, membership.PaidShares),
		ExpiresAt:           &expiresAt,
		RateLimit5h:         cafeSnapshotLimit(round.RateLimit5hPerShareSnapshot, membership.PaidShares),
		RateLimit1d:         cafeSnapshotLimit(round.RateLimit1dPerShareSnapshot, membership.PaidShares),
		RateLimit7d:         cafeSnapshotLimit(round.RateLimit7dPerShareSnapshot, membership.PaidShares),
		ManagedSourceType:   APIKeyManagedSourceCafeRoomMembership,
		ManagedSourceID:     &sourceID,
	}
	if err := s.apiKeyRepo.Create(ctx, created); err != nil {
		return nil, fmt.Errorf("create cafe membership key: %w", err)
	}
	key, err := tx.APIKey.Query().Where(apikey.IDEQ(created.ID), apikey.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("reload cafe membership key: %w", err)
	}
	return key, nil
}

func (s *CafeRoomActivationService) liveCafeAccountIDs(ctx context.Context, excludeRoundID int64) (map[int64]struct{}, error) {
	rounds, err := s.entClient.GroupBuyRound.Query().Where(
		groupbuyround.IDNEQ(excludeRoundID),
		groupbuyround.AssignedAccountIDNotNil(),
		groupbuyround.StatusIn(GroupBuyRoundStatusActivating, GroupBuyRoundStatusActive),
	).Select(groupbuyround.FieldAssignedAccountID).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active cafe account assignments: %w", err)
	}
	result := make(map[int64]struct{}, len(rounds))
	for _, round := range rounds {
		if round.AssignedAccountID != nil {
			result[*round.AssignedAccountID] = struct{}{}
		}
	}
	return result, nil
}

func normalizeCafeFulfillmentPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}

func cafeFulfillmentPagination(page, pageSize, total int) *pagination.PaginationResult {
	pages := 0
	if total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}
	return &pagination.PaginationResult{Page: page, PageSize: pageSize, Total: int64(total), Pages: pages}
}

func cafePendingRoundFromEntity(round *dbent.GroupBuyRound) CafePendingRound {
	item := CafePendingRound{ID: round.ID, Status: round.Status, SubscriptionTier: cafeRoundSubscriptionTier(round), PaidShares: round.PaidShares, TotalShares: round.TotalShares, MaxBuyers: round.TotalShares, FulfillmentDeadlineAt: round.FulfillmentDeadlineAt}
	if round.MaxBuyers != nil {
		item.MaxBuyers = *round.MaxBuyers
	}
	if round.FulfillmentDeadlineAt != nil && round.FulfillmentTimeoutMinutes != nil {
		paidFullAt := round.FulfillmentDeadlineAt.Add(-time.Duration(*round.FulfillmentTimeoutMinutes) * time.Minute)
		item.PaidFullAt = &paidFullAt
	}
	if room := round.Edges.CafeRoom; room != nil {
		item.RoomID, item.RoomCode, item.RoomName = room.ID, room.Code, room.Name
	} else if round.CafeRoomID != nil {
		item.RoomID = *round.CafeRoomID
		if round.RoomCodeSnapshot != nil {
			item.RoomCode = *round.RoomCodeSnapshot
		}
		if round.RoomNameSnapshot != nil {
			item.RoomName = *round.RoomNameSnapshot
		}
	}
	for _, membership := range round.Edges.CafeMemberships {
		if membership.PaidShares > 0 {
			item.JoinedBuyers++
		}
	}
	return item
}

func validateCafeRoundAwaitingAccount(round *dbent.GroupBuyRound, now time.Time) error {
	if round == nil || round.CafeRoomID == nil || round.CafeFulfillmentVersion != "membership_share" || round.Status != GroupBuyRoundStatusAwaitingAccount {
		return ErrCafeRoundNotAwaitingAccount
	}
	if round.FulfillmentDeadlineAt == nil || !now.Before(*round.FulfillmentDeadlineAt) {
		return ErrCafeFulfillmentDeadlineExpired
	}
	return nil
}

func cafeRoundSubscriptionTier(round *dbent.GroupBuyRound) string {
	if round == nil || round.SubscriptionTier == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*round.SubscriptionTier))
}

func normalizeCafeAccountPlanType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch strings.ReplaceAll(strings.ReplaceAll(normalized, "_", ""), "-", "") {
	case "plus", "chatgptplus":
		return "plus"
	case "pro", "chatgptpro":
		return "pro"
	default:
		return ""
	}
}

func cafeAccountTierMatches(tier, planType string) bool {
	return (tier == "plus" && planType == "plus") || (tier == "pro" && planType == "pro")
}

func cafeAccountCredentialString(item *dbent.Account, key string) string {
	if item == nil || item.Credentials == nil {
		return ""
	}
	value, ok := item.Credentials[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func cafeSnapshotLimit(perShare *float64, shares int) float64 {
	if perShare == nil || *perShare <= 0 || shares <= 0 {
		return 0
	}
	return *perShare * float64(shares)
}

func cafeMembershipManagedKeyName(room *dbent.CafeRoom) string {
	if room == nil || strings.TrimSpace(room.Name) == "" {
		return "Pixel Cafe Membership"
	}
	return strings.TrimSpace(room.Name) + " / Membership"
}

func translateCafeAccountAssignmentError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "idx_group_buy_rounds_assigned_account_live") || strings.Contains(message, "unique constraint") && strings.Contains(message, "assigned_account") {
		return ErrCafeAccountAlreadyInUse.WithCause(err)
	}
	return fmt.Errorf("assign cafe account: %w", err)
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
