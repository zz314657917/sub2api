package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
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
	ID           int64   `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	PricePerSeat float64 `json:"price_per_seat"`
	PriceLabel   string  `json:"price_label"`
	ValidityDays int     `json:"validity_days"`
	TotalSeats   int     `json:"total_seats"`
}

type CafePublicRound struct {
	ID             int64      `json:"id"`
	Status         string     `json:"status"`
	PaidSeats      int        `json:"paid_seats"`
	ReservedSeats  int        `json:"reserved_seats"`
	RemainingSeats int        `json:"remaining_seats"`
	DeadlineAt     time.Time  `json:"deadline_at"`
	ActivatedAt    *time.Time `json:"activated_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type CafePublicSeatVisual struct {
	SeatNo     int    `json:"seat_no"`
	State      string `json:"state"`
	AvatarSeed string `json:"avatar_seed,omitempty"`
	IsMine     bool   `json:"is_mine"`
}

type CafePublicRoom struct {
	ID            int64                  `json:"id"`
	Code          string                 `json:"code"`
	Name          string                 `json:"name"`
	ZoneKey       string                 `json:"zone_key"`
	ThemeKey      string                 `json:"theme_key"`
	SceneSlotKey  string                 `json:"scene_slot_key"`
	Featured      bool                   `json:"featured"`
	Plan          CafePublicPlan         `json:"plan"`
	Round         *CafePublicRound       `json:"round,omitempty"`
	SeatVisuals   []CafePublicSeatVisual `json:"seat_visuals"`
	PurchaseState string                 `json:"purchase_state"`
}

type CafePublicZone struct {
	Key           string `json:"key"`
	Name          string `json:"name"`
	RoomCount     int    `json:"room_count"`
	OpenSeatCount int    `json:"open_seat_count"`
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
	Activation     string `json:"activation"`
	Refund         string `json:"refund"`
	OneSeatPerUser bool   `json:"one_seat_per_user"`
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
	Room          CafeMyRoomRoom        `json:"room"`
	Plan          CafeMyRoomPlan        `json:"plan"`
	Round         CafeMyRoomRound       `json:"round"`
	Seat          CafeMyRoomSeat        `json:"seat"`
	ManagedAPIKey *CafeMyRoomManagedKey `json:"managed_api_key"`
}

type CafeMyRoomRoom struct {
	ID       int64  `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	ZoneKey  string `json:"zone_key"`
	ThemeKey string `json:"theme_key"`
}

type CafeMyRoomPlan struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	ValidityDays int    `json:"validity_days"`
}

type CafeMyRoomRound struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	PaidSeats  int    `json:"paid_seats"`
	TotalSeats int    `json:"total_seats"`
}

type CafeMyRoomSeat struct {
	ID          int64      `json:"id"`
	SeatNo      *int       `json:"seat_no"`
	Status      string     `json:"status"`
	ActivatedAt *time.Time `json:"activated_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type CafeMyRoomManagedKey struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Quota       float64 `json:"quota"`
	QuotaUsed   float64 `json:"quota_used"`
	RateLimit5h float64 `json:"rate_limit_5h"`
	RateLimit1d float64 `json:"rate_limit_1d"`
	RateLimit7d float64 `json:"rate_limit_7d"`
	Protected   bool    `json:"protected"`
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
			Activation:     "full_only",
			Refund:         "pending_configuration",
			OneSeatPerUser: true,
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
	count, err := s.myRoomSeatQuery(userID, statuses, now).Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count cafe my rooms: %w", err)
	}
	seats, err := s.myRoomSeatQuery(userID, statuses, now).
		Order(dbent.Desc(groupbuyseat.FieldCreatedAt), dbent.Desc(groupbuyseat.FieldID)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		WithRound(func(roundQuery *dbent.GroupBuyRoundQuery) {
			roundQuery.WithCafeRoom()
		}).
		WithPlan().
		WithBoundAPIKey().
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list cafe my rooms: %w", err)
	}
	items := make([]CafeMyRoom, 0, len(seats))
	for _, seat := range seats {
		item, ok := cafeMyRoomFromSeat(seat)
		if ok {
			items = append(items, item)
		}
	}
	return items, &pagination.PaginationResult{
		Page:     page,
		PageSize: pageSize,
		Total:    int64(count),
		Pages:    int((count + pageSize - 1) / pageSize),
	}, nil
}

