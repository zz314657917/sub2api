package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
	"github.com/Wei-Shaw/sub2api/ent/caferoundmembership"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/predicate"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	cafePublicDefaultRoomLimit = 8
	cafePublicMaxRoomLimit     = 24
	cafePublicMaxPageSize      = 100
)

var (
	ErrCafeDisabled             = infraerrors.NotFound("CAFE_DISABLED", "pixel cafe is disabled")
	ErrCafePublicUnavailable    = infraerrors.InternalServer("CAFE_SERVICE_UNAVAILABLE", "pixel cafe service is unavailable")
	ErrCafePublicRoomNotFound   = infraerrors.NotFound("CAFE_ROOM_NOT_FOUND", "cafe room not found")
	ErrCafeMyRoomsInvalidStatus = infraerrors.BadRequest("CAFE_MY_ROOMS_INVALID_STATUS", "invalid cafe my rooms status")
)

// CafePublicSettings limits the user-facing service to the existing public setting view.
type CafePublicSettings interface {
	GetPublicSettings(context.Context) (*PublicSettings, error)
}

// CafePublicService creates redacted, read-only projections for the user Cafe API.
// It intentionally does not reuse the administrator CafeRoom DTO, which includes
// operational fields such as account and target-group identifiers.
type CafePublicService struct {
	entClient *dbent.Client
	settings  CafePublicSettings
	lobby     *CafeLobbyActivityService
	now       func() time.Time
}

func NewCafePublicService(entClient *dbent.Client, settings CafePublicSettings) *CafePublicService {
	return newCafePublicService(entClient, settings, nil)
}

func ProvideCafePublicService(entClient *dbent.Client, settings CafePublicSettings, lobby *CafeLobbyActivityService) *CafePublicService {
	return newCafePublicService(entClient, settings, lobby)
}

func newCafePublicService(entClient *dbent.Client, settings CafePublicSettings, lobby *CafeLobbyActivityService) *CafePublicService {
	return &CafePublicService{entClient: entClient, settings: settings, lobby: lobby, now: time.Now}
}

type CafePublicPlan struct {
	ID               int64   `json:"id"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	PricePerShare    float64 `json:"price_per_share"`
	PriceLabel       string  `json:"price_label"`
	ValidityDays     int     `json:"validity_days"`
	SubscriptionTier string  `json:"subscription_tier"`
	TotalShares      int     `json:"total_shares"`
	MaxBuyers        int     `json:"max_buyers"`
	MaxSharesPerUser int     `json:"max_shares_per_user"`
}

type CafePublicRound struct {
	ID                    int64      `json:"id"`
	Status                string     `json:"status"`
	PaidShares            int        `json:"paid_shares"`
	ReservedShares        int        `json:"reserved_shares"`
	RemainingShares       int        `json:"remaining_shares"`
	MaxBuyers             int        `json:"max_buyers"`
	JoinedBuyers          int        `json:"joined_buyers"`
	RemainingBuyerSlots   int        `json:"remaining_buyer_slots"`
	DeadlineAt            time.Time  `json:"deadline_at"`
	FulfillmentDeadlineAt *time.Time `json:"fulfillment_deadline_at,omitempty"`
	ActivatedAt           *time.Time `json:"activated_at,omitempty"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
}

type CafePublicMemberAvatar struct {
	AvatarSeed string `json:"avatar_seed"`
}

// CafePublicSeatVisual is retained only for in-process legacy compatibility.
// It is deliberately excluded from JSON; new public clients use member_avatars.
type CafePublicSeatVisual struct {
	SeatNo     int
	State      string
	AvatarSeed string
	IsMine     bool
}

type CafePublicRoom struct {
	ID            int64                    `json:"id"`
	Code          string                   `json:"code"`
	Name          string                   `json:"name"`
	ZoneKey       string                   `json:"zone_key"`
	ThemeKey      string                   `json:"theme_key"`
	SceneSlotKey  string                   `json:"scene_slot_key"`
	Featured      bool                     `json:"featured"`
	Plan          CafePublicPlan           `json:"plan"`
	Round         *CafePublicRound         `json:"round,omitempty"`
	MemberAvatars []CafePublicMemberAvatar `json:"member_avatars"`
	PurchaseState string                   `json:"purchase_state"`
	MyPaidShares  int                      `json:"my_paid_shares,omitempty"`
	SeatVisuals   []CafePublicSeatVisual   `json:"-"`
}

type CafePublicZone struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	RoomCount      int    `json:"room_count"`
	OpenShareCount int    `json:"open_share_count"`
}

type CafePublicOverview struct {
	APIVersion string            `json:"api_version"`
	ServerTime time.Time         `json:"server_time"`
	Zones      []CafePublicZone  `json:"zones"`
	Rooms      []CafePublicRoom  `json:"rooms"`
	Lobby      CafeLobbyActivity `json:"lobby"`
}