func (s *CafePublicService) myRoomSeatQuery(userID int64, statuses []string, now time.Time) *dbent.GroupBuySeatQuery {
	query := s.entClient.GroupBuySeat.Query().Where(
		groupbuyseat.UserIDEQ(userID),
		groupbuyseat.HasRoundWith(groupbuyround.CafeRoomIDNotNil(), groupbuyround.HasCafeRoom()),
	)
	if len(statuses) == 0 {
		return query
	}
	predicates := make([]predicate.GroupBuySeat, 0, len(statuses))
	for _, status := range statuses {
		switch status {
		case CafeMyRoomStatusActive:
			predicates = append(predicates, cafeMyRoomActivePredicate(now))
		case CafeMyRoomStatusWaiting:
			predicates = append(predicates, cafeMyRoomWaitingPredicate(now))
		case CafeMyRoomStatusHistory:
			predicates = append(predicates, groupbuyseat.Not(groupbuyseat.Or(
				cafeMyRoomActivePredicate(now),
				cafeMyRoomWaitingPredicate(now),
			)))
		}
	}
	return query.Where(groupbuyseat.Or(predicates...))
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

func cafeMyRoomFromSeat(seat *dbent.GroupBuySeat) (CafeMyRoom, bool) {
	if seat == nil || seat.Edges.Round == nil || seat.Edges.Round.Edges.CafeRoom == nil || seat.Edges.Plan == nil {
		return CafeMyRoom{}, false
	}
	round := seat.Edges.Round
	room := round.Edges.CafeRoom
	item := CafeMyRoom{
		MembershipID: seat.ID,
		Room: CafeMyRoomRoom{
			ID: room.ID, Code: room.Code, Name: room.Name, ZoneKey: room.ZoneKey, ThemeKey: room.ThemeKey,
		},
		Plan:  CafeMyRoomPlan{ID: seat.Edges.Plan.ID, Title: seat.Edges.Plan.Title, ValidityDays: seat.Edges.Plan.ValidityDays},
		Round: CafeMyRoomRound{ID: round.ID, Status: round.Status, PaidSeats: round.PaidSeats, TotalSeats: round.TotalSeats},
		Seat:  CafeMyRoomSeat{ID: seat.ID, SeatNo: seat.SeatNo, Status: seat.Status, ActivatedAt: seat.ActivatedAt, ExpiresAt: seat.ExpiresAt},
	}
	if key := seat.Edges.BoundAPIKey; key != nil && key.UserID == seat.UserID && key.ManagedSourceType == APIKeyManagedSourceCafeRoomSeat && key.ManagedSourceID != nil && *key.ManagedSourceID == seat.ID {
		item.ManagedAPIKey = &CafeMyRoomManagedKey{
			ID: key.ID, Name: key.Name, Status: key.Status, Quota: key.Quota, QuotaUsed: key.QuotaUsed,
			RateLimit5h: key.RateLimit5h, RateLimit1d: key.RateLimit1d, RateLimit7d: key.RateLimit7d, Protected: true,
		}
	}
	return item, true
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
	var rows []struct {
		ZoneKey string `json:"zone_key"`
	}
	err := s.visibleRoomQuery().Unique(true).Select(caferoom.FieldZoneKey).Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("list public cafe zones: %w", err)
	}
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		if key := strings.TrimSpace(row.ZoneKey); key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	zones := make([]CafePublicZone, 0, len(keys))
	for _, key := range keys {
		rooms, result, err := s.list(ctx, userID, CafePublicListParams{Page: 1, PageSize: cafePublicMaxPageSize, Zone: key})
		if err != nil {
			return nil, err
		}
		openSeats := 0
		for _, room := range rooms {
			if room.Round != nil {
				openSeats += room.Round.RemainingSeats
			}
		}
		zone := CafePublicZone{Key: key, Name: cafeZoneName(key), OpenSeatCount: openSeats}
		if result != nil {
			zone.RoomCount = int(result.Total)
		}
		zones = append(zones, zone)
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
	query.Where(groupbuyround.StatusIn(CafeRoundStatusOpen, "activating", "active")).
		Order(dbent.Asc(groupbuyround.FieldDeadlineAt)).
		Limit(1).
		WithSeats(func(seatQuery *dbent.GroupBuySeatQuery) {
			seatQuery.Order(dbent.Asc(groupbuyseat.FieldSeatNo), dbent.Asc(groupbuyseat.FieldID))
		})
}

func publicCafeRoom(room *dbent.CafeRoom, userID int64, now time.Time) CafePublicRoom {
	if room == nil || room.Edges.Plan == nil {
		return CafePublicRoom{SeatVisuals: []CafePublicSeatVisual{}, PurchaseState: "unavailable"}
	}
	plan := room.Edges.Plan
	publicPlan := CafePublicPlan{
		ID:           plan.ID,
		Title:        plan.Title,
		Description:  cafeString(plan.Description),
		PricePerSeat: plan.PricePerShare,
		PriceLabel:   plan.PriceLabel,
		ValidityDays: plan.ValidityDays,
		TotalSeats:   publicCafeSeatCount(plan.TotalShares, plan.SeatCount),
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
		SeatVisuals:   []CafePublicSeatVisual{},
		PurchaseState: "unavailable",
	}
	if len(room.Edges.Rounds) == 0 {
		return result
	}
	round := room.Edges.Rounds[0]
	remaining := round.TotalSeats - round.PaidSeats - round.ReservedSeats
	if remaining < 0 {
		remaining = 0
	}
	result.Round = &CafePublicRound{
		ID:             round.ID,
		Status:         round.Status,
		PaidSeats:      round.PaidSeats,
		ReservedSeats:  round.ReservedSeats,
		RemainingSeats: remaining,
		DeadlineAt:     round.DeadlineAt,
		ActivatedAt:    round.ActivatedAt,
		ExpiresAt:      round.EntitlementExpiresAt,
	}
	result.SeatVisuals = publicCafeSeatVisuals(round, userID, now)
	result.PurchaseState = publicCafePurchaseState(round, remaining)
	return result
}

func publicCafeSeatVisuals(round *dbent.GroupBuyRound, userID int64, now time.Time) []CafePublicSeatVisual {
	if round == nil {
		return []CafePublicSeatVisual{}
	}
	totalSeats := round.TotalSeats
	if totalSeats <= 0 {
		totalSeats = round.TotalShares
	}
	if totalSeats <= 0 {
		return []CafePublicSeatVisual{}
	}
	visuals := make([]CafePublicSeatVisual, totalSeats)
	for index := range visuals {
		visuals[index] = CafePublicSeatVisual{SeatNo: index + 1, State: "empty"}
	}
	for _, seat := range round.Edges.Seats {
		if seat.SeatNo == nil || *seat.SeatNo < 1 || *seat.SeatNo > totalSeats {
			continue
		}
		state := publicCafeSeatState(seat, now)
		if state == "empty" {
			continue
		}
		seatNo := *seat.SeatNo
		visuals[seatNo-1] = CafePublicSeatVisual{
			SeatNo:     seatNo,
			State:      state,
			AvatarSeed: publicCafeAvatarSeed(round.ID, seatNo),
			IsMine:     userID > 0 && seat.UserID == userID,
		}
	}
	return visuals
}

func publicCafeSeatState(seat *dbent.GroupBuySeat, now time.Time) string {
	if seat == nil {
		return "empty"
	}
	switch seat.Status {
	case GroupBuySeatStatusActive:
		return "active"
	case GroupBuySeatStatusPaid:
		return "paid"
	case GroupBuySeatStatusLocked:
		if seat.LockedUntil != nil && seat.LockedUntil.After(now) {
			return "locked"
		}
	}
	return "empty"
}

func publicCafePurchaseState(round *dbent.GroupBuyRound, remaining int) string {
	if round == nil {
		return "unavailable"
	}
	switch round.Status {
	case CafeRoundStatusOpen:
		if remaining > 0 {
			return "available"
		}
		return "full"
	case "activating":
		return "activating"
	case "active":
		return "active"
	default:
		return "unavailable"
	}
}

func publicCafeAvatarSeed(roundID int64, seatNo int) string {
	digest := sha256.Sum256([]byte(fmt.Sprintf("cafe-seat:%d:%d", roundID, seatNo)))
	return hex.EncodeToString(digest[:8])
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