type CafePublicRoomDetail struct {
	APIVersion string          `json:"api_version"`
	Room       CafePublicRoom  `json:"room"`
	Rules      CafePublicRules `json:"rules"`
	ServerTime time.Time       `json:"server_time"`
}

type CafePublicRules struct {
	Activation      string `json:"activation"`
	Refund          string `json:"refund"`
	OneKeyPerMember bool   `json:"one_key_per_member"`
}

type CafePublicListParams struct {
	Page     int
	PageSize int
	Zone     string
	Featured *bool
}

const (
	CafeMyRoomStatusActive  = "active"
	CafeMyRoomStatusWaiting = "waiting"
	CafeMyRoomStatusHistory = "history"
)

type CafeMyRoomsListParams struct {
	Page     int
	PageSize int
	Statuses []string
}

type CafeMyRoom struct {
	MembershipID  int64                 `json:"membership_id"`
	Status        string                `json:"status"`
	PaidShares    int                   `json:"paid_shares"`
	ActivatedAt   *time.Time            `json:"activated_at,omitempty"`
	ExpiresAt     *time.Time            `json:"expires_at,omitempty"`
	Room          CafeMyRoomRoom        `json:"room"`
	Account       *CafeMyRoomAccount    `json:"account,omitempty"`
	Plan          CafeMyRoomPlan        `json:"plan"`
	Round         CafeMyRoomRound       `json:"round"`
	ManagedAPIKey *CafeMyRoomManagedKey `json:"managed_api_key"`
	Seat          CafeMyRoomSeat        `json:"-"`
}

type CafeMyRoomRoom struct {
	ID       int64  `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	ZoneKey  string `json:"zone_key"`
	ThemeKey string `json:"theme_key"`
}

type CafeMyRoomAccount struct {
	Name               string   `json:"name"`
	Platform           string   `json:"platform"`
	EmailMasked        string   `json:"email_masked,omitempty"`
	Remaining7dPercent *float64 `json:"remaining_7d_percent,omitempty"`
}

type CafeMyRoomPlan struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	SubscriptionTier string `json:"subscription_tier,omitempty"`
	ValidityDays     int    `json:"validity_days"`
}

type CafeMyRoomRound struct {
	ID          int64  `json:"id"`
	Status      string `json:"status"`
	PaidShares  int    `json:"paid_shares"`
	TotalShares int    `json:"total_shares"`
}

type CafeMyRoomSeat struct {
	ID          int64
	SeatNo      *int
	Status      string
	ActivatedAt *time.Time
	ExpiresAt   *time.Time
}

type CafeMyRoomManagedKey struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Quota       float64    `json:"quota"`
	QuotaUsed   float64    `json:"quota_used"`
	RateLimit5h float64    `json:"rate_limit_5h"`
	RateLimit1d float64    `json:"rate_limit_1d"`
	RateLimit7d float64    `json:"rate_limit_7d"`
	Usage5h     float64    `json:"usage_5h"`
	Usage7d     float64    `json:"usage_7d"`
	ResetAt5h   *time.Time `json:"reset_at_5h,omitempty"`
	ResetAt7d   *time.Time `json:"reset_at_7d,omitempty"`
	Protected   bool       `json:"protected"`
}

func (s *CafePublicService) Overview(ctx context.Context, userID int64, roomLimit int) (*CafePublicOverview, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	if roomLimit <= 0 {
		roomLimit = cafePublicDefaultRoomLimit
	}
	if roomLimit > cafePublicMaxRoomLimit {
		roomLimit = cafePublicMaxRoomLimit
	}

	featured := true
	rooms, _, err := s.list(ctx, userID, CafePublicListParams{Page: 1, PageSize: roomLimit, Featured: &featured})
	if err != nil {
		return nil, err
	}
	zones, err := s.listZones(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &CafePublicOverview{
		APIVersion: "cafe.v1",
		ServerTime: s.now().UTC(),
		Zones:      zones,
		Rooms:      rooms,
		Lobby:      s.lobbyActivity(ctx),
	}, nil
}

// LobbyActivity returns an anonymous availability projection without affecting room discovery.
func (s *CafePublicService) LobbyActivity(ctx context.Context, _ int64) (*CafeLobbyActivity, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	activity := s.lobbyActivity(ctx)
	return &activity, nil
}

func (s *CafePublicService) lobbyActivity(ctx context.Context) CafeLobbyActivity {
	if s == nil {
		return unavailableCafeLobbyActivity(time.Now())
	}
	if s.lobby == nil {
		return unavailableCafeLobbyActivity(s.now())
	}
	return s.lobby.Snapshot(ctx)
}

func unavailableCafeLobbyActivity(now time.Time) CafeLobbyActivity {
	return CafeLobbyActivity{
		Available:  false,
		Date:       now.Format("2006-01-02"),
		Timezone:   "Local",
		Label:      cafeLobbyLabel,
		DisplayMax: cafeLobbyDisplayMax,
		Avatars:    []CafeLobbyAvatar{},
	}
}

func (s *CafePublicService) List(ctx context.Context, userID int64, params CafePublicListParams) ([]CafePublicRoom, *pagination.PaginationResult, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, nil, err
	}
	return s.list(ctx, userID, params)
}

func (s *CafePublicService) Get(ctx context.Context, userID, roomID int64) (*CafePublicRoomDetail, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	if roomID <= 0 {
		return nil, ErrCafePublicRoomNotFound
	}
	room, err := s.visibleRoomQuery().
		Where(caferoom.IDEQ(roomID)).
		WithPlan().
		WithRounds(currentCafeRoundQuery).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrCafePublicRoomNotFound
		}
		return nil, fmt.Errorf("get public cafe room: %w", err)
	}
	publicRoom := publicCafeRoom(room, userID, s.now())
	return &CafePublicRoomDetail{
		APIVersion: "cafe.v1",
		Room:       publicRoom,
		Rules: CafePublicRules{
			Activation:      "full_then_assign_account",
			Refund:          "automatic_after_fulfillment_timeout",
			OneKeyPerMember: true,
		},
		ServerTime: s.now().UTC(),
	}, nil
}

// ParseCafeMyRoomStatuses accepts the public status filter without adding a cursor contract.
func ParseCafeMyRoomStatuses(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	return normalizeCafeMyRoomStatuses(strings.Split(raw, ","))
}

func (s *CafePublicService) MyRooms(ctx context.Context, userID int64, params CafeMyRoomsListParams) ([]CafeMyRoom, *pagination.PaginationResult, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, nil, err
	}
	if s == nil || s.entClient == nil {
		return nil, nil, ErrCafePublicUnavailable
	}
	statuses, err := normalizeCafeMyRoomStatuses(params.Statuses)
	if err != nil {
		return nil, nil, err
	}
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > cafePublicMaxPageSize {
		pageSize = cafePublicMaxPageSize
	}
	now := s.now()
	type datedMyRoom struct {
		item CafeMyRoom
		at   time.Time
	}
	allItems := make([]datedMyRoom, 0)
	memberships, err := s.entClient.CafeRoundMembership.Query().Where(
		caferoundmembership.UserIDEQ(userID),
		caferoundmembership.HasRoundWith(groupbuyround.CafeFulfillmentVersionEQ("membership_share"), groupbuyround.CafeRoomIDNotNil(), groupbuyround.HasCafeRoom()),
	).WithRound(func(roundQuery *dbent.GroupBuyRoundQuery) {
		roundQuery.WithCafeRoom().WithAssignedAccount().WithPlan()
	}).All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list cafe memberships: %w", err)
	}
	keyIDs := make([]int64, 0, len(memberships))
	for _, membership := range memberships {
		if membership.BoundAPIKeyID != nil {
			keyIDs = append(keyIDs, *membership.BoundAPIKeyID)
		}
	}
	keysByID := make(map[int64]*dbent.APIKey, len(keyIDs))
	if len(keyIDs) > 0 {
		keys, err := s.entClient.APIKey.Query().Where(apikey.IDIn(keyIDs...), apikey.DeletedAtIsNil()).All(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("load cafe membership keys: %w", err)
		}
		for _, key := range keys {
			keysByID[key.ID] = key
		}
	}
	for _, membership := range memberships {
		if !cafeMyRoomMembershipMatchesStatuses(membership, statuses, now) {
			continue
		}
		item, ok := cafeMyRoomFromMembership(membership, keysByID, now)
		if ok {
			allItems = append(allItems, datedMyRoom{item: item, at: membership.CreatedAt})
		}
	}
	legacySeats, err := s.entClient.GroupBuySeat.Query().Where(
		groupbuyseat.UserIDEQ(userID),
		groupbuyseat.HasRoundWith(groupbuyround.CafeFulfillmentVersionEQ("legacy_seat"), groupbuyround.CafeRoomIDNotNil(), groupbuyround.HasCafeRoom()),
	).WithRound(func(roundQuery *dbent.GroupBuyRoundQuery) {
		roundQuery.WithCafeRoom(func(roomQuery *dbent.CafeRoomQuery) { roomQuery.WithAccount() })
	}).WithPlan().WithBoundAPIKey().All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list legacy cafe rooms: %w", err)
	}
	for _, seat := range legacySeats {
		if !cafeMyRoomLegacySeatMatchesStatuses(seat, statuses, now) {
			continue
		}
		if item, ok := cafeMyRoomFromSeat(seat, now); ok {
			allItems = append(allItems, datedMyRoom{item: item, at: seat.CreatedAt})
		}
	}
	sort.SliceStable(allItems, func(i, j int) bool {
		if allItems[i].at.Equal(allItems[j].at) {
			return allItems[i].item.MembershipID > allItems[j].item.MembershipID
		}
		return allItems[i].at.After(allItems[j].at)
	})
	count := len(allItems)
	start := (page - 1) * pageSize
	if start > count {
		start = count
	}
	end := start + pageSize
	if end > count {
		end = count
	}
	items := make([]CafeMyRoom, 0, end-start)
	for _, entry := range allItems[start:end] {
		items = append(items, entry.item)
	}
	return items, &pagination.PaginationResult{
		Page:     page,
		PageSize: pageSize,
		Total:    int64(count),
		Pages:    int((count + pageSize - 1) / pageSize),
	}, nil
}

func normalizeCafeMyRoomStatuses(statuses []string) ([]string, error) {
	if len(statuses) == 0 {
		return nil, nil
	}
	result := make([]string, 0, len(statuses))
	seen := make(map[string]struct{}, len(statuses))
	for _, status := range statuses {
		status = strings.TrimSpace(status)
		if status != CafeMyRoomStatusActive && status != CafeMyRoomStatusWaiting && status != CafeMyRoomStatusHistory {
			return nil, ErrCafeMyRoomsInvalidStatus
		}
		if _, exists := seen[status]; exists {
			return nil, ErrCafeMyRoomsInvalidStatus
		}
		seen[status] = struct{}{}
		result = append(result, status)
	}
	return result, nil
}

func cafeMyRoomActivePredicate(now time.Time) predicate.GroupBuySeat {
	return groupbuyseat.And(
		groupbuyseat.StatusEQ(GroupBuySeatStatusActive),
		groupbuyseat.Or(groupbuyseat.ExpiresAtIsNil(), groupbuyseat.ExpiresAtGT(now)),
	)
}

func cafeMyRoomWaitingPredicate(now time.Time) predicate.GroupBuySeat {
	return groupbuyseat.Or(
		groupbuyseat.StatusEQ(GroupBuySeatStatusPaid),
		groupbuyseat.And(
			groupbuyseat.StatusEQ(GroupBuySeatStatusLocked),
			groupbuyseat.LockedUntilNotNil(),
			groupbuyseat.LockedUntilGT(now),
		),
	)
}

func cafeMyRoomMembershipMatchesStatuses(membership *dbent.CafeRoundMembership, statuses []string, now time.Time) bool {
	if len(statuses) == 0 {
		return true
	}
	active := membership != nil && membership.Status == GroupBuySeatStatusActive && (membership.ExpiresAt == nil || membership.ExpiresAt.After(now))
	waiting := membership != nil && membership.Edges.Round != nil && (membership.Status == GroupBuySeatStatusPaid || membership.Status == GroupBuySeatStatusLocked) &&
		(membership.Edges.Round.Status == GroupBuyRoundStatusOpen || membership.Edges.Round.Status == GroupBuyRoundStatusAwaitingAccount || membership.Edges.Round.Status == GroupBuyRoundStatusActivating)
	for _, status := range statuses {
		if status == CafeMyRoomStatusActive && active || status == CafeMyRoomStatusWaiting && waiting || status == CafeMyRoomStatusHistory && !active && !waiting {
			return true
		}
	}
	return false
}

func cafeMyRoomLegacySeatMatchesStatuses(seat *dbent.GroupBuySeat, statuses []string, now time.Time) bool {
	if len(statuses) == 0 {
		return true
	}
	active := seat != nil && seat.Status == GroupBuySeatStatusActive && (seat.ExpiresAt == nil || seat.ExpiresAt.After(now))
	waiting := seat != nil && (seat.Status == GroupBuySeatStatusPaid || seat.Status == GroupBuySeatStatusLocked && seat.LockedUntil != nil && seat.LockedUntil.After(now))
	for _, status := range statuses {
		if status == CafeMyRoomStatusActive && active || status == CafeMyRoomStatusWaiting && waiting || status == CafeMyRoomStatusHistory && !active && !waiting {
			return true
		}
	}
	return false
}

func cafeMyRoomFromMembership(membership *dbent.CafeRoundMembership, keys map[int64]*dbent.APIKey, now time.Time) (CafeMyRoom, bool) {
	if membership == nil || membership.Edges.Round == nil || membership.Edges.Round.Edges.CafeRoom == nil || membership.Edges.Round.Edges.Plan == nil {
		return CafeMyRoom{}, false
	}
	round := membership.Edges.Round
	room := round.Edges.CafeRoom
	plan := round.Edges.Plan
	item := CafeMyRoom{
		MembershipID: membership.ID,
		Status:       membership.Status,
		PaidShares:   membership.PaidShares,
		ActivatedAt:  membership.ActivatedAt,
		ExpiresAt:    membership.ExpiresAt,
		Room:         CafeMyRoomRoom{ID: room.ID, Code: room.Code, Name: room.Name, ZoneKey: room.ZoneKey, ThemeKey: room.ThemeKey},
		Plan:         CafeMyRoomPlan{ID: plan.ID, Title: plan.Title, SubscriptionTier: cafeRoundSubscriptionTier(round), ValidityDays: cafeRoundValidityDays(round, plan.ValidityDays)},
		Round:        CafeMyRoomRound{ID: round.ID, Status: round.Status, PaidShares: round.PaidShares, TotalShares: round.TotalShares},
	}
	if round.Status == GroupBuyRoundStatusActive && membership.Status == GroupBuySeatStatusActive {
		if assigned := round.Edges.AssignedAccount; assigned != nil {
			item.Account = safeCafeMyRoomAccount(assigned, now)
		}
		if membership.BoundAPIKeyID != nil {
			if key := keys[*membership.BoundAPIKeyID]; key != nil && key.UserID == membership.UserID && key.ManagedSourceType == APIKeyManagedSourceCafeRoomMembership && key.ManagedSourceID != nil && *key.ManagedSourceID == membership.ID {
				item.ManagedAPIKey = cafeMyRoomManagedKey(key, now)
			}
		}
	}
	return item, true
}

func cafeMyRoomFromSeat(seat *dbent.GroupBuySeat, now time.Time) (CafeMyRoom, bool) {
	if seat == nil || seat.Edges.Round == nil || seat.Edges.Round.Edges.CafeRoom == nil || seat.Edges.Plan == nil {
		return CafeMyRoom{}, false
	}
	round := seat.Edges.Round
	room := round.Edges.CafeRoom
	item := CafeMyRoom{
		MembershipID: seat.ID,
		Status:       seat.Status,
		PaidShares:   seat.ShareCount,
		ActivatedAt:  seat.ActivatedAt,
		ExpiresAt:    seat.ExpiresAt,
		Room: CafeMyRoomRoom{
			ID: room.ID, Code: room.Code, Name: room.Name, ZoneKey: room.ZoneKey, ThemeKey: room.ThemeKey,
		},
		Plan:  CafeMyRoomPlan{ID: seat.Edges.Plan.ID, Title: seat.Edges.Plan.Title, SubscriptionTier: "plus", ValidityDays: seat.Edges.Plan.ValidityDays},
		Round: CafeMyRoomRound{ID: round.ID, Status: round.Status, PaidShares: round.PaidShares, TotalShares: round.TotalShares},
		Seat:  CafeMyRoomSeat{ID: seat.ID, SeatNo: seat.SeatNo, Status: seat.Status, ActivatedAt: seat.ActivatedAt, ExpiresAt: seat.ExpiresAt},
	}
	if account := room.Edges.Account; account != nil {
		item.Account = safeCafeMyRoomAccount(account, now)
	}
	if key := seat.Edges.BoundAPIKey; key != nil && key.UserID == seat.UserID && key.ManagedSourceType == APIKeyManagedSourceCafeRoomSeat && key.ManagedSourceID != nil && *key.ManagedSourceID == seat.ID {
		item.ManagedAPIKey = cafeMyRoomManagedKey(key, now)
	}
	return item, true
}

func safeCafeMyRoomAccount(account *dbent.Account, now time.Time) *CafeMyRoomAccount {
	if account == nil {
		return nil
	}
	return &CafeMyRoomAccount{
		Name:               cafeAccountDisplayName(account.Name),
		Platform:           account.Platform,
		EmailMasked:        cafeAccountEmailMasked(account),
		Remaining7dPercent: cafeAccountRemaining7dPercent(account, now),
	}
}

func cafeAccountRemaining7dPercent(account *dbent.Account, now time.Time) *float64 {
	if account == nil || account.Platform != PlatformOpenAI {
		return nil
	}
	progress := buildCodexUsageProgressFromExtra(account.Extra, "7d", now)
	if progress == nil || math.IsNaN(progress.Utilization) || math.IsInf(progress.Utilization, 0) {
		return nil
	}
	remaining := 100 - math.Min(100, math.Max(0, progress.Utilization))
	return &remaining
}

func cafeMyRoomManagedKey(key *dbent.APIKey, now time.Time) *CafeMyRoomManagedKey {
	if key == nil {
		return nil
	}
	usage5h, resetAt5h := cafeMyRoomWindowProjection(key.Usage5h, key.RateLimit5h, key.Window5hStart, 5*time.Hour, now)
	usage7d, resetAt7d := cafeMyRoomWindowProjection(key.Usage7d, key.RateLimit7d, key.Window7dStart, 7*24*time.Hour, now)
	return &CafeMyRoomManagedKey{
		ID: key.ID, Name: key.Name, Status: key.Status, Quota: key.Quota, QuotaUsed: key.QuotaUsed,
		RateLimit5h: key.RateLimit5h, RateLimit1d: key.RateLimit1d, RateLimit7d: key.RateLimit7d,
		Usage5h: usage5h, Usage7d: usage7d, ResetAt5h: resetAt5h, ResetAt7d: resetAt7d, Protected: true,
	}
}

func cafeMyRoomWindowProjection(usage, limit float64, windowStart *time.Time, windowSize time.Duration, now time.Time) (float64, *time.Time) {
	if limit <= 0 {
		return usage, nil
	}
	if windowStart == nil {
		return 0, nil
	}
	resetAt := windowStart.Add(windowSize).UTC()
	if !resetAt.After(now) {
		return 0, nil
	}
	return usage, &resetAt
}

func cafeRoundValidityDays(round *dbent.GroupBuyRound, fallback int) int {
	if round != nil && round.ValidityDaysSnapshot != nil && *round.ValidityDaysSnapshot > 0 {
		return *round.ValidityDaysSnapshot
	}
	return fallback
}

func cafeAccountEmailMasked(account *dbent.Account) string {
	if account == nil {
		return ""
	}
	value, ok := account.Credentials["email"].(string)
	return MaskCafeEmail(value, ok)
}

func cafeAccountDisplayName(name string) string {
	if masked := MaskCafeEmail(name, true); masked != "" {
		return masked
	}
	return name
}

// MaskCafeEmail fails closed for incomplete or malformed values before masking.
func MaskCafeEmail(value string, isString bool) string {
	value = strings.TrimSpace(value)
	if !isString || value == "" || strings.Count(value, "@") != 1 || strings.ContainsAny(value, " \t\r\n") {
		return ""
	}
	parts := strings.SplitN(value, "@", 2)
	if parts[0] == "" || parts[1] == "" || strings.Contains(parts[1], ".") == false {
		return ""
	}
	return MaskEmail(value)
}

func (s *CafePublicService) list(ctx context.Context, userID int64, params CafePublicListParams) ([]CafePublicRoom, *pagination.PaginationResult, error) {
	if s == nil || s.entClient == nil {
		return nil, nil, ErrCafePublicUnavailable
	}
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = cafePublicDefaultRoomLimit
	}
	if pageSize > cafePublicMaxPageSize {
		pageSize = cafePublicMaxPageSize
	}

	query := s.visibleRoomQuery()
	if zone := strings.TrimSpace(params.Zone); zone != "" {
		query.Where(caferoom.ZoneKeyEQ(zone))
	}
	if params.Featured != nil {
		query.Where(caferoom.FeaturedEQ(*params.Featured))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count public cafe rooms: %w", err)
	}
	rooms, err := s.listVisibleRoomEntities(ctx, page, pageSize, strings.TrimSpace(params.Zone), params.Featured)
	if err != nil {
		return nil, nil, err
	}
	result := make([]CafePublicRoom, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, publicCafeRoom(room, userID, s.now()))
	}
	return result, &pagination.PaginationResult{
		Page:     page,
		PageSize: pageSize,
		Total:    int64(total),
		Pages:    int((total + pageSize - 1) / pageSize),
	}, nil
}

func (s *CafePublicService) listVisibleRoomEntities(ctx context.Context, page, pageSize int, zone string, featured *bool) ([]*dbent.CafeRoom, error) {
	query := s.visibleRoomQuery()
	if zone != "" {
		query.Where(caferoom.ZoneKeyEQ(zone))
	}
	if featured != nil {
		query.Where(caferoom.FeaturedEQ(*featured))
	}
	rooms, err := query.
		Order(dbent.Desc(caferoom.FieldFeatured), dbent.Asc(caferoom.FieldSortOrder), dbent.Asc(caferoom.FieldID)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		WithPlan().
		WithRounds(currentCafeRoundQuery).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list public cafe rooms: %w", err)
	}
	return rooms, nil
}

func (s *CafePublicService) listZones(ctx context.Context, userID int64) ([]CafePublicZone, error) {
	rooms, err := s.visibleRoomQuery().WithRounds(currentCafeRoundQuery).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list public cafe zones: %w", err)
	}
	zoneCounts := make(map[string]*CafePublicZone)
	for _, room := range rooms {
		key := strings.TrimSpace(room.ZoneKey)
		if key == "" {
			continue
		}
		zone := zoneCounts[key]
		if zone == nil {
			zone = &CafePublicZone{Key: key, Name: cafeZoneName(key)}
			zoneCounts[key] = zone
		}
		zone.RoomCount++
		if len(room.Edges.Rounds) > 0 {
			round := room.Edges.Rounds[0]
			remaining := round.TotalShares - round.PaidShares - round.ReservedShares
			if remaining > 0 {
				zone.OpenShareCount += remaining
			}
		}
	}
	keys := make([]string, 0, len(zoneCounts))
	for key := range zoneCounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	zones := make([]CafePublicZone, 0, len(keys))
	for _, key := range keys {
		zones = append(zones, *zoneCounts[key])
	}
	return zones, nil
}

func (s *CafePublicService) visibleRoomQuery() *dbent.CafeRoomQuery {
	return s.entClient.CafeRoom.Query().Where(
		caferoom.StatusEQ(CafeRoomStatusEnabled),
		caferoom.DeletedAtIsNil(),
		caferoom.HasPlanWith(
			groupbuyplan.DeletedAtIsNil(),
			groupbuyplan.StatusEQ(GroupBuyPlanStatusActive),
			groupbuyplan.FulfillmentModeEQ(CafeRoomFulfillmentMode),
		),
	)
}

func (s *CafePublicService) requireEnabled(ctx context.Context) error {
	if s == nil || s.settings == nil {
		return ErrCafePublicUnavailable
	}
	settings, err := s.settings.GetPublicSettings(ctx)
	if err != nil {
		return fmt.Errorf("load public cafe settings: %w", err)
	}
	if settings == nil || !settings.PixelCafeEnabled {
		return ErrCafeDisabled
	}
	return nil
}

func currentCafeRoundQuery(query *dbent.GroupBuyRoundQuery) {
	query.Where(groupbuyround.StatusIn(CafeRoundStatusOpen, GroupBuyRoundStatusAwaitingAccount, GroupBuyRoundStatusActivating, GroupBuyRoundStatusActive, GroupBuyRoundStatusRefunding, GroupBuyRoundStatusRefunded)).
		Order(dbent.Desc(groupbuyround.FieldCreatedAt), dbent.Desc(groupbuyround.FieldID)).
		Limit(1).
		WithSeats().
		WithCafeMemberships()
}

func publicCafeRoom(room *dbent.CafeRoom, userID int64, now time.Time) CafePublicRoom {
	if room == nil || room.Edges.Plan == nil {
		return CafePublicRoom{MemberAvatars: []CafePublicMemberAvatar{}, PurchaseState: "unavailable"}
	}
	plan := room.Edges.Plan
	totalShares, tier, maxBuyers, maxSharesPerUser := publicCafePlanPolicy(plan)
	publicPlan := CafePublicPlan{
		ID:               plan.ID,
		Title:            plan.Title,
		Description:      cafeString(plan.Description),
		PricePerShare:    plan.PricePerShare,
		PriceLabel:       plan.PriceLabel,
		ValidityDays:     plan.ValidityDays,
		SubscriptionTier: tier,
		TotalShares:      totalShares,
		MaxBuyers:        maxBuyers,
		MaxSharesPerUser: maxSharesPerUser,
	}
	result := CafePublicRoom{
		ID:            room.ID,
		Code:          room.Code,
		Name:          room.Name,
		ZoneKey:       room.ZoneKey,
		ThemeKey:      room.ThemeKey,
		SceneSlotKey:  room.SceneSlotKey,
		Featured:      room.Featured,
		Plan:          publicPlan,
		MemberAvatars: []CafePublicMemberAvatar{},
		PurchaseState: "unavailable",
	}
	if len(room.Edges.Rounds) == 0 {
		return result
	}
	round := room.Edges.Rounds[0]
	remaining := round.TotalShares - round.PaidShares - round.ReservedShares
	if remaining < 0 {
		remaining = 0
	}
	joinedBuyers, myPaidShares, isParticipant := publicCafeMembershipCounts(round, userID, now)
	roundMaxBuyers := maxBuyers
	if round.MaxBuyers != nil {
		roundMaxBuyers = *round.MaxBuyers
	}
	remainingBuyerSlots := roundMaxBuyers - joinedBuyers
	if remainingBuyerSlots < 0 {
		remainingBuyerSlots = 0
	}
	result.Round = &CafePublicRound{ID: round.ID, Status: round.Status, PaidShares: round.PaidShares, ReservedShares: round.ReservedShares, RemainingShares: remaining, MaxBuyers: roundMaxBuyers, JoinedBuyers: joinedBuyers, RemainingBuyerSlots: remainingBuyerSlots, DeadlineAt: round.DeadlineAt, FulfillmentDeadlineAt: round.FulfillmentDeadlineAt, ActivatedAt: round.ActivatedAt, ExpiresAt: round.EntitlementExpiresAt}
	result.MemberAvatars = publicCafeMemberAvatars(round)
	result.SeatVisuals = publicCafeSeatVisuals(round, userID, now)
	result.MyPaidShares = myPaidShares
	result.PurchaseState = publicCafePurchaseState(round, remaining, remainingBuyerSlots, isParticipant)
	return result
}

func publicCafeMembershipCounts(round *dbent.GroupBuyRound, userID int64, now time.Time) (joinedBuyers, myPaidShares int, isParticipant bool) {
	if round == nil {
		return 0, 0, false
	}
	if round.CafeFulfillmentVersion == "membership_share" {
		for _, membership := range round.Edges.CafeMemberships {
			if membership.PaidShares > 0 || membership.ReservedShares > 0 {
				joinedBuyers++
			}
			if userID > 0 && membership.UserID == userID {
				myPaidShares = membership.PaidShares
				isParticipant = membership.PaidShares > 0 || membership.ReservedShares > 0
			}
		}
		return joinedBuyers, myPaidShares, isParticipant
	}
	for _, seat := range round.Edges.Seats {
		if publicCafeLegacySeatVisible(seat, now) {
			joinedBuyers++
			if userID > 0 && seat.UserID == userID {
				myPaidShares += seat.ShareCount
				isParticipant = true
			}
		}
	}
	return joinedBuyers, myPaidShares, isParticipant
}

func publicCafeMemberAvatars(round *dbent.GroupBuyRound) []CafePublicMemberAvatar {
	if round == nil {
		return []CafePublicMemberAvatar{}
	}
	avatars := make([]CafePublicMemberAvatar, 0)
	if round.CafeFulfillmentVersion == "membership_share" {
		memberships := append([]*dbent.CafeRoundMembership(nil), round.Edges.CafeMemberships...)
		sort.Slice(memberships, func(i, j int) bool { return memberships[i].ID < memberships[j].ID })
		for _, membership := range memberships {
			if membership.PaidShares > 0 {
				avatars = append(avatars, CafePublicMemberAvatar{AvatarSeed: publicCafeAvatarSeed(round.ID, int(membership.ID))})
			}
		}
		return avatars
	}
	for _, seat := range round.Edges.Seats {
		if seat.Status == GroupBuySeatStatusPaid || seat.Status == GroupBuySeatStatusActive || seat.Status == GroupBuySeatStatusRefundPending || seat.Status == GroupBuySeatStatusRefundProcessing {
			avatars = append(avatars, CafePublicMemberAvatar{AvatarSeed: publicCafeAvatarSeed(round.ID, int(seat.ID))})
		}
	}
	return avatars
}

func publicCafeLegacySeatVisible(seat *dbent.GroupBuySeat, now time.Time) bool {
	return seat != nil && (seat.Status == GroupBuySeatStatusPaid || seat.Status == GroupBuySeatStatusActive || seat.Status == GroupBuySeatStatusLocked && seat.LockedUntil != nil && seat.LockedUntil.After(now))
}

func publicCafeSeatVisuals(round *dbent.GroupBuyRound, userID int64, now time.Time) []CafePublicSeatVisual {
	if round == nil || round.CafeFulfillmentVersion == "membership_share" {
		return []CafePublicSeatVisual{}
	}
	totalSeats := publicCafeSeatCount(round.TotalShares, round.TotalSeats)
	if totalSeats <= 0 {
		return []CafePublicSeatVisual{}
	}
	visuals := make([]CafePublicSeatVisual, totalSeats)
	for index := range visuals {
		visuals[index] = CafePublicSeatVisual{SeatNo: index + 1, State: "empty"}
	}
	for _, seat := range round.Edges.Seats {
		if seat.SeatNo == nil || *seat.SeatNo < 1 || *seat.SeatNo > totalSeats || !publicCafeLegacySeatVisible(seat, now) {
			continue
		}
		state := seat.Status
		seatNo := *seat.SeatNo
		visuals[seatNo-1] = CafePublicSeatVisual{SeatNo: seatNo, State: state, AvatarSeed: publicCafeAvatarSeed(round.ID, seatNo), IsMine: userID > 0 && seat.UserID == userID}
	}
	return visuals
}

func publicCafePurchaseState(round *dbent.GroupBuyRound, remaining, remainingBuyerSlots int, isParticipant bool) string {
	if round == nil {
		return "unavailable"
	}
	switch round.Status {
	case CafeRoundStatusOpen:
		if remaining > 0 && (remainingBuyerSlots > 0 || isParticipant) {
			return "available"
		}
		if remaining > 0 {
			return "buyers_full"
		}
		return "awaiting_account"
	case GroupBuyRoundStatusAwaitingAccount:
		return "awaiting_account"
	case GroupBuyRoundStatusActivating:
		return "activating"
	case GroupBuyRoundStatusActive:
		return "active"
	case GroupBuyRoundStatusRefunding:
		return "refunding"
	case GroupBuyRoundStatusRefunded:
		return "refunded"
	default:
		return "unavailable"
	}
}

func publicCafeAvatarSeed(roundID int64, membershipID int) string {
	digest := sha256.Sum256([]byte(fmt.Sprintf("cafe-member:%d:%d", roundID, membershipID)))
	return hex.EncodeToString(digest[:8])
}

func publicCafePlanPolicy(plan *dbent.GroupBuyPlan) (int, string, int, int) {
	if plan == nil {
		return 0, "", 0, 0
	}
	totalShares := plan.TotalShares
	if totalShares <= 0 {
		totalShares = plan.SeatCount
	}
	tier := strings.ToLower(strings.TrimSpace(plan.SubscriptionTier))
	if tier == "" {
		tier = "plus"
	}
	maxBuyers := plan.MaxBuyers
	if maxBuyers <= 0 {
		maxBuyers = totalShares
		if maxBuyers > 4 {
			maxBuyers = 4
		}
	}
	maxSharesPerUser := plan.MaxSharesPerUser
	if maxSharesPerUser <= 0 {
		maxSharesPerUser = totalShares
	}
	return totalShares, tier, maxBuyers, maxSharesPerUser
}

func publicCafeSeatCount(totalShares, seatCount int) int {
	if totalShares > 0 {
		return totalShares
	}
	return seatCount
}

func cafeString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cafeZoneName(key string) string {
	switch key {
	case "featured":
		return "精选大厅"
	case "claude":
		return "Claude 区"
	case "openai":
		return "OpenAI 区"
	case "gemini":
		return "Gemini 区"
	default:
		return key
	}
}
